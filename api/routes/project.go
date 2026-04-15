package routes

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prism/database"
	"strconv"
)

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

	// Map database projects to simplified Project struct
	projects := make([]struct {
		ID          uint      `json:"ID"`
		Name        string    `json:"ProjectName"`
		CreatedAt   time.Time `json:"CreatedAt"`
		IsBugBounty bool      `json:"IsBugBounty"`
		ClientEmail string    `json:"ClientEmail"`
	}, len(dbProjects))

	for i, p := range dbProjects {
		projects[i].ID = p.ID
		projects[i].Name = p.ProjectName
		projects[i].CreatedAt = p.CreatedAt
		projects[i].IsBugBounty = p.IsBugBounty
		projects[i].ClientEmail = p.ClientEmail
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
		ID   uint   `json:"ID"`
		Name string `json:"ProjectName"`
	}{
		ID:   dbProject.ID,
		Name: dbProject.ProjectName,
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
