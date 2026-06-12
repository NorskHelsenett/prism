package database

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// VulnerabilityDraft is a private, server-side work-in-progress vulnerability.
// Drafts replace the old localStorage draft: they survive browser changes and
// give attachments (videos, files) a parent to upload against before the
// vulnerability itself exists. Access is owner-only — drafts never appear in
// project or visibility queries, which is why this is a separate table rather
// than a flag on JSONData.
type VulnerabilityDraft struct {
	gorm.Model
	Owner         string `gorm:"index;not null"`
	Vulnerability datatypes.JSON
	ProjectID     *uint
}

func CreateDraft(d *VulnerabilityDraft) error {
	return db.Create(d).Error
}

// GetDraft fetches a draft only when the caller owns it. Missing and
// not-owned are indistinguishable to the caller (both ErrRecordNotFound).
func GetDraft(id uint, owner string) (*VulnerabilityDraft, error) {
	var d VulnerabilityDraft
	if err := db.Where("id = ? AND owner = ?", id, owner).First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

// ListDrafts returns the caller's drafts, newest first, without payloads —
// callers fetch the one they want via GetDraft.
func ListDrafts(owner string) ([]VulnerabilityDraft, error) {
	var list []VulnerabilityDraft
	if err := db.Select("id", "created_at", "updated_at", "owner", "project_id").
		Where("owner = ?", owner).
		Order("updated_at DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func UpdateDraft(d *VulnerabilityDraft) error {
	return db.Model(&VulnerabilityDraft{}).
		Where("id = ? AND owner = ?", d.ID, d.Owner).
		Updates(map[string]any{
			"vulnerability": d.Vulnerability,
			"project_id":    d.ProjectID,
		}).Error
}

// DeleteDraft removes a draft and its uploaded attachments. Hard delete on
// both: an abandoned draft's blobs must not linger as soft-deleted rows.
func DeleteDraft(id uint, owner string) error {
	return db.Transaction(func(tx *gorm.DB) error {
		res := tx.Unscoped().Where("id = ? AND owner = ?", id, owner).
			Delete(&VulnerabilityDraft{})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return tx.Unscoped().Where("draft_id = ?", id).
			Delete(&VulnerabilityAttachment{}).Error
	})
}

// PublishDraft turns a draft into a real vulnerability: creates the JSONData
// row, re-parents the draft's attachments to it, rewrites draft-scoped
// attachment URLs in the markdown to vulnerability-scoped ones, and consumes
// the draft — atomically. Inline data: URIs are NOT extracted here; callers
// run MigrateVulnAttachments afterwards, mirroring PostVulnerability.
func PublishDraft(draft *VulnerabilityDraft) (*JSONData, error) {
	vuln := &JSONData{
		Vulnerability: draft.Vulnerability,
		FoundBy:       draft.Owner,
		ProjectID:     draft.ProjectID,
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(vuln).Error; err != nil {
			return fmt.Errorf("create vulnerability: %w", err)
		}
		if err := tx.Model(&VulnerabilityAttachment{}).
			Where("draft_id = ?", draft.ID).
			Updates(map[string]any{"draft_id": 0, "vulnerability_id": vuln.ID}).Error; err != nil {
			return fmt.Errorf("claim attachments: %w", err)
		}
		rewritten, changed, err := rewriteDraftAttachmentURLs(vuln.Vulnerability, draft.ID, vuln.ID)
		if err != nil {
			return fmt.Errorf("rewrite attachment urls: %w", err)
		}
		if changed {
			vuln.Vulnerability = rewritten
			if err := tx.Save(vuln).Error; err != nil {
				return fmt.Errorf("save rewritten vulnerability: %w", err)
			}
		}
		res := tx.Unscoped().Where("id = ? AND owner = ?", draft.ID, draft.Owner).
			Delete(&VulnerabilityDraft{})
		if res.Error != nil {
			return fmt.Errorf("consume draft: %w", res.Error)
		}
		if res.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return vuln, nil
}

// rewriteDraftAttachmentURLs swaps /api/drafts/<draftID>/attachments/ for
// /api/vulnerability/<vulnID>/attachments/ in the markdown fields, keeping
// keys and cosmetic extensions intact.
func rewriteDraftAttachmentURLs(payload datatypes.JSON, draftID, vulnID uint) (datatypes.JSON, bool, error) {
	if len(payload) == 0 {
		return payload, false, nil
	}
	var data map[string]any
	if err := json.Unmarshal(payload, &data); err != nil {
		return payload, false, err
	}
	oldPrefix := DraftAttachmentURLPrefix(draftID)
	newPrefix := "/api/vulnerability/" + strconv.FormatUint(uint64(vulnID), 10) + "/attachments/"

	changed := false
	for _, field := range []string{"evidence", "remediation"} {
		raw, ok := data[field].(string)
		if !ok || raw == "" {
			continue
		}
		rewritten := strings.ReplaceAll(raw, oldPrefix, newPrefix)
		if rewritten != raw {
			data[field] = rewritten
			changed = true
		}
	}
	if !changed {
		return payload, false, nil
	}
	merged, err := json.Marshal(data)
	if err != nil {
		return payload, false, err
	}
	return merged, true, nil
}

// DraftAttachmentURLPrefix is the URL prefix draft attachment summaries hand
// out; PublishDraft rewrites it, so both sides must agree on the shape.
func DraftAttachmentURLPrefix(draftID uint) string {
	return "/api/drafts/" + strconv.FormatUint(uint64(draftID), 10) + "/attachments/"
}

// ListDraftAttachments returns a draft's attachments, oldest first.
func ListDraftAttachments(draftID uint) ([]VulnerabilityAttachment, error) {
	var list []VulnerabilityAttachment
	if err := db.Where("draft_id = ?", draftID).
		Order("created_at ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// GetDraftAttachment fetches an attachment scoped to a draft.
func GetDraftAttachment(draftID uint, key string) (*VulnerabilityAttachment, error) {
	var a VulnerabilityAttachment
	if err := db.Where("draft_id = ? AND key = ?", draftID, key).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func DeleteDraftAttachment(draftID uint, key string) error {
	res := db.Unscoped().Where("draft_id = ? AND key = ?", draftID, key).
		Delete(&VulnerabilityAttachment{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
