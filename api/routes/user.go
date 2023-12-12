package routes

import (
	"prism/database"

	"net/http"
	"github.com/gin-gonic/gin"
)

func GetUserInfo(c *gin.Context){
	email := c.Param("email")
	user,_ := database.GetUserDataByEmail(email)
	c.JSON(http.StatusOK, user)
}