package database

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDraftTestDB(t *testing.T) func() {
	t.Helper()

	originalDB := db

	testDB, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("opening test database: %v", err)
	}
	if err := testDB.AutoMigrate(&JSONData{}, &VulnerabilityDraft{}, &VulnerabilityAttachment{}); err != nil {
		t.Fatalf("migrating schema: %v", err)
	}

	db = testDB
	return func() { db = originalDB }
}

func TestGetDraftIsOwnerScoped(t *testing.T) {
	cleanup := setupDraftTestDB(t)
	defer cleanup()

	draft := &VulnerabilityDraft{Owner: "owner@example.com"}
	if err := CreateDraft(draft); err != nil {
		t.Fatalf("create draft: %v", err)
	}

	if _, err := GetDraft(draft.ID, "owner@example.com"); err != nil {
		t.Fatalf("owner should see own draft: %v", err)
	}
	if _, err := GetDraft(draft.ID, "other@example.com"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("non-owner access should look like not-found, got %v", err)
	}
}

func TestPublishDraftClaimsAttachmentsAndRewritesURLs(t *testing.T) {
	cleanup := setupDraftTestDB(t)
	defer cleanup()

	owner := "owner@example.com"
	draft := &VulnerabilityDraft{Owner: owner}
	if err := CreateDraft(draft); err != nil {
		t.Fatalf("create draft: %v", err)
	}

	att := &VulnerabilityAttachment{
		DraftID:      draft.ID,
		Key:          "11111111-2222-3333-4444-555555555555",
		Filename:     "clip.mp4",
		Mime:         "video/mp4",
		OriginalData: []byte("not really a video"),
		UploadedBy:   owner,
	}
	if err := CreateAttachment(att); err != nil {
		t.Fatalf("create attachment: %v", err)
	}

	evidence := fmt.Sprintf("PoC: ![clip.mp4](%s%s.mp4)", DraftAttachmentURLPrefix(draft.ID), att.Key)
	payload, _ := json.Marshal(map[string]any{"evidence": evidence, "title": "demo"})
	draft.Vulnerability = payload
	if err := UpdateDraft(draft); err != nil {
		t.Fatalf("update draft: %v", err)
	}

	vuln, err := PublishDraft(draft)
	if err != nil {
		t.Fatalf("publish: %v", err)
	}
	if vuln.ID == 0 {
		t.Fatal("published vulnerability has no ID")
	}
	if vuln.FoundBy != owner {
		t.Errorf("FoundBy = %q, want %q", vuln.FoundBy, owner)
	}

	// The markdown must now reference the vulnerability-scoped URL.
	var data map[string]any
	if err := json.Unmarshal(vuln.Vulnerability, &data); err != nil {
		t.Fatalf("unmarshal published payload: %v", err)
	}
	want := fmt.Sprintf("PoC: ![clip.mp4](/api/vulnerability/%d/attachments/%s.mp4)", vuln.ID, att.Key)
	if data["evidence"] != want {
		t.Errorf("evidence = %q, want %q", data["evidence"], want)
	}

	// The attachment must be re-parented to the vulnerability.
	claimed, err := GetAttachment(vuln.ID, att.Key)
	if err != nil {
		t.Fatalf("claimed attachment not reachable via vuln scope: %v", err)
	}
	if claimed.DraftID != 0 {
		t.Errorf("claimed attachment still has DraftID %d", claimed.DraftID)
	}
	if _, err := GetDraftAttachment(draft.ID, att.Key); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("attachment still reachable via draft scope: %v", err)
	}

	// The draft must be consumed.
	if _, err := GetDraft(draft.ID, owner); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("draft still exists after publish: %v", err)
	}
}

func TestDeleteDraftRemovesAttachments(t *testing.T) {
	cleanup := setupDraftTestDB(t)
	defer cleanup()

	owner := "owner@example.com"
	draft := &VulnerabilityDraft{Owner: owner}
	if err := CreateDraft(draft); err != nil {
		t.Fatalf("create draft: %v", err)
	}
	att := &VulnerabilityAttachment{
		DraftID:      draft.ID,
		Key:          "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
		Mime:         "application/octet-stream",
		OriginalData: []byte{0x00, 0x01},
		UploadedBy:   owner,
	}
	if err := CreateAttachment(att); err != nil {
		t.Fatalf("create attachment: %v", err)
	}

	if err := DeleteDraft(draft.ID, "other@example.com"); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("non-owner delete should look like not-found, got %v", err)
	}
	if err := DeleteDraft(draft.ID, owner); err != nil {
		t.Fatalf("delete draft: %v", err)
	}
	if _, err := GetDraftAttachment(draft.ID, att.Key); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Errorf("attachment survived draft deletion: %v", err)
	}
}
