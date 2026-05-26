package routes

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"prism/database"
	"prism/models"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"
)

// PushMessage is the JSON payload delivered to service-worker.js. Mirrors the
// fields the existing SW reads (title/body/url); keep new fields optional so
// older service workers still render the basics.
type PushMessage struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Url   string `json:"url"`
}

var (
	vapidPublicKey  string
	vapidPrivateKey string
)

const (
	maxRetries     = 5
	initialBackoff = 2 * time.Second
)

func InitNotification() {
	var err error
	for retries := 0; retries < maxRetries; retries++ {
		var vapidKey string
		vapidKey, err = database.GetVAAPIpublicKey()
		if err == nil {
			vapidPublicKey = vapidKey
			break
		}

		log.Printf("Unable to retrieve VAAPI keys, will retry: %v", err)
		time.Sleep(initialBackoff * time.Duration(retries+1))
	}

	if err != nil {
		log.Fatalf("Unable to retrieve VAAPI keys after retries: %v", err)
	}

	if vapidPublicKey == "" {
		if err := createAndPersistVAAPIKeys(); err != nil {
			log.Fatalf("Unable to generate VAAPI keys: %v", err)
		}
	} else {
		vapidPrivateKey, err = database.GetVAAPIprivateKey()
		if err != nil {
			log.Fatalf("Unable to retrieve VAAPI private key: %v", err)
		}
	}
}

func createAndPersistVAAPIKeys() error {
	priv, pub, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		return fmt.Errorf("Unable to generate VAAPI keys: %v", err)
	}
	vapidPrivateKey = priv
	vapidPublicKey = pub
	if err := database.CreateVAAPIprivateKey(vapidPrivateKey, vapidPublicKey); err != nil {
		return fmt.Errorf("Unable to store VAAPI keys: %v", err)
	}
	return nil
}

