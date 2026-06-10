package database

import (
	"bytes"
	"encoding/json"
	"errors"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"strings"
	"time"

	webpencoder "github.com/HugoSmits86/nativewebp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
	"gorm.io/gorm"
)

// MaxAttachmentBytes caps the upload size at 512 MiB (sized for screen
// recordings). Proxy long-edge is now configurable via the admin setting
// Settings.AttachmentMaxEdge, with the fallback in DefaultAttachmentMaxEdge.
const MaxAttachmentBytes = 512 << 20

// attachmentKind controls how the bytes are served. Images get a downscaled
// proxy for inline rendering; videos stream the original with Range support;
// documents are served inline with their real MIME; generic files are only
// ever served as a forced download. Everything non-image carries nosniff so
// the browser cannot upgrade content-type into a script context.
type attachmentKind int

const (
	kindImage attachmentKind = iota
	kindVideo
	kindDocument
	kindFile
)

type mimeRule struct {
	kind   attachmentKind
	verify func([]byte) bool
}

var attachmentMimeRules = map[string]mimeRule{
	"image/png":          {kind: kindImage, verify: bytesPrefix([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A})},
	"image/jpeg":         {kind: kindImage, verify: bytesPrefix([]byte{0xFF, 0xD8, 0xFF})},
	"image/gif":          {kind: kindImage, verify: verifyGIF},
	"image/webp":         {kind: kindImage, verify: verifyWebP},
	"video/mp4":          {kind: kindVideo, verify: verifyMP4},
	"video/webm":         {kind: kindVideo, verify: verifyEBML},
	"application/pdf":    {kind: kindDocument, verify: bytesPrefix([]byte("%PDF-"))},
	"text/plain":         {kind: kindDocument, verify: verifyTextLike},
	"application/json":   {kind: kindDocument, verify: verifyJSON},
	"application/zip":    {kind: kindFile, verify: bytesPrefix([]byte{'P', 'K'})},
	"application/x-gzip": {kind: kindFile, verify: bytesPrefix([]byte{0x1F, 0x8B})},
	// Catch-all for everything the sniffer cannot positively identify. Safe
	// because kindFile is never served inline: forced download + nosniff means
	// there is no rendering context for the bytes to abuse.
	"application/octet-stream": {kind: kindFile},
}

// VulnerabilityAttachment is an image or document attached to exactly one
// vulnerability. Access is derived from access to the parent vulnerability;
// there is no per-attachment ACL beyond that. The bytes never round-trip
// through a global blob namespace — every URL encodes the owning vuln id.
type VulnerabilityAttachment struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	VulnerabilityID uint   `gorm:"index;not null"`
	Key             string `gorm:"uniqueIndex;size:64;not null"`

	Filename string
	Mime     string // sniffed + magic-verified; one of attachmentMimeRules

	OriginalData []byte
	// ProxyData is the downscaled image (WebP, JPEG fallback) for inline
	// rendering. Empty for non-image attachments — those are served by
	// streaming OriginalData with nosniff.
	ProxyData []byte
	ProxyMime string

	UploadedBy string `gorm:"index"`
}

// IsImage reports whether the attachment has a renderable image proxy.
func (a *VulnerabilityAttachment) IsImage() bool {
	return len(a.ProxyData) > 0
}

// AllowedAttachmentMime reports whether the sniffed MIME is in the whitelist.
// Pass the raw value from SniffAttachmentMime; parameter normalisation is
// handled internally.
func AllowedAttachmentMime(mime string) bool {
	_, ok := attachmentMimeRules[normaliseMime(mime)]
	return ok
}

// VerifyAttachmentMagic confirms the file's leading bytes actually match the
// claimed MIME. Defends against renaming a `.html` to `.png` to smuggle a
// content-type past the sniffer (the F-3 path).
func VerifyAttachmentMagic(mime string, data []byte) bool {
	rule, ok := attachmentMimeRules[normaliseMime(mime)]
	if !ok {
		return false
	}
	if rule.verify == nil {
		return true
	}
	return rule.verify(data)
}

// AttachmentKind classifies the MIME. Returns "image", "video", "document" or
// "file"; empty if the MIME is not allowed.
func AttachmentKind(mime string) string {
	rule, ok := attachmentMimeRules[normaliseMime(mime)]
	if !ok {
		return ""
	}
	switch rule.kind {
	case kindImage:
		return "image"
	case kindVideo:
		return "video"
	case kindFile:
		return "file"
	default:
		return "document"
	}
}

// ResolveAttachmentMime sniffs the bytes and returns the MIME to store. A
// type only resolves to itself when the magic check confirms it; anything
// unrecognised (or markup masquerading as text) falls back to
// application/octet-stream, which is stored as a generic file and only ever
// served as a forced download.
func ResolveAttachmentMime(data []byte) string {
	mime := SniffAttachmentMime(data)
	if mime != "application/octet-stream" && AllowedAttachmentMime(mime) && VerifyAttachmentMagic(mime, data) {
		return mime
	}
	// net/http only sniffs video/mp4 when an ftyp brand starts with "mp4";
	// real-world recordings (brands isom/avc1/qt) sniff as octet-stream.
	// Our own ftyp check covers those.
	if verifyMP4(data) {
		return "video/mp4"
	}
	return "application/octet-stream"
}

