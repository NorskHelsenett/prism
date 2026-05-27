package database

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/SherClockHolmes/webpush-go"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// testDBCounter gives each setup call a unique in-memory database. Using
// `file::memory:?cache=shared` without a unique name shares one backing
// store across every open() in the process, so tests in the same package
// leak state into each other — which is exactly the failure mode the
// dashboard tests dodge by having only one test per file.
var testDBCounter uint64

// setupNotificationsTestDB swaps the package-level db for an isolated
// in-memory sqlite, auto-migrates the subscriber + notification + user_data
// schemas, and runs the notification migration so the UNIQUE endpoint index
// is installed — which is what UpsertSubscriber's take-over semantics
// actually rely on.
func setupNotificationsTestDB(t *testing.T) func() {
	t.Helper()
	original := db
	dsn := fmt.Sprintf("file:notif_test_%d?mode=memory&cache=shared",
		atomic.AddUint64(&testDBCounter, 1))
	testDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	// SQLite is single-writer at the file/cache level, and the in-memory
	// shared-cache mode used here aggravates SQLITE_BUSY when GORM opens
	// multiple connections. Pin to one connection so concurrent goroutines
	// in the race-regression test serialize cleanly through the driver
	// instead of fighting the lock. Production InitDB caps to 10 + sets
	// busy_timeout via WAL pragmas; the equivalent for an in-memory DB is
	// just "one connection".
	sqlDB, err := testDB.DB()
	if err != nil {
		t.Fatalf("get sql.DB: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	if err := testDB.Exec("PRAGMA busy_timeout = 5000;").Error; err != nil {
		t.Fatalf("busy_timeout: %v", err)
	}
	if err := testDB.AutoMigrate(&Subscriber{}, &Notification{}, &UserData{}); err != nil {
		t.Fatalf("migrate schema: %v", err)
	}
	db = testDB
	if _, err := MigrateNotifications(); err != nil {
		t.Fatalf("notification migration: %v", err)
	}
	return func() { db = original }
}

func makeSub(endpoint string) webpush.Subscription {
	return webpush.Subscription{
		Endpoint: endpoint,
		Keys: webpush.Keys{
			P256dh: "p256_" + endpoint,
			Auth:   "auth_" + endpoint,
		},
	}
}

func TestUpsertSubscriber_CreatesRow(t *testing.T) {
	defer setupNotificationsTestDB(t)()

	if err := UpsertSubscriber("alice@x", makeSub("https://push/alice-1"), "ua/1"); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	rows, err := ListSubscribersForEmail("alice@x")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("want 1 row, got %d", len(rows))
	}
	got := rows[0]
	if got.Endpoint != "https://push/alice-1" || got.P256dh != "p256_https://push/alice-1" || got.UserAgent != "ua/1" {
		t.Fatalf("row not populated as expected: %+v", got)
	}
	if got.LastSeenAt.IsZero() {
		t.Fatalf("last_seen_at should be set")
	}
}

func TestUpsertSubscriber_SameOwnerUpdatesNotDuplicates(t *testing.T) {
	defer setupNotificationsTestDB(t)()

	endpoint := "https://push/alice-2"
	if err := UpsertSubscriber("alice@x", makeSub(endpoint), "ua/old"); err != nil {
		t.Fatalf("first upsert: %v", err)
	}
	// Resubscribe with the same endpoint but new keys (this is what the
	// browser hands the server on every page load after permission was
	// granted; the data shouldn't accumulate rows).
	rotated := makeSub(endpoint)
	rotated.Keys.P256dh = "rotated-p256"
	rotated.Keys.Auth = "rotated-auth"
	if err := UpsertSubscriber("alice@x", rotated, "ua/new"); err != nil {
		t.Fatalf("second upsert: %v", err)
	}

	rows, err := ListSubscribersForEmail("alice@x")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row after re-subscribe, got %d", len(rows))
	}
	if rows[0].P256dh != "rotated-p256" || rows[0].Auth != "rotated-auth" || rows[0].UserAgent != "ua/new" {
		t.Fatalf("keys not rotated: %+v", rows[0])
	}
}

