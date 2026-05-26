package routes

import (
	"fmt"
	"reflect"
	"sort"
	"sync"
	"sync/atomic"
	"testing"

	"prism/database"
	"prism/models"

	"github.com/SherClockHolmes/webpush-go"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// dispatchTestDBCounter — see comment in database/subscribers_test.go.
var dispatchTestDBCounter uint64

// --- pure-function tests (no DB) ---

func TestNormaliseRecipients_TrimsLowercasesAndDedupes(t *testing.T) {
	got := normaliseRecipients(
		[]string{"  Alice@x ", "alice@x", "BOB@x", "", "  ", "carol@x"},
		"",
	)
	sort.Strings(got)
	want := []string{"alice@x", "bob@x", "carol@x"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalise = %v, want %v", got, want)
	}
}

func TestNormaliseRecipients_DropsActor(t *testing.T) {
	got := normaliseRecipients(
		[]string{"alice@x", "bob@x", "Alice@X"},
		"  Alice@x  ",
	)
	if !reflect.DeepEqual(got, []string{"bob@x"}) {
		t.Fatalf("normalise = %v, want only bob", got)
	}
}

func TestInAppPushEnabled_HonourPrefs(t *testing.T) {
	off := false
	prefs := models.NotificationPrefs{
		InAppNewVuln:    &off,
		PushNewComment:  &off,
		InAppNewComment: nil,
		PushNewVuln:     nil,
	}.Effective()

	if inAppEnabled(prefs, models.NotificationKindNewVuln) {
		t.Fatalf("in-app new-vuln should be off")
	}
	if !inAppEnabled(prefs, models.NotificationKindNewComment) {
		t.Fatalf("in-app new-comment defaults to on")
	}
	if !pushEnabled(prefs, models.NotificationKindNewVuln) {
		t.Fatalf("push new-vuln defaults to on")
	}
	if pushEnabled(prefs, models.NotificationKindNewComment) {
		t.Fatalf("push new-comment should be off")
	}
}

func TestHashEndpoint_DeterministicAndShort(t *testing.T) {
	a := hashEndpoint("https://push.example/abc")
	b := hashEndpoint("https://push.example/abc")
	c := hashEndpoint("https://push.example/xyz")
	if a != b {
		t.Fatalf("hash should be deterministic")
	}
	if a == c {
		t.Fatalf("different endpoints should hash differently")
	}
	if len(a) != 16 {
		t.Fatalf("expected 16-char hex prefix, got %d", len(a))
	}
}

// --- Dispatch integration test (uses in-memory db + mocked push sender) ---

// recordedPush is what the test push sender captures so the test can assert
// "this email's endpoint got a payload" without standing up VAPID + a fake
// provider.
type recordedPush struct {
	Endpoint string
	Payload  string
}

func setupDispatchTestDB(t *testing.T) func() {
	t.Helper()
	dsn := fmt.Sprintf("file:dispatch_test_%d?mode=memory&cache=shared",
		atomic.AddUint64(&dispatchTestDBCounter, 1))
	testDB, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	// Same MaxOpenConns(1) trick as the database package tests — SQLite
	// in-memory shared-cache is single-writer and GORM otherwise opens
	// multiple connections, causing spurious SQLITE_BUSY.
	if sqlDB, err := testDB.DB(); err != nil {
		t.Fatalf("sql.DB: %v", err)
	} else {
		sqlDB.SetMaxOpenConns(1)
	}
	if err := testDB.Exec("PRAGMA busy_timeout = 5000;").Error; err != nil {
		t.Fatalf("busy_timeout: %v", err)
	}
	if err := testDB.AutoMigrate(
		&database.Subscriber{},
		&database.Notification{},
		&database.UserData{},
	); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	original := database.SetDBForTest(testDB)
	if _, err := database.MigrateNotifications(); err != nil {
		t.Fatalf("migrate notif: %v", err)
	}
	return func() { database.SetDBForTest(original) }
}

func TestDispatch_SkipsActorAndDeliversToOthers(t *testing.T) {
	defer setupDispatchTestDB(t)()

	if err := database.UpsertSubscriber("alice@x", webpush.Subscription{
		Endpoint: "https://push/alice",
		Keys:     webpush.Keys{P256dh: "p", Auth: "a"},
	}, ""); err != nil {
		t.Fatalf("alice subscribe: %v", err)
	}
	if err := database.UpsertSubscriber("bob@x", webpush.Subscription{
		Endpoint: "https://push/bob",
		Keys:     webpush.Keys{P256dh: "p", Auth: "a"},
	}, ""); err != nil {
		t.Fatalf("bob subscribe: %v", err)
	}

	var recorded []recordedPush
	var mu sync.Mutex
	original := pushSender
	pushSender = func(s database.Subscriber, payload []byte) error {
		mu.Lock()
		defer mu.Unlock()
		recorded = append(recorded, recordedPush{
			Endpoint: s.Endpoint,
			Payload:  string(payload),
		})
		return nil
	}
	defer func() { pushSender = original }()

	if err := Dispatch(DispatchRequest{
		Kind:       models.NotificationKindNewVuln,
		ActorEmail: "alice@x",
		Recipients: []string{"alice@x", "bob@x"},
		Title:      "Test",
		Body:       "hello",
		URL:        "/vulnerability/1/view",
		// VulnerabilityID intentionally 0 to skip the ACL hop — the ACL
		// path is exercised separately in TestDispatch_FiltersByACL.
	}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	// Bob received a push, Alice did not (she's the actor).
	if len(recorded) != 1 || recorded[0].Endpoint != "https://push/bob" {
		t.Fatalf("expected one push to bob's endpoint, got %+v", recorded)
	}
	// Bob has an in-app row, Alice does not.
	bobRows, _ := database.GetNotifications("bob@x", 10)
	aliceRows, _ := database.GetNotifications("alice@x", 10)
	if len(bobRows) != 1 {
		t.Fatalf("bob should have 1 in-app row, got %d", len(bobRows))
	}
	if bobRows[0].Who != "alice@x" || bobRows[0].What != "hello" {
		t.Fatalf("in-app row content wrong: %+v", bobRows[0])
	}
	if len(aliceRows) != 0 {
		t.Fatalf("actor should not receive their own notification, got %d", len(aliceRows))
	}
}

func TestDispatch_HonoursInAppOptOut(t *testing.T) {
	defer setupDispatchTestDB(t)()

	// Bob exists with an opted-out in-app pref for new comments.
	active := true
	if err := database.DBForTest().Create(&database.UserData{
		Email:  "bob@x",
		Role:   "visitor",
		Active: &active,
	}).Error; err != nil {
		t.Fatalf("seed bob: %v", err)
	}
	off := false
	if err := database.PatchSettingsForUser("bob@x", models.UserSettings{
		NotificationPrefs: models.NotificationPrefs{InAppNewComment: &off},
	}); err != nil {
		t.Fatalf("patch prefs: %v", err)
	}

	original := pushSender
	pushSender = func(database.Subscriber, []byte) error { return nil }
	defer func() { pushSender = original }()

	if err := Dispatch(DispatchRequest{
		Kind:       models.NotificationKindNewComment,
		ActorEmail: "alice@x",
		Recipients: []string{"bob@x"},
		Title:      "PRISM",
		Body:       "💬 hi",
		URL:        "/vulnerability/1/view#cid",
	}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	rows, _ := database.GetNotifications("bob@x", 10)
	if len(rows) != 0 {
		t.Fatalf("bob opted out of in-app for comments; expected 0 rows, got %d", len(rows))
	}
}

func TestDispatch_HonoursPushOptOut(t *testing.T) {
	defer setupDispatchTestDB(t)()

	active := true
	if err := database.DBForTest().Create(&database.UserData{
		Email:  "bob@x",
		Role:   "visitor",
		Active: &active,
	}).Error; err != nil {
		t.Fatalf("seed bob: %v", err)
	}
	off := false
	if err := database.PatchSettingsForUser("bob@x", models.UserSettings{
		NotificationPrefs: models.NotificationPrefs{PushNewVuln: &off},
	}); err != nil {
		t.Fatalf("patch prefs: %v", err)
	}
	if err := database.UpsertSubscriber("bob@x", webpush.Subscription{
		Endpoint: "https://push/bob",
		Keys:     webpush.Keys{P256dh: "p", Auth: "a"},
	}, ""); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	pushed := 0
	original := pushSender
	pushSender = func(database.Subscriber, []byte) error { pushed++; return nil }
	defer func() { pushSender = original }()

	if err := Dispatch(DispatchRequest{
		Kind:       models.NotificationKindNewVuln,
		ActorEmail: "alice@x",
		Recipients: []string{"bob@x"},
		Title:      "Project",
		Body:       "New vuln Foo",
		URL:        "/vulnerability/9/view",
	}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	if pushed != 0 {
		t.Fatalf("bob opted out of push; expected 0 sends, got %d", pushed)
	}
	// In-app still landed (in-app default is on).
	rows, _ := database.GetNotifications("bob@x", 10)
	if len(rows) != 1 {
		t.Fatalf("expected in-app to still fire when only push is off, got %d", len(rows))
	}
}

func TestDispatch_PushesToAllEndpointsForOneUser(t *testing.T) {
	defer setupDispatchTestDB(t)()

	for _, ep := range []string{"https://push/bob-laptop", "https://push/bob-phone"} {
		if err := database.UpsertSubscriber("bob@x", webpush.Subscription{
			Endpoint: ep,
			Keys:     webpush.Keys{P256dh: "p", Auth: "a"},
		}, ""); err != nil {
			t.Fatalf("subscribe %s: %v", ep, err)
		}
	}

	var recorded []string
	var mu sync.Mutex
	original := pushSender
	pushSender = func(s database.Subscriber, _ []byte) error {
		mu.Lock()
		recorded = append(recorded, s.Endpoint)
		mu.Unlock()
		return nil
	}
	defer func() { pushSender = original }()

	if err := Dispatch(DispatchRequest{
		Kind:       models.NotificationKindNewVuln,
		ActorEmail: "alice@x",
		Recipients: []string{"bob@x"},
		Title:      "T",
		Body:       "B",
		URL:        "/x",
	}); err != nil {
		t.Fatalf("dispatch: %v", err)
	}
	sort.Strings(recorded)
	want := []string{"https://push/bob-laptop", "https://push/bob-phone"}
	if !reflect.DeepEqual(recorded, want) {
		t.Fatalf("expected push to both devices, got %v", recorded)
	}
}

func TestDispatch_ContinuesAfterPushError(t *testing.T) {
	defer setupDispatchTestDB(t)()

	if err := database.UpsertSubscriber("bob@x", webpush.Subscription{
		Endpoint: "https://push/broken",
		Keys:     webpush.Keys{P256dh: "p", Auth: "a"},
	}, ""); err != nil {
		t.Fatalf("subscribe bob: %v", err)
	}
	if err := database.UpsertSubscriber("carol@x", webpush.Subscription{
		Endpoint: "https://push/ok",
		Keys:     webpush.Keys{P256dh: "p", Auth: "a"},
	}, ""); err != nil {
		t.Fatalf("subscribe carol: %v", err)
	}

	carolGotPushed := false
	original := pushSender
	pushSender = func(s database.Subscriber, _ []byte) error {
		if s.Endpoint == "https://push/broken" {
			return errInjected{}
		}
		if s.Endpoint == "https://push/ok" {
			carolGotPushed = true
		}
		return nil
	}
	defer func() { pushSender = original }()

	if err := Dispatch(DispatchRequest{
		Kind:       models.NotificationKindNewVuln,
		ActorEmail: "alice@x",
		Recipients: []string{"bob@x", "carol@x"},
		Title:      "T",
		Body:       "B",
		URL:        "/x",
	}); err != nil {
		t.Fatalf("dispatch should not bubble per-recipient push errors, got %v", err)
	}
	if !carolGotPushed {
		t.Fatalf("carol should have been pushed even though bob's send failed")
	}
}

type errInjected struct{}

func (errInjected) Error() string { return "injected" }
