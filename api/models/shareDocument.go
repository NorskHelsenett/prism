package models

import (
	"time"
)

type SharedDocument struct {
	CreatedAt      time.Time  `gorm:"not null" json:"createdAt"`
	DocumentID     uint       `gorm:"primaryKey;autoIncrement:false" json:"documentId"`
	InvitedEmails  string     `gorm:"null" json:"invitedEmails"`                     // a comma-separated list of email addresses
	ShareToken     string     `gorm:"not null;uniqueIndex;size:8" json:"shareToken"` // 8-character random string
	SharedByEmail  string     `gorm:"not null" json:"sharedByEmail"`                 // email address of the person who shared the document
	ExpirationDate *time.Time `gorm:"index" json:"expirationDate"`
	Passphrase     string     `gorm:"null" json:"passphrase"`     // stored in plain text
	AccessType     string     `gorm:"not null" json:"accessType"` // 'anyone-with-link', 'organization', 'passphrase-protected', or 'only-invited'
}
