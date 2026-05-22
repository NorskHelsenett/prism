package routes

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"prism/config"
	"prism/database"
	"prism/models"
	"prism/report"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// hasGlobalProjectAccess mirrors auth.globalAccess for the report routes,
// avoiding a cross-package export. A role has global project access when
// it lists "/project/:id" in its permissions (and matches the request
// action implicitly via RBACMiddleware, which already gated the route).
func hasGlobalProjectAccess(role string) bool {
	if role == "admin" {
		return true
	}
	r, ok := config.AppConfig.Roles[role]
	if !ok {
		return false
	}
	for _, p := range r.Permissions {
		if p.Resource == "/project/:id" {
			return true
		}
	}
	return false
}

// generateReportShareToken returns an 8-char base62 token, retrying on
// collision. Mirrors share.generateRandomString but keeps it package-local
// so we can guarantee uniqueness in the report table.
func generateReportShareToken() (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	for attempt := 0; attempt < 8; attempt++ {
		buf := make([]byte, 8)
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		for i := range buf {
			buf[i] = letters[int(buf[i])%len(letters)]
		}
		token := string(buf)
		exists, err := database.TokenExists(token)
		if err != nil {
			return "", err
		}
		if !exists {
			return token, nil
		}
	}
	return "", errors.New("failed to generate unique share token")
}

// notFound responds with a generic 404 per the project-wide rule of never
// returning 5xx for ACL or input-shape failures (errors are logged
// internally; the caller sees nothing implementation-specific).
func notFound(c *gin.Context) {
	c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
}

type reportCreateInput struct {
	Title            string                          `json:"title" binding:"required"`
	ExecutiveSummary string                          `json:"executiveSummary"`
	ProjectIDs       []uint                          `json:"projectIds" binding:"required"`
	VulnerabilityIDs []uint                          `json:"vulnerabilityIds"`
	FindingOverrides map[uint]models.FindingOverride `json:"findingOverrides"`
}

type reportUpdateInput struct {
	Title            *string                          `json:"title"`
	ExecutiveSummary *string                          `json:"executiveSummary"`
	ProjectIDs       *[]uint                          `json:"projectIds"`
	VulnerabilityIDs *[]uint                          `json:"vulnerabilityIds"`
	FindingOverrides *map[uint]models.FindingOverride `json:"findingOverrides"`
}

type reportShareInput struct {
	InvitedEmails []string `json:"invitedEmails"`
}

func CreateReport(c *gin.Context) {
	email, _ := c.Get("email")
	role, _ := c.Get("role")

	var input reportCreateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(input.ProjectIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least one project is required"})
		return
	}

	ok, err := database.HasWriteOnAllProjects(email.(string), role.(string), input.ProjectIDs, hasGlobalProjectAccess(role.(string)))
	if err != nil {
		fmt.Printf("HasWriteOnAllProjects failed: %v\n", err)
		notFound(c)
		return
	}
	if !ok {
		notFound(c)
		return
	}

	token, err := generateReportShareToken()
	if err != nil {
		fmt.Printf("generateReportShareToken failed: %v\n", err)
		notFound(c)
		return
	}

	r := &models.Report{
		Title:                input.Title,
		ExecutiveSummary:     input.ExecutiveSummary,
		ProjectIDsList:       input.ProjectIDs,
		VulnerabilityIDsList: input.VulnerabilityIDs,
		FindingOverridesData: input.FindingOverrides,
		OwnerEmail:           email.(string),
		ShareToken:           token,
		InvitedEmailsList:    []string{},
	}
	if err := r.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.CreateReport(r); err != nil {
		fmt.Printf("CreateReport failed: %v\n", err)
		notFound(c)
		return
	}
	c.JSON(http.StatusOK, r)
}

func ListReports(c *gin.Context) {
	email, _ := c.Get("email")
	role, _ := c.Get("role")
	global := hasGlobalProjectAccess(role.(string))

	all, err := database.ListReports()
	if err != nil {
		fmt.Printf("ListReports failed: %v\n", err)
		notFound(c)
		return
	}
	visible := make([]models.Report, 0, len(all))
	for _, r := range all {
		ok, err := database.HasReadOnAnyProject(email.(string), role.(string), r.ProjectIDsList, global)
		if err != nil {
			fmt.Printf("HasReadOnAnyProject failed for report %d: %v\n", r.ID, err)
			continue
		}
		if ok {
			visible = append(visible, r)
		}
	}
	c.JSON(http.StatusOK, visible)
}

