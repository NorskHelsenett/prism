package routes

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prism/database"
)

type projectGroupRequest struct {
	Name      string `json:"name"`
	Color     string `json:"color"`
	SortOrder int    `json:"sortOrder"`
}

type projectGroupResponse struct {
	ID        uint   `json:"ID"`
	Name      string `json:"Name"`
	Color     string `json:"Color"`
	SortOrder int    `json:"SortOrder"`
}

func GetProjectGroups(c *gin.Context) {
	groups, err := database.ListProjectGroups()
	if err != nil {
		log.Printf("ListProjectGroups failed: %v", err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	out := make([]projectGroupResponse, len(groups))
	for i, g := range groups {
		out[i] = projectGroupResponse{
			ID:        g.ID,
			Name:      g.Name,
			Color:     g.Color,
			SortOrder: g.SortOrder,
		}
	}
	c.JSON(http.StatusOK, out)
}

func PostProjectGroup(c *gin.Context) {
	var req projectGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Name == "" {
		req.Name = "New group"
	}
	group := database.ProjectGroup{
		Name:      req.Name,
		Color:     req.Color,
		SortOrder: req.SortOrder,
	}
	if err := database.CreateProjectGroup(&group); err != nil {
		log.Printf("CreateProjectGroup failed: %v", err)
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	c.JSON(http.StatusCreated, projectGroupResponse{
		ID:        group.ID,
		Name:      group.Name,
		Color:     group.Color,
		SortOrder: group.SortOrder,
	})
}

func PutProjectGroup(c *gin.Context) {
	idStr := c.Param("groupID")
	id, err := strconv.ParseUint(idStr, 10, strconv.IntSize)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var req projectGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	group := database.ProjectGroup{
		Name:      req.Name,
		Color:     req.Color,
		SortOrder: req.SortOrder,
	}
	group.ID = uint(id)
	if err := database.UpdateProjectGroup(&group); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		log.Printf("UpdateProjectGroup(%d) failed: %v", id, err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Project group updated"})
}

func DeleteProjectGroupHandler(c *gin.Context) {
	idStr := c.Param("groupID")
	id, err := strconv.ParseUint(idStr, 10, strconv.IntSize)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	if err := database.DeleteProjectGroup(uint(id)); err != nil {
		log.Printf("DeleteProjectGroup(%d) failed: %v", id, err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, gin.H{})
}

type assignProjectGroupRequest struct {
	GroupID   *uint `json:"groupId"`
	SortOrder *int  `json:"sortOrder"`
}

func PatchProjectGroupAssignment(c *gin.Context) {
	idStr := c.Param("projectID")
	id, err := strconv.ParseUint(idStr, 10, strconv.IntSize)
	if err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}
	var req assignProjectGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.SetProjectGroup(uint(id), req.GroupID, req.SortOrder); err != nil {
		log.Printf("SetProjectGroup(%d) failed: %v", id, err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Project group assignment updated"})
}
