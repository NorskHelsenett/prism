package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"prism/auth"
	"fmt"
)

func HandleProjectPost(c *gin.Context){
    email, exists := c.Request.Context().Value(auth.EmailContextKey).(string)
    if !exists {
        // Handle missing email
        c.AbortWithStatus(http.StatusInternalServerError)
        return
    }
    fmt.Println(email)
}