// DeleteNotificationsHandler clears all notification rows for the caller. The
// dropdown's "Clear all" button now uses PUT /api/notification/read-all
// instead, which preserves history; this endpoint stays for completeness
// (e.g. account-wipe flows).
func DeleteNotificationsHandler(c *gin.Context) {
	email, _ := c.Get("email")
	if err := database.DeleteNotifications(email.(string)); err != nil {
		log.Printf("notifications: delete for %q: %v", email, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error fetching notifications"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "notifications deleted"})
}

// MarkAllReadHandler flips IsRead on every unread row for the caller in one
// SQL UPDATE, replacing the old "delete to clear the badge" hack.
func MarkAllReadHandler(c *gin.Context) {
	email, _ := c.Get("email")
	if err := database.MarkAllNotificationsRead(email.(string)); err != nil {
		log.Printf("notifications: mark-all-read for %q: %v", email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not mark notifications read"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "marked read"})
}

func GetNotificationPublicKey(c *gin.Context) {
	publicKey, err := database.GetVAAPIpublicKey()
	if err != nil {
		log.Printf("notifications: get public key: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error fetching vaapi public key"})
		return
	}
	c.JSON(http.StatusOK, publicKey)
}

func GetNotificationsHandler(c *gin.Context) {
	email, _ := c.Get("email")
	notifications, err := database.GetNotifications(email.(string), 50)
	if err != nil {
		log.Printf("notifications: get for %q: %v", email, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "error fetching notifications"})
		return
	}
	c.JSON(http.StatusOK, notifications)
}

// MarkNotificationReadHandler now takes the row id rather than the RFC3339
// timestamp. The old timestamp-based lookup round-tripped through time.Parse
// and was unstable across precision boundaries; ids are the natural key.
func MarkNotificationReadHandler(c *gin.Context) {
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid notification id"})
		return
	}
	if err := database.MarkNotificationAsRead(email.(string), uint(id)); err != nil {
		// Surface 404 rather than leaking ownership info — matches the
		// project-wide policy of not returning 500 for normal misses.
		log.Printf("notifications: mark read id=%d email=%q: %v", id, email, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "notification not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}

// SubscribeNotification accepts a PushSubscription JSON and binds the device
// to the current user. The actual take-over (rebinding the endpoint away from
// any previous owner on the same browser) happens inside
// database.UpsertSubscriber.
func SubscribeNotification(c *gin.Context) {
	email, _ := c.Get("email")
	var sub webpush.Subscription
	if err := c.BindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}
	userAgent := c.GetHeader("User-Agent")
	if err := database.UpsertSubscriber(email.(string), sub, userAgent); err != nil {
		log.Printf("notifications: upsert subscriber for %q: %v", email, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not register subscription"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"status": "created"})
}

// UnsubscribeNotification removes one device's subscription. The frontend
// calls this from the explicit "disable notifications" button — *not* on
// logout. Logout intentionally leaves the row in place so the user keeps
// receiving pushes after signing out, which is the chosen behaviour.
func UnsubscribeNotification(c *gin.Context) {
	email, _ := c.Get("email")
	var body struct {
		Endpoint string `json:"endpoint"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}
	if err := database.DeleteSubscriberByEndpoint(email.(string), body.Endpoint); err != nil {
		log.Printf("notifications: delete subscriber for %q: %v", email, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not remove subscription"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "removed"})
}

// ListDevices returns the caller's push subscriptions for a future "your
// devices" management UI. Endpoints are hashed because the raw value is a
// secret URL that lets anyone push to that device.
func ListDevices(c *gin.Context) {
	email, _ := c.Get("email")
	rows, err := database.ListSubscribersForEmail(email.(string))
	if err != nil {
		log.Printf("notifications: list devices for %q: %v", email, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not list devices"})
		return
	}
	out := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		out = append(out, gin.H{
			"endpointHash": hashEndpoint(r.Endpoint),
			"userAgent":    r.UserAgent,
			"lastSeenAt":   r.LastSeenAt,
		})
	}
	c.JSON(http.StatusOK, out)
}

func hashEndpoint(endpoint string) string {
	sum := sha256.Sum256([]byte(endpoint))
	return hex.EncodeToString(sum[:8])
}

func ResetNotifications(c *gin.Context) {
	if err := database.ResetNotifications(); err != nil {
		log.Printf("Unable to reset VAAPI keys %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to perform requested task"})
		return
	}
	if err := createAndPersistVAAPIKeys(); err != nil {
		log.Printf("Unable to persist VAAPI keys %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to perform requested task"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": "VAAPI keys reseted and created"})
}

// DispatchRequest is the new entry point the event queue hands to the
// dispatcher. Replaces SendMessage's positional-arg soup and makes the
// distinction between "actor" (skipped from delivery, recorded as Who) and
// recipients explicit.
type DispatchRequest struct {
	Kind            string
	VulnerabilityID uint
	ActorEmail      string
	Recipients      []string
	Title           string
	Body            string
	URL             string
}

// pushSender is the function Dispatch uses to deliver one push. The real
// implementation in sendPush talks to a webpush provider over HTTPS; tests
// swap this for an in-memory recorder so they can assert delivery shape
// without standing up VAPID + a fake provider.
var pushSender = sendPush

// Dispatch is the single fan-out point for both push and in-app delivery. It:
//   - normalises and dedupes the recipient list (trim/lowercase)
//   - drops the actor and any recipient whose access has been revoked
//   - honours the per-user prefs matrix for each channel
//   - inserts one notifications row per surviving recipient (race-free single
//     INSERT — no read-modify-write)
//   - pushes to every endpoint the recipient has registered (no `break` on
//     first match like the old code)
//   - cleans up dead endpoints (404/410) so the table can't grow stale
//
// Push and in-app errors are logged and skipped per-recipient; one bad device
// no longer halts delivery to the rest of the list.
func Dispatch(req DispatchRequest) error {
	recipients := normaliseRecipients(req.Recipients, req.ActorEmail)
	if len(recipients) == 0 {
		return nil
	}

	subs, err := database.ListSubscribersForEmails(recipients)
	if err != nil {
		log.Printf("notifications: list subscribers: %v", err)
		subs = nil // Continue with in-app delivery even if push lookup fails.
	}
	subsByEmail := map[string][]database.Subscriber{}
	for _, s := range subs {
		s := s
		subsByEmail[s.Email] = append(subsByEmail[s.Email], s)
	}

	pushPayload, err := json.Marshal(&PushMessage{
		Title: req.Title,
		Body:  req.Body,
		Url:   req.URL,
	})
	if err != nil {
		return fmt.Errorf("marshal push payload: %w", err)
	}

	vulnID := req.VulnerabilityID
	var vulnIDPtr *uint
	if vulnID != 0 {
		v := vulnID
		vulnIDPtr = &v
	}

	for _, recipient := range recipients {
		canSee := true
		if vulnID != 0 {
			ok, err := database.CanRecipientSeeVulnerability(recipient, vulnID)
			if err != nil {
				log.Printf("notifications: acl check for %q: %v", recipient, err)
				continue
			}
			canSee = ok
		}
		if !canSee {
			continue
		}

		prefs, err := loadPrefs(recipient)
		if err != nil {
			log.Printf("notifications: load prefs for %q: %v", recipient, err)
			prefs = models.NotificationPrefs{}.Effective()
		}

		if inAppEnabled(prefs, req.Kind) {
			if err := database.CreateNotification(recipient, models.Notification{
				Kind:            req.Kind,
				Who:             req.ActorEmail,
				What:            req.Body,
				Where:           req.URL,
				VulnerabilityID: vulnIDPtr,
			}); err != nil {
				log.Printf("notifications: create in-app for %q: %v", recipient, err)
			}
		}

		if !pushEnabled(prefs, req.Kind) {
			continue
		}
		for _, s := range subsByEmail[recipient] {
			s := s
			if err := pushSender(s, pushPayload); err != nil {
				log.Printf("notifications: push to %q (endpoint %s): %v",
					recipient, hashEndpoint(s.Endpoint), err)
			}
		}
	}
	return nil
}

func inAppEnabled(p models.ResolvedNotificationPrefs, kind string) bool {
	switch kind {
	case models.NotificationKindNewVuln:
		return p.InAppNewVuln
	case models.NotificationKindNewComment:
		return p.InAppNewComment
	default:
		return true
	}
}

func pushEnabled(p models.ResolvedNotificationPrefs, kind string) bool {
	switch kind {
	case models.NotificationKindNewVuln:
		return p.PushNewVuln
	case models.NotificationKindNewComment:
		return p.PushNewComment
	default:
		return true
	}
}

func loadPrefs(email string) (models.ResolvedNotificationPrefs, error) {
	settings, err := database.GetPreferencesForUser(email)
	if err != nil {
		return models.NotificationPrefs{}.Effective(), err
	}
	return settings.NotificationPrefs.Effective(), nil
}

// normaliseRecipients trims whitespace, lowercases, drops the actor, and
// dedupes. The dispatcher always uses the canonical lowercased form so push
// lookups and ACL checks agree on identity.
func normaliseRecipients(in []string, actor string) []string {
	actor = strings.ToLower(strings.TrimSpace(actor))
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, raw := range in {
		e := strings.ToLower(strings.TrimSpace(raw))
		if e == "" || e == actor {
			continue
		}
		if _, dup := seen[e]; dup {
			continue
		}
		seen[e] = struct{}{}
		out = append(out, e)
	}
	return out
}

// sendPush dispatches one push and reaps dead endpoints. Returning nil for
// 404/410 lets the caller continue down the recipient list — the row is gone
// from the next dispatch onward.
func sendPush(s database.Subscriber, payload []byte) error {
	sub := s.AsPushSubscription()
	resp, err := webpush.SendNotification(payload, &sub, &webpush.Options{
		Subscriber:      "cat@nhn.no",
		VAPIDPublicKey:  vapidPublicKey,
		VAPIDPrivateKey: vapidPrivateKey,
		TTL:             30,
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusGone {
		if delErr := database.DeleteDeadSubscriber(s.Endpoint); delErr != nil {
			log.Printf("notifications: prune dead endpoint %s: %v",
				hashEndpoint(s.Endpoint), delErr)
		}
		return nil
	}
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		body, _ := io.ReadAll(resp.Body)
		return errors.New("push provider returned " + resp.Status + ": " + string(body))
	}
	return nil
}
