package routes

import (
	"io"
	"log"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"prism/config"
	"prism/database"
)

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

	// Serve the proxy to non-creators, original to creators.
	data := img.ProxyData
	mime := img.ProxyMime
	if len(data) == 0 {
		// No proxy yet (pre-backfill row). Fall back to the original but only
		// if the requester is the creator — otherwise refuse.
		if !isBlobCreator(c, img) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		data = img.Data
		mime = img.Mime
	}
	if mime == "" {
		mime = "application/octet-stream"
	}
	c.Data(http.StatusOK, mime, data)
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
	if mime == "" {
		mime = "application/octet-stream"
	}
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
		// "Global viewer" = role has a broad /vulnerability/:id permission
		// entry, mirroring the logic in auth.ACLMiddleware.
		globalViewer := false
		if r, ok := config.AppConfig.Roles[roleStr]; ok {
			for _, p := range r.Permissions {
				if p.Resource == "/vulnerability/:id" {
					globalViewer = true
					break
				}
			}
		}
		ok, err := database.CanAccessVulnerability(img.OwnerID, emailStr, roleStr, globalViewer)
		if err != nil {
			log.Printf("canReadBlob: CanAccessVulnerability: %v", err)
			return false
		}
		return ok
	default:
		// Orphan blobs (not yet bound to a vulnerability by the POST/PUT
		// handler) — only the uploader or admin can see.
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

func HandleBlobUpload(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	email, _ := c.Get("email")
	uploader, _ := email.(string)

	ownerType := c.PostForm("context")
	ownerIDStr := c.PostForm("id")
	var ownerID uint
	if ownerIDStr != "" {
		if v, err := strconv.ParseUint(ownerIDStr, 10, 64); err == nil {
			ownerID = uint(v)
		}
	}
	// Only "vulnerability" is a supported context today; anything else becomes
	// an orphan (uploader-only) until the vulnerability create/update handler
	// claims it via BindOrphanBlobs.
	if ownerType != "vulnerability" {
		ownerType = ""
		ownerID = 0
	}

	files := form.File["image"]
	var filenames []string
	for _, file := range files {
		filename := generateUniqueFilename(file.Filename)

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

		mime := database.DetectMime(data)
		proxyBytes, proxyMime, perr := database.GenerateProxy(data, mime)
		if perr != nil {
			// Non-image or unsupported format — skip proxy, but still store
			// so existing flows (e.g. non-image attachments) don't break.
			proxyBytes = nil
			proxyMime = ""
		}

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
	if err := database.DeleteImage(filename); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

func generateUniqueFilename(originalName string) string {
	newUUID := uuid.New()
	extension := filepath.Ext(originalName)
	return newUUID.String() + extension
}
