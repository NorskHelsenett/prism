package models

import (
	"time"

	"gorm.io/gorm"
)

type Note struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`

	UserEmail string `gorm:"index;not null" json:"-"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Preview   string `json:"preview"`
	Tags      string `gorm:"index" json:"-"`
}

type NoteListItem struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Preview   string    `json:"preview"`
	Tags      []string  `json:"tags"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt,omitempty"`
}
