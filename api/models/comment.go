package models

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID        string    `json:"id"`
	UserEmail string    `json:"email"`
	Text      string    `json:"text"`
	ParentID  string    `json:"parentId,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type Revision struct {
	UserEmail string    `json:"email"`
	Property  string    `json:"property"`
	OldValue  string    `json:"oldValue"`
	NewValue  string    `json:"newValue"`
	CreatedAt time.Time `json:"createdAt"`
}

func (c *Comment) BeforeCreate(tx *gorm.DB) (err error) {
	c.CreatedAt = time.Now()
	return
}

func (r *Revision) BeforeCreate(tx *gorm.DB) (err error) {
	r.CreatedAt = time.Now()
	return
}
