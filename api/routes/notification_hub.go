package routes

import (
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"prism/models"
)

// -- SSE pub/sub for notifications ----------------------------------------
//
// Mirrors the noteBroadcaster in notes.go: subscribers are keyed by the
// authenticated user's email, so an event published for one recipient can
// never reach another user's stream. That per-user keying is the isolation
// boundary — there is no broadcast channel.

type NotificationEvent struct {
	Type         string               `json:"type"` // notification.created | notification.read | notification.readAll
	ID           uint                 `json:"id,omitempty"`
	Notification *models.Notification `json:"notification,omitempty"`
}

type notificationSubscriber struct {
	ch   chan NotificationEvent
	done chan struct{}
}

type notificationBroadcaster struct {
	mu   sync.RWMutex
	subs map[string]map[*notificationSubscriber]struct{} // keyed by user email
}

var notificationHub = &notificationBroadcaster{subs: map[string]map[*notificationSubscriber]struct{}{}}

func (b *notificationBroadcaster) subscribe(userEmail string) *notificationSubscriber {
	sub := &notificationSubscriber{
		ch:   make(chan NotificationEvent, 16),
		done: make(chan struct{}),
	}
	b.mu.Lock()
	if _, ok := b.subs[userEmail]; !ok {
		b.subs[userEmail] = map[*notificationSubscriber]struct{}{}
	}
	b.subs[userEmail][sub] = struct{}{}
	b.mu.Unlock()
	return sub
}

func (b *notificationBroadcaster) unsubscribe(userEmail string, sub *notificationSubscriber) {
	b.mu.Lock()
	if set, ok := b.subs[userEmail]; ok {
		delete(set, sub)
		if len(set) == 0 {
			delete(b.subs, userEmail)
		}
	}
	b.mu.Unlock()
	close(sub.done)
}

func (b *notificationBroadcaster) publish(userEmail string, evt NotificationEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	set, ok := b.subs[userEmail]
	if !ok {
		return
	}
	for sub := range set {
		// Non-blocking send — if a subscriber is slow we drop the event for
		// them rather than stalling other tabs. The client re-fetches the
		// full list on reconnect, so a dropped event self-heals.
		select {
		case sub.ch <- evt:
		default:
		}
	}
}

// StreamNotificationsHandler is an SSE endpoint. It streams
// NotificationEvents scoped to the authenticated user across their open
// tabs/devices, replacing the old 60s polling loop.
func StreamNotificationsHandler(c *gin.Context) {
	email, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	sub := notificationHub.subscribe(email)
	defer notificationHub.unsubscribe(email, sub)

	// Opening hello so EventSource on the client transitions to OPEN quickly.
	if _, err := io.WriteString(c.Writer, ": connected\n\n"); err != nil {
		return
	}
	c.Writer.Flush()

	ctx := c.Request.Context()
	keepalive := time.NewTicker(25 * time.Second)
	defer keepalive.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-keepalive.C:
			if _, err := io.WriteString(c.Writer, ": ping\n\n"); err != nil {
				return
			}
			c.Writer.Flush()
		case evt := <-sub.ch:
			payload, err := json.Marshal(evt)
			if err != nil {
				continue
			}
			if _, err := io.WriteString(c.Writer, "event: "+evt.Type+"\ndata: "+string(payload)+"\n\n"); err != nil {
				return
			}
			c.Writer.Flush()
		}
	}
}
