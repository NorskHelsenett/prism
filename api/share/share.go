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

func extractMarkdownImagesWithBase64(vulnerability map[string]interface{}) error {
	// Fields to process
	markdownFields := []string{"evidence", "remediation"}

	// Regex pattern for UUID-style files with 3 or 4 char extensions
	pattern := `/api/blob/[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}\.[a-zA-Z]{3,4}`
	re := regexp.MustCompile(pattern)

	// Map to store image data (prevent fetching same image multiple times)
	imageBase64Map := make(map[string]string)

	// Process each markdown field
	for _, field := range markdownFields {
		markdown, ok := vulnerability[field].(string)
		if !ok {
			continue // Skip if field doesn't exist or isn't a string
		}

		// Find all matches in the markdown
		matches := re.FindAllString(markdown, -1)

		// Process each matched image path
		for _, match := range matches {
			// Extract just the filename from the path
			imageName := match[len("/api/blob/"):]

			// Get base64 data if we haven't already
			if _, exists := imageBase64Map[imageName]; !exists {
				// Get the image data from database
				imageData, err := database.GetImage(imageName)
				if err != nil {
					fmt.Printf("failed to get image %s: %v\n", imageName, err)
					continue
				}

				// Convert to base64
				base64Data := base64.StdEncoding.EncodeToString(imageData.Data)
				imageBase64Map[imageName] = base64Data
			}

			// Replace the URL with base64 data
			markdown = strings.Replace(
				markdown,
				match,
				fmt.Sprintf("data:image/jpeg;base64,%s", imageBase64Map[imageName]),
				-1,
			)
		}

		// Update the field in vulnerability map
		vulnerability[field] = markdown
	}

	// Update images array with base64 data
	if images, ok := vulnerability["images"].([]interface{}); ok {
		for i, img := range images {
			if imageName, ok := img.(string); ok {
				// Skip if we already have this image
				if base64Data, exists := imageBase64Map[imageName]; exists {
					images[i] = base64Data
					continue
				}

				// Get the image data from database
				imageData, err := database.GetImage(imageName)
				if err != nil {
					fmt.Printf("failed to get image %s: %v\n", imageName, err)
					continue
				}

				// Convert to base64
				base64Data := base64.StdEncoding.EncodeToString(imageData.Data)
				images[i] = base64Data
				imageBase64Map[imageName] = base64Data
			}
		}
		vulnerability["images"] = images
	}

	return nil
}
