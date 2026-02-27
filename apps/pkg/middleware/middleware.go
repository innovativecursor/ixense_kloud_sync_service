package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/config"
)

func InternalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		cfg, err := config.Env()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"status":  false,
				"message": "Failed to load config",
			})
			return
		}

		clientKey := c.GetHeader("X-INTERNAL-KEY")

		if clientKey == "" || clientKey != cfg.Internal.ServiceKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"status":  false,
				"message": "Unauthorized - Invalid Service Key",
			})
			return
		}

		c.Next()
	}
}
