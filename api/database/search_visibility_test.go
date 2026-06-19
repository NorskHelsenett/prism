package database

import (
	"fmt"
	"strings"
	"sync/atomic"
	"testing"

	"prism/config"

	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupVulnSearchTestDB swaps the package-level db for an isolated in-memory
// sqlite, migrates the vulnerability-related schema and builds the
// accessible_vulnerabilities view that SearchVulnerabilities / CanAccessVulnerability
// join against in production.
func setupVulnSearchTestDB(t *testing.T) func() {
	t.Helper()
	original := db
	dsn := fmt.Sprintf("file:vuln_search_test_%d?mode=memory&cache=shared",
		atomic.AddUint64(&testDBCounter, 1))
	testDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	sqlDB, err := testDB.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := testDB.AutoMigrate(&ProjectData{}, &JSONData{}, &UserData{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}
	db = testDB
	initAccessibleVulnerabilitiesView()
	return func() { db = original }
}

func insertVuln(t *testing.T, foundBy, visibility, assignedTo, status string, projectID *uint) uint {
	t.Helper()
	payload := fmt.Sprintf(
		`{"title":"t","criticality":"high","date":"2025-01-01","category":"A1","visibility":"%s","assignedTo":"%s"}`,
		visibility, assignedTo)
	rec := JSONData{
		Vulnerability: datatypes.JSON([]byte(payload)),
		FoundBy:       foundBy,
		Status:        status,
		ProjectID:     projectID,
	}
	if err := db.Create(&rec).Error; err != nil {
		t.Fatalf("create vuln: %v", err)
	}
	return rec.ID
}

func idSet(res []MinimalJSONData) map[uint]bool {
	m := make(map[uint]bool, len(res))
	for _, r := range res {
		m[r.ID] = true
	}
	return m
}

func mustSearch(t *testing.T, global bool, email string, admin bool, p VulnerabilitySearchParams) map[uint]bool {
	t.Helper()
	res, err := SearchVulnerabilities(global, email, admin, p)
	if err != nil {
		t.Fatalf("SearchVulnerabilities(%s): %v", email, err)
	}
	return idSet(res)
}

// TestSearchVulnerabilities_DefaultHidesUndisclosed guards the table default:
// restricted-visibility findings are excluded unless the reveal toggle
// (IncludeUndisclosed) is set — even for an admin who is otherwise allowed to
// see everything.
func TestSearchVulnerabilities_DefaultHidesUndisclosed(t *testing.T) {
	defer setupVulnSearchTestDB(t)()

	admin := "admin@example.com"
	pubID := insertVuln(t, "reporter@example.com", "public", "", "Reported", nil)
	undID := insertVuln(t, "reporter@example.com", "undisclosed", "assignee@example.com", "Reported", nil)

	// Default: undisclosed is hidden.
	ids := mustSearch(t, true, admin, true, VulnerabilitySearchParams{})
	if !ids[pubID] {
		t.Fatalf("default search should return the public vuln %d; got %v", pubID, ids)
	}
	if ids[undID] {
		t.Fatalf("default search must hide the undisclosed vuln %d; got %v", undID, ids)
	}

	// Reveal: both returned.
	ids = mustSearch(t, true, admin, true, VulnerabilitySearchParams{IncludeUndisclosed: true})
	if !ids[pubID] || !ids[undID] {
		t.Fatalf("reveal search should return both vulns; got %v", ids)
	}
}

// TestSearchVulnerabilities_RevealRespectsAccess guards that revealing
// undisclosed findings never widens visibility past the reporter and assignee:
// a plain project member (or an outsider) still must not see them.
func TestSearchVulnerabilities_RevealRespectsAccess(t *testing.T) {
	defer setupVulnSearchTestDB(t)()

	reporter := "reporter@example.com"
	assignee := "assignee@example.com"
	member := "member@example.com"
	outsider := "outsider@example.com"

	// member is a project client but neither the reporter nor the assignee.
	proj := ProjectData{ProjectName: "P", ClientEmail: member}
	if err := db.Create(&proj).Error; err != nil {
		t.Fatalf("create project: %v", err)
	}
	pid := proj.ID
	undID := insertVuln(t, reporter, "undisclosed", assignee, "Reported", &pid)

	reveal := VulnerabilitySearchParams{IncludeUndisclosed: true}

	if ids := mustSearch(t, false, reporter, false, reveal); !ids[undID] {
		t.Fatalf("reporter should see their own undisclosed vuln when revealed")
	}
	if ids := mustSearch(t, false, assignee, false, reveal); !ids[undID] {
		t.Fatalf("assignee should see the assigned undisclosed vuln when revealed")
	}
	if ids := mustSearch(t, false, member, false, reveal); ids[undID] {
		t.Fatalf("project member who is neither reporter nor assignee must not see undisclosed vuln")
	}
	if ids := mustSearch(t, false, outsider, false, reveal); ids[undID] {
		t.Fatalf("outsider must not see undisclosed vuln")
	}
}

// TestCanRecipientSeeVulnerability_RestrictedNotificationGate guards the
// notification path: a restricted finding must only be delivered to its
// reporter or assignee, never to other members of the project — while
// non-restricted findings still reach project members.
func TestCanRecipientSeeVulnerability_RestrictedNotificationGate(t *testing.T) {
	defer setupVulnSearchTestDB(t)()

	origCfg := config.AppConfig
	config.AppConfig = &config.Config{Roles: map[string]config.Role{"hacker": {}}}
	defer func() { config.AppConfig = origCfg }()

	reporter := "reporter@example.com"
	assignee := "assignee@example.com"
	member := "member@example.com"

	active := true
	for _, e := range []string{reporter, assignee, member} {
		if err := db.Create(&UserData{Email: e, Role: "hacker", Active: &active}).Error; err != nil {
			t.Fatalf("create user %s: %v", e, err)
		}
	}

	// All three are project members, so CanAccessVulnerability passes for each;
	// the visibility gate is what must tell them apart.
	proj := ProjectData{ProjectName: "P", HackerName: strings.Join([]string{reporter, assignee, member}, ",")}
	if err := db.Create(&proj).Error; err != nil {
		t.Fatalf("create project: %v", err)
	}
	pid := proj.ID

	undID := insertVuln(t, reporter, "undisclosed", assignee, "Reported", &pid)
	pubID := insertVuln(t, reporter, "public", "", "Reported", &pid)

	cases := []struct {
		name    string
		email   string
		vulnID  uint
		wantSee bool
	}{
		{"reporter sees restricted", reporter, undID, true},
		{"assignee sees restricted", assignee, undID, true},
		{"plain member blocked from restricted", member, undID, false},
		{"plain member still notified of public", member, pubID, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := CanRecipientSeeVulnerability(c.email, c.vulnID)
			if err != nil {
				t.Fatalf("CanRecipientSeeVulnerability(%s, %d): %v", c.email, c.vulnID, err)
			}
			if got != c.wantSee {
				t.Fatalf("CanRecipientSeeVulnerability(%s, vuln=%d) = %v, want %v", c.email, c.vulnID, got, c.wantSee)
			}
		})
	}
}
