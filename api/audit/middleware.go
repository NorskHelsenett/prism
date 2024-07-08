package audit

import (
	"prism/auth"
	"prism/database"
	"prism/routes"

	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type WrappedResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

func NewWrappedResponseWriter(w http.ResponseWriter) *WrappedResponseWriter {
	return &WrappedResponseWriter{ResponseWriter: w, StatusCode: http.StatusOK}
}

func (wrw *WrappedResponseWriter) WriteHeader(code int) {
	wrw.StatusCode = code
	wrw.ResponseWriter.WriteHeader(code)
}

func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		settings, _ := database.GetSettings(false)
		if settings == nil || !settings.AuditLog.Enabled {
			c.Next()
			return
		}

		var email string
		method := c.Request.Method

		apiKey := c.GetHeader("x-api-key")
		if apiKey != "" {
			emailFound, _ := routes.ValidateAPIKey(apiKey)
			method += " [x-api-key]"
			email = emailFound
		}

		userInfo, err := auth.GetSignedCookie(c, "session_cookie")
		if err != nil {
			fmt.Println(err)
		} else {
			email = userInfo.Email
		}

		// Initialize your audit log with HTTP method
		auditLog := database.AuditLog{
			Timestamp: time.Now(),
			Action:    c.Request.URL.Path,
			Method:    method, // Capture the HTTP method
			UserEmail: email,
		}

		// Process request
		c.Next()

		// Get status code
		statusCode := c.Writer.Status()
		auditLog.Status = fmt.Sprintf("HTTP %d", statusCode)

		// Record the audit log asynchronously
		go database.RecordAuditLog(auditLog)
	}
}

func GetAllAudits(c *gin.Context) {
	limitQuery := c.DefaultQuery("limit", "0")
	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		// Handle error if the limit is not an integer
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid limit value"})
		return
	}

	audit, err := database.GetAllAudits(limit)
	if err != nil {
		log.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load audits"})
		return
	}

	countAudit, err := database.CountAllAudits()
	if err != nil {
		log.Printf("%v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count audits"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"total": countAudit, "audits": audit})
}
