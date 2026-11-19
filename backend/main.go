package main

import (
	"log"
	"net/http"
	"os"
	"time"

	kratosmiddleware "backend/internal/middleware"
	// "backend/internal/wire"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func main() {
	// Read config from environment variables
	kratosInternalURL := getEnv("KRATOS_INTERNAL_URL", "http://localhost:4433")
	kratosPublicURL := getEnv("KRATOS_PUBLIC_URL", "http://localhost:4433")
	frontendURL := getEnv("FRONTEND_URL", "http://localhost:3000")
	port := getEnv("BE_PORT", "8080")

	// Initialise a Gin Router
	router := gin.Default()

	// Configure router to use CORS
	// This is needed so our frontend can share Ory Kratos session cookies cross-origin
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Accept", "Authorization", "X-Session-Token"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           time.Hour,
	}))

	// Public routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Instantiate the auth middleware
	kratosAuth := kratosmiddleware.NewKratosMiddleware(kratosInternalURL, kratosPublicURL)

	protected := router.Group("/")
	protected.Use(kratosAuth.RequireAuth())
	{
		// /me is a mock route for testing, will replace later
		protected.GET("/me", func(c *gin.Context) {
			session, ok := kratosmiddleware.GetSession(c)
			if !ok {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "session not found in context"})
				return
			}
			identity, ok := session.GetIdentityOk()
			if !ok || identity == nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "identity not found in session"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"identity_id": identity.Id,
				"traits":      identity.Traits,
			})
		})

		// place all protected routes and router instantiations here
		// wire.WireEvent(queries, protected)
	}

	log.Printf("Server listening on http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