// loadReportForRead returns the report if the caller can read it via project
// ACL, else writes 404 and returns nil.
func loadReportForRead(c *gin.Context) *models.Report {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		notFound(c)
		return nil
	}
	r, err := database.GetReport(uint(id))
	if err != nil {
		notFound(c)
		return nil
	}
	email, _ := c.Get("email")
	role, _ := c.Get("role")
	ok, err := database.HasReadOnAnyProject(email.(string), role.(string), r.ProjectIDsList, hasGlobalProjectAccess(role.(string)))
	if err != nil {
		fmt.Printf("HasReadOnAnyProject failed: %v\n", err)
		notFound(c)
		return nil
	}
	if !ok {
		notFound(c)
		return nil
	}
	return r
}

// loadReportForWrite is like loadReportForRead but requires write-on-all.
func loadReportForWrite(c *gin.Context) *models.Report {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		notFound(c)
		return nil
	}
	r, err := database.GetReport(uint(id))
	if err != nil {
		notFound(c)
		return nil
	}
	email, _ := c.Get("email")
	role, _ := c.Get("role")
	ok, err := database.HasWriteOnAllProjects(email.(string), role.(string), r.ProjectIDsList, hasGlobalProjectAccess(role.(string)))
	if err != nil {
		fmt.Printf("HasWriteOnAllProjects failed: %v\n", err)
		notFound(c)
		return nil
	}
	if !ok {
		notFound(c)
		return nil
	}
	return r
}

func GetReport(c *gin.Context) {
	r := loadReportForRead(c)
	if r == nil {
		return
	}
	c.JSON(http.StatusOK, r)
}

func PatchReport(c *gin.Context) {
	r := loadReportForWrite(c)
	if r == nil {
		return
	}
	var input reportUpdateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Title != nil {
		r.Title = *input.Title
	}
	if input.ExecutiveSummary != nil {
		r.ExecutiveSummary = *input.ExecutiveSummary
	}
	if input.ProjectIDs != nil {
		// If projects change, re-check write access against the *new* set so
		// you can't escalate by handing yourself a project you can't write to.
		email, _ := c.Get("email")
		role, _ := c.Get("role")
		ok, err := database.HasWriteOnAllProjects(email.(string), role.(string), *input.ProjectIDs, hasGlobalProjectAccess(role.(string)))
		if err != nil || !ok {
			notFound(c)
			return
		}
		r.ProjectIDsList = *input.ProjectIDs
	}
	if input.VulnerabilityIDs != nil {
		r.VulnerabilityIDsList = *input.VulnerabilityIDs
	}
	if input.FindingOverrides != nil {
		r.FindingOverridesData = *input.FindingOverrides
	}

	if err := r.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := database.UpdateReport(r); err != nil {
		fmt.Printf("UpdateReport failed: %v\n", err)
		notFound(c)
		return
	}
	c.JSON(http.StatusOK, r)
}

func DeleteReport(c *gin.Context) {
	r := loadReportForWrite(c)
	if r == nil {
		return
	}
	if err := database.DeleteReport(r.ID); err != nil {
		fmt.Printf("DeleteReport failed: %v\n", err)
		notFound(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "deleted"})
}

func UpdateReportShare(c *gin.Context) {
	r := loadReportForWrite(c)
	if r == nil {
		return
	}
	var input reportShareInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	cleaned := make([]string, 0, len(input.InvitedEmails))
	for _, e := range input.InvitedEmails {
		e = strings.TrimSpace(strings.ToLower(e))
		if e != "" {
			cleaned = append(cleaned, e)
		}
	}
	r.InvitedEmailsList = cleaned
	if err := database.UpdateReport(r); err != nil {
		fmt.Printf("UpdateReport(share) failed: %v\n", err)
		notFound(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"shareToken":    r.ShareToken,
		"invitedEmails": r.InvitedEmailsList,
	})
}

func PublishReport(c *gin.Context) {
	r := loadReportForWrite(c)
	if r == nil {
		return
	}
	email, _ := c.Get("email")

	payload, err := buildSnapshot(r, email.(string))
	if err != nil {
		fmt.Printf("buildSnapshot failed: %v\n", err)
		notFound(c)
		return
	}

	version, err := database.PublishReportVersion(r.ID, func(versionNumber int) (*models.ReportVersion, error) {
		payload.Version = versionNumber
		payload.PublishedAt = time.Now().UTC()
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		pdfBytes, err := report.RenderPDF(payload)
		if err != nil {
			return nil, err
		}
		return &models.ReportVersion{
			Data:        data,
			PDF:         pdfBytes,
			PublishedAt: payload.PublishedAt,
			PublishedBy: email.(string),
		}, nil
	})
	if err != nil {
		fmt.Printf("PublishReportVersion failed: %v\n", err)
		notFound(c)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"id":          version.ID,
		"version":     version.VersionNumber,
		"publishedAt": version.PublishedAt,
		"publishedBy": version.PublishedBy,
	})
}

