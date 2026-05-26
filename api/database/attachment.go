package database

import (
	"bytes"
	"errors"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"net/http"
	"time"

	webpencoder "github.com/HugoSmits86/nativewebp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
	"gorm.io/gorm"
)

const (
	maxAttachmentLongEdge = 1080
	MaxAttachmentBytes    = 25 << 20
)

var allowedAttachmentMimes = map[string]struct{}{
	"image/png":  {},
	"image/jpeg": {},
	"image/gif":  {},
	"image/webp": {},
}

// VulnerabilityAttachment is an image attached to exactly one vulnerability.
// Access is derived from access to the parent vulnerability; there is no
// per-attachment ACL beyond that. The bytes never round-trip through a global
// blob namespace — every URL encodes the owning vuln id.
type VulnerabilityAttachment struct {
	ID        uint `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	VulnerabilityID uint   `gorm:"index;not null"`
	Key             string `gorm:"uniqueIndex;size:64;not null"`

	Filename string
	Mime     string

	OriginalData []byte
	ProxyData    []byte
	ProxyMime    string

	UploadedBy string `gorm:"index"`
}

func AllowedAttachmentMime(mime string) bool {
	_, ok := allowedAttachmentMimes[mime]
	return ok
}

func SniffAttachmentMime(data []byte) string {
	return http.DetectContentType(data)
}

// EncodeAttachmentProxy downscales the long edge to 1080px and re-encodes as
// WebP. Falls back to JPEG if the WebP encoder rejects the image. Never
// upscales.
func EncodeAttachmentProxy(original []byte) ([]byte, string, error) {
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
	if longEdge > maxAttachmentLongEdge {
		scale := float64(maxAttachmentLongEdge) / float64(longEdge)
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
