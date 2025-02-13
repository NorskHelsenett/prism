package routes

import (
	"net/http"
	"prism/database"

	"github.com/gin-gonic/gin"
)

func HandleDashboard(c *gin.Context) {
	year := c.Query("year") // Get year from query parameter

	total, _ := database.CountJSONData(year)
	criticalities, _ := database.CountCriticalities(year)
	owasp, _ := database.CountOWASPCategories(year)
	owaspCriticalities, _ := database.FetchOWASPCriticalities(year)
	projects, _ := database.CountProjects()
	bugBounties, _ := database.CountBugBounties(year)
	statuses, _ := database.CountByStatus(year)

	c.JSON(http.StatusOK, gin.H{
		"total":              total,
		"projects":           projects,
		"bugBounties":        bugBounties,
		"statuses":           statuses,
		"criticalities":      criticalities,
		"owasp":              owasp,
		"owaspCriticalities": owaspCriticalities,
	})
}
