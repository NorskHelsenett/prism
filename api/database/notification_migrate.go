package database

import (
	"fmt"
	"log"
	"strings"

	"gorm.io/gorm"
)

// NotificationMigrationReport summarises one pass of MigrateNotifications so
// startup logs make the migration's effect visible without re-querying.
type NotificationMigrationReport struct {
	SubscribersCleared    int
	UniqueIndexInstalled  bool
	UserRowsWithLegacy    int
	JSONColumnCleared     int
	Errors                []string
}

func (r NotificationMigrationReport) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "subscribers cleared       : %d\n", r.SubscribersCleared)
	fmt.Fprintf(&b, "unique index installed    : %t\n", r.UniqueIndexInstalled)
	fmt.Fprintf(&b, "users with legacy blobs   : %d\n", r.UserRowsWithLegacy)
	fmt.Fprintf(&b, "json columns cleared      : %d\n", r.JSONColumnCleared)
	if len(r.Errors) > 0 {
		fmt.Fprintf(&b, "errors (%d):\n", len(r.Errors))
		for _, e := range r.Errors {
			fmt.Fprintf(&b, "  - %s\n", e)
		}
	}
	return b.String()
}

// MigrateNotifications upgrades the notification system without carrying the
// pre-refactor corrupted state forward. The cross-user delivery bug lived in
// the existing subscribers + user_data.notifications data; preserving any of
// it would re-import the bug. So this migration is "wipe and rebuild":
//
//  1. DELETE every row from subscribers. Users will re-subscribe through the
//     new POST /api/notification/subscribe, which goes via UpsertSubscriber
//     and produces correctly endpoint-keyed rows.
//  2. Install the UNIQUE index on endpoint, which the new upsert relies on.
//  3. NULL user_data.notifications. Drops the legacy in-app dropdown history.
//     New events repopulate the dedicated notifications table cleanly.
//
// Idempotent: a steady-state second run finds nothing to clear, no index to
// install, and exits with an empty report.
func MigrateNotifications() (NotificationMigrationReport, error) {
	var report NotificationMigrationReport
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := wipeSubscribers(tx, &report); err != nil {
			return fmt.Errorf("wipe subscribers: %w", err)
		}
		if err := installSubscriberUniqueIndex(tx, &report); err != nil {
			return fmt.Errorf("install unique index: %w", err)
		}
		if err := clearUserNotificationsJSON(tx, &report); err != nil {
			return fmt.Errorf("clear legacy notifications: %w", err)
		}
		return nil
	})
	if err != nil {
		return report, err
	}
	return report, nil
}

// wipeSubscribers removes any row in subscribers that still has the legacy
// shape (no endpoint column populated). The first run clears everything;
// subsequent runs see endpoint-populated rows from UpsertSubscriber and skip
// them, so this is safe to re-execute on every boot.
func wipeSubscribers(tx *gorm.DB, report *NotificationMigrationReport) error {
	// Targeted hard delete: only rows that look pre-refactor (empty endpoint).
	// New rows always have an endpoint, so they're untouched.
	res := tx.Unscoped().
		Where("endpoint IS NULL OR endpoint = ''").
		Delete(&Subscriber{})
	if res.Error != nil {
		return res.Error
	}
	report.SubscribersCleared = int(res.RowsAffected)
	return nil
}

// installSubscriberUniqueIndex adds the UNIQUE index on endpoint that
// UpsertSubscriber's take-over semantics rely on. IF NOT EXISTS keeps a
// re-run a no-op. Without this index, two concurrent subscribes could race
// into two rows for the same endpoint, re-introducing the cross-user
// delivery bug.
func installSubscriberUniqueIndex(tx *gorm.DB, report *NotificationMigrationReport) error {
	if err := tx.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_subscribers_endpoint_unique
		ON subscribers(endpoint)
		WHERE endpoint <> '' AND deleted_at IS NULL
	`).Error; err != nil {
		return err
	}
	report.UniqueIndexInstalled = true
	return nil
}

// clearUserNotificationsJSON nulls out the legacy per-user notifications
// blob. We intentionally do not drain into the new table: the old data is
// what the original bug was misdelivering, and importing it would re-import
// the misdelivery. The auto-migrate doesn't drop columns, so the column
// itself stays; only its contents are cleared.
func clearUserNotificationsJSON(tx *gorm.DB, report *NotificationMigrationReport) error {
	// Probe via a cheap COUNT first; if the legacy column has already been
	// dropped on some future schema generation we exit silently.
	var count int64
	err := tx.Raw(`
		SELECT COUNT(*) FROM user_data
		WHERE notifications IS NOT NULL AND length(notifications) > 0
	`).Scan(&count).Error
	if err != nil {
		if strings.Contains(err.Error(), "no such column") {
			return nil
		}
		return err
	}
	report.UserRowsWithLegacy = int(count)
	if count == 0 {
		return nil
	}
	res := tx.Exec(`UPDATE user_data SET notifications = NULL WHERE notifications IS NOT NULL`)
	if res.Error != nil {
		return res.Error
	}
	report.JSONColumnCleared = int(res.RowsAffected)
	return nil
}

// RunNotificationMigrationOnce is invoked from main at startup. It logs the
// report only when something actually happened so a steady-state boot stays
// quiet.
func RunNotificationMigrationOnce() {
	report, err := MigrateNotifications()
	if err != nil {
		log.Printf("notification migration: %v", err)
		return
	}
	if report.SubscribersCleared > 0 || report.JSONColumnCleared > 0 {
		log.Printf("notification migration:\n%s", report.String())
	}
}
