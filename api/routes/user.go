package routes

import (
	"prism/config"
	"prism/database"
	"prism/session"

	"net/http"

	"github.com/gin-gonic/gin"
)

func GetUserInfo(c *gin.Context) {
	email := c.Param("email")
	user, _ := database.GetUserDataByEmail(email)
	if user != nil {
		user.Role = "" // limit data to viewer
	}
	c.Header("Cache-Control", "public, max-age=3600")
	c.JSON(http.StatusOK, user)
}

func GetAccessListRoutes(c *gin.Context) {
	// Retrieve the role from the context
	role, exists := c.Get("role")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Find the role in the appConfig
	roleConfig, roleExists := config.AppConfig.Roles[role.(string)]
	if !roleExists {
		c.JSON(http.StatusForbidden, gin.H{"error": "role not found"})
		return
	}
	// Return the accessible paths with their actions
	c.JSON(http.StatusOK, roleConfig)
}

// Helper function to check if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func UpdateUser(c *gin.Context, s *session.SessionStore) {
	var user database.UserData

	// Parse the incoming JSON to newSettings
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
		return
	}

	if user.Role == "admin" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := database.UpdateUser(&user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	err = s.UpdateUser(&user)

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

func GetAllUsers(c *gin.Context, s *session.SessionStore) {
	users, _ := session.GetAllUsers(s)
	c.JSON(http.StatusOK, users)
}

func GetAllProfilesEmailOnly(c *gin.Context, s *session.SessionStore) {
	users, _ := session.GetAllProfiles(s)
	c.JSON(http.StatusOK, users)
}