// buildSnapshot freezes the current draft into a payload. Vulnerabilities are
// resolved by ID and copied by value — later edits to the underlying vuln
// won't change a published report. FindingOverrides win over the raw vuln
// values when present.
func buildSnapshot(r *models.Report, _ string) (models.ReportVersionPayload, error) {
	payload := models.ReportVersionPayload{
		Title:            r.Title,
		ExecutiveSummary: r.ExecutiveSummary,
	}

	projectNames := make(map[uint]string, len(r.ProjectIDsList))
	for _, pid := range r.ProjectIDsList {
		proj, err := database.GetProject(pid)
		if err != nil {
			return payload, fmt.Errorf("load project %d: %w", pid, err)
		}
		projectNames[pid] = proj.ProjectName
		payload.Projects = append(payload.Projects, models.ReportSnapshotProject{
			ID:   pid,
			Name: proj.ProjectName,
		})
	}

	for _, vid := range r.VulnerabilityIDsList {
		vuln, err := database.GetJSONData(vid)
		if err != nil {
			// Skip vulnerabilities that disappeared; don't abort the publish.
			fmt.Printf("publish: skipping missing vulnerability %d: %v\n", vid, err)
			continue
		}
		var parsed map[string]interface{}
		if len(vuln.Vulnerability) > 0 {
			if err := json.Unmarshal(vuln.Vulnerability, &parsed); err != nil {
				fmt.Printf("publish: bad vulnerability JSON %d: %v\n", vid, err)
				continue
			}
		}
		finding := models.ReportSnapshotFinding{
			ID:            vid,
			Title:         stringField(parsed, "title"),
			Severity:      stringField(parsed, "severity"),
			Summary:       stringField(parsed, "description"),
			Status:        vuln.Status,
			Vulnerability: parsed,
		}
		if vuln.ProjectID != nil {
			finding.ProjectID = *vuln.ProjectID
			if name, ok := projectNames[*vuln.ProjectID]; ok {
				finding.ProjectName = name
			}
		}
		if ov, ok := r.FindingOverridesData[vid]; ok {
			if strings.TrimSpace(ov.Title) != "" {
				finding.Title = ov.Title
			}
			if strings.TrimSpace(ov.Summary) != "" {
				finding.Summary = ov.Summary
			}
		}
		payload.Findings = append(payload.Findings, finding)
	}
	return payload, nil
}

func stringField(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func ListReportVersions(c *gin.Context) {
	r := loadReportForRead(c)
	if r == nil {
		return
	}
	versions, err := database.ListReportVersions(r.ID)
	if err != nil {
		fmt.Printf("ListReportVersions failed: %v\n", err)
		notFound(c)
		return
	}
	c.JSON(http.StatusOK, versions)
}

func GetReportVersion(c *gin.Context) {
	r := loadReportForRead(c)
	if r == nil {
		return
	}
	n, err := strconv.Atoi(c.Param("version"))
	if err != nil || n < 1 {
		notFound(c)
		return
	}
	v, err := database.GetReportVersion(r.ID, n)
	if err != nil {
		notFound(c)
		return
	}
	c.JSON(http.StatusOK, v)
}

func GetReportVersionPDF(c *gin.Context) {
	r := loadReportForRead(c)
	if r == nil {
		return
	}
	n, err := strconv.Atoi(c.Param("version"))
	if err != nil || n < 1 {
		notFound(c)
		return
	}
	v, err := database.GetReportVersion(r.ID, n)
	if err != nil {
		notFound(c)
		return
	}
	servePDF(c, r, v)
}

func servePDF(c *gin.Context, r *models.Report, v *models.ReportVersion) {
	filename := fmt.Sprintf("%s-v%d.pdf", slugify(r.Title), v.VersionNumber)
	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/pdf", v.PDF)
}

func slugify(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			out = append(out, r)
		case r == ' ', r == '-', r == '_':
			out = append(out, '-')
		}
	}
	if len(out) == 0 {
		return "report"
	}
	return string(out)
}
