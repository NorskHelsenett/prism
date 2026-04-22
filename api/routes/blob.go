package routes

import (
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"prism/config"
	"prism/database"
)

// allowedImageMimes is the exhaustive set of MIME types accepted by the blob
// upload endpoint. Anything else is rejected so the endpoint can't be used as
// a same-origin HTML/SVG/JS host (pentest finding F-1/F-3).
var allowedImageMimes = map[string]string{
	"image/png":  ".png",
	"image/jpeg": ".jpg",
	"image/gif":  ".gif",
	"image/webp": ".webp",
}

func GetBlob(c *gin.Context) {
	filename := c.Param("filename")

	img, err := database.GetImage(filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	if !canReadBlob(c, img) {
		// Don't leak the distinction between "doesn't exist" and "no access".
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	// Always serve the proxy. If none exists (legacy non-image row that
	// couldn't be decoded during backfill), refuse — do NOT fall back to
	// img.Data, which could be an arbitrary byte stream with an attacker-
	// controlled MIME. Pentest finding F-1 was live XSS via this path.
	if len(img.ProxyData) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	mime := img.ProxyMime
	if mime == "" {
		mime = "application/octet-stream"
	}
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, mime, img.ProxyData)
}

// GetBlobOriginal serves the untouched original. Only the creator of the
// resource the blob was uploaded for may access it.
func GetBlobOriginal(c *gin.Context) {
	filename := c.Param("filename")

	img, err := database.GetImage(filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if !isBlobCreator(c, img) {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	mime := img.Mime
	if _, ok := allowedImageMimes[mime]; !ok {
		// Defence-in-depth: a legacy row could have a MIME we no longer
		// trust to render inline. Force download in that case.
		mime = "application/octet-stream"
	}
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Disposition", "attachment; filename=\""+filename+"\"")
	c.Data(http.StatusOK, mime, img.Data)
}

// canReadBlob returns true if the user can see at least the proxy.
func canReadBlob(c *gin.Context, img *database.ImageData) bool {
	email, _ := c.Get("email")
	role, _ := c.Get("role")
	emailStr, _ := email.(string)
	roleStr, _ := role.(string)
	if roleStr == "admin" {
		return true
	}

	switch img.OwnerType {
	case "vulnerability":
		ok, err := database.CanAccessVulnerability(img.OwnerID, emailStr, roleStr, roleHasGlobalVulnerability(roleStr))
		if err != nil {
			log.Printf("canReadBlob: CanAccessVulnerability: %v", err)
			return false
		}
		return ok
	default:
		// Orphan blobs (not yet bound to a vulnerability by the POST/PUT
		// handler) — only the uploader can see.
		return emailStr != "" && img.UploaderEmail == emailStr
	}
}

// isBlobCreator returns true if the user is the creator of the resource the
// blob is attached to — vulnerability FoundBy, or uploader for orphans.
func isBlobCreator(c *gin.Context, img *database.ImageData) bool {
	email, _ := c.Get("email")
	role, _ := c.Get("role")
	emailStr, _ := email.(string)
	roleStr, _ := role.(string)
	if roleStr == "admin" {
		return true
	}
	if img.OwnerType == "vulnerability" {
		vuln, err := database.GetJSONData(img.OwnerID)
		if err != nil {
			return false
		}
		return vuln.FoundBy == emailStr
	}
	return emailStr != "" && img.UploaderEmail == emailStr
}

// roleHasGlobalVulnerability mirrors auth.ACLMiddleware's check: if a role
// declares /vulnerability/:id in its permissions, it gets the "global viewer"
// flag for vulnerability ACL evaluation.
func roleHasGlobalVulnerability(roleStr string) bool {
	r, ok := config.AppConfig.Roles[roleStr]
	if !ok {
		return false
	}
	for _, p := range r.Permissions {
		if p.Resource == "/vulnerability/:id" {
			return true
		}
	}
	return false
}

func HandleBlobUpload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email, _ := c.Get("email")
	uploader, _ := email.(string)
	role, _ := c.Get("role")
	roleStr, _ := role.(string)

	// context= binds the upload to a specific resource so the read-side ACL
	// can be derived from the parent resource. Currently only "vulnerability"
	// is supported; anything else falls through to orphan (uploader-only).
	ownerType := c.PostForm("context")
	ownerIDStr := c.PostForm("id")
	var ownerID uint
	if ownerType == "vulnerability" {
		if ownerIDStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "id is required when context=vulnerability"})
			return
		}
		v, err := strconv.ParseUint(ownerIDStr, 10, 64)
		if err != nil || v == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
			return
		}
		ownerID = uint(v)
		// Pentest finding F-2: the uploader must actually have access to the
		// vulnerability they're claiming to attach the blob to. Without this
		// check, any /blob:write role can plant bytes in any vulnerability's
		// namespace — and the vulnerability creator would download them via
		// /original believing they were pristine.
		canAccess, err := database.CanAccessVulnerability(ownerID, uploader, roleStr, roleHasGlobalVulnerability(roleStr))
		if err != nil || !canAccess {
			c.JSON(http.StatusForbidden, gin.H{"error": "not allowed to attach to this resource"})
			return
		}
	} else {
		ownerType = ""
		ownerID = 0
	}

	files := form.File["image"]
	if len(files) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no image file in request"})
		return
	}

	var filenames []string
	for _, file := range files {
		fileData, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		data, err := io.ReadAll(fileData)
		fileData.Close()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Pentest finding F-1/F-3: reject anything that isn't a known image
		// MIME (sniffed from the bytes, not the attacker-controlled multipart
		// header). The extension is derived from the sniffed MIME so the
		// stored filename can't carry a surprise extension like .html.
		mime := database.DetectMime(data)
		ext, ok := allowedImageMimes[mime]
		if !ok {
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "unsupported image type"})
			return
		}

		proxyBytes, proxyMime, perr := database.GenerateProxy(data, mime)
		if perr != nil {
			// The MIME said image but the decoder disagreed. Refuse.
			c.JSON(http.StatusUnsupportedMediaType, gin.H{"error": "image could not be decoded"})
			return
		}

		filename := uuid.New().String() + ext
		img := &database.ImageData{
			Filename:      filename,
			Data:          data,
			Mime:          mime,
			ProxyData:     proxyBytes,
			ProxyMime:     proxyMime,
			UploaderEmail: uploader,
			OwnerType:     ownerType,
			OwnerID:       ownerID,
		}
		if err := database.SaveImageFull(img); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		filenames = append(filenames, filename)
	}

	c.JSON(http.StatusOK, gin.H{"fileNames": filenames})
}

func HandleBlobDelete(c *gin.Context) {
	filename := c.Param("filename")
	// Guard against path-traversal in the filename param (Gin's :filename
	// already limits the match but belt-and-suspenders).
	if strings.ContainsAny(filename, "/\\") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}
	if err := database.DeleteImage(filename); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}
