package models

import (
	"encoding/json"
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Report is the mutable draft of a vulnerability report.
// Published immutable snapshots live in ReportVersion.
type Report struct {
	gorm.Model

	Title            string `gorm:"not null"`
	ExecutiveSummary string `gorm:"type:text"`

	ProjectIDs     datatypes.JSON `gorm:"null" json:"-"`
	ProjectIDsList []uint         `gorm:"-" json:"projectIds"`

	VulnerabilityIDs     datatypes.JSON `gorm:"null" json:"-"`
	VulnerabilityIDsList []uint         `gorm:"-" json:"vulnerabilityIds"`

	FindingOverrides     datatypes.JSON              `gorm:"null" json:"-"`
	FindingOverridesData map[uint]FindingOverride    `gorm:"-" json:"findingOverrides"`

	OwnerEmail string `gorm:"not null;index"`

	ShareToken string `gorm:"uniqueIndex;size:8"`

	InvitedEmails     datatypes.JSON `gorm:"null" json:"-"`
	InvitedEmailsList []string       `gorm:"-" json:"invitedEmails"`

	LatestPublishedVersionID *uint `json:"latestPublishedVersionId"`
}

// FindingOverride lets the report author edit a finding's surface fields
// without mutating the underlying vulnerability.
type FindingOverride struct {
	Title   string `json:"title"`
	Summary string `json:"summary"`
}

func (r *Report) Validate() error {
	if r.Title == "" {
		return errors.New("title is required")
	}
	if len(r.ProjectIDsList) == 0 {
		return errors.New("at least one project is required")
	}
	return nil
}

func (r *Report) AfterFind(tx *gorm.DB) error {
	if len(r.ProjectIDs) > 0 {
		if err := json.Unmarshal(r.ProjectIDs, &r.ProjectIDsList); err != nil {
			return err
		}
	} else {
		r.ProjectIDsList = []uint{}
	}
	if len(r.VulnerabilityIDs) > 0 {
		if err := json.Unmarshal(r.VulnerabilityIDs, &r.VulnerabilityIDsList); err != nil {
			return err
		}
	} else {
		r.VulnerabilityIDsList = []uint{}
	}
	if len(r.FindingOverrides) > 0 {
		if err := json.Unmarshal(r.FindingOverrides, &r.FindingOverridesData); err != nil {
			return err
		}
	} else {
		r.FindingOverridesData = map[uint]FindingOverride{}
	}
	if len(r.InvitedEmails) > 0 {
		if err := json.Unmarshal(r.InvitedEmails, &r.InvitedEmailsList); err != nil {
			return err
		}
	} else {
		r.InvitedEmailsList = []string{}
	}
	return nil
}

func (r *Report) BeforeSave(tx *gorm.DB) error {
	var err error
	if r.ProjectIDsList == nil {
		r.ProjectIDs = datatypes.JSON([]byte("[]"))
	} else {
		r.ProjectIDs, err = json.Marshal(r.ProjectIDsList)
		if err != nil {
			return err
		}
	}
	if r.VulnerabilityIDsList == nil {
		r.VulnerabilityIDs = datatypes.JSON([]byte("[]"))
	} else {
		r.VulnerabilityIDs, err = json.Marshal(r.VulnerabilityIDsList)
		if err != nil {
			return err
		}
	}
	if r.FindingOverridesData == nil {
		r.FindingOverrides = datatypes.JSON([]byte("{}"))
	} else {
		r.FindingOverrides, err = json.Marshal(r.FindingOverridesData)
		if err != nil {
			return err
		}
	}
	if r.InvitedEmailsList == nil {
		r.InvitedEmails = datatypes.JSON([]byte("[]"))
	} else {
		r.InvitedEmails, err = json.Marshal(r.InvitedEmailsList)
		if err != nil {
			return err
		}
	}
	return nil
}

// ReportVersion is an immutable snapshot of a Report at publish time.
// Data is the frozen JSON payload (title, exec summary, snapshotted
// vulnerabilities by value, project names, applied overrides).
// PDF holds the rendered bytes — never re-rendered after publish.
type ReportVersion struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	ReportID      uint   `gorm:"not null;index" json:"reportId"`
	VersionNumber int    `gorm:"not null" json:"version"`
	Data          datatypes.JSON `gorm:"not null" json:"data"`
	PDF           []byte `gorm:"not null" json:"-"`
	PublishedAt   time.Time `gorm:"not null" json:"publishedAt"`
	PublishedBy   string `gorm:"not null" json:"publishedBy"`
}

// ReportVersionPayload is the frozen JSON shape stored in ReportVersion.Data.
// Vulnerabilities are snapshotted by value: if a vuln title later changes,
// the historic report keeps the original wording.
type ReportVersionPayload struct {
	Title            string                      `json:"title"`
	ExecutiveSummary string                      `json:"executiveSummary"`
	Projects         []ReportSnapshotProject     `json:"projects"`
	Findings         []ReportSnapshotFinding     `json:"findings"`
	PublishedAt      time.Time                   `json:"publishedAt"`
	PublishedBy      string                      `json:"publishedBy"`
	Version          int                         `json:"version"`
}

type ReportSnapshotProject struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ReportSnapshotFinding struct {
	ID            uint                   `json:"id"`
	Title         string                 `json:"title"`
	Severity      string                 `json:"severity"`
	Status        string                 `json:"status"`
	Summary       string                 `json:"summary"`
	ProjectID     uint                   `json:"projectId"`
	ProjectName   string                 `json:"projectName"`
	Vulnerability map[string]interface{} `json:"vulnerability,omitempty"`
}
