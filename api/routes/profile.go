package routes

import (
	"errors"
	"net/http"
	"prism/database"
	"prism/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetUserPreferences returns the current user's preferences
func GetUserPreferences(c *gin.Context) {
	email, _ := c.Get("email")

	preferences, err := database.GetPreferencesForUser(email.(string))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusOK, models.UserSettings{})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch preferences"})
		return
	}

	c.JSON(http.StatusOK, preferences)
}

// UpdateUserPreferences updates the user's preferences
func UpdateUserPreferences(c *gin.Context) {
	email, _ := c.Get("email")

	var preferences models.UserSettings
	if err := c.ShouldBindJSON(&preferences); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if err := database.PatchSettingsForUser(email.(string), preferences); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preferences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Preferences updated successfully"})
}
