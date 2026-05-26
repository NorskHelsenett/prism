package models

import (
	"time"
)

// Notification is the wire shape returned to the frontend. The DB row lives in
// the database package; this struct mirrors the JSON the dropdown consumes.
type Notification struct {
	ID              uint      `json:"id"`
	Who             string    `json:"who"`
	What            string    `json:"what"`
	IsRead          bool      `json:"read"`
	Where           string    `json:"where"`
	When            time.Time `json:"when"`
	Kind            string    `json:"kind,omitempty"`
	VulnerabilityID *uint     `json:"vulnerabilityId,omitempty"`
}

// NotificationKind values are stored on the row so the UI and the dispatcher
// can filter behaviour by event type without re-parsing `what`.
const (
	NotificationKindNewVuln    = "new_vuln"
	NotificationKindNewComment = "new_comment"
)
