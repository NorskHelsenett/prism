package event

import (
	"testing"
	"time"

	"prism/database"
	"prism/models"
)

func commentAt(id, email string, at time.Time) models.Comment {
	return models.Comment{ID: id, UserEmail: email, Text: "t", CreatedAt: at}
}

// The comments trigger fires on every UPDATE of the comments column. Only an
// event caused by a NEW (or freshly edited) comment should notify; deletions
// leave the newest surviving comment older than the event and must be silent.
func TestCommentForEvent_NewCommentNotifies(t *testing.T) {
	now := time.Now()
	event := database.EventQueue{CreatedAt: now}
	comments := []models.Comment{
		commentAt("c1", "alice@x", now.Add(-1*time.Hour)),
		commentAt("c2", "bob@x", now.Add(-2*time.Second)),
	}
	last, ok := commentForEvent(event, comments)
	if !ok {
		t.Fatalf("fresh comment should notify")
	}
	if last.ID != "c2" {
		t.Fatalf("picked %q, want the newest comment c2", last.ID)
	}
}

func TestCommentForEvent_DeletionDoesNotNotify(t *testing.T) {
	now := time.Now()
	event := database.EventQueue{CreatedAt: now}
	// Newest surviving comment is an hour old — this event was a deletion.
	comments := []models.Comment{
		commentAt("c1", "alice@x", now.Add(-3*time.Hour)),
		commentAt("c2", "bob@x", now.Add(-1*time.Hour)),
	}
	if _, ok := commentForEvent(event, comments); ok {
		t.Fatalf("deletion-triggered event must not notify")
	}
}

func TestCommentForEvent_LastCommentDeletedDoesNotNotify(t *testing.T) {
	event := database.EventQueue{CreatedAt: time.Now()}
	if _, ok := commentForEvent(event, nil); ok {
		t.Fatalf("event with no comments left must not notify")
	}
}
