package routes

import (
	"prism/database"

	"net/http"
	"github.com/gin-gonic/gin"
)

func GetUserInfo(c *gin.Context){
	email := c.Param("email")
	user,_ := database.GetUserDataByEmail(email)
	c.Header("Cache-Control", "public, max-age=3600")
	c.JSON(http.StatusOK, user)
}

func GetAllUsers(c *gin.Context) {
	users,_ := database.GetAllUsers()
	c.JSON(http.StatusOK, users)
}