package routes

import (
	"encoding/json"
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

// UpdateUserPreferences merges incoming preference fields into the user's stored settings.
// Keys not present in the request body are preserved, so different feature areas (board,
// project list, calendar) can each PATCH only the fields they own without clobbering others.
func UpdateUserPreferences(c *gin.Context) {
	email, _ := c.Get("email")

	var patch map[string]any
	if err := c.ShouldBindJSON(&patch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	existing, err := database.GetPreferencesForUser(email.(string))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load existing preferences"})
		return
	}

	existingBytes, err := json.Marshal(existing)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode existing preferences"})
		return
	}

	var merged map[string]any
	if err := json.Unmarshal(existingBytes, &merged); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode existing preferences"})
		return
	}
	if merged == nil {
		merged = map[string]any{}
	}
	for k, v := range patch {
		merged[k] = v
	}

	mergedBytes, err := json.Marshal(merged)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode merged preferences"})
		return
	}
	var final models.UserSettings
	if err := json.Unmarshal(mergedBytes, &final); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid preference fields"})
		return
	}

	if err := database.PatchSettingsForUser(email.(string), final); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update preferences"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Preferences updated successfully"})
}
