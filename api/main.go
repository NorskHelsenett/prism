package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prism/audit"
	"prism/auth"
	"prism/config"
	"prism/database"
	"prism/event"
	"prism/middleware"
	"prism/routes"
	"prism/session"
	"prism/share"
	"prism/webauthn"

	"fmt"
	"net/http"
	"time"
)

var session_db *gorm.DB // Global variable for the database

func main() {
	config.LoadConfig()

	if config.AppConfig == nil {
		log.Fatalf("Configuration is loaded from auth.OIDC.init()")
	}
	database.InitDB()

	// Convert any legacy /api/blob/<id> URLs and inline data: URIs in
	// vulnerability evidence/remediation into per-vuln attachment rows.
	// Idempotent: a no-op once everything is on the new scheme.
	//
	// Run in a goroutine AFTER InitDB so the HTTP and health servers can
	// start immediately. A large first-time migration can take a while
	// (it re-encodes every legacy image), and if we ran it synchronously
	// the kubelet's liveness/readiness probes would fail before any
	// listener was up and recycle the pod mid-migration.
	go func() {
		log.Printf("attachment migration: starting in background")
		report, err := database.MigrateAllAttachments(false)
		if err != nil {
			log.Printf("attachment migration: %v", err)
			return
		}
		if report.VulnerabilitiesChanged > 0 || report.LegacyBlobsConverted > 0 || report.DataURIsConverted > 0 {
			log.Printf("attachment migration: %s", report.String())
		} else {
			log.Printf("attachment migration: no work to do")
		}
	}()

	// Move legacy per-user notification JSON blobs into the new notifications
	// table, backfill typed columns on the subscribers table, dedupe shared
	// endpoints across users (the cross-user bug, expressed as data), and
	// install the UNIQUE endpoint index. Idempotent.
	database.RunNotificationMigrationOnce()

	initSessionDatabase()
	sessionStore := session.NewSessionStore(session_db)
	session.LoadSessionStore(sessionStore)
	webauthn.Init()
	routes.InitNotification()
	// Set up the primary Gin router for the main application
	r := gin.Default()

	r.GET("/.well-known/config.json", routes.HandleClientConfig)

	// Serve static files from web/build directory
	r.Static("/_app", "./web/build/_app")
	r.Static("/assets", "./web/build/assets")
	r.Static("/img", "./web/build/img")
	r.StaticFile("/favicon.png", "./web/build/favicon.png")
	r.StaticFile("/robots.txt", "./web/build/robots.txt")
	r.StaticFile("/service-worker.js", "./web/build/service-worker.js")

	// Serve specific .well-known files (config.json is served dynamically via API)
	r.StaticFile("/.well-known/CHANGELOG.md", "./web/build/.well-known/CHANGELOG.md")
	r.StaticFile("/.well-known/buildinfo.json", "./web/build/.well-known/buildinfo.json")

	// SPA fallback - serve index.html for all non-API routes
	r.NoRoute(func(c *gin.Context) {
		// Don't serve index.html for API routes
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}
		c.File("./web/build/index.html")
	})

	r.Use(CORSMiddleware())
	r.Use(audit.AuditMiddleware())

	// Public API endpoints (no auth required)
	r.GET("/api/login", auth.HandleLogin)
	r.GET("/api/callback", func(c *gin.Context) {
		auth.HandleCallback(c, sessionStore)
	})
	r.GET("/api/logout", func(c *gin.Context) { auth.HandleLogout(c, sessionStore) })

	shareGroup := r.Group("/api")
	{
		shareGroup.Use(middleware.RateLimiter())
		shareGroup.POST("/share/:token", func(c *gin.Context) { share.GetPublicVulnerability(c, sessionStore) })
	}

	// Authenticated users (API only)
	apiRoutes := r.Group("/api")
	apiRoutes.Use(auth.AuthMiddleware(sessionStore))

	{
		apiRoutes.GET("/notification/publicKey", routes.GetNotificationPublicKey)
		apiRoutes.GET("/notification", routes.GetNotificationsHandler)
		apiRoutes.GET("/notification/events", routes.StreamNotificationsHandler)
		apiRoutes.DELETE("/notification", routes.DeleteNotificationsHandler)
		apiRoutes.PUT("/notification/read-all", routes.MarkAllReadHandler)
		apiRoutes.PUT("/notification/:id/read", routes.MarkNotificationReadHandler)
		apiRoutes.POST("/notification/subscribe", routes.SubscribeNotification)
		apiRoutes.DELETE("/notification/subscribe", routes.UnsubscribeNotification)
		apiRoutes.GET("/notification/devices", routes.ListDevices)

		apiRoutes.GET("/dashboard", routes.HandleDashboard)

		//-- RBAC MIDDLEWARE FROM HERE ON --//
		apiRoutes.Use(auth.RBACMiddleware())

		apiRoutes.GET("/profile", func(c *gin.Context) { auth.HandleUserRequest(c, sessionStore) })
		apiRoutes.GET("/profile/:email", routes.GetUserInfo) //handle when a user logs in, store in both databases

		apiRoutes.POST("/profile/apikey", routes.CreateAPIKey)
		apiRoutes.GET("/profile/apikey", routes.GetAPIKey)
		apiRoutes.PATCH("/profile/preferences", routes.UpdateUserPreferences)
		apiRoutes.GET("/profile/preferences", routes.GetUserPreferences)
		apiRoutes.DELETE("/profile/apikey/:id", routes.DeleteAPIKey)

		apiRoutes.GET("/session/otp/generate", session.HandleOTPGenerate)
		apiRoutes.PATCH("/session/otp/generate", middleware.RateLimiter(), session.HandleOTPGenerateConfirm)
		apiRoutes.POST("/session/otp/validate", middleware.RateLimiter(), func(c *gin.Context) { session.HandleOTPValidate(c, sessionStore) })
		apiRoutes.GET("/session/otp/reset", func(c *gin.Context) { session.HandleOTPReset(c, sessionStore) })
		apiRoutes.GET("/session/all", func(c *gin.Context) { session.GetUserSessions(c, sessionStore) })
		apiRoutes.DELETE("/session/:uuid", func(c *gin.Context) { session.DeleteUserSession(c, sessionStore) })

		// Passkey (WebAuthn) 2FA endpoints
		apiRoutes.POST("/session/passkey/register/begin", webauthn.BeginRegistration)
		apiRoutes.POST("/session/passkey/register/finish", webauthn.FinishRegistration)
		apiRoutes.POST("/session/passkey/begin", middleware.RateLimiter(), webauthn.BeginAuthentication)
		apiRoutes.POST("/session/passkey/finish", middleware.RateLimiter(), func(c *gin.Context) { webauthn.FinishAuthentication(c, sessionStore) })
		apiRoutes.GET("/session/passkey/credentials", webauthn.GetCredentials)
		apiRoutes.DELETE("/session/passkey/credentials/:id", webauthn.DeleteCredential)
		apiRoutes.GET("/session/passkey/has", webauthn.HasPasskeys)

		apiRoutes.GET("/profile/access-list", routes.GetAccessListRoutes)

		apiRoutes.POST("/planning/new", routes.NewAssassment)
		apiRoutes.GET("/planning", routes.RetrieveAssessmentsHandler)
		apiRoutes.GET("/planning/:id", routes.RetrieveAssessmentsHandler)
		apiRoutes.PATCH("/planning/:id", routes.PatchAssessmentsHandler)
		apiRoutes.GET("/planning/:id/assignedHackers", routes.FindNonAvailablePersons)
		apiRoutes.PUT("/planning/:id", routes.PutAssessmentsHandler)
		apiRoutes.DELETE("/planning/:id", routes.DeleteAssessmentsHandler)

		// Group with ACL middleware
		protectedRoutes := apiRoutes.Group("/")
		{
			protectedRoutes.Use(auth.ACLMiddleware())

			protectedRoutes.GET("/project/:projectID", routes.GetProject)
			protectedRoutes.GET("/project/:projectID/vulnerabilities/total", routes.GetProjectVulnerabilitiesTotal)
			protectedRoutes.GET("/project/:projectID/vulnerabilities", routes.GetProjectVulnerabilitiesForProject)
			protectedRoutes.GET("/vulnerability/:findingsID", routes.GetVulnerability)

			protectedRoutes.PUT("/project/:projectID", routes.HandleProjectPut)
			protectedRoutes.DELETE("/project/:projectID", routes.DeleteProject)

			protectedRoutes.PUT("/vulnerability/:findingsID", routes.PutVulnerability)
			protectedRoutes.PATCH("/vulnerability/:findingsID", routes.PatchVulnerability)
			protectedRoutes.DELETE("/vulnerability/:findingsID", routes.DeleteVulnerability)

			protectedRoutes.POST("/vulnerability/:findingsID/comment", routes.NewComment)
			protectedRoutes.PUT("/vulnerability/:findingsID/comment", routes.UpdateComment)
			protectedRoutes.DELETE("/vulnerability/:findingsID/comment/:cid", routes.DeleteComment)
			protectedRoutes.PUT("/vulnerability/:findingsID/status/:status", routes.ChangeStatusVulnerability)

			protectedRoutes.GET("/vulnerability/:findingsID/attachments", routes.ListAttachmentsHandler)
			protectedRoutes.HEAD("/vulnerability/:findingsID/attachments", routes.ListAttachmentsHandler)
			protectedRoutes.POST("/vulnerability/:findingsID/attachments", routes.PostAttachment)
			protectedRoutes.GET("/vulnerability/:findingsID/attachments/:key", routes.GetAttachmentProxy)
			protectedRoutes.HEAD("/vulnerability/:findingsID/attachments/:key", routes.GetAttachmentProxy)
			protectedRoutes.GET("/vulnerability/:findingsID/attachments/:key/original", routes.GetAttachmentOriginal)
			protectedRoutes.HEAD("/vulnerability/:findingsID/attachments/:key/original", routes.GetAttachmentOriginal)
			protectedRoutes.DELETE("/vulnerability/:findingsID/attachments/:key", routes.DeleteAttachmentHandler)

			shareRoutes := apiRoutes.Group("/")
			{

				shareRoutes.Use(auth.ShareMiddleware())

				shareRoutes.GET("/vulnerability/share/all", share.GetAll)
				shareRoutes.GET("/vulnerability/share/:findingsID", share.GetShareVulnerability)
				shareRoutes.POST("/vulnerability/share/:findingsID", share.ShareVulnerability)
				shareRoutes.DELETE("/vulnerability/share/:findingsID", share.DeleteShareVulnerability)
				shareRoutes.PUT("/vulnerability/share/:findingsID", share.ShareVulnerability)
			}
		}

		apiRoutes.POST("/project", routes.HandleProjectPost)
		apiRoutes.GET("/project/all", routes.GetProjects)

		apiRoutes.GET("/project-group", routes.GetProjectGroups)
		apiRoutes.POST("/project-group", routes.PostProjectGroup)
		apiRoutes.PUT("/project-group/:groupID", routes.PutProjectGroup)
		apiRoutes.DELETE("/project-group/:groupID", routes.DeleteProjectGroupHandler)
		apiRoutes.PATCH("/project/:projectID/group", routes.PatchProjectGroupAssignment)
		apiRoutes.GET("/vulnerability/search", routes.SearchVulnerabilities)

		apiRoutes.GET("/profile/all", routes.GetAllProfilesEmailOnly)
		apiRoutes.GET("/slack/channels", routes.GetSlackChannels)

		apiRoutes.GET("/notes", routes.ListNotes)
		apiRoutes.GET("/notes/tags", routes.ListNoteTags)
		apiRoutes.GET("/notes/events", routes.StreamNotes)
		apiRoutes.POST("/notes", routes.CreateNote)
		apiRoutes.GET("/notes/:id", routes.GetNote)
		apiRoutes.PATCH("/notes/:id", routes.UpdateNote)
		apiRoutes.DELETE("/notes/:id", routes.DeleteNote)
		apiRoutes.POST("/notes/:id/restore", routes.RestoreNote)
		apiRoutes.DELETE("/notes/:id/permanent", routes.HardDeleteNote)
		apiRoutes.DELETE("/notes/trash", routes.EmptyNoteTrash)

		apiRoutes.POST("/vulnerability", routes.PostVulnerability)

		// Server-side vulnerability drafts (owner-scoped; no ACL middleware
		// needed — every handler checks the session email itself).
		apiRoutes.POST("/drafts", routes.CreateDraftHandler)
		apiRoutes.GET("/drafts", routes.ListDraftsHandler)
		apiRoutes.GET("/drafts/:draftID", routes.GetDraftHandler)
		apiRoutes.PUT("/drafts/:draftID", routes.UpdateDraftHandler)
		apiRoutes.DELETE("/drafts/:draftID", routes.DeleteDraftHandler)
		apiRoutes.POST("/drafts/:draftID/publish", routes.PublishDraftHandler)
		apiRoutes.GET("/drafts/:draftID/attachments", routes.ListDraftAttachmentsHandler)
		apiRoutes.POST("/drafts/:draftID/attachments", routes.PostDraftAttachment)
		apiRoutes.GET("/drafts/:draftID/attachments/:key", routes.GetDraftAttachmentProxy)
		apiRoutes.HEAD("/drafts/:draftID/attachments/:key", routes.GetDraftAttachmentProxy)
		apiRoutes.DELETE("/drafts/:draftID/attachments/:key", routes.DeleteDraftAttachmentHandler)

		apiRoutes.GET("/settings/teams", routes.GetTeams)
		apiRoutes.GET("/settings/teams/:id", routes.GetTeam)
		apiRoutes.POST("/settings/teams", routes.PostTeam)
		apiRoutes.PUT("/settings/teams/:id", routes.UpdateTeam)
		apiRoutes.DELETE("/settings/teams/:id", routes.DeleteTeam)
		apiRoutes.POST("/settings/teams/:id/archive", routes.ArchiveTeam)
		apiRoutes.POST("/settings/teams/:id/members", routes.AddMemberToTeam)
		apiRoutes.DELETE("/settings/teams/:id/members", routes.RemoveMemberFromTeam)
		apiRoutes.GET("/profile/memberof", routes.GetUserTeams)

		apiRoutes.GET("/settings/users/all", func(c *gin.Context) { routes.GetAllUsers(c, sessionStore) })
		apiRoutes.DELETE("/settings/user/:id", func(c *gin.Context) { routes.DeleteUser(c, sessionStore) })
		apiRoutes.PATCH("/settings/user/:id/active", func(c *gin.Context) { routes.ToggleUserActive(c, sessionStore) })
		apiRoutes.PUT("/settings/profile", func(c *gin.Context) { routes.UpdateUserRole(c, sessionStore) })
		apiRoutes.GET("/settings/events", event.EventQueues)
		apiRoutes.GET("/settings", routes.GetSettings)
		apiRoutes.POST("/settings", routes.PostSettings)
		apiRoutes.POST("/settings/attachments/regenerate-proxies", routes.RegenerateAttachmentProxies)
		apiRoutes.GET("/settings/roles-list", routes.GetAllRoles)
		apiRoutes.GET("/settings/cleanup", routes.CleanUpDatabase)
		apiRoutes.PUT("/settings/events/:id/update/:status", event.UpdateEventQueues)
		apiRoutes.DELETE("/settings/events/:id", event.DeleteEventQueue)
		apiRoutes.DELETE("/settings/notification", routes.ResetNotifications)
		apiRoutes.GET("/settings/export", routes.ExportAllData)
		apiRoutes.GET("/settings/audit", audit.GetAllAudits)
		apiRoutes.POST("/settings/import", routes.ImportData)
		apiRoutes.DELETE("/settings/session/otp/reset/:email", func(c *gin.Context) { session.HandleOTPResetForUser(c, sessionStore) })
		apiRoutes.DELETE("/settings/session/passkey/reset/:email", webauthn.ResetPasskeysForUser)
		apiRoutes.GET("/settings/session/mfa-status/:email", webauthn.GetMFAStatusForUser)
	}

	go event.PollEventQueue()

	// Run the main application in a separate goroutine
	go func() {
		if err := r.Run(":8080"); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Main server failed to start: %v\n", err)
		}
	}()

	// Set up a second server for health checks
	health := gin.Default()
	health.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Run the health check server on a different port
	healthServer := &http.Server{
		Addr:         ":8888",
		Handler:      health,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Printf("Health check server failed to start: %v", err)
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Origin", config.AppConfig.Cors.Origin)
		c.Writer.Header().Set("Access-Control-Allow-Origin", config.AppConfig.Cors.Origin)
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, Cookie")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")

		if os.Getenv("GO_ENV") == "dev" {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func initSessionDatabase() {
	var err error // Declare err separately
	session_db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to the database")
	}

	// Perform database migrations
	session_db.AutoMigrate(&session.Session{})
	session_db.AutoMigrate(&database.UserData{})
}
