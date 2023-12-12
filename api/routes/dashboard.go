package routes

import (
	"prism/database"
	"net/http"

	"github.com/gin-gonic/gin"
)

func HandleDashboard(c *gin.Context){
	total,_ := database.CountJSONData()
	criticalities, _ := database.CountCriticalities()
	owasp,_ := database.CountOWASPCategories()
	projects := 10
	bugBounties := 5
	unresolvedTasks := 3
	c.JSON(http.StatusOK, gin.H{"total": total,"projects": projects,"bugBounties": bugBounties, "unresolved": unresolvedTasks, "criticalities": criticalities, "owasp": owasp})
}