// AttachmentURLSuffix returns a cosmetic extension appended to scoped
// attachment URLs so the renderer can pick the right element (<video> vs
// <img>) without a metadata round-trip. The routes strip it before the key
// lookup; keys are UUIDs and never contain dots.
func AttachmentURLSuffix(mime string) string {
	switch normaliseMime(mime) {
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	default:
		return ""
	}
}

// SniffAttachmentMime returns the sniffed MIME of the bytes, ignoring any
// client-supplied content-type. Always normalised (no charset parameter).
func SniffAttachmentMime(data []byte) string {
	return normaliseMime(http.DetectContentType(data))
}

func normaliseMime(mime string) string {
	if i := strings.Index(mime, ";"); i >= 0 {
		mime = mime[:i]
	}
	return strings.ToLower(strings.TrimSpace(mime))
}

func bytesPrefix(prefix []byte) func([]byte) bool {
	return func(b []byte) bool { return bytes.HasPrefix(b, prefix) }
}

func verifyGIF(b []byte) bool {
	return bytes.HasPrefix(b, []byte("GIF87a")) || bytes.HasPrefix(b, []byte("GIF89a"))
}

func verifyWebP(b []byte) bool {
	return len(b) >= 12 && bytes.HasPrefix(b, []byte("RIFF")) && bytes.Equal(b[8:12], []byte("WEBP"))
}

// verifyMP4 checks for the ISO BMFF "ftyp" box that every MP4 starts with:
// a 4-byte box size followed by the literal "ftyp".
func verifyMP4(b []byte) bool {
	return len(b) >= 12 && bytes.Equal(b[4:8], []byte("ftyp"))
}

// verifyEBML checks the EBML magic that prefixes WebM (and Matroska) files.
func verifyEBML(b []byte) bool {
	return bytes.HasPrefix(b, []byte{0x1A, 0x45, 0xDF, 0xA3})
}

// verifyTextLike rejects bytes containing NUL or stray control characters,
// and also bytes that look like markup. http.DetectContentType sniffs
// `<svg>…</svg>` or `<?xml …?>` as text/plain (not unambiguous HTML), so
// without this guard a caller could upload SVG/HTML markup under the
// text/plain whitelist. Storing markup as text/plain is harmless on read
// (nosniff prevents promotion), but it creates a UX surprise (an "image"
// upload shows as text) and leaves a defence-in-depth gap if anything ever
// served the bytes with a renderable MIME.
func verifyTextLike(b []byte) bool {
	sample := b
	if len(sample) > 4096 {
		sample = sample[:4096]
	}
	trimmed := bytes.TrimLeft(sample, " \t\r\n")
	if len(trimmed) > 0 && trimmed[0] == '<' {
		return false
	}
	for _, c := range sample {
		if c == 0 {
			return false
		}
		// permit common whitespace; reject other control bytes
		if c < 0x09 || (c > 0x0D && c < 0x20) {
			return false
		}
	}
	return true
}

func verifyJSON(b []byte) bool {
	var x any
	return json.Unmarshal(b, &x) == nil
}

// EncodeAttachmentProxy re-encodes the original as WebP (lossless), down-
// scaling first if the long edge exceeds maxEdge. Falls back to JPEG if the
// WebP encoder rejects the image. Never upscales. Returns empty bytes if the
// MIME isn't an image type; callers should branch on AttachmentKind before
// invoking. Pass maxEdge=0 to use DefaultAttachmentMaxEdge.
func EncodeAttachmentProxy(original []byte, maxEdge int) ([]byte, string, error) {
	if maxEdge <= 0 {
		maxEdge = DefaultAttachmentMaxEdge
	}
	img, _, err := image.Decode(bytes.NewReader(original))
	if err != nil {
		return nil, "", err
	}
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return nil, "", errors.New("invalid image dimensions")
	}
	longEdge := w
	if h > longEdge {
		longEdge = h
	}

	target := img
	if longEdge > maxEdge {
		scale := float64(maxEdge) / float64(longEdge)
		nw := int(float64(w) * scale)
		nh := int(float64(h) * scale)
		if nw < 1 {
			nw = 1
		}
		if nh < 1 {
			nh = 1
		}
		scaled := image.NewRGBA(image.Rect(0, 0, nw, nh))
		draw.CatmullRom.Scale(scaled, scaled.Bounds(), img, bounds, draw.Over, nil)
		target = scaled
	}

	var buf bytes.Buffer
	if err := webpencoder.Encode(&buf, target, nil); err == nil {
		return buf.Bytes(), "image/webp", nil
	}
	buf.Reset()
	if err := jpeg.Encode(&buf, target, &jpeg.Options{Quality: 85}); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), "image/jpeg", nil
}

func CreateAttachment(a *VulnerabilityAttachment) error {
	return db.Create(a).Error
}

func GetAttachment(vulnID uint, key string) (*VulnerabilityAttachment, error) {
	var a VulnerabilityAttachment
	if err := db.Where("vulnerability_id = ? AND key = ?", vulnID, key).First(&a).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

func ListAttachments(vulnID uint) ([]VulnerabilityAttachment, error) {
	var list []VulnerabilityAttachment
	if err := db.Where("vulnerability_id = ?", vulnID).
		Order("created_at ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func DeleteAttachment(vulnID uint, key string) error {
	res := db.Where("vulnerability_id = ? AND key = ?", vulnID, key).
		Delete(&VulnerabilityAttachment{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
