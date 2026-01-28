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
	"prism/ws"

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
	initSessionDatabase()
	sessionStore := session.NewSessionStore(session_db)
	session.LoadSessionStore(sessionStore)
	routes.InitNotification()
	// Set up the primary Gin router for the main application
	r := gin.Default()

	r.GET("/ws", ws.WSHandler)

	// Serve static files from web/build directory
	r.Static("/_app", "./web/build/_app")
	r.Static("/assets", "./web/build/assets")
	r.Static("/img", "./web/build/img")
	r.StaticFile("/favicon.png", "./web/build/favicon.png")
	r.StaticFile("/robots.txt", "./web/build/robots.txt")
	r.StaticFile("/service-worker.js", "./web/build/service-worker.js")

	// Serve .well-known directory for build info
	r.Static("/.well-known", "./web/build/.well-known")

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

	r.GET("/api/login", auth.HandleLogin)
	r.GET("/api/callback", func(c *gin.Context) {
		auth.HandleCallback(c, sessionStore)
	})

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
		apiRoutes.DELETE("/notification", routes.DeleteNotificationsHandler)
		apiRoutes.PUT("/notification/:time/read", routes.MarkNotificationReadHandler)
		apiRoutes.POST("/notification/subscribe", routes.SubscribeNotification)

		apiRoutes.GET("/logout", func(c *gin.Context) { auth.HandleLogout(c, sessionStore) })
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
		apiRoutes.GET("/vulnerability/all", routes.GetAllVulnerabilities)

		apiRoutes.GET("/profile/all", routes.GetAllProfilesEmailOnly)
		apiRoutes.GET("/slack/channels", routes.GetSlackChannels)

		apiRoutes.GET("/blob/:filename", routes.GetBlob)
		apiRoutes.POST("/blob/upload", routes.HandleBlobUpload)
		apiRoutes.DELETE("/blob/:filename", routes.HandleBlobDelete)

		apiRoutes.POST("/vulnerability", routes.PostVulnerability)

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
		apiRoutes.PUT("/settings/profile", func(c *gin.Context) { routes.UpdateUserRole(c, sessionStore) })
		apiRoutes.GET("/settings/events", event.EventQueues)
		apiRoutes.GET("/settings", routes.GetSettings)
		apiRoutes.POST("/settings", routes.PostSettings)
		apiRoutes.GET("/settings/roles-list", routes.GetAllRoles)
		apiRoutes.GET("/settings/cleanup", routes.CleanUpDatabase)
		apiRoutes.PUT("/settings/events/:id/update/:status", event.UpdateEventQueues)
		apiRoutes.DELETE("/settings/events/:id", event.DeleteEventQueue)
		apiRoutes.DELETE("/settings/notification", routes.ResetNotifications)
		apiRoutes.GET("/settings/export", routes.ExportAllData)
		apiRoutes.GET("/settings/audit", audit.GetAllAudits)
		apiRoutes.POST("/settings/import", routes.ImportData)
		apiRoutes.DELETE("/settings/session/otp/reset/:email", func(c *gin.Context) { session.HandleOTPResetForUser(c, sessionStore) })
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
