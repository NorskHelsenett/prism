package share

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"prism/auth"
	"prism/database"
	"prism/models"
	"prism/session"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func GetShareVulnerability(c *gin.Context) {
	var sharedDocument models.SharedDocument

	IDStr := c.Param("findingsID")
	if IDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	id64, err := strconv.ParseUint(IDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}
	id := uint(id64)

	// ShareMiddleware already enforces role-level write on /vulnerability.
	// Combined with the per-vulnerability access check below, this matches
	// the project's existing definition of effective write access on the
	// resource — preventing any authenticated pentester from reading the
	// passphrase of a share for a vulnerability they cannot access.
	email, _ := c.Get("email")
	role, _ := c.Get("role")
	isGlobal, _ := c.Get("isGlobalVulnerability")

	allowed, err := database.CanAccessVulnerability(id, email.(string), role.(string), isGlobal.(bool))
	if err != nil {
		fmt.Printf("CanAccessVulnerability failed for vuln %d: %v\n", id, err)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	if !allowed {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	sharedDocument.DocumentID = id

	document, err := database.GetShareDocument(id)
	if err != nil {
		c.JSON(http.StatusOK, sharedDocument)
		return
	}

	c.JSON(http.StatusOK, document)
}

func DeleteShareVulnerability(c *gin.Context) {
	IDStr := c.Param("findingsID")
	if IDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	id64, err := strconv.ParseUint(IDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	id := uint(id64)

	email, _ := c.Get("email")
	role, _ := c.Get("role")
	isGlobal, _ := c.Get("isGlobalVulnerability")

	allowed, err := database.CanAccessVulnerability(id, email.(string), role.(string), isGlobal.(bool))
	if err != nil {
		fmt.Printf("CanAccessVulnerability failed for vuln %d: %v\n", id, err)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	if !allowed {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	err = database.DeleteShareDocument(id)
	if err != nil {
		c.AbortWithStatus(http.StatusConflict)
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func ShareVulnerability(c *gin.Context) {
	var input struct {
		InvitedEmailsJSON []string `json:"invitedEmails"`
		ExpirationDate    string   `json:"expirationDate"`
		Passphrase        string   `json:"passphrase"`
		AccessType        string   `json:"accessType" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	IDStr := c.Param("findingsID")
	if IDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID is required"})
		return
	}

	id64, err := strconv.ParseUint(IDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	id := uint(id64)

	email, _ := c.Get("email")
	role, _ := c.Get("role")
	isGlobal, _ := c.Get("isGlobalVulnerability")

	allowed, err := database.CanAccessVulnerability(id, email.(string), role.(string), isGlobal.(bool))
	if err != nil {
		fmt.Printf("CanAccessVulnerability failed for vuln %d: %v\n", id, err)
		c.AbortWithStatus(http.StatusForbidden)
		return
	}
	if !allowed {
		c.AbortWithStatus(http.StatusForbidden)
		return
	}

	var expirationDate *time.Time
	if input.ExpirationDate != "" {
		parsedDate, err := time.Parse("2006-01-02", input.ExpirationDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Expected YYYY-MM-DD"})
			return
		}
		if parsedDate.Before(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Expiration date cannot be in the past"})
			return
		}
		expirationDate = &parsedDate
	}

	sharedDocument, err := database.GetShareDocument(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create a new shared document
			shareToken, err := generateRandomString(8)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}

			sharedDocument = &models.SharedDocument{
				CreatedAt:         time.Now(),
				DocumentID:        id,
				ShareToken:        shareToken,
				SharedByEmail:     email.(string),
				ExpirationDate:    expirationDate,
				AccessType:        input.AccessType,
				InvitedEmailsJSON: input.InvitedEmailsJSON,
				Passphrase:        input.Passphrase,
			}
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		// Update existing document
		sharedDocument.ExpirationDate = expirationDate
		sharedDocument.AccessType = input.AccessType
		sharedDocument.InvitedEmailsJSON = input.InvitedEmailsJSON
		sharedDocument.Passphrase = input.Passphrase
	}

	// Validate input
	if sharedDocument.AccessType == "passphrase-protected" && sharedDocument.Passphrase == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passphrase is required for passphrase-protected access type"})
		return
	}

	if sharedDocument.AccessType == "only-invited" && len(sharedDocument.InvitedEmailsJSON) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "InvitedEmails are required for only-invited access type"})
		return
	}

	// Persist the shared document
	err = database.PersistShareDocument(sharedDocument)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, sharedDocument)
}

func GetAll(c *gin.Context) {
	email, _ := c.Get("email")

	userShares, err := database.GetAllShares(email.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting all shares"})
		return
	}

	if len(*userShares) == 0 {
		c.JSON(http.StatusOK, []models.SharedDocument{}) // Return an empty list if no shares
		return
	}

	// Strip passphrase from the list response; it's only exposed via the
	// per-finding endpoint that backs the share modal dialog.
	sanitized := make([]models.SharedDocument, 0, len(*userShares))
	for _, share := range *userShares {
		share.Passphrase = ""
		sanitized = append(sanitized, share)
	}

	c.JSON(http.StatusOK, sanitized)
}

func generateRandomString(n int) (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i := 0; i < n; i++ {
		bytes[i] = letters[int(bytes[i])%len(letters)]
	}

	return string(bytes), nil
}

func GetPublicVulnerability(c *gin.Context, store *session.SessionStore) {
	token := c.Param("token")
	if token == "" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	exists, err := database.ExistsShare(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "notFound"})
		return
	}

	if exists.ExpirationDate != nil && exists.ExpirationDate.Before(time.Now()) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "expired"})
		return
	}

	if exists.AccessType == "only-invited" || exists.AccessType == "organization" {
		userInfo, err := auth.GetSignedCookie(c, "session_cookie")
		if err != nil {
			// Handle error or invalid session
			c.AbortWithStatus(http.StatusUnauthorized)
			fmt.Println(err)
			return
		}

		validation, err := store.ValidateSession(userInfo.Email, userInfo.SessionID)
		if err != nil {
			// Handle invalid session
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - invalid session"})
			return
		}

		if !validation.IsValid {
			// Handle completely invalid session
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized - session not found"})
			return
		}

		if exists.AccessType == "only-invited" {
			emailAllowed := false
			for _, invitedEmail := range exists.InvitedEmailsJSON {
				if strings.TrimSpace(userInfo.Email) == strings.TrimSpace(invitedEmail) {
					emailAllowed = true
					break
				}
			}

			if !emailAllowed {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
				return
			}
		}
	}

	if exists.AccessType == "passphrase-protected" {
		var input struct {
			Passphrase string `json:"passphrase" binding:"required"`
		}
		if err := c.ShouldBindJSON(&input); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Passphrase is required"})
			return
		}

		// Verify the passphrase
		if input.Passphrase != exists.Passphrase {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid passphrase"})
			return
		}
	}

	// Retrieve the vulnerability data from the database
	vuln, err := database.GetJSONData(exists.DocumentID)
	if err != nil {
		fmt.Printf("failed to retrieve vulnerability data, %v", err)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "notFound"})
		return
	}

	if vuln.Status == "Resolved" || vuln.Status == "Rejected" {
		fmt.Printf("Status is closed")
		c.AbortWithStatusJSON(http.StatusOK, gin.H{"error": "closed"})
		return
	}

	var vulnerability map[string]interface{}
	if err := json.Unmarshal(vuln.Vulnerability, &vulnerability); err != nil {
		fmt.Printf("failed to unmarshal vulnerability data, %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to unmarshal vulnerability data"})
		return
	}

	// Extract images and get their base64 values
	if err := extractMarkdownImagesWithBase64(vulnerability); err != nil {
		fmt.Printf("failed to process images: %v", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to process images"})
		return
	}

	// Get the evidence markdown from the vulnerability map
	evidenceMarkdown, ok := vulnerability["evidence"].(string)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "evidence is not a string"})
		return
	}

	// Update the evidence in the vulnerability map
	vulnerability["evidence"] = evidenceMarkdown

	// Update the vulnerability data with images embedded as base64 strings
	vulnBytes, err := json.Marshal(vulnerability)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal updated vulnerability data"})
		return
	}
	vuln.Vulnerability = datatypes.JSON(vulnBytes)

	var foundby struct {
		Avatar string `json:"avatar"`
		Name   string `json:"name"`
	}

	user, _ := database.GetUserDataByEmail(vuln.FoundBy)
	if user != nil {
		foundby.Avatar = user.Picture
		foundby.Name = user.Name
	}

	foundBytes, err := json.Marshal(foundby)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal updated vulnerability data"})
		return
	}

	vuln.FoundBy = string(foundBytes) //doesnt work now does it!

	// Return the updated vulnerability data
	c.JSON(http.StatusOK, vuln)
}

// Match either the new scoped attachment URL or the legacy global blob URL.
// New scoped pattern is matched first so it wins in the alternation when
// markdown happens to embed both (unlikely, but defensive).
var (
	scopedAttachmentURLRegex = regexp.MustCompile(`/api/vulnerability/(\d+)/attachments/([a-zA-Z0-9\-]+)`)
	legacyBlobShareURLRegex  = regexp.MustCompile(`/api/blob/([a-zA-Z0-9\-]+(?:\.[a-zA-Z0-9]+)?)`)
)

// extractMarkdownImagesWithBase64 walks evidence + remediation markdown and
// the images[] array, replacing both legacy /api/blob/<filename> URLs and
// new /api/vulnerability/<id>/attachments/<key> URLs with base64 data: URIs.
//
// Shared documents render in unauthenticated browsers; baking the bytes into
// the response is how the share viewer sees the images without ever hitting
// the auth-gated attachment endpoint.
func extractMarkdownImagesWithBase64(vulnerability map[string]interface{}) error {
	resolved := map[string]string{} // url → data: URI

	resolve := func(url string) string {
		if cached, ok := resolved[url]; ok {
			return cached
		}
		if m := scopedAttachmentURLRegex.FindStringSubmatch(url); len(m) == 3 {
			id, err := strconv.ParseUint(m[1], 10, 64)
			if err != nil {
				return ""
			}
			att, err := database.GetAttachment(uint(id), m[2])
			if err != nil {
				fmt.Printf("share: load attachment %s: %v\n", url, err)
				return ""
			}
			mime, data := att.ProxyMime, att.ProxyData
			if len(data) == 0 {
				mime, data = att.Mime, att.OriginalData
			}
			if len(data) == 0 || mime == "" {
				return ""
			}
			out := "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
			resolved[url] = out
			return out
		}
		if m := legacyBlobShareURLRegex.FindStringSubmatch(url); len(m) == 2 {
			img, err := database.GetImage(m[1])
			if err != nil {
				fmt.Printf("share: load legacy blob %s: %v\n", url, err)
				return ""
			}
			mime := database.SniffAttachmentMime(img.Data)
			if !database.AllowedAttachmentMime(mime) {
				mime = "application/octet-stream"
			}
			out := "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(img.Data)
			resolved[url] = out
			return out
		}
		return ""
	}

	combined := regexp.MustCompile(scopedAttachmentURLRegex.String() + `|` + legacyBlobShareURLRegex.String())

	for _, field := range []string{"evidence", "remediation"} {
		raw, ok := vulnerability[field].(string)
		if !ok || raw == "" {
			continue
		}
		vulnerability[field] = combined.ReplaceAllStringFunc(raw, func(match string) string {
			if dataURI := resolve(match); dataURI != "" {
				return dataURI
			}
			return match
		})
	}

	// The legacy images[] array stored raw filenames (no /api/blob/ prefix).
	// Resolve those against the legacy path so old shared documents still
	// render their attachments.
	if images, ok := vulnerability["images"].([]interface{}); ok {
		for i, img := range images {
			name, ok := img.(string)
			if !ok {
				continue
			}
			if dataURI := resolve("/api/blob/" + name); dataURI != "" {
				images[i] = dataURI
			}
		}
		vulnerability["images"] = images
	}

	return nil
}
