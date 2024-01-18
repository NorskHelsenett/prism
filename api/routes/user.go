package routes

import (
	"prism/database"

	"github.com/gin-gonic/gin"
	"net/http"
)

func GetUserInfo(c *gin.Context) {
	email := c.Param("email")
	user, _ := database.GetUserDataByEmail(email)
	c.Header("Cache-Control", "public, max-age=3600")
	c.JSON(http.StatusOK, user)
}

func UpdateUser(c *gin.Context) {
	var user database.UserData
	email, _ := c.Request.Context().Value("email").(string)

	// Parse the incoming JSON to newSettings
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid data format"})
		return
	}

	if user.Email != email {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Wrong data values"})
	}

	err := database.UpdateUser(&user)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

func GetAllUsers(c *gin.Context) {
	users, _ := database.GetAllUsers()
	c.JSON(http.StatusOK, users)
}