// TestUpsertSubscriber_TakesOverFromDifferentOwner is the core regression
// guard for the cross-user push delivery bug. Alice subscribes on a browser
// (endpoint X bound to alice@x). Bob then logs in on the same browser and
// subscribes — the same endpoint comes back from pushManager because it's
// per-browser, not per-user. After Bob's upsert, only Bob should own X.
func TestUpsertSubscriber_TakesOverFromDifferentOwner(t *testing.T) {
	defer setupNotificationsTestDB(t)()

	endpoint := "https://push/shared-browser"
	if err := UpsertSubscriber("alice@x", makeSub(endpoint), "shared"); err != nil {
		t.Fatalf("alice upsert: %v", err)
	}
	if err := UpsertSubscriber("bob@x", makeSub(endpoint), "shared"); err != nil {
		t.Fatalf("bob upsert: %v", err)
	}

	aliceRows, _ := ListSubscribersForEmail("alice@x")
	bobRows, _ := ListSubscribersForEmail("bob@x")
	if len(aliceRows) != 0 {
		t.Fatalf("alice should no longer own endpoint, got %d rows", len(aliceRows))
	}
	if len(bobRows) != 1 || bobRows[0].Endpoint != endpoint {
		t.Fatalf("bob should own endpoint after take-over, got %+v", bobRows)
	}

	// And the (endpoint -> recipients) view used by the dispatcher must not
	// see any alice rows for this endpoint.
	all, _ := ListSubscribersForEmails([]string{"alice@x", "bob@x"})
	for _, r := range all {
		if r.Endpoint == endpoint && r.Email == "alice@x" {
			t.Fatalf("stale alice row for endpoint still in result set: %+v", r)
		}
	}
}

func TestUpsertSubscriber_RejectsMissingEndpoint(t *testing.T) {
	defer setupNotificationsTestDB(t)()
	err := UpsertSubscriber("alice@x", webpush.Subscription{}, "")
	if err == nil {
		t.Fatalf("expected error for empty endpoint")
	}
}

// TestUpsertSubscriber_RejectsTakeOverWithoutMatchingKeys is the regression
// guard for the pentest F-01 finding: an authed user must not be able to
// yank another user's subscription by knowing only the endpoint URL. The
// take-over is allowed only when the requesting browser can prove it holds
// the underlying subscription by presenting the same p256dh/auth keys.
func TestUpsertSubscriber_RejectsTakeOverWithoutMatchingKeys(t *testing.T) {
	defer setupNotificationsTestDB(t)()

	endpoint := "https://push/contested"
	original := webpush.Subscription{
		Endpoint: endpoint,
		Keys:     webpush.Keys{P256dh: "real-p256", Auth: "real-auth"},
	}
	if err := UpsertSubscriber("alice@x", original, "alice-ua"); err != nil {
		t.Fatalf("alice upsert: %v", err)
	}

	// Attacker knows the URL but not the keys.
	urlOnly := webpush.Subscription{
		Endpoint: endpoint,
		// p256dh + auth empty (the live pentest attack)
	}
	if err := UpsertSubscriber("bob@x", urlOnly, "bob-ua"); !errors.Is(err, ErrSubscriberEndpointClaimed) {
		t.Fatalf("expected ErrSubscriberEndpointClaimed for url-only take-over, got %v", err)
	}
	// Attacker guesses wrong keys.
	wrongKeys := webpush.Subscription{
		Endpoint: endpoint,
		Keys:     webpush.Keys{P256dh: "guessed", Auth: "guessed"},
	}
	if err := UpsertSubscriber("bob@x", wrongKeys, "bob-ua"); !errors.Is(err, ErrSubscriberEndpointClaimed) {
		t.Fatalf("expected ErrSubscriberEndpointClaimed for wrong-keys take-over, got %v", err)
	}

	// Alice's row is intact.
	rows, err := ListSubscribersForEmail("alice@x")
	if err != nil {
		t.Fatalf("list alice: %v", err)
	}
	if len(rows) != 1 || rows[0].P256dh != "real-p256" || rows[0].Auth != "real-auth" {
		t.Fatalf("alice's row was clobbered by failed take-over: %+v", rows)
	}
	// Bob has no row for this endpoint.
	bobRows, _ := ListSubscribersForEmail("bob@x")
	if len(bobRows) != 0 {
		t.Fatalf("bob should have no row after rejected take-over, got %+v", bobRows)
	}
}

// TestUpsertSubscriber_AllowsTakeOverWhenKeysMatch covers the legitimate
// shared-browser path: Alice logs out, Bob logs in on the same browser, the
// browser returns the same endpoint *and* the same keys from
// pushManager.getSubscription, take-over succeeds.
func TestUpsertSubscriber_AllowsTakeOverWhenKeysMatch(t *testing.T) {
	defer setupNotificationsTestDB(t)()

	endpoint := "https://push/shared"
	sub := webpush.Subscription{
		Endpoint: endpoint,
		Keys:     webpush.Keys{P256dh: "shared-p256", Auth: "shared-auth"},
	}
	if err := UpsertSubscriber("alice@x", sub, ""); err != nil {
		t.Fatalf("alice upsert: %v", err)
	}
	if err := UpsertSubscriber("bob@x", sub, ""); err != nil {
		t.Fatalf("bob take-over with matching keys should succeed: %v", err)
	}
	aliceRows, _ := ListSubscribersForEmail("alice@x")
	bobRows, _ := ListSubscribersForEmail("bob@x")
	if len(aliceRows) != 0 || len(bobRows) != 1 || bobRows[0].Endpoint != endpoint {
		t.Fatalf("matching-keys take-over didn't rebind cleanly: alice=%v bob=%v", aliceRows, bobRows)
	}
}

