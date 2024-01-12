package main

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "prism/auth"
    "prism/config"
    "prism/routes"
    "prism/database"
    "prism/event"

    "time"
    "fmt"
)

func main() {
    database.InitDB()
    // Set up the primary Gin router for the main application
    r := gin.Default()
    r.Use(CORSMiddleware())

    r.GET("/api/login", auth.HandleLogin)
    r.GET("/api/callback", auth.HandleCallback)

    // Authenticated users
    r.Use(auth.AuthMiddleware())

    apiRoutes := r.Group("/api")
    {
        apiRoutes.GET("/user", auth.HandleUserRequest)
        apiRoutes.GET("/userinfo/:email", routes.GetUserInfo)
        apiRoutes.GET("/logout", auth.HandleLogout)
        apiRoutes.GET("/dashboard", routes.HandleDashboard)

        // ACL Routes
        apiRoutes.GET("/blob/:filename", routes.GetBlob)
        apiRoutes.GET("/project/:projectID", routes.GetProject)
        apiRoutes.GET("/vulnerabilityfinding/:id", routes.GetVulnerability)
        apiRoutes.GET("/project/:projectID/vulnerabilities/total", routes.GetProjectVulnerabilitiesTotal)
        apiRoutes.GET("/project/:projectID/vulnerabilities", routes.GetProjectVulnerabilitiesForProject)

        // Admin users
        apiRoutes.Use(auth.AdminMiddleware())

        apiRoutes.GET("/user/all", routes.GetAllUsers)
        apiRoutes.DELETE("/vulnerability/:id", routes.DeleteVulnerability)
        apiRoutes.DELETE("/project/:projectID", routes.DeleteProject)
        apiRoutes.PUT("/project/:projectID", routes.HandleProjectPut)
        apiRoutes.GET("/project", routes.GetProjects)
        apiRoutes.POST("/project", routes.HandleProjectPost)
        apiRoutes.POST("/blob/upload", routes.HandleBlobUpload)
        apiRoutes.DELETE("/blob/:filename", routes.HandleBlobDelete)

        apiRoutes.POST("/vulnerabilityfinding", routes.PostVulnerability )
        apiRoutes.PUT("/vulnerabilityfinding/:id", routes.PutVulnerability)
        apiRoutes.PUT("/vulnerabilityfinding/:id/status/:status", routes.ChangeStatusVulnerability)
        apiRoutes.GET("/vulnerabilityfinding/all", routes.GetAllVulnerabilities)

        apiRoutes.GET("/settings/events", event.EventQueues)
        apiRoutes.PUT("/settings/events/:id/update/:status", event.UpdateEventQueues)
        apiRoutes.GET("/settings/export", routes.ExportAllData)
        apiRoutes.POST("/settings/import", routes.ImportData)
    }

    go event.PollEventQueue()

    // Run the main application in a separate goroutine
    go func() {
        if err := r.Run(":8080"); err != nil && err != http.ErrServerClosed {
            fmt.Println("Main server failed to start: %v", err)
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
    appConfig, err := config.LoadConfig()
    if err != nil { }
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", appConfig.Cors.Origin)
        c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
        c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Cookie")
        c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
        c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

        if c.Request.Method == "OPTIONS" {
            c.AbortWithStatus(http.StatusNoContent)
            return
        }

        c.Next()
    }
}
