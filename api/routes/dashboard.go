package routes

import (
	"net/http"
	"prism/database"

	"github.com/gin-gonic/gin"
)

func HandleDashboard(c *gin.Context) {
	year := c.Query("year")

	emailValue, _ := c.Get("email")
	email, _ := emailValue.(string)

	globalProjectValue, _ := c.Get("isGlobalProject")
	isGlobalProject, _ := globalProjectValue.(bool)

	globalVulnerabilityValue, _ := c.Get("isGlobalVulnerability")
	isGlobalVulnerability, _ := globalVulnerabilityValue.(bool)

	metrics, err := database.GetDashboardMetrics(year, email, isGlobalVulnerability, isGlobalProject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":              metrics.Total,
		"projects":           metrics.Projects,
		"bugBounties":        metrics.BugBounties,
		"statuses":           metrics.Statuses,
		"criticalities":      metrics.Criticalities,
		"owasp":              metrics.OWASP,
		"owaspCriticalities": metrics.OWASPCriticalities,
		"years":              metrics.Years,
	})
}
