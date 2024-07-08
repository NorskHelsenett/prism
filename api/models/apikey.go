package models

import (
	"time"

	"gorm.io/gorm"
)

type APIKey struct {
	gorm.Model
	Name         string    `json:"name"`
	ExpiryAt     time.Time `json:"expire"`
	HashedAPIKey string    `json:"-"`
	Email        string    `json:"email"`
	ApiKey       string    `gorm:"-" json:"apikey"`
}
