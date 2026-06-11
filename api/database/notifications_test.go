package database

import (
	"sync"
	"testing"

	"prism/models"
)

func TestCreateNotification_InsertsRow(t *testing.T) {
	defer setupNotificationsTestDB(t)()

	if _, err := CreateNotification("alice@x", models.Notification{
		Kind:  models.NotificationKindNewVuln,
		Who:   "bob@x",
		What:  "New vuln",
		Where: "/vulnerability/42/view",
	}); err != nil {
		t.Fatalf("create: %v", err)
	}
	rows, err := GetNotifications("alice@x", 10)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	got := rows[0]
	if got.Who != "bob@x" || got.What != "New vuln" || got.Where != "/vulnerability/42/view" {
		t.Fatalf("row fields not preserved: %+v", got)
	}
	if got.IsRead {
		t.Fatalf("new row should be unread")
	}
	if got.When.IsZero() {
		t.Fatalf("created_at should be set")
	}
}

func TestCreateNotification_EmptyRecipientIsNoop(t *testing.T) {
	defer setupNotificationsTestDB(t)()
	if _, err := CreateNotification("", models.Notification{What: "ghost"}); err != nil {
		t.Fatalf("empty recipient should be no-op, got error: %v", err)
	}
	var count int64
	if err := db.Model(&Notification{}).Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no rows after empty-recipient create, got %d", count)
	}
}

func TestGetNotifications_OnlyReturnsRecipientsRows(t *testing.T) {
	defer setupNotificationsTestDB(t)()
	for _, who := range []string{"alice@x", "alice@x", "alice@x"} {
		if _, err := CreateNotification(who, models.Notification{What: "alice msg"}); err != nil {
			t.Fatalf("create: %v", err)
		}
	}
	if _, err := CreateNotification("bob@x", models.Notification{What: "bob msg"}); err != nil {
		t.Fatalf("create bob: %v", err)
	}

	rows, err := GetNotifications("alice@x", 10)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(rows) != 3 {
		t.Fatalf("expected 3 alice rows, got %d", len(rows))
	}
	for _, r := range rows {
		if r.What != "alice msg" {
			t.Fatalf("leaked non-alice row: %+v", r)
		}
	}
}

func TestGetNotifications_HonoursLimitAndOrdersDesc(t *testing.T) {
	defer setupNotificationsTestDB(t)()
	for i := 0; i < 5; i++ {
		if _, err := CreateNotification("alice@x", models.Notification{What: "msg"}); err != nil {
			t.Fatalf("create: %v", err)
		}
	}
	rows, err := GetNotifications("alice@x", 2)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows under limit, got %d", len(rows))
	}
	// IDs auto-increment so DESC ordering is observable through them.
	if rows[0].ID < rows[1].ID {
		t.Fatalf("expected DESC by created_at (proxied by id), got %d before %d", rows[0].ID, rows[1].ID)
	}
}

func TestMarkNotificationAsRead_ScopedByRecipient(t *testing.T) {
	defer setupNotificationsTestDB(t)()
	if _, err := CreateNotification("alice@x", models.Notification{What: "for alice"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	var row Notification
	if err := db.Where("recipient_email = ?", "alice@x").First(&row).Error; err != nil {
		t.Fatalf("load: %v", err)
	}

	// Bob tries to flip Alice's row — should be a no-op (record not found
	// at the scope, so the function returns gorm.ErrRecordNotFound).
	if err := MarkNotificationAsRead("bob@x", row.ID); err == nil {
		t.Fatalf("bob should not be able to mark alice's row read")
	}
	var unchanged Notification
	_ = db.First(&unchanged, row.ID).Error
	if unchanged.IsRead {
		t.Fatalf("alice's row was flipped by bob, that's the bug we're guarding against")
	}

	// Alice can flip her own row.
	if err := MarkNotificationAsRead("alice@x", row.ID); err != nil {
		t.Fatalf("alice mark read: %v", err)
	}
	var after Notification
	_ = db.First(&after, row.ID).Error
	if !after.IsRead {
		t.Fatalf("row should be marked read")
	}
}

func TestMarkAllNotificationsRead_FlipsUnreadOnly(t *testing.T) {
	defer setupNotificationsTestDB(t)()
	if _, err := CreateNotification("alice@x", models.Notification{What: "one"}); err != nil {
		t.Fatalf("create one: %v", err)
	}
	if _, err := CreateNotification("alice@x", models.Notification{What: "two"}); err != nil {
		t.Fatalf("create two: %v", err)
	}
	if _, err := CreateNotification("bob@x", models.Notification{What: "bob"}); err != nil {
		t.Fatalf("create bob: %v", err)
	}

	if err := MarkAllNotificationsRead("alice@x"); err != nil {
		t.Fatalf("mark all: %v", err)
	}

	var unreadAlice int64
	_ = db.Model(&Notification{}).
		Where("recipient_email = ? AND is_read = ?", "alice@x", false).
		Count(&unreadAlice).Error
	if unreadAlice != 0 {
		t.Fatalf("alice should have no unread rows, got %d", unreadAlice)
	}

	// Bob's row is untouched.
	var bobRow Notification
	_ = db.Where("recipient_email = ?", "bob@x").First(&bobRow).Error
	if bobRow.IsRead {
		t.Fatalf("bob's row was wrongly flipped: %+v", bobRow)
	}
}

func TestDeleteNotifications_ScopedByRecipient(t *testing.T) {
	defer setupNotificationsTestDB(t)()
	if _, err := CreateNotification("alice@x", models.Notification{What: "a"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, err := CreateNotification("bob@x", models.Notification{What: "b"}); err != nil {
		t.Fatalf("create bob: %v", err)
	}
	if err := DeleteNotifications("alice@x"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	aliceRows, _ := GetNotifications("alice@x", 10)
	if len(aliceRows) != 0 {
		t.Fatalf("alice should have no rows, got %d", len(aliceRows))
	}
	bobRows, _ := GetNotifications("bob@x", 10)
	if len(bobRows) != 1 {
		t.Fatalf("bob should still have his row, got %d", len(bobRows))
	}
}

// TestCreateNotification_ConcurrentWritesAllPersist is the regression guard
// for the lost-update race that the old UserData.Notifications JSON column
// had. N concurrent inserts must produce N rows; the old read-modify-write
// JSON path lost all but one.
func TestCreateNotification_ConcurrentWritesAllPersist(t *testing.T) {
	defer setupNotificationsTestDB(t)()
	const n = 25
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		i := i
		go func() {
			defer wg.Done()
			_, _ = CreateNotification("alice@x", models.Notification{
				What: "concurrent",
				Who:  "bob@x",
			})
			_ = i
		}()
	}
	wg.Wait()

	var count int64
	if err := db.Model(&Notification{}).
		Where("recipient_email = ?", "alice@x").
		Count(&count).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if int(count) != n {
		t.Fatalf("expected all %d concurrent rows to land, got %d (this is the old JSON-blob race)", n, count)
	}
}
