package models

import (
	"errors"
	"time"
)

type Assessment struct {
	Responsible string    `json:"responsible_hacker"`
	Title       string    `json:"title"`
	Projects    []Project `json:"projects"`
	DateFrom    string    `json:"dateFrom" gorm:"type:date"`
	DateTo      string    `json:"dateTo" gorm:"type:date"`
	Note        string    `json:"note"`
	Hackers     []Hacker  `json:"hackers"`
	ID          uint      `json:"id" gorm:"-"`
	Estimate    uint      `json:"estimate"`
	WorkOrder   string    `json:"workorder"`
	Status      string    `json:"status"`
	Requester   string    `json:"requester"`
}

type Hacker struct {
	Email string `json:"email"`
}

type Project struct {
	Id   uint   `json:"id"`
	Name string `json:"name"`
}

func (a *Assessment) Validate() error {
	if a.Title == "" {
		return errors.New("Title is required")
	}
	if len(a.Projects) == 0 {
		return errors.New("at least one project is required")
	}
	for _, project := range a.Projects {
		if project.Id == 0 {
			return errors.New("project ID is required")
		}
	}

	const layout = "2006-01-02"
	if a.DateFrom == "" {
		return errors.New("dateFrom is required")
	}
	if _, err := time.Parse(layout, a.DateFrom); err != nil {
		return errors.New("dateFrom is not in the format yyyy-mm-dd")
	}
	if a.DateTo == "" {
		return errors.New("dateTo is required")
	}
	if _, err := time.Parse(layout, a.DateTo); err != nil {
		return errors.New("dateTo is not in the format yyyy-mm-dd")
	}
	// if len(a.Hackers) == 0 {
	// 	return errors.New("at least one hacker is required")
	// }
	for _, hacker := range a.Hackers {
		if hacker.Email == "" {
			return errors.New("hacker email is required")
		}
	}
	return nil
}