func TestDeleteSubscriberByEndpoint_ScopedToOwner(t *testing.T) {
	defer setupNotificationsTestDB(t)()
	if err := UpsertSubscriber("alice@x", makeSub("https://push/a"), ""); err != nil {
		t.Fatalf("alice upsert: %v", err)
	}
	if err := UpsertSubscriber("bob@x", makeSub("https://push/b"), ""); err != nil {
		t.Fatalf("bob upsert: %v", err)
	}

	// Bob tries to delete by an endpoint he doesn't own — should be a no-op,
	// not a leak that nukes Alice's row.
	if err := DeleteSubscriberByEndpoint("bob@x", "https://push/a"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	aliceRows, _ := ListSubscribersForEmail("alice@x")
	if len(aliceRows) != 1 {
		t.Fatalf("alice's row was wrongly removed by bob, got %d rows", len(aliceRows))
	}

	// Bob deletes his own row — should disappear.
	if err := DeleteSubscriberByEndpoint("bob@x", "https://push/b"); err != nil {
		t.Fatalf("delete own: %v", err)
	}
	bobRows, _ := ListSubscribersForEmail("bob@x")
	if len(bobRows) != 0 {
		t.Fatalf("bob's own row should be gone, got %d", len(bobRows))
	}
}

func TestDeleteDeadSubscriber_RemovesAllRowsForEndpoint(t *testing.T) {
	defer setupNotificationsTestDB(t)()
	if err := UpsertSubscriber("alice@x", makeSub("https://push/dead"), ""); err != nil {
		t.Fatalf("upsert: %v", err)
	}
	if err := DeleteDeadSubscriber("https://push/dead"); err != nil {
		t.Fatalf("dead delete: %v", err)
	}
	rows, _ := ListSubscribersForEmail("alice@x")
	if len(rows) != 0 {
		t.Fatalf("dead endpoint not pruned: %d rows", len(rows))
	}
}

func TestListSubscribersForEmail_OnlyReturnsOwnersRows(t *testing.T) {
	defer setupNotificationsTestDB(t)()
	if err := UpsertSubscriber("alice@x", makeSub("https://push/a1"), ""); err != nil {
		t.Fatalf("a1: %v", err)
	}
	if err := UpsertSubscriber("alice@x", makeSub("https://push/a2"), ""); err != nil {
		t.Fatalf("a2: %v", err)
	}
	if err := UpsertSubscriber("bob@x", makeSub("https://push/b1"), ""); err != nil {
		t.Fatalf("b1: %v", err)
	}
	rows, err := ListSubscribersForEmail("alice@x")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 alice rows, got %d", len(rows))
	}
	for _, r := range rows {
		if r.Email != "alice@x" {
			t.Fatalf("leaked non-alice row: %+v", r)
		}
	}
}

// TestUpsertSubscriber_ConcurrentSameEndpointConvergesToOneOwner stresses the
// take-over path: two users racing to claim the same endpoint must end with
// exactly one row bound to whichever upsert won. The UNIQUE index installed
// by the migration plus the transaction inside UpsertSubscriber is what
// makes this safe.
func TestUpsertSubscriber_ConcurrentSameEndpointConvergesToOneOwner(t *testing.T) {
	defer setupNotificationsTestDB(t)()

	endpoint := "https://push/race"
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		_ = UpsertSubscriber("alice@x", makeSub(endpoint), "")
	}()
	go func() {
		defer wg.Done()
		_ = UpsertSubscriber("bob@x", makeSub(endpoint), "")
	}()
	wg.Wait()

	var rows []Subscriber
	if err := db.Where("endpoint = ?", endpoint).Find(&rows).Error; err != nil {
		t.Fatalf("query: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected exactly one row after race, got %d: %+v", len(rows), rows)
	}
	if rows[0].Email != "alice@x" && rows[0].Email != "bob@x" {
		t.Fatalf("unexpected owner after race: %+v", rows[0])
	}
}
