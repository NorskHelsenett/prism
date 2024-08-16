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
	isGlobal, _ := c.Get("isGlobalVulnerability")

	allShares, err := database.GetAllShares()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting all shares"})
		return
	}

	// Extract document IDs from shares
	var shareDocumentIds []uint
	for _, share := range *allShares {
		shareDocumentIds = append(shareDocumentIds, share.DocumentID)
	}

	requestedIds, err := database.GetVulnerabilityIds(isGlobal.(bool), email.(string), shareDocumentIds)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "error getting vulnerabilities and shares"})
		return
	}

	// Get the intersection of requested IDs and share document IDs
	intersectedIds := intersect(requestedIds, shareDocumentIds)

	var filteredShares []models.SharedDocument
	for _, share := range *allShares {
		if contains(intersectedIds, share.DocumentID) {
			filteredShares = append(filteredShares, share)
		}
	}

	if len(filteredShares) == 0 {
		c.JSON(http.StatusOK, []models.SharedDocument{}) // Return an empty list if no shares
		return
	}

	c.JSON(http.StatusOK, filteredShares)
}

func contains(slice []uint, item uint) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}

func intersect(a, b []uint) []uint {
	set := make(map[uint]bool)
	var result []uint
	for _, item := range a {
		set[item] = true
	}
	for _, item := range b {
		if set[item] {
			result = append(result, item)
		}
	}
	return result
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

	imageBase64Map := make(map[string]string)
	if images, ok := vulnerability["images"].([]interface{}); ok {
		for i, img := range images {
			if imageName, ok := img.(string); ok {
				// Retrieve the image data from the database
				imageData, err := database.GetImage(imageName)
				if err != nil {
					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to read image data"})
					return
				}

				// Convert the image data to a base64 string and store it in the map
				base64Data := base64.StdEncoding.EncodeToString(imageData.Data)
				imageBase64Map[imageName] = base64Data

				// Update the image entry in the images slice
				images[i] = imageData.Data
			}
		}
		vulnerability["images"] = images
	}

	// Get the evidence markdown from the vulnerability map
	evidenceMarkdown, ok := vulnerability["evidence"].(string)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "evidence is not a string"})
		return
	}

	// Replace all markdown image links with base64 data
	for imageName, base64Data := range imageBase64Map {
		// Use a regular expression to replace the image link with a base64 string
		re := regexp.MustCompile(`!\[` + regexp.QuoteMeta(imageName) + `\]\((http[s]?:\/\/.*?/api/blob/` + regexp.QuoteMeta(imageName) + `)\)`)
		evidenceMarkdown = re.ReplaceAllString(evidenceMarkdown, fmt.Sprintf("![%s](data:image/jpeg;base64,%s)", imageName, base64Data))
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

	// Return the updated vulnerability data
	c.JSON(http.StatusOK, vuln)
}
