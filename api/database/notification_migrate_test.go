package database

import (
	"testing"
	"time"

	"github.com/SherClockHolmes/webpush-go"
)

func TestMigrateNotifications_Idempotent(t *testing.T) {
	defer setupNotificationsTestDB(t)()
	// The setup helper already ran the migration once. A second run on a
	// clean steady-state schema must succeed with an effectively empty
	// report.
	report, err := MigrateNotifications()
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if report.SubscribersCleared != 0 || report.JSONColumnCleared != 0 {
		t.Fatalf("idempotent run should clear nothing, got %+v", report)
	}
	if !report.UniqueIndexInstalled {
		t.Fatalf("unique index should still report installed on idempotent run")
	}
}

// TestMigrateNotifications_WipesPreRefactorSubscribers verifies the
// "wipe the bad state" behaviour the user asked for: rows that pre-date the
// typed-endpoint schema (endpoint column NULL or empty) get hard-deleted on
// migration, so the new endpoint-keyed upsert path starts from a clean slate.
func TestMigrateNotifications_WipesPreRefactorSubscribers(t *testing.T) {
	defer setupNotificationsTestDB(t)()

	// Insert a row that looks pre-refactor: no endpoint. The migration
	// previously installed the unique index, so we add this row by going
	// around the normal upsert path.
	preRefactor := Subscriber{
		Email:      "legacy@x",
		Endpoint:   "",
		LastSeenAt: time.Now(),
	}
	if err := db.Create(&preRefactor).Error; err != nil {
		t.Fatalf("seed legacy row: %v", err)
	}
	// And a properly-shaped post-refactor row.
	if err := UpsertSubscriber("modern@x", webpush.Subscription{
		Endpoint: "https://push/modern",
		Keys:     webpush.Keys{P256dh: "p", Auth: "a"},
	}, "ua"); err != nil {
		t.Fatalf("modern upsert: %v", err)
	}

	report, err := MigrateNotifications()
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if report.SubscribersCleared != 1 {
		t.Fatalf("expected 1 legacy row cleared, got %d (report: %+v)", report.SubscribersCleared, report)
	}

	// Legacy row gone, modern row preserved.
	var rows []Subscriber
	if err := db.Find(&rows).Error; err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected exactly 1 surviving row, got %d", len(rows))
	}
	if rows[0].Email != "modern@x" {
		t.Fatalf("wrong row survived: %+v", rows[0])
	}
}

func TestMigrateNotifications_DoesNotImportLegacyInAppHistory(t *testing.T) {
	defer setupNotificationsTestDB(t)()
	// AutoMigrate of UserData (post-refactor) doesn't include the
	// `notifications` JSON column anymore. The migration's
	// clearUserNotificationsJSON branch gracefully no-ops when the column
	// doesn't exist — which is the steady state we end up in here. The
	// guard we're asserting is that no rows leak into the new notifications
	// table from migration: it's a wipe, not a drain.
	if _, err := MigrateNotifications(); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	var count int64
	if err := db.Model(&Notification{}).Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("migration should never insert into notifications, got %d", count)
	}
}

// TestMigrateNotifications_UniqueIndexBlocksDuplicateEndpoints proves the
// installed index actually prevents two rows from ever sharing an endpoint.
// Without this constraint, the take-over logic in UpsertSubscriber can race
// itself.
func TestMigrateNotifications_UniqueIndexBlocksDuplicateEndpoints(t *testing.T) {
	defer setupNotificationsTestDB(t)()

	if err := db.Create(&Subscriber{
		Email:      "first@x",
		Endpoint:   "https://push/dup",
		LastSeenAt: time.Now(),
	}).Error; err != nil {
		t.Fatalf("first insert: %v", err)
	}
	err := db.Create(&Subscriber{
		Email:      "second@x",
		Endpoint:   "https://push/dup",
		LastSeenAt: time.Now(),
	}).Error
	if err == nil {
		t.Fatalf("expected unique-index violation on duplicate endpoint")
	}
}
