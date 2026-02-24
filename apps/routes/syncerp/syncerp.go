package syncerp

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/syncbyerp"
	"github.com/innovativecursor/ixense_kloud_sync_service/apps/routes/getapiroutes"
	"gorm.io/gorm"
)

func SyncERP(db *gorm.DB) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "10004"
		log.Printf("Defaulting to port %s", port)
	}

	apiV1, router := getapiroutes.GetApiRoutes()

	// Define handlers oauth
	apiV1.GET("/sync-erp", func(c *gin.Context) {
		c.String(http.StatusOK, "sync Service Healthy")
	})

	apiV1.GET("sync/get", func(c *gin.Context) {
		syncbyerp.SyncERPHandler(c, db)
	})

	apiV1.GET("sync/get-all-itemcodes", func(c *gin.Context) {
		syncbyerp.GetAllItemCodesHandler(c, db)
	})
	// Listen and serve on defined port
	log.Printf("Application started, Listening on Port %s", port)
	router.Run(":" + port)
}

//done
