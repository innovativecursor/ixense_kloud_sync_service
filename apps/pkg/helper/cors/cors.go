package cors

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/config"
)

// Helper function to use coors
// func CORSMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {

// 		origin := c.Request.Header.Get("Origin")

// 		// Allow only specific origins
// 		if origin == "http://localhost:3004" || origin == "https://admin-kloudpx.innovativecursor.com" {
// 			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
// 		}

// 		c.Writer.Header().Set("Content-Type", "application/json")
// 		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
// 		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
// 		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-CSRF-Token")

// 		// Handle preflight
// 		if c.Request.Method == "OPTIONS" {
// 			c.AbortWithStatus(204)
// 			return
// 		}

// 		c.Next()
// 	}
// }

func CORSMiddleware() gin.HandlerFunc {
	cfg, err := config.Env()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	allowedOrigins := make(map[string]bool)
	for _, o := range cfg.Endpoints {
		allowedOrigins[o] = true
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
