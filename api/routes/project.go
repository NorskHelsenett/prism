package routes

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prism/database"
)

// normalizeDateOnly returns a YYYY-MM-DD string for plain dates and clears values
// that represent "unset" — including Go's time.Time zero (0001-01-01...) which can
// leak in if a date was once serialized as a time.Time before this column became a
// plain string field.
func normalizeDateOnly(value string) string {
	if value == "" {
		return ""
	}
	if len(value) >= 10 {
		prefix := value[:10]
		if prefix == "0001-01-01" {
			return ""
		}
		return prefix
	}
	return value
}

func GetProjects(c *gin.Context) {
	isGlobal, _ := c.Get("isGlobalProject")

	var dbProjects []database.ProjectData
	var err error

	if isGlobal == true {
		dbProjects, err = database.GetProjects()
		if err != nil {
			log.Printf("GetProjects() failed: %v", err)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
	} else {
		email, _ := c.Get("email")
		dbProjects, err = database.GetProjectsFor(email.(string))
		if err != nil {
			log.Printf("GetProjectsFor(%q) failed: %v", email, err)
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
	}

	rangeStart := c.Query("startDate")
	rangeEnd := c.Query("endDate")

	type projectSummary struct {
		ID           uint      `json:"ID"`
		Name         string    `json:"ProjectName"`
		CreatedAt    time.Time `json:"CreatedAt"`
		IsBugBounty  bool      `json:"IsBugBounty"`
		ClientEmail  string    `json:"ClientEmail"`
		HackerName   string    `json:"HackerName"`
		SlackChannel string    `json:"SlackChannel"`
		StartDate    string    `json:"StartDate"`
		EndDate      string    `json:"EndDate"`
		Color        string    `json:"Color"`
		GroupID      *uint     `json:"GroupID"`
		SortOrder    int       `json:"SortOrder"`
	}

	projects := make([]projectSummary, 0, len(dbProjects))
	for _, p := range dbProjects {
		startDate := normalizeDateOnly(p.StartDate)
		endDate := normalizeDateOnly(p.EndDate)
		if rangeStart != "" && rangeEnd != "" {
			// Only return projects whose [StartDate, EndDate] overlaps the requested range.
			// Projects with missing dates are excluded from windowed queries.
			if startDate == "" || endDate == "" {
				continue
			}
			if endDate < rangeStart || startDate > rangeEnd {
				continue
			}
		}
		projects = append(projects, projectSummary{
			ID:           p.ID,
			Name:         p.ProjectName,
			CreatedAt:    p.CreatedAt,
			IsBugBounty:  p.IsBugBounty,
			ClientEmail:  p.ClientEmail,
			HackerName:   p.HackerName,
			SlackChannel: p.SlackChannel,
			StartDate:    startDate,
			EndDate:      endDate,
			Color:        p.Color,
			GroupID:      p.GroupID,
			SortOrder:    p.SortOrder,
		})
	}

	c.JSON(http.StatusOK, projects)
}

func HandleProjectPut(c *gin.Context) {
	projectIDStr := c.Param("projectID")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, strconv.IntSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Project ID"})
		return
	}

	var updatedProjectData database.ProjectData
	// Step 2: Bind incoming JSON data to the struct.
	if err := c.ShouldBindJSON(&updatedProjectData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check that projectID matches updatedProjectData.ID
	if uint(projectID) != updatedProjectData.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID mismatch"})
		return
	}

	updatedProjectData.StartDate = normalizeDateOnly(updatedProjectData.StartDate)
	updatedProjectData.EndDate = normalizeDateOnly(updatedProjectData.EndDate)

	err = database.UpdateProject(&updatedProjectData)
	if err != nil {
		if !errors.Is(err, database.ErrNotFound) && !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("UpdateProject(%d) failed: %v", updatedProjectData.ID, err)
		}
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Project updated successfully"})
}

func HandleProjectPost(c *gin.Context) {
	var projectData database.ProjectData
	if err := c.ShouldBindJSON(&projectData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	projectData.StartDate = normalizeDateOnly(projectData.StartDate)
	projectData.EndDate = normalizeDateOnly(projectData.EndDate)

	if err := database.CreateProject(&projectData); err != nil {
		log.Printf("CreateProject failed: %v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": projectData.ID})
}

func GetProject(c *gin.Context) {

	projectIDStr := c.Param("projectID")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, strconv.IntSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Project ID"})
		return
	}

	id := uint(projectID)

	dbProject, err := database.GetProject(id)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("GetProject(%d) failed: %v", id, err)
		}
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	// Check if full data is requested
	full := c.Query("full") == "true"

	if full {
		// Return full project data
		c.JSON(http.StatusOK, dbProject)
		return
	}

	// Map database project to simplified Project struct
	project := struct {
		ID        uint   `json:"ID"`
		Name      string `json:"ProjectName"`
		StartDate string `json:"StartDate"`
		EndDate   string `json:"EndDate"`
		Color     string `json:"Color"`
	}{
		ID:        dbProject.ID,
		Name:      dbProject.ProjectName,
		StartDate: normalizeDateOnly(dbProject.StartDate),
		EndDate:   normalizeDateOnly(dbProject.EndDate),
		Color:     dbProject.Color,
	}

	c.JSON(http.StatusOK, project)
}

func GetProjectVulnerabilitiesForProject(c *gin.Context) {
	isGlobalVal, _ := c.Get("isGlobalVulnerability")
	isGlobal, _ := isGlobalVal.(bool)
	emailVal, _ := c.Get("email")
	email, _ := emailVal.(string)

	projectIDStr := c.Param("projectID")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, strconv.IntSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Project ID"})
		return
	}

	// Retrieve optional dateFrom and dateTo query parameters
	dateFrom := c.Query("from")
	dateTo := c.Query("to")

	vulnerabilites, err := database.GetProjectVulnerabilities(uint(projectID), dateFrom, dateTo)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("GetProjectVulnerabilities(%d) failed: %v", projectID, err)
		}
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	vulnerabilites = database.FilterJSONDataForUser(vulnerabilites, email, isGlobal)

	c.JSON(http.StatusOK, vulnerabilites)
}

func GetProjectVulnerabilitiesTotal(c *gin.Context) {
	projectIDStr := c.Param("projectID")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, strconv.IntSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Project ID"})
		return
	}

	total, err := database.CountProjectVulnerabilities(uint(projectID))
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("CountProjectVulnerabilities(%d) failed: %v", projectID, err)
		}
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{"total_vulnerabilities": total})
}

func DeleteProject(c *gin.Context) {
	projectIDStr := c.Param("projectID")
	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Project ID is required"})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, strconv.IntSize)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Project ID"})
		return
	}

	err = database.DeleteProjectAndAssets(uint(projectID))
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("DeleteProjectAndAssets(%d) failed: %v", projectID, err)
		}
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.JSON(http.StatusOK, gin.H{})
}
