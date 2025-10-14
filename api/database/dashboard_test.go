package database

import (
	"fmt"
	"testing"
	"time"

	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDashboardTestDB(t *testing.T) func() {
	t.Helper()

	originalDB := db

	testDB, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("opening test database: %v", err)
	}

	if err := testDB.AutoMigrate(&ProjectData{}, &JSONData{}); err != nil {
		t.Fatalf("migrating schema: %v", err)
	}

	db = testDB

	return func() {
		db = originalDB
	}
}

func TestGetDashboardMetricsRespectsProjectAccess(t *testing.T) {
	cleanup := setupDashboardTestDB(t)
	defer cleanup()

	userEmail := "user@example.com"

	accessibleProject := ProjectData{ProjectName: "Accessible", ClientEmail: userEmail}
	if err := db.Create(&accessibleProject).Error; err != nil {
		t.Fatalf("creating accessible project: %v", err)
	}

	bugBountyProject := ProjectData{ProjectName: "Bounty", ClientEmail: userEmail, IsBugBounty: true}
	if err := db.Create(&bugBountyProject).Error; err != nil {
		t.Fatalf("creating bug bounty project: %v", err)
	}

	restrictedProject := ProjectData{ProjectName: "Restricted", ClientEmail: "other@example.com"}
	if err := db.Create(&restrictedProject).Error; err != nil {
		t.Fatalf("creating restricted project: %v", err)
	}

	insertVulnerability := func(project ProjectData, foundBy, visibility, assignedTo, status, criticality, category string, createdAt time.Time) {
		payload := fmt.Sprintf(`{"criticality":"%s","category":"%s","visibility":"%s","assignedTo":"%s"}`,
			criticality, category, visibility, assignedTo)

		pid := project.ID
		record := JSONData{
			Vulnerability: datatypes.JSON([]byte(payload)),
			FoundBy:       foundBy,
			Status:        status,
			ProjectID:     &pid,
		}
		record.CreatedAt = createdAt

		if err := db.Create(&record).Error; err != nil {
			t.Fatalf("creating vulnerability for project %s: %v", project.ProjectName, err)
		}
	}

	date2025 := time.Date(2025, time.January, 10, 0, 0, 0, 0, time.UTC)

	// Accessible project vulnerability (should be counted)
	insertVulnerability(accessibleProject, userEmail, "public", "", "open", "high", "A1", date2025)

	// Bug bounty project vulnerability (should be counted and contribute to bug bounty total)
	insertVulnerability(bugBountyProject, "reporter@example.com", "public", "", "closed", "medium", "A2", date2025)

	// Restricted project vulnerability (should not be counted)
	insertVulnerability(restrictedProject, "other@example.com", "public", "", "triaged", "critical", "A3", date2025)

	// Hidden vulnerability in accessible project assigned to someone else (should be filtered out)
	insertVulnerability(accessibleProject, "other@example.com", "hidden", "another@example.com", "open", "low", "A1", date2025)

	metrics, err := GetDashboardMetrics("2025", userEmail, false, false)
	if err != nil {
		t.Fatalf("GetDashboardMetrics returned error: %v", err)
	}

	if metrics.Total != 2 {
		t.Fatalf("expected total 2, got %d", metrics.Total)
	}

	if metrics.Projects != 2 {
		t.Fatalf("expected projects 2, got %d", metrics.Projects)
	}

	if metrics.BugBounties != 1 {
		t.Fatalf("expected bug bounty count 1, got %d", metrics.BugBounties)
	}

	if got := metrics.Statuses["open"]; got != 1 {
		t.Fatalf("expected open status count 1, got %d", got)
	}

	if _, exists := metrics.Statuses["triaged"]; exists {
		t.Fatalf("expected triaged status to be excluded, but it was present")
	}

	if got := metrics.Criticalities["high"]; got != 1 {
		t.Fatalf("expected high criticality count 1, got %d", got)
	}

	if got := metrics.Criticalities["medium"]; got != 1 {
		t.Fatalf("expected medium criticality count 1, got %d", got)
	}

	if _, exists := metrics.Criticalities["critical"]; exists {
		t.Fatalf("expected critical criticality to be excluded, but it was present")
	}

	if got := metrics.OWASP["A1"]; got != 1 {
		t.Fatalf("expected OWASP A1 count 1, got %d", got)
	}

	if got := metrics.OWASP["A2"]; got != 1 {
		t.Fatalf("expected OWASP A2 count 1, got %d", got)
	}

	if _, exists := metrics.OWASP["A3"]; exists {
		t.Fatalf("expected OWASP A3 category to be excluded, but it was present")
	}

	criticalityBreakdown, ok := metrics.OWASPCriticalities["A2"]
	if !ok {
		t.Fatalf("expected breakdown for category A2 to exist")
	}

	if got := criticalityBreakdown["medium"]; got != 1 {
		t.Fatalf("expected OWASP A2 medium count 1, got %d", got)
	}

	if breakdownA1, ok := metrics.OWASPCriticalities["A1"]; ok {
		if got := breakdownA1["high"]; got != 1 {
			t.Fatalf("expected OWASP A1 high count 1, got %d", got)
		}

		if got := breakdownA1["low"]; got != 0 {
			t.Fatalf("expected OWASP A1 low count 0, got %d", got)
		}
	} else {
		t.Fatalf("expected breakdown for category A1 to exist")
	}
}
