package syncerp

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/deduct"
	"github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/maperpandkp"
	"github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/middleware"
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

	apiV1.GET("sync/get-all-data", func(c *gin.Context) {
		syncbyerp.GetAllERPSyncMedicinesHandler(c, db)
	})

	apiV1.GET("sync/get-all-itemcodes", middleware.InternalAuthMiddleware(), func(c *gin.Context) {
		syncbyerp.GetAllItemCodesHandler(c, db)
	})

	// apiV1.GET("sync/get-all-mapped-item", middleware.InternalAuthMiddleware(), func(c *gin.Context) {
	// 	maperpandkp.GetAllERPItemMappingsHandler(c, db)
	// })

	apiV1.PUT("sync/map-erp-kp-products", middleware.InternalAuthMiddleware(), func(c *gin.Context) {
		maperpandkp.UpdateERPMappingHandler(c, db)
	})

	apiV1.POST("sync/recovery", middleware.InternalAuthMiddleware(), func(c *gin.Context) {
		maperpandkp.RecoverySyncHandler(c, db)
	})

	apiV1.POST("erp/webhook/product", middleware.InternalAuthMiddleware(), func(c *gin.Context) {
		syncbyerp.ERPProductWebhookHandler(c, db)
	})
	apiV1.POST("erp/webhook/stock", middleware.InternalAuthMiddleware(), func(c *gin.Context) {
		syncbyerp.ERPStockWebhookHandler(c, db)
	})

	apiV1.POST("sync/deduct-stock", middleware.InternalAuthMiddleware(), func(c *gin.Context) {
		deduct.DeductStockHandler(c, db)
	})
	// Listen and serve on defined port
	log.Printf("Application started, Listening on Port %s", port)
	router.Run(":" + port)
}

//done
