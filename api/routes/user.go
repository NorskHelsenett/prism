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

func UpdateUserRole(c *gin.Context, s *session.SessionStore) {
	var user database.UserData

	isAdmin, _ := c.Get("isAdmin")
	if !isAdmin.(bool) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Parse the incoming JSON to newSettings
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
		return
	}

	if user.Role == "admin" {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	err := database.UpdateUserRole(&user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	_ = s.UpdateUserRole(&user)

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

func GetAllUsers(c *gin.Context, s *session.SessionStore) {
	users, _ := session.GetAllUsers(s)
	c.JSON(http.StatusOK, users)
}

func DeleteUser(c *gin.Context, s *session.SessionStore) {
	id := c.Param("id")

	err := s.DB.Where("id = ?", id).Delete(&database.UserData{}).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user from session store"})
		return
	}

	err = database.DeleteUser(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Successfully deleted user"})
}

func ToggleUserActive(c *gin.Context, s *session.SessionStore) {
	id := c.Param("id")

	isAdmin, _ := c.Get("isAdmin")
	if !isAdmin.(bool) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := database.ToggleUserActive(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle user active status"})
		return
	}

	// Sync to session store
	_ = s.DB.Model(&database.UserData{}).Where("email = ?", user.Email).Update("active", user.Active).Error

	c.JSON(http.StatusOK, user)
}

func GetAllProfilesEmailOnly(c *gin.Context) {
	roleName, _ := c.Get("role")
	role, exists := config.AppConfig.Roles[roleName.(string)]
	if !exists {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Only roles with write access to projects or planning need user listings
	if !roleHasWriteAccess(role, "/project", "/planning") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	users, err := database.GetAllProfilesWithTeams()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

func roleHasWriteAccess(role config.Role, resources ...string) bool {
	for _, perm := range role.Permissions {
		for _, res := range resources {
			if perm.Resource == res {
				for _, action := range perm.Action {
					if action == "write" || action == "*" {
						return true
					}
				}
			}
		}
	}
	return false
}
