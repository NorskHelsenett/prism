package database

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"regexp"
	"strings"

	_ "image/gif" // register gif decoder
	_ "image/png" // register png decoder

	webpencoder "github.com/HugoSmits86/nativewebp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // register webp decoder
)

const (
	proxyMaxEdge     = 1080
	proxyJPEGQuality = 82
)

var ErrUnsupportedImage = errors.New("unsupported image format")

// DetectMime sniffs the image MIME from the first bytes (avoids trusting
// the multipart header). Returns "" if not a recognised image type.
func DetectMime(data []byte) string {
	if len(data) == 0 {
		return ""
	}
	m := http.DetectContentType(data)
	// http.DetectContentType handles png/jpeg/gif/webp. Return only the
	// content-type prefix, stripping any charset info.
	if i := strings.Index(m, ";"); i >= 0 {
		m = strings.TrimSpace(m[:i])
	}
	return m
}

// GenerateProxy decodes the input image, downscales the long edge to at most
// proxyMaxEdge (no upscale), strips EXIF by re-encoding, and returns WebP
// bytes. If WebP encoding fails it falls back to JPEG.
func GenerateProxy(data []byte, mime string) ([]byte, string, error) {
	img, err := decodeImage(data, mime)
	if err != nil {
		return nil, "", err
	}

	resized := resizeLongEdge(img, proxyMaxEdge)

	// Prefer WebP (nativewebp is lossless VP8L — quality isn't configurable,
	// but the resize step caps size).
	var buf bytes.Buffer
	if err := webpencoder.Encode(&buf, resized, nil); err == nil {
		return buf.Bytes(), "image/webp", nil
	}

	// Fallback: JPEG (lossy) for photos, PNG for small diagrams would be nicer
	// but JPEG is a safe single-branch fallback.
	buf.Reset()
	if err := jpeg.Encode(&buf, resized, &jpeg.Options{Quality: proxyJPEGQuality}); err != nil {
		return nil, "", err
	}
	return buf.Bytes(), "image/jpeg", nil
}

func decodeImage(data []byte, mime string) (image.Image, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, ErrUnsupportedImage
	}
	return img, nil
}

// BackfillImages scans for legacy ImageData rows that lack owner metadata
// and/or a proxy and fills them in:
//   - Owner: for each row, search vulnerabilities' JSON content for the
//     filename; if found, bind the blob to that vulnerability and the
//     vulnerability's creator (FoundBy) as uploader.
//   - Proxy: decode the original bytes and generate a WebP proxy.
//
// Designed to be called once in a background goroutine from main(). Logs but
// does not return errors per-row.
func BackfillImages() {
	rows, err := ListImagesMissingProxy()
	if err != nil {
		log.Printf("BackfillImages: list failed: %v", err)
		return
	}
	if len(rows) == 0 {
		return
	}
	log.Printf("BackfillImages: starting on %d image(s)", len(rows))
	processed := 0
	for _, row := range rows {
		fields := map[string]interface{}{}

		// Owner resolution.
		if row.OwnerType == "" {
			vuln, found := findVulnReferencing(row.Filename)
			if found {
				fields["owner_type"] = "vulnerability"
				fields["owner_id"] = vuln.ID
				if row.UploaderEmail == "" {
					fields["uploader_email"] = vuln.FoundBy
				}
			}
		}

		// Proxy generation.
		if len(row.ProxyData) == 0 && len(row.Data) > 0 {
			mime := row.Mime
			if mime == "" {
				mime = DetectMime(row.Data)
				if mime != "" {
					fields["mime"] = mime
				}
			}
			proxy, proxyMime, perr := GenerateProxy(row.Data, mime)
			if perr == nil {
				fields["proxy_data"] = proxy
				fields["proxy_mime"] = proxyMime
			}
		}

		if len(fields) == 0 {
			continue
		}
		if err := UpdateImageMeta(row.Filename, fields); err != nil {
			log.Printf("BackfillImages: update %s: %v", row.Filename, err)
			continue
		}
		processed++
	}
	log.Printf("BackfillImages: finished, %d row(s) updated", processed)
}

var blobRefRegex = regexp.MustCompile(`/api/blob/([a-zA-Z0-9\-]+(?:\.[a-zA-Z0-9]+)?)`)

// BindOrphanBlobs finds blob references in the given content string and, for
// any matching ImageData rows that are currently orphans (empty OwnerType) and
// were uploaded by the same user, binds them to the given owner. Used by the
// vulnerability create/update path so screenshots pasted during registration
// get their ACL set when the vuln is saved.
func BindOrphanBlobs(ownerType string, ownerID uint, uploaderEmail, content string) {
	if ownerType == "" || ownerID == 0 || uploaderEmail == "" || content == "" {
		return
	}
	matches := blobRefRegex.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return
	}
	filenames := make([]string, 0, len(matches))
	seen := map[string]struct{}{}
	for _, m := range matches {
		f := m[1]
		if _, ok := seen[f]; ok {
			continue
		}
		seen[f] = struct{}{}
		filenames = append(filenames, f)
	}
	err := db.Model(&ImageData{}).
		Where("filename IN ? AND uploader_email = ? AND (owner_type = '' OR owner_type IS NULL)", filenames, uploaderEmail).
		Updates(map[string]interface{}{
			"owner_type": ownerType,
			"owner_id":   ownerID,
		}).Error
	if err != nil {
		log.Printf("BindOrphanBlobs: %v", err)
	}
}

func findVulnReferencing(filename string) (*JSONData, bool) {
	var jd JSONData
	// Match the filename anywhere in the vulnerability JSON. Fast enough for
	// a one-shot backfill; SQLite's LIKE is table-scan but we only run once.
	err := db.Where("vulnerability LIKE ?", "%"+filename+"%").First(&jd).Error
	if err != nil {
		return nil, false
	}
	return &jd, true
}

func resizeLongEdge(src image.Image, maxEdge int) image.Image {
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	long := w
	if h > long {
		long = h
	}
	// Never upscale.
	if long <= maxEdge {
		return src
	}
	ratio := float64(maxEdge) / float64(long)
	nw := int(float64(w) * ratio)
	nh := int(float64(h) * ratio)
	dst := image.NewRGBA(image.Rect(0, 0, nw, nh))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)
	return dst
}
