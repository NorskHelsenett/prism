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
	"prism/routes"
	"prism/session"

	"fmt"
	"net/http"
	"time"
)

var session_db *gorm.DB // Global variable for the database

func main() {
	if config.AppConfig == nil {
		log.Fatalf("Configuration is loaded from auth.OIDC.init()")
	}
	database.InitDB()
	initSessionDatabase()
	sessionStore := session.NewSessionStore(session_db)
	session.LoadSessionStore(sessionStore)
	// Set up the primary Gin router for the main application
	r := gin.Default()
	r.Use(CORSMiddleware())
	r.Use(audit.AuditMiddleware())

	r.GET("/api/login", auth.HandleLogin)
	r.GET("/api/callback", func(c *gin.Context) {
		auth.HandleCallback(c, sessionStore)
	})

	// Authenticated users
	r.Use(auth.AuthMiddleware(sessionStore))

	apiRoutes := r.Group("/api")
	{
		apiRoutes.GET("/profile", func(c *gin.Context) { auth.HandleUserRequest(c, sessionStore) })
		apiRoutes.GET("/profile/session/all", func(c *gin.Context) { session.GetUserSessions(c, sessionStore) })
		apiRoutes.DELETE("/profile/session/:uuid", func(c *gin.Context) { session.DeleteUserSession(c, sessionStore) })
		apiRoutes.GET("/profile/:email", routes.GetUserInfo) //handle when a user logs in, store in both databases
		apiRoutes.GET("/logout", func(c *gin.Context) { auth.HandleLogout(c, sessionStore) })
		apiRoutes.GET("/dashboard", routes.HandleDashboard)

		apiRoutes.GET("/session/otp/generate", session.HandleOTPGenerate)
		apiRoutes.POST("/session/otp/validate", func(c *gin.Context) { session.HandleOTPValidate(c, sessionStore) })
		apiRoutes.GET("/session/otp/reset", func(c *gin.Context) { session.HandleOTPReset(c, sessionStore) })

		apiRoutes.Use(auth.RBACMiddleware())
		apiRoutes.GET("/profile/access-list", routes.GetAccessListRoutes)

		apiRoutes.POST("/planning/new", routes.NewAssassment)
		apiRoutes.GET("/planning", routes.RetrieveAssessmentsHandler)
		apiRoutes.GET("/planning/:id", routes.RetrieveAssessmentsHandler)
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

			apiRoutes.GET("/project/all", routes.GetProjects)
			apiRoutes.GET("/vulnerability/all", routes.GetAllVulnerabilities)
		}

		apiRoutes.GET("/profile/all", func(c *gin.Context) { routes.GetAllProfilesEmailOnly(c, sessionStore) })
		apiRoutes.GET("/slack/channels", routes.GetSlackChannels)

		apiRoutes.GET("/blob/:filename", routes.GetBlob)
		apiRoutes.POST("/blob/upload", routes.HandleBlobUpload)
		apiRoutes.DELETE("/blob/:filename", routes.HandleBlobDelete)

		apiRoutes.POST("/project", routes.HandleProjectPost)
		apiRoutes.PUT("/project/:projectID", routes.HandleProjectPut)
		apiRoutes.DELETE("/project/:projectID", routes.DeleteProject)

		apiRoutes.POST("/vulnerability", routes.PostVulnerability)
		apiRoutes.PUT("/vulnerability/:id", routes.PutVulnerability)
		apiRoutes.DELETE("/vulnerability/:id", routes.DeleteVulnerability)
		apiRoutes.POST("/vulnerability/:id/comment", routes.NewComment)
		apiRoutes.PUT("/vulnerability/:id/comment", routes.UpdateComment)
		apiRoutes.DELETE("/vulnerability/:id/comment/:cid", routes.DeleteComment)
		apiRoutes.PUT("/vulnerability/:id/status/:status", routes.ChangeStatusVulnerability)

		apiRoutes.GET("/settings/users/all", func(c *gin.Context) { routes.GetAllUsers(c, sessionStore) })
		apiRoutes.PUT("/settings/profile", func(c *gin.Context) { routes.UpdateUser(c, sessionStore) })
		apiRoutes.GET("/settings/events", event.EventQueues)
		apiRoutes.GET("/settings", routes.GetSettings)
		apiRoutes.POST("/settings", routes.PostSettings)
		apiRoutes.GET("/settings/roles-list", routes.GetAllRoles)
		apiRoutes.GET("/settings/cleanup", routes.CleanUpDatabase)
		apiRoutes.PUT("/settings/events/:id/update/:status", event.UpdateEventQueues)
		apiRoutes.DELETE("/settings/events/:id", event.DeleteEventQueue)
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
		fmt.Println("Health check server failed to start: %v", err)
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
