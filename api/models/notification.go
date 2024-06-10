package models

import (
	"time"
)

type Notification struct {
	Who    string    `json:"who"`
	What   string    `json:"what"`
	IsRead bool      `json:"read"`
	Where  string    `json:"where"`
	When   time.Time `json:"when"`
}
