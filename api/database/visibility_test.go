package database

import (
	"fmt"
	"testing"

	"gorm.io/datatypes"
)

func TestCanViewVulnerability(t *testing.T) {
	reporterEmail := "reporter@example.com"
	assigneeEmail := "assignee@example.com"
	otherEmail := "viewer@example.com"

	makeEntry := func(visibility, assignedTo, foundBy string) JSONData {
		payload := fmt.Sprintf(`{"visibility":"%s","assignedTo":"%s"}`, visibility, assignedTo)
		return JSONData{
			Vulnerability: datatypes.JSON([]byte(payload)),
			FoundBy:       foundBy,
		}
	}

	tests := []struct {
		name       string
		entry      JSONData
		email      string
		isGlobal   bool
		wantAccess bool
	}{
		{
			name:       "global user has access",
			entry:      makeEntry("undisclosed", assigneeEmail, reporterEmail),
			email:      otherEmail,
			isGlobal:   true,
			wantAccess: true,
		},
		{
			name:       "published visible to project member",
			entry:      makeEntry("published", "", reporterEmail),
			email:      otherEmail,
			isGlobal:   false,
			wantAccess: true,
		},
		{
			name:       "reporter can see undisclosed",
			entry:      makeEntry("undisclosed", assigneeEmail, reporterEmail),
			email:      reporterEmail,
			isGlobal:   false,
			wantAccess: true,
		},
		{
			name:       "assignee can see undisclosed",
			entry:      makeEntry("undisclosed", assigneeEmail, reporterEmail),
			email:      assigneeEmail,
			isGlobal:   false,
			wantAccess: true,
		},
		{
			name:       "non privileged cannot see undisclosed",
			entry:      makeEntry("undisclosed", assigneeEmail, reporterEmail),
			email:      otherEmail,
			isGlobal:   false,
			wantAccess: false,
		},
		{
			name:       "hidden treated as restricted",
			entry:      makeEntry("hidden", "", reporterEmail),
			email:      otherEmail,
			isGlobal:   false,
			wantAccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanViewVulnerability(tt.entry, tt.email, tt.isGlobal); got != tt.wantAccess {
				t.Fatalf("CanViewVulnerability() = %v, want %v", got, tt.wantAccess)
			}
		})
	}
}
