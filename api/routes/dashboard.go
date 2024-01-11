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
	owaspCriticalities, _ := database.FetchOWASPCriticalities()
	projects,_ := database.CountProjects()
	bugBounties,_ := database.CountBugBounties()
	unresolvedTasks,_ := database.CountUnresolvedTasks()
	c.JSON(http.StatusOK, gin.H{"total": total,"projects": projects,"bugBounties": bugBounties, "unresolved": unresolvedTasks, "criticalities": criticalities, "owasp": owasp, "owaspCriticalities": owaspCriticalities})
}