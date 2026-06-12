package routes

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	"prism/database"
)

// Server-side vulnerability drafts. A draft is private to its owner: every
// handler scopes by the session email, and both "doesn't exist" and "not
// yours" map to 404 so the endpoints do not leak existence. Drafts replace
// the old localStorage draft and give attachments (videos, files) a parent
// to upload against before the vulnerability itself exists.

type draftPayload struct {
	Vulnerability datatypes.JSON `json:"vulnerability"`
	ProjectID     *uint          `json:"projectID"`
}

func emailFromContext(c *gin.Context) string {
	emailVal, _ := c.Get("email")
	email, _ := emailVal.(string)
	return email
}

// loadOwnDraft resolves :draftID and fetches the draft if the caller owns it.
func loadOwnDraft(c *gin.Context) (*database.VulnerabilityDraft, bool) {
	email := emailFromContext(c)
	if email == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return nil, false
	}
	id64, err := strconv.ParseUint(c.Param("draftID"), 10, strconv.IntSize)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return nil, false
	}
	draft, err := database.GetDraft(uint(id64), email)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("draft: load %d: %v", id64, err)
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return nil, false
	}
	return draft, true
}

func draftSummary(d *database.VulnerabilityDraft) gin.H {
	return gin.H{
		"id":        d.ID,
		"projectID": d.ProjectID,
		"createdAt": d.CreatedAt.UTC().Format(time.RFC3339),
		"updatedAt": d.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

// CreateDraftHandler creates a draft, optionally seeded with a payload.
//
//	POST /api/drafts
func CreateDraftHandler(c *gin.Context) {
	email := emailFromContext(c)
	if email == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	var body draftPayload
	// Body is optional; an empty draft is a valid starting point.
	_ = c.ShouldBindJSON(&body)

	draft := &database.VulnerabilityDraft{
		Owner:         email,
		Vulnerability: body.Vulnerability,
		ProjectID:     body.ProjectID,
	}
	if err := database.CreateDraft(draft); err != nil {
		log.Printf("draft: create for %s: %v", email, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "could not create draft"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": draft.ID})
}

// ListDraftsHandler lists the caller's drafts (newest first, no payloads).
//
//	GET /api/drafts
func ListDraftsHandler(c *gin.Context) {
	email := emailFromContext(c)
	if email == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	list, err := database.ListDrafts(email)
	if err != nil {
		log.Printf("draft: list for %s: %v", email, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	out := make([]gin.H, 0, len(list))
	for i := range list {
		out = append(out, draftSummary(&list[i]))
	}
	c.JSON(http.StatusOK, out)
}

// GetDraftHandler returns a single draft with its payload.
//
//	GET /api/drafts/:draftID
func GetDraftHandler(c *gin.Context) {
	draft, ok := loadOwnDraft(c)
	if !ok {
		return
	}
	out := draftSummary(draft)
	out["vulnerability"] = draft.Vulnerability
	c.JSON(http.StatusOK, out)
}

// UpdateDraftHandler replaces the draft payload (autosave).
//
//	PUT /api/drafts/:draftID
func UpdateDraftHandler(c *gin.Context) {
	draft, ok := loadOwnDraft(c)
	if !ok {
		return
	}
	var body draftPayload
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid draft payload"})
		return
	}
	draft.Vulnerability = body.Vulnerability
	draft.ProjectID = body.ProjectID
	if err := database.UpdateDraft(draft); err != nil {
		log.Printf("draft: update %d: %v", draft.ID, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "could not save draft"})
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

// DeleteDraftHandler discards a draft and its uploaded attachments.
//
//	DELETE /api/drafts/:draftID
func DeleteDraftHandler(c *gin.Context) {
	draft, ok := loadOwnDraft(c)
	if !ok {
		return
	}
	if err := database.DeleteDraft(draft.ID, draft.Owner); err != nil {
		log.Printf("draft: delete %d: %v", draft.ID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

// PublishDraftHandler turns a draft into a real vulnerability: validates the
// payload, creates the JSONData row, re-parents the draft's attachments,
// rewrites draft attachment URLs in the markdown, consumes the draft, and
// finally extracts any inline data: URIs — mirroring PostVulnerability.
//
//	POST /api/drafts/:draftID/publish
func PublishDraftHandler(c *gin.Context) {
	draft, ok := loadOwnDraft(c)
	if !ok {
		return
	}

	var vulnData VulnerabilityData
	if err := json.Unmarshal(draft.Vulnerability, &vulnData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid vulnerability data format"})
		return
	}
	if err := validateVulnerabilityData(vulnData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vuln, err := database.PublishDraft(draft)
	if err != nil {
		log.Printf("draft: publish %d: %v", draft.ID, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "could not publish draft"})
		return
	}

	if changed, err := database.MigrateVulnAttachments(vuln); err != nil {
		log.Printf("attachment normalise on publish vuln=%d: %v", vuln.ID, err)
	} else if changed {
		if err := database.UpdateVulnerability(vuln); err != nil {
			log.Printf("attachment normalise save vuln=%d: %v", vuln.ID, err)
		}
	}

	c.JSON(http.StatusCreated, gin.H{"id": vuln.ID})
}

// PostDraftAttachment uploads an attachment parented to a draft.
//
//	POST /api/drafts/:draftID/attachments  (multipart/form-data, field "file")
func PostDraftAttachment(c *gin.Context) {
	draft, ok := loadOwnDraft(c)
	if !ok {
		return
	}
	att, ok := readAttachmentUpload(c)
	if !ok {
		return
	}
	att.DraftID = draft.ID
	if err := database.CreateAttachment(att); err != nil {
		log.Printf("attachment: persist draft=%d: %v", draft.ID, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "could not save attachment"})
		return
	}
	c.JSON(http.StatusOK, draftAttachmentSummary(draft.ID, att))
}

// ListDraftAttachmentsHandler returns the attachment metadata for a draft.
//
//	GET /api/drafts/:draftID/attachments
func ListDraftAttachmentsHandler(c *gin.Context) {
	draft, ok := loadOwnDraft(c)
	if !ok {
		return
	}
	list, err := database.ListDraftAttachments(draft.ID)
	if err != nil {
		log.Printf("attachment: list draft=%d: %v", draft.ID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	out := make([]gin.H, 0, len(list))
	for i := range list {
		out = append(out, draftAttachmentSummary(draft.ID, &list[i]))
	}
	c.JSON(http.StatusOK, out)
}

// GetDraftAttachmentProxy serves a draft attachment with the same kind
// handling as the vulnerability endpoint.
//
//	GET /api/drafts/:draftID/attachments/:key
func GetDraftAttachmentProxy(c *gin.Context) {
	draft, ok := loadOwnDraft(c)
	if !ok {
		return
	}
	att, err := database.GetDraftAttachment(draft.ID, attachmentKeyParam(c))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	serveAttachment(c, att)
}

// DeleteDraftAttachmentHandler removes a draft attachment.
//
//	DELETE /api/drafts/:draftID/attachments/:key
func DeleteDraftAttachmentHandler(c *gin.Context) {
	draft, ok := loadOwnDraft(c)
	if !ok {
		return
	}
	if err := database.DeleteDraftAttachment(draft.ID, attachmentKeyParam(c)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func draftAttachmentSummary(draftID uint, a *database.VulnerabilityAttachment) gin.H {
	displayMime := a.ProxyMime
	if displayMime == "" {
		displayMime = a.Mime
	}
	return gin.H{
		"key":       a.Key,
		"url":       database.DraftAttachmentURLPrefix(draftID) + a.Key + database.AttachmentURLSuffix(a.Mime),
		"filename":  a.Filename,
		"mime":      displayMime,
		"kind":      database.AttachmentKind(a.Mime),
		"createdAt": a.CreatedAt.UTC().Format(time.RFC3339),
	}
}
