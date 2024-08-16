package models

import (
	"encoding/json"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type Team struct {
	gorm.Model
	Name        string         `gorm:"not null" json:"name"`
	Role        string         `json:"role"`
	Archived    bool           `json:"archived"`
	Members     datatypes.JSON `gorm:"type:json" json:"-"`
	MembersJSON []string       `gorm:"-" json:"members"`
}

func (t *Team) AfterFind(tx *gorm.DB) (err error) {
	return json.Unmarshal(t.Members, &t.MembersJSON)
}

func (t *Team) BeforeSave(tx *gorm.DB) (err error) {
	t.Members, err = json.Marshal(t.MembersJSON)
	return
}
