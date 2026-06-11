package routes

import (
	"testing"

	"prism/config"
	"prism/database"
	"prism/models"

	"gorm.io/datatypes"
)

// accessibleVulnerabilitiesViewSQL mirrors initAccessibleVulnerabilitiesView
// in the database package (unexported, so the routes tests recreate the view
// themselves on the in-memory test DB).
const accessibleVulnerabilitiesViewSQL = `
	CREATE VIEW accessible_vulnerabilities AS
	SELECT
			json_data.id AS id,
			json_extract(json_data.vulnerability, '$.visibility') AS visibility,
			json_extract(json_data.vulnerability, '$.assignedTo') AS assigned_to,
			json_data.found_by,
			json_data.project_id,
			project_data.client_email,
			project_data.hacker_name
	FROM json_data
	LEFT JOIN project_data ON json_data.project_id = project_data.id;
`

// TestDispatch_FiltersByACLAndPublishesSSEOnlyToSurvivors is the isolation
// proof for the streaming path: a recipient whose access to the vulnerability
// has been revoked gets neither an in-app row nor an SSE event, while a
// project member gets both. (This also covers the ACL hop through
// CanRecipientSeeVulnerability that the other Dispatch tests skip by leaving
// VulnerabilityID at 0.)
func TestDispatch_FiltersByACLAndPublishesSSEOnlyToSurvivors(t *testing.T) {
	defer setupDispatchTestDB(t)()

	testDB := database.DBForTest()
	if err := testDB.AutoMigrate(&database.JSONData{}, &database.ProjectData{}); err != nil {
		t.Fatalf("migrate vuln tables: %v", err)
	}
	if err := testDB.Exec(accessibleVulnerabilitiesViewSQL).Error; err != nil {
		t.Fatalf("create view: %v", err)
	}

	// Role config: "hacker" has no global /vulnerability/:id permission, so
	// access is decided per-vulnerability by the view. AppConfig is nil in
	// tests (no config file is loaded), so install a minimal one.
	originalConfig := config.AppConfig
	config.AppConfig = &config.Config{Roles: map[string]config.Role{
		"hacker": {Permissions: []config.Permission{}},
	}}
	defer func() { config.AppConfig = originalConfig }()

	active := true
	for _, email := range []string{"alice@x", "revoked@x"} {
		if err := testDB.Create(&database.UserData{
			Email: email, Role: "hacker", Active: &active,
		}).Error; err != nil {
			t.Fatalf("seed user %s: %v", email, err)
		}
	}

	// Alice is on the project; revoked@x is not (any more).
	project := database.ProjectData{ProjectName: "P", HackerName: "alice@x"}
	if err := testDB.Create(&project).Error; err != nil {
		t.Fatalf("seed project: %v", err)
	}
	vuln := database.JSONData{
		Vulnerability: datatypes.JSON([]byte(`{"title":"XSS"}`)),
		FoundBy:       "carol@x",
		ProjectID:     &project.ID,
	}
	if err := testDB.Create(&vuln).Error; err != nil {
		t.Fatalf("seed vuln: %v", err)
	}

	original := pushSender
	pushSender = func(s database.Subscriber, payload []byte) error { return nil }
	defer func() { pushSender = original }()

	aliceStream := notificationHub.subscribe("alice@x")
	defer notificationHub.unsubscribe("alice@x", aliceStream)
	revokedStream := notificationHub.subscribe("revoked@x")
	defer notificationHub.unsubscribe("revoked@x", revokedStream)

	if err := Dispatch(DispatchRequest{
		Kind:            models.NotificationKindNewVuln,
		VulnerabilityID: vuln.ID,
		ActorEmail:      "carol@x",
		Recipients:      []string{"alice@x", "revoked@x"},
		Title:           "P",
		Body:            "New vulnerability XSS",
		URL:             "/vulnerability/1/view",
	}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	// Alice's stream received exactly the created event, with the row id.
	select {
	case evt := <-aliceStream.ch:
		if evt.Type != "notification.created" {
			t.Fatalf("alice event type = %q, want notification.created", evt.Type)
		}
		if evt.Notification == nil || evt.Notification.ID == 0 {
			t.Fatalf("alice event missing notification payload: %+v", evt)
		}
		if evt.Notification.What != "New vulnerability XSS" || evt.Notification.Who != "carol@x" {
			t.Fatalf("alice event content wrong: %+v", evt.Notification)
		}
	default:
		t.Fatalf("alice should have received an SSE event")
	}

	// The revoked user's stream stays silent.
	select {
	case evt := <-revokedStream.ch:
		t.Fatalf("revoked user must not receive SSE events, got %+v", evt)
	default:
	}

	// And the database agrees: one row for alice, none for revoked.
	aliceRows, _ := database.GetNotifications("alice@x", 10)
	revokedRows, _ := database.GetNotifications("revoked@x", 10)
	if len(aliceRows) != 1 {
		t.Fatalf("alice should have 1 in-app row, got %d", len(aliceRows))
	}
	if len(revokedRows) != 0 {
		t.Fatalf("revoked user must have 0 in-app rows, got %d", len(revokedRows))
	}
}

// TestMarkReadPublishesToOwnStreamOnly verifies the cross-tab sync events are
// also scoped per user: marking a row read publishes to the owner's stream
// and nobody else's.
func TestMarkReadPublishesToOwnStreamOnly(t *testing.T) {
	defer setupDispatchTestDB(t)()

	created, err := database.CreateNotification("alice@x", models.Notification{What: "hi"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	aliceStream := notificationHub.subscribe("alice@x")
	defer notificationHub.unsubscribe("alice@x", aliceStream)
	bobStream := notificationHub.subscribe("bob@x")
	defer notificationHub.unsubscribe("bob@x", bobStream)

	if err := database.MarkNotificationAsRead("alice@x", created.ID); err != nil {
		t.Fatalf("mark read: %v", err)
	}
	// The HTTP handler publishes after the DB call; emulate that here since
	// we exercise the hub directly rather than spinning up gin.
	notificationHub.publish("alice@x", NotificationEvent{Type: "notification.read", ID: created.ID})

	select {
	case evt := <-aliceStream.ch:
		if evt.Type != "notification.read" || evt.ID != created.ID {
			t.Fatalf("alice read event wrong: %+v", evt)
		}
	default:
		t.Fatalf("alice should have received the read event")
	}
	select {
	case evt := <-bobStream.ch:
		t.Fatalf("bob must not see alice's read event, got %+v", evt)
	default:
	}
}
