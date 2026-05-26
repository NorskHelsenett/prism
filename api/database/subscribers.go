package database

import (
	"errors"
	"strings"
	"time"

	"github.com/SherClockHolmes/webpush-go"
	"gorm.io/gorm"
)

// ErrSubscriberMissingEndpoint is returned when an upsert is attempted with a
// subscription that lacks an endpoint. Endpoint is the cross-user uniqueness
// key; without it, the take-over semantics cannot be enforced.
var ErrSubscriberMissingEndpoint = errors.New("push subscription is missing endpoint")

// UpsertSubscriber binds the given push endpoint to the given email, taking it
// over from any other email that previously owned it.
//
// This is the core fix for the cross-user push delivery bug: a single browser
// only ever has one row in subscribers, owned by the user who most recently
// enabled push on that device. When Alice subscribes on browser B then logs
// out, Alice's row keeps endpoint X. If Bob then logs in on browser B and
// enables push, browser B returns the same endpoint X (pushManager is per
// browser-installation, not per user), and this upsert rebinds X to Bob.
// Alice's row for X is gone — server pushes for Alice no longer land on
// Bob's screen.
//
// We don't use ON CONFLICT here because GORM/sqlite ON CONFLICT requires a
// UNIQUE index, and the migration installs that index. We instead delete any
// other row owning this endpoint inside a single transaction, then upsert by
// (email, endpoint).
func UpsertSubscriber(email string, sub webpush.Subscription, userAgent string) error {
	endpoint := strings.TrimSpace(sub.Endpoint)
	if endpoint == "" {
		return ErrSubscriberMissingEndpoint
	}
	return db.Transaction(func(tx *gorm.DB) error {
		// Take-over: delete any row pointing at this endpoint that doesn't
		// belong to the current user. Same email/endpoint is left in place and
		// re-fetched below so the upsert is a single round-trip.
		if err := tx.Unscoped().
			Where("endpoint = ? AND email <> ?", endpoint, email).
			Delete(&Subscriber{}).Error; err != nil {
			return err
		}

		var existing Subscriber
		err := tx.Where("email = ? AND endpoint = ?", email, endpoint).
			First(&existing).Error
		now := time.Now().UTC()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return tx.Create(&Subscriber{
				Email:      email,
				Endpoint:   endpoint,
				P256dh:     sub.Keys.P256dh,
				Auth:       sub.Keys.Auth,
				UserAgent:  userAgent,
				LastSeenAt: now,
			}).Error
		}
		if err != nil {
			return err
		}
		return tx.Model(&existing).Updates(map[string]any{
			"p256dh":       sub.Keys.P256dh,
			"auth":         sub.Keys.Auth,
			"user_agent":   userAgent,
			"last_seen_at": now,
		}).Error
	})
}

// DeleteSubscriberByEndpoint removes the current user's subscription for one
// device. Scoped by email so users can't yank each other's subscriptions by
// guessing endpoints.
func DeleteSubscriberByEndpoint(email, endpoint string) error {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ErrSubscriberMissingEndpoint
	}
	return db.Unscoped().
		Where("email = ? AND endpoint = ?", email, endpoint).
		Delete(&Subscriber{}).Error
}

// DeleteDeadSubscriber removes a row when the push provider says the endpoint
// is gone (404/410). Called from the dispatcher so dead devices age out of
// the table on their own — no manual cleanup needed.
func DeleteDeadSubscriber(endpoint string) error {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return ErrSubscriberMissingEndpoint
	}
	return db.Unscoped().Where("endpoint = ?", endpoint).Delete(&Subscriber{}).Error
}

// ListSubscribersForEmail returns every subscription belonging to one user
// (one user can have several devices). The dispatcher pushes to all of them.
func ListSubscribersForEmail(email string) ([]Subscriber, error) {
	var rows []Subscriber
	err := db.Where("email = ? AND endpoint <> ''", email).Find(&rows).Error
	return rows, err
}

// ListSubscribersForEmails is the bulk variant used by SendMessage so the
// dispatcher does a single SELECT instead of one per recipient.
func ListSubscribersForEmails(emails []string) ([]Subscriber, error) {
	if len(emails) == 0 {
		return nil, nil
	}
	var rows []Subscriber
	err := db.Where("email IN ? AND endpoint <> ''", emails).Find(&rows).Error
	return rows, err
}

// AsPushSubscription rebuilds the webpush-go subscription struct from the
// typed columns so the dispatcher can hand it straight to SendNotification.
func (s *Subscriber) AsPushSubscription() webpush.Subscription {
	return webpush.Subscription{
		Endpoint: s.Endpoint,
		Keys: webpush.Keys{
			Auth:   s.Auth,
			P256dh: s.P256dh,
		},
	}
}
