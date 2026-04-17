package routes

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"prism/database"
)

// -- SSE pub/sub ----------------------------------------------------------

type NoteEvent struct {
	Type      string    `json:"type"` // note.updated | note.created | note.deleted | note.restored | note.purged
	ID        uint      `json:"id"`
	UpdatedAt time.Time `json:"updatedAt,omitempty"`
	Source    string    `json:"source,omitempty"` // client-provided tab id so originator can ignore
}

type noteSubscriber struct {
	ch   chan NoteEvent
	done chan struct{}
}

type noteBroadcaster struct {
	mu   sync.RWMutex
	subs map[string]map[*noteSubscriber]struct{} // keyed by user email
}

var notes = &noteBroadcaster{subs: map[string]map[*noteSubscriber]struct{}{}}

func (b *noteBroadcaster) subscribe(userEmail string) *noteSubscriber {
	sub := &noteSubscriber{
		ch:   make(chan NoteEvent, 16),
		done: make(chan struct{}),
	}
	b.mu.Lock()
	if _, ok := b.subs[userEmail]; !ok {
		b.subs[userEmail] = map[*noteSubscriber]struct{}{}
	}
	b.subs[userEmail][sub] = struct{}{}
	b.mu.Unlock()
	return sub
}

func (b *noteBroadcaster) unsubscribe(userEmail string, sub *noteSubscriber) {
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

func (b *noteBroadcaster) publish(userEmail string, evt NoteEvent) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	set, ok := b.subs[userEmail]
	if !ok {
		return
	}
	for sub := range set {
		// Non-blocking send — if a subscriber is slow we drop the event for
		// them rather than stalling other tabs. The client can re-fetch on
		// reconnect.
		select {
		case sub.ch <- evt:
		default:
		}
	}
}

// -- Helpers --------------------------------------------------------------

func currentUser(c *gin.Context) (string, bool) {
	v, ok := c.Get("email")
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}

func parseNoteID(c *gin.Context) (uint, bool) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return 0, false
	}
	return uint(id), true
}

// -- Handlers -------------------------------------------------------------

func ListNotes(c *gin.Context) {
	email, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	opts := database.ListNotesOptions{
		Query: c.Query("q"),
		Tag:   c.Query("tag"),
		Trash: c.Query("trash") == "true",
	}
	list, err := database.ListNotes(email, opts)
	if err != nil {
		log.Printf("ListNotes: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, list)
}

func ListNoteTags(c *gin.Context) {
	email, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	tags, err := database.ListAllTags(email)
	if err != nil {
		log.Printf("ListNoteTags: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, tags)
}

func GetNote(c *gin.Context) {
	email, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	id, ok := parseNoteID(c)
	if !ok {
		return
	}
	note, err := database.GetNote(email, id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		log.Printf("GetNote: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, note)
}

type notePayload struct {
	Content string `json:"content"`
	Source  string `json:"source,omitempty"`
}

func CreateNote(c *gin.Context) {
	email, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	var p notePayload
	// A fresh note can be empty — don't require a body.
	_ = c.ShouldBindJSON(&p)

	note, err := database.CreateNote(email, p.Content)
	if err != nil {
		log.Printf("CreateNote: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "unable to create note"})
		return
	}
	notes.publish(email, NoteEvent{Type: "note.created", ID: note.ID, UpdatedAt: note.UpdatedAt, Source: p.Source})
	c.JSON(http.StatusCreated, note)
}

func UpdateNote(c *gin.Context) {
	email, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	id, ok := parseNoteID(c)
	if !ok {
		return
	}
	var p notePayload
	if err := c.ShouldBindJSON(&p); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	note, err := database.UpdateNote(email, id, p.Content)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		log.Printf("UpdateNote: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	notes.publish(email, NoteEvent{Type: "note.updated", ID: note.ID, UpdatedAt: note.UpdatedAt, Source: p.Source})
	c.JSON(http.StatusOK, note)
}

func DeleteNote(c *gin.Context) {
	email, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	id, ok := parseNoteID(c)
	if !ok {
		return
	}
	source := c.Query("source")
	if err := database.SoftDeleteNote(email, id); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		log.Printf("DeleteNote: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	notes.publish(email, NoteEvent{Type: "note.deleted", ID: id, Source: source})
	c.JSON(http.StatusOK, gin.H{"status": "trashed"})
}

func RestoreNote(c *gin.Context) {
	email, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	id, ok := parseNoteID(c)
	if !ok {
		return
	}
	note, err := database.RestoreNote(email, id)
	if err != nil {
		if errors.Is(err, database.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		log.Printf("RestoreNote: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	notes.publish(email, NoteEvent{Type: "note.restored", ID: note.ID, UpdatedAt: note.UpdatedAt})
	c.JSON(http.StatusOK, note)
}

func HardDeleteNote(c *gin.Context) {
	email, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	id, ok := parseNoteID(c)
	if !ok {
		return
	}
	if err := database.HardDeleteNote(email, id); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		log.Printf("HardDeleteNote: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	notes.publish(email, NoteEvent{Type: "note.purged", ID: id})
	c.JSON(http.StatusOK, gin.H{"status": "purged"})
}

func EmptyNoteTrash(c *gin.Context) {
	email, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	n, err := database.EmptyTrash(email)
	if err != nil {
		log.Printf("EmptyNoteTrash: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	notes.publish(email, NoteEvent{Type: "note.purged"})
	c.JSON(http.StatusOK, gin.H{"purged": n})
}

// StreamNotes is an SSE endpoint. It streams NoteEvents scoped to the
// authenticated user across their open tabs/devices.
func StreamNotes(c *gin.Context) {
	email, ok := currentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	sub := notes.subscribe(email)
	defer notes.unsubscribe(email, sub)

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
