package routes

import (
	"fmt"
	"net/http"
	"prism/crypto"
	"prism/database"
	"prism/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

func ValidateAPIKey(apiKey string) (email string, valid bool) {
	hashedAPIKey, err := crypto.HashAPIKey(apiKey)
	if err != nil {
		return "", false
	}
	return database.ValidateAPIKey(hashedAPIKey)
}

func DeleteAPIKey(c *gin.Context) {
	email, _ := c.Get("email")

	idStr := c.Param("id") // Get ID as string
	if idStr != "" {
		// Convert string ID to uint
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID format"})
			return
		}
		err = database.DeleteApiKey(uint(id), email.(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "unable to delete assessment"})
		}
		c.JSON(http.StatusOK, gin.H{"status": "deleted successfully"})
	} else {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "id is required"})
	}
}

func GetAPIKey(c *gin.Context) {
	email, _ := c.Get("email")

	apiKeys, err := database.GetApiKeys(email.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unable to retrieve apikeys"})
		return
	}

	c.JSON(http.StatusOK, apiKeys)
}

func CreateAPIKey(c *gin.Context) {
	email, _ := c.Get("email")
	var apikey models.APIKey

	if err := c.BindJSON(&apikey); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
		return
	}

	apikey.Email = email.(string)

	apikeyRef, err := crypto.CreateAPIKey(&apikey)
	if err != nil {
		fmt.Printf("Error creating new API key to database")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
		return
	}

	newapikey, err := database.PersistApiKey(apikeyRef)
	if err != nil {
		fmt.Printf("Error creating new API key to database")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
		return
	}

	c.JSON(http.StatusOK, newapikey)
}
