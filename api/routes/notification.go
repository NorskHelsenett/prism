package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"prism/database"
	"prism/models"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/gin-gonic/gin"
)

type PushMessage struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Url   string `json:"url"`
}

var (
	vapidPublicKey  string
	vapidPrivateKey string
	err             error
)

const (
	maxRetries     = 5
	initialBackoff = 2 * time.Second
)

func InitNotification() {

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
		err = createAndPersistVAAPIKeys()
		if err != nil {
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
	vapidPrivateKey, vapidPublicKey, err = webpush.GenerateVAPIDKeys()
	if err != nil {
		return fmt.Errorf("Unable to generate VAAPI keys: %v", err)
	}
	err = database.CreateVAAPIprivateKey(vapidPrivateKey, vapidPublicKey)
	if err != nil {
		return fmt.Errorf("Unable to store VAAPI keys: %v", err)
	}
	return nil
}

type Subscriber struct {
	Subscription webpush.Subscription
	Email        string
}

func DeleteNotificationsHandler(c *gin.Context) {
	email, _ := c.Get("email")

	if database.DeleteNotifications(email.(string)) != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error fetching notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"error": "invalid time format"})
}

func GetNotificationPublicKey(c *gin.Context) {
	publicKey, err := database.GetVAAPIpublicKey()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error fetching vaapi public key"})
		return
	}

	c.JSON(http.StatusOK, publicKey)
}

func GetNotificationsHandler(c *gin.Context) {
	email, _ := c.Get("email")

	notifications, err := database.GetNotifications(email.(string))

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "error fetching notifications"})
		return
	}

	c.JSON(http.StatusOK, notifications)
}

func MarkNotificationReadHandler(c *gin.Context) {
	email, exists := c.Get("email")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Parse the "time" parameter
	timeParam := c.Param("time")
	notificationTime, err := time.Parse(time.RFC3339, timeParam) // Assuming RFC3339 format for the timestamp
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid time format"})
		return
	}

	// Mark the notification as read
	err = database.MarkNotificationAsRead(email.(string), notificationTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error marking notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "notification marked as read"})
}

func SubscribeNotification(c *gin.Context) {
	email, _ := c.Get("email")
	var sub webpush.Subscription
	if err := c.BindJSON(&sub); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad Request"})
		return
	}

	jsonSub, err := json.Marshal(&sub)
	if err != nil {
		log.Printf("Error marshalling sub %v", err)
		return
	}

	if database.AppendSubscriber(email.(string), jsonSub) != nil {
		log.Printf("Error appending subscriber %v", err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"status": "created"})
}

func ResetNotifications(c *gin.Context) {
	err := database.ResetNotifications()
	if err != nil {
		log.Printf("Unable to reset VAAPI keys %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to perform requested task"})
		return
	}

	err = createAndPersistVAAPIKeys()
	if err != nil {
		log.Printf("Unable to persist VAAPI keys %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to perform requested task"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"success": "VAAPI keys reseted and created"})
}

func SendMessage(title, body, url, foundBy, ignoreEmail string, usersToNotify map[string]bool) error {
	// Placeholder for the logic to find subscribers who have read access to the resource
	//subscribers := findSubscribersWithReadAccess(url)

	// Map to keep track of the subscribers who have already received the notification
	sentTo := make(map[string]bool)

	subscribers, err := database.GetAllSubscribers()
	if err != nil {
		return err
	}

	// notification each subscriber in PRISM
	for userEmail := range usersToNotify {
		// Skip if the subscriber is the one who registered the vulnerability
		if userEmail == ignoreEmail {
			continue
		}

		// Web push notification to each subscriber
		for _, s := range subscribers {
			if s.Email == userEmail {
				err = sendPushNotification(s.Subscription, title, body, url)
				if err != nil {
					return err
				}

				break
			}
		}

		// Skip if the subscriber has already been sent this notification
		if _, alreadySent := sentTo[userEmail]; alreadySent {
			continue
		}

		// Create the notification object
		notification := models.Notification{
			Who:    ignoreEmail,
			What:   body,
			IsRead: false,
			Where:  url,
			When:   time.Now(),
		}

		// Save the notification to the database
		err := database.CreateNotification(userEmail, notification)
		if err != nil {
			log.Printf("Error creating notification for %s: %v", userEmail, err)
			return err
		}

		// Mark this subscriber as having been sent the notification
		sentTo[userEmail] = true
	}
	return nil
}

func sendPushNotification(subscriptionJson []byte, title, body, url string) error {
	// Construct your push notification payload by marshaling your data structure to JSON
	data := &PushMessage{
		Title: title,
		Body:  body,
		Url:   url,
	}
	jsonPayload, err := json.Marshal(data)
	if err != nil {
		log.Print(http.StatusInternalServerError, gin.H{"error": "Error marshaling push notification payload"})
		return err
	}
	var subscription webpush.Subscription

	if err := json.Unmarshal(subscriptionJson, &subscription); err != nil {
		log.Printf("Error unmarshalling subscribers %v", err)
		return err
	}

	resp, err := webpush.SendNotification(jsonPayload, &subscription, &webpush.Options{
		Subscriber:      "cat@nhn.no",
		VAPIDPublicKey:  vapidPublicKey,
		VAPIDPrivateKey: vapidPrivateKey,
		TTL:             30,
	})
	if err != nil {
		// Handle error
		log.Printf("Error sending web push %v", err)
		return err
	}
	if resp.StatusCode != 200 && resp.StatusCode != 202 {
		// Read response body for additional error details
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
		}
		bodyString := string(bodyBytes)

		// Log detailed error message
		log.Printf("Error sending web push. Status Code: %d, Status: %s, Response Body: %s",
			resp.StatusCode, resp.Status, bodyString)
	}
	defer resp.Body.Close() // Close the response body when done reading
	return nil
}
