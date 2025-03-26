package models

type UserSettings struct{
	SwimlaneUsers []string `json:"swimlaneUsers" gorm:"swimlaneUsers"`
}