package routes

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"prism/database"
)

// loadVulnIfReadable resolves :findingsID, fetches the vulnerability, and
// confirms the caller can view it. Both "doesn't exist" and "no access" map
// to 404 so the endpoint does not leak existence.
func loadVulnIfReadable(c *gin.Context) (database.JSONData, uint, bool) {
	idStr := c.Param("findingsID")
	id64, err := strconv.ParseUint(idStr, 10, strconv.IntSize)
	if err != nil || id64 == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return database.JSONData{}, 0, false
	}
	id := uint(id64)

	vuln, err := database.GetJSONData(id)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("attachment: GetJSONData(%d): %v", id, err)
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return database.JSONData{}, 0, false
	}

	emailVal, _ := c.Get("email")
	email, _ := emailVal.(string)
	isGlobalVal, _ := c.Get("isGlobalVulnerability")
	isGlobal, _ := isGlobalVal.(bool)

	if !database.CanViewVulnerability(vuln, email, isGlobal) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return database.JSONData{}, 0, false
	}
	return vuln, id, true
}

// canModifyAttachments mirrors the upload spec: write-class role (admin or
// global vulnerability access) OR the vulnerability's FoundBy.
func canModifyAttachments(c *gin.Context, vuln database.JSONData) bool {
	if v, _ := c.Get("isAdmin"); v != nil {
		if b, _ := v.(bool); b {
			return true
		}
	}
	if v, _ := c.Get("isGlobalVulnerability"); v != nil {
		if b, _ := v.(bool); b {
			return true
		}
	}
	emailVal, _ := c.Get("email")
	email, _ := emailVal.(string)
	if email != "" && strings.EqualFold(strings.TrimSpace(vuln.FoundBy), strings.TrimSpace(email)) {
		return true
	}
	return false
}

// readAttachmentUpload pulls the multipart "file" field, resolves + verifies
// the MIME from the bytes, and builds the attachment row (parent IDs are the
// caller's job). Writes the error response itself when it returns !ok.
func readAttachmentUpload(c *gin.Context) (*database.VulnerabilityAttachment, bool) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, database.MaxAttachmentBytes)

	f, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing file"})
		return nil, false
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "could not read file"})
		return nil, false
	}

	// Sniffed from the bytes and magic-verified; anything unrecognised is
	// stored as application/octet-stream and only served as a forced download.
	mime := database.ResolveAttachmentMime(data)

	var (
		proxy     []byte
		proxyMime string
	)
	if database.AttachmentKind(mime) == "image" {
		proxy, proxyMime, err = database.EncodeAttachmentProxy(data, database.EffectiveAttachmentMaxEdge())
		if err != nil {
			log.Printf("attachment: encode proxy: %v", err)
			// Same body as the sniff/magic failure path: do not let the
			// uploader distinguish "rejected at sniff" from "rejected at
			// decode" — the distinction discloses internal pipeline shape.
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "unsupported file type"})
			return nil, false
		}
	}

	emailVal, _ := c.Get("email")
	email, _ := emailVal.(string)

	return &database.VulnerabilityAttachment{
		Key:          uuid.New().String(),
		Filename:     sanitizeAttachmentFilename(header.Filename),
		Mime:         mime,
		OriginalData: data,
		ProxyData:    proxy,
		ProxyMime:    proxyMime,
		UploadedBy:   email,
	}, true
}

// PostAttachment uploads an image attachment for a vulnerability.
//
//	POST /api/vulnerability/:findingsID/attachments  (multipart/form-data, field "file")
func PostAttachment(c *gin.Context) {
	vuln, vulnID, ok := loadVulnIfReadable(c)
	if !ok {
		return
	}
	if !canModifyAttachments(c, vuln) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	att, ok := readAttachmentUpload(c)
	if !ok {
		return
	}
	att.VulnerabilityID = vulnID
	if err := database.CreateAttachment(att); err != nil {
		log.Printf("attachment: persist vuln=%d: %v", vulnID, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "could not save attachment"})
		return
	}
	c.JSON(http.StatusOK, attachmentSummary(vulnID, att))
}

// ListAttachmentsHandler returns the attachment metadata for a vulnerability.
//
//	GET /api/vulnerability/:findingsID/attachments
func ListAttachmentsHandler(c *gin.Context) {
	_, vulnID, ok := loadVulnIfReadable(c)
	if !ok {
		return
	}
	list, err := database.ListAttachments(vulnID)
	if err != nil {
		log.Printf("attachment: list vuln=%d: %v", vulnID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	out := make([]gin.H, 0, len(list))
	for _, a := range list {
		a := a
		out = append(out, attachmentSummary(vulnID, &a))
	}
	c.JSON(http.StatusOK, out)
}

// GetAttachmentProxy serves the downscaled proxy.
//
//	GET /api/vulnerability/:findingsID/attachments/:key
//
// Cache: private, no-cache, must-revalidate + ETag — every request reaches the
// server (so revocation is immediate); the server returns 304 when the access
// check still passes and the bytes have not changed.
func GetAttachmentProxy(c *gin.Context) {
	_, vulnID, ok := loadVulnIfReadable(c)
	if !ok {
		return
	}
	att, err := database.GetAttachment(vulnID, attachmentKeyParam(c))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	serveAttachment(c, att)
}

// serveAttachment streams an attachment by kind. Images: the downscaled WebP
// proxy. Videos: the original with Range support so <video> can seek.
// Documents (pdf, txt, json): the original inline — magic-byte verified on
// upload, MIME on the whitelist, nosniff prevents the browser from upgrading
// it into a script context. Generic files: forced download, never rendered.
func serveAttachment(c *gin.Context, att *database.VulnerabilityAttachment) {
	if att.IsImage() {
		writeBytesWithETag(c, att.ProxyMime, att.ProxyData, "")
		return
	}
	switch database.AttachmentKind(att.Mime) {
	case "video":
		serveBytesWithRanges(c, att.Mime, att.OriginalData, filenameForDownload(att))
	case "file":
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Cache-Control", "private, no-cache, must-revalidate")
		c.Header("Content-Disposition", `attachment; filename="`+filenameForDownload(att)+`"`)
		c.Data(http.StatusOK, "application/octet-stream", att.OriginalData)
	default:
		writeBytesWithETag(c, att.Mime, att.OriginalData, filenameForDownload(att))
	}
}

// GetAttachmentOriginal serves the untouched original. Only the writer/finder
// of the vulnerability may access. Forced download to neutralise any legacy
// MIME funkiness.
//
//	GET /api/vulnerability/:findingsID/attachments/:key/original
func GetAttachmentOriginal(c *gin.Context) {
	vuln, vulnID, ok := loadVulnIfReadable(c)
	if !ok {
		return
	}
	if !canModifyAttachments(c, vuln) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	att, err := database.GetAttachment(vulnID, attachmentKeyParam(c))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	mime := att.Mime
	if !database.AllowedAttachmentMime(mime) {
		mime = "application/octet-stream"
	}
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Disposition", `attachment; filename="`+filenameForDownload(att)+`"`)
	c.Header("Cache-Control", "private, no-cache, must-revalidate")
	c.Data(http.StatusOK, mime, att.OriginalData)
}

// DeleteAttachmentHandler removes an attachment.
//
//	DELETE /api/vulnerability/:findingsID/attachments/:key
func DeleteAttachmentHandler(c *gin.Context) {
	vuln, vulnID, ok := loadVulnIfReadable(c)
	if !ok {
		return
	}
	if !canModifyAttachments(c, vuln) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err := database.DeleteAttachment(vulnID, attachmentKeyParam(c)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

// attachmentKeyParam returns :key with any cosmetic extension stripped
// (URLs carry e.g. ".mp4" so the markdown renderer can pick <video> without
// a metadata round-trip; stored keys are UUIDs and never contain dots).
func attachmentKeyParam(c *gin.Context) string {
	key := c.Param("key")
	if i := strings.IndexByte(key, '.'); i > 0 {
		key = key[:i]
	}
	return key
}

func attachmentSummary(vulnID uint, a *database.VulnerabilityAttachment) gin.H {
	displayMime := a.ProxyMime
	if displayMime == "" {
		displayMime = a.Mime
	}
	return gin.H{
		"key":       a.Key,
		"url":       "/api/vulnerability/" + strconv.FormatUint(uint64(vulnID), 10) + "/attachments/" + a.Key + database.AttachmentURLSuffix(a.Mime),
		"filename":  a.Filename,
		"mime":      displayMime,
		"kind":      database.AttachmentKind(a.Mime),
		"createdAt": a.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// serveBytesWithRanges streams with HTTP Range support (video seeking) plus
// the same nosniff/private-revalidation posture as writeBytesWithETag.
// http.ServeContent handles Range, If-Range and If-None-Match (via the ETag
// header set below).
func serveBytesWithRanges(c *gin.Context, mime string, data []byte, downloadName string) {
	sum := sha256.Sum256(data)
	etag := `"` + hex.EncodeToString(sum[:]) + `"`

	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "private, no-cache, must-revalidate")
	c.Header("ETag", etag)
	if downloadName != "" {
		c.Header("Content-Disposition", `inline; filename="`+downloadName+`"`)
	}
	if mime == "" {
		mime = "application/octet-stream"
	}
	c.Header("Content-Type", mime)
	http.ServeContent(c.Writer, c.Request, "", time.Time{}, bytes.NewReader(data))
}

// writeBytesWithETag streams the response with ETag + private revalidation
// caching. Pass a non-empty downloadName to add Content-Disposition: inline
// with the given filename (so PDFs/text files open with a sensible name).
func writeBytesWithETag(c *gin.Context, mime string, data []byte, downloadName string) {
	sum := sha256.Sum256(data)
	etag := `"` + hex.EncodeToString(sum[:]) + `"`

	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Cache-Control", "private, no-cache, must-revalidate")
	c.Header("ETag", etag)
	if downloadName != "" {
		c.Header("Content-Disposition", `inline; filename="`+downloadName+`"`)
	}

	if match := c.GetHeader("If-None-Match"); match != "" {
		for _, part := range strings.Split(match, ",") {
			candidate := strings.TrimSpace(part)
			if candidate == etag || candidate == "*" {
				c.Status(http.StatusNotModified)
				return
			}
		}
	}
	if mime == "" {
		mime = "application/octet-stream"
	}
	c.Data(http.StatusOK, mime, data)
}

func sanitizeAttachmentFilename(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "attachment"
	}
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "\"", "_")
	name = strings.ReplaceAll(name, "\x00", "_")
	if len(name) > 200 {
		name = name[:200]
	}
	return name
}

func filenameForDownload(a *database.VulnerabilityAttachment) string {
	if a.Filename != "" {
		return sanitizeAttachmentFilename(a.Filename)
	}
	ext := ".bin"
	switch a.Mime {
	case "image/png":
		ext = ".png"
	case "image/jpeg":
		ext = ".jpg"
	case "image/gif":
		ext = ".gif"
	case "image/webp":
		ext = ".webp"
	case "video/mp4":
		ext = ".mp4"
	case "video/webm":
		ext = ".webm"
	case "application/zip":
		ext = ".zip"
	case "application/x-gzip":
		ext = ".gz"
	}
	return a.Key + ext
}
