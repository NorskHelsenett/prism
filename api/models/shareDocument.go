package models

import (
	"encoding/json"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type SharedDocument struct {
	CreatedAt         time.Time      `gorm:"not null" json:"createdAt"`
	DocumentID        uint           `gorm:"primaryKey;autoIncrement:false" json:"documentId"`
	InvitedEmails     datatypes.JSON `gorm:"null" json:"-"`
	InvitedEmailsJSON []string       `gorm:"-" json:"invitedEmails"`
	ShareToken        string         `gorm:"not null;uniqueIndex;size:8" json:"shareToken"`
	SharedByEmail     string         `gorm:"not null" json:"sharedByEmail"`
	ExpirationDate    *time.Time     `gorm:"index" json:"expirationDate"`
	Passphrase        string         `gorm:"null" json:"passphrase"`
	AccessType        string         `gorm:"not null" json:"accessType"`
}

func (sd *SharedDocument) AfterFind(tx *gorm.DB) (err error) {
	if len(sd.InvitedEmails) > 0 {
		return json.Unmarshal(sd.InvitedEmails, &sd.InvitedEmailsJSON)
	}
	sd.InvitedEmailsJSON = []string{}
	return nil
}

func (sd *SharedDocument) BeforeSave(tx *gorm.DB) (err error) {
	if len(sd.InvitedEmailsJSON) > 0 {
		sd.InvitedEmails, err = json.Marshal(sd.InvitedEmailsJSON)
	} else {
		sd.InvitedEmails = datatypes.JSON([]byte("[]"))
	}
	return
}
