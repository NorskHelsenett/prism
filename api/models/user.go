package models

type UserSettings struct {
	SwimlaneUsers      []string `json:"swimlaneUsers" gorm:"swimlaneUsers"`
	CollapsedGroups    []uint   `json:"collapsedGroups"`
	ProjectShowAll     bool     `json:"projectShowAll"`
	ProjectShowInactive bool    `json:"projectShowInactive"`
}
