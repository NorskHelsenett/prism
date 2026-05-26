package share

import (
	"fmt"
	"net/http"
	"prism/auth"
	"prism/database"
	"prism/models"
	"prism/session"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetPublicReport resolves a short token to the report's latest published
// version. Access is invite-by-email: the visitor must hold a valid session
// AND their email must appear in Report.InvitedEmailsList OR they must have
// read access to one of the report's projects (project-derived read for
// internal viewers).
//
// All denials respond with 404 / 401 — never 500 — so the response shape
// reveals nothing about which condition failed.
func GetPublicReport(c *gin.Context, store *session.SessionStore) {
	r, version, ok := resolvePublicReport(c, store)
	if !ok {
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"reportId":    r.ID,
		"title":       r.Title,
		"version":     version.VersionNumber,
		"publishedAt": version.PublishedAt,
		"publishedBy": version.PublishedBy,
		"data":        version.Data,
	})
}

// GetPublicReportPDF streams the stored PDF bytes for the latest version.
func GetPublicReportPDF(c *gin.Context, store *session.SessionStore) {
	r, version, ok := resolvePublicReport(c, store)
	if !ok {
		return
	}
	filename := fmt.Sprintf("%s-v%d.pdf", strings.ReplaceAll(strings.ToLower(r.Title), " ", "-"), version.VersionNumber)
	c.Header("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Data(http.StatusOK, "application/pdf", version.PDF)
}

func resolvePublicReport(c *gin.Context, store *session.SessionStore) (*models.Report, *models.ReportVersion, bool) {
	token := c.Param("token")
	if token == "" {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
		return nil, nil, false
	}
	r, err := database.GetReportByToken(token)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
		return nil, nil, false
	}
	if r.LatestPublishedVersionID == nil {
		// Report exists but nothing published yet — same 404 to avoid leaking.
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
		return nil, nil, false
	}

	userInfo, err := auth.GetSignedCookie(c, "session_cookie")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return nil, nil, false
	}
	validation, err := store.ValidateSession(userInfo.Email, userInfo.SessionID)
	if err != nil || !validation.IsValid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return nil, nil, false
	}

	viewerEmail := strings.TrimSpace(strings.ToLower(userInfo.Email))
	viewerRole := lookupRole(userInfo.Email)
	if !emailAllowed(viewerEmail, r) {
		if !hasProjectRead(userInfo.Email, viewerRole, r) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return nil, nil, false
		}
	}

	version, err := database.GetReportVersionByID(*r.LatestPublishedVersionID)
	if err != nil {
		fmt.Printf("GetReportVersionByID(%d) failed: %v\n", *r.LatestPublishedVersionID, err)
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "not found"})
		return nil, nil, false
	}
	return r, version, true
}

func emailAllowed(viewer string, r *models.Report) bool {
	for _, e := range r.InvitedEmailsList {
		if strings.TrimSpace(strings.ToLower(e)) == viewer {
			return true
		}
	}
	return false
}

// lookupRole resolves a session email to the stored UserData.Role used by
// the project ACL helpers. Failures default to "visitor" — the most
// restrictive named role — so an unknown user can't pick up elevated
// rights via the lookup path.
func lookupRole(email string) string {
	u, err := database.GetUserDataByEmail(email)
	if err != nil || u == nil {
		return "visitor"
	}
	return u.Role
}

// hasProjectRead lets any internal viewer who can read at least one of the
// report's projects see the disclosed report — matching the user's brief
// ("readers of that project should be able to view this as well when it's
// published"). Falls back to false on any error rather than blocking access
// noisily.
func hasProjectRead(email, role string, r *models.Report) bool {
	if role == "admin" {
		return true
	}
	ok, err := database.HasReadOnAnyProject(email, role, r.ProjectIDsList, false)
	if err != nil {
		fmt.Printf("hasProjectRead: %v\n", err)
		return false
	}
	return ok
}
