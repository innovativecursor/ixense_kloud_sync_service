package syncbyerp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/config"
	"github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/maperpandkp"
	"github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/models"
	cf "github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/syncbyerp/config"
	"gorm.io/gorm"
)

func SyncERPProducts(db *gorm.DB) error {

	cfg, err := config.Env()
	if err != nil {
		return err
	}

	client := &http.Client{}

	page := 1
	for {
		url := fmt.Sprintf("%s/items-info?page=%d", cfg.ERP.BaseURL, page)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return err
		}

		req.Header.Set("ERP-API-KEY", cfg.ERP.APIKey)

		resp, err := client.Do(req)
		if err != nil {
			return err
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}

		var erpResponse cf.ERPItemsResponse
		if err := json.Unmarshal(body, &erpResponse); err != nil {
			return err
		}

		if !erpResponse.Status {
			return fmt.Errorf("ERP API returned false status")
		}

		//  Sync Products
		for _, product := range erpResponse.Products {

			normalizedStock := product.Stock
			if product.UnitOfMeasurement == "per box" && product.NumberOfPiecesPerBox > 0 {
				normalizedStock = product.Stock * product.NumberOfPiecesPerBox
			}

			var existing models.ERPSyncMedicine

			err := db.Where("item_code = ?", product.ItemCode).
				First(&existing).Error

			if err == gorm.ErrRecordNotFound {

				newItem := models.ERPSyncMedicine{
					ERPProductID:         product.ID,
					ItemCode:             product.ItemCode,
					BrandName:            product.BrandName,
					GenericName:          product.GenericName,
					Power:                product.Power,
					DosageForm:           product.DosageForm,
					Packaging:            product.Packaging,
					Description:          product.Description,
					Category:             product.Category,
					SubCategory:          product.SubCategory,
					DirectionsForUse:     product.DirectionsForUse,
					Manufacturer:         product.Manufacturer,
					Distributor:          product.Distributor,
					OriginCountry:        product.OriginCountry,
					ERPUnit:              product.UnitOfMeasurement,
					UnitOfMeasurement:    product.UnitOfMeasurement,
					Stock:                normalizedStock,
					NumberOfPiecesPerBox: product.NumberOfPiecesPerBox,
					SellingPricePerPiece: product.SellingPricePerPiece,
					CostPricePerBox:      product.CostPricePerBox,
					Discount:             product.Discount,
					VATClassification:    product.VATClassification,
					VAT:                  product.VAT,
					MinThreshold:         product.MinThreshold,
					MaxThreshold:         product.MaxThreshold,
					Image:                product.Image,
					Prescription:         product.Prescription,
				}

				if err := db.Create(&newItem).Error; err != nil {
					return err
				}

			} else if err == nil {

				db.Model(&existing).Updates(models.ERPSyncMedicine{
					ERPProductID:         product.ID,
					Stock:                normalizedStock,
					ERPUnit:              product.UnitOfMeasurement,
					SellingPricePerPiece: product.SellingPricePerPiece,
					CostPricePerBox:      product.CostPricePerBox,
					Discount:             product.Discount,
					VATClassification:    product.VATClassification,
					VAT:                  product.VAT,
				})
			}
		}

		if page >= erpResponse.Meta.LastPage {
			break
		}

		page++
	}

	return nil
}

func SyncERPHandler(c *gin.Context, db *gorm.DB) {

	start := time.Now()

	err := SyncERPProducts(db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "ERP sync failed",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":      true,
		"message":     "ERP sync completed successfully",
		"synced_at":   time.Now(),
		"duration_ms": time.Since(start).Milliseconds(),
	})
}

func GetAllItemCodesHandler(c *gin.Context, db *gorm.DB) {

	var itemCodes []string

	err := db.Model(&models.ERPSyncMedicine{}).
		Pluck("item_code", &itemCodes).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to fetch item codes",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":     true,
		"total":      len(itemCodes),
		"item_codes": itemCodes,
	})
}

func GetAllERPSyncMedicinesHandler(c *gin.Context, db *gorm.DB) {

	var medicines []models.ERPSyncMedicine
	var total int64

	// Only page from query
	page := c.DefaultQuery("page", "1")
	search := c.Query("search")

	pageInt := 1
	fmt.Sscan(page, &pageInt)

	if pageInt < 1 {
		pageInt = 1
	}

	// Fixed limit
	limitInt := 20
	offset := (pageInt - 1) * limitInt

	query := db.Model(&models.ERPSyncMedicine{})

	// Optional search
	if search != "" {
		query = query.Where(
			"item_code LIKE ? OR brand_name LIKE ? OR generic_name LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%",
		)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to count ERP medicines",
			"error":   err.Error(),
		})
		return
	}

	// Fetch paginated data
	if err := query.
		Limit(limitInt).
		Offset(offset).
		Order("created_at DESC").
		Find(&medicines).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Failed to fetch ERP medicines",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": true,
		"page":   pageInt,
		"limit":  limitInt, // always 20
		"total":  total,
		"data":   medicines,
	})
}

func ERPProductWebhookHandler(c *gin.Context, db *gorm.DB) {

	var payload cf.ERPProductWebhookRequest

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "Invalid payload",
		})
		return
	}

	product := payload.Product

	var existing models.ERPSyncMedicine

	err := db.Where("item_code = ?", product.ItemCode).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {

		newItem := models.ERPSyncMedicine{
			ERPProductID:         product.ID,
			ItemCode:             product.ItemCode,
			BrandName:            product.BrandName,
			GenericName:          product.GenericName,
			Power:                product.Power,
			DosageForm:           product.DosageForm,
			Packaging:            product.Packaging,
			Description:          product.Description,
			Category:             product.Category,
			SubCategory:          product.SubCategory,
			DirectionsForUse:     product.DirectionsForUse,
			Manufacturer:         product.Manufacturer,
			Distributor:          product.Distributor,
			OriginCountry:        product.OriginCountry,
			UnitOfMeasurement:    product.UnitOfMeasurement,
			NumberOfPiecesPerBox: product.NumberOfPiecesPerBox,
			SellingPricePerPiece: product.SellingPricePerPiece,
			CostPricePerBox:      product.CostPricePerBox,
			Discount:             product.Discount,
			VATClassification:    product.VATClassification,
			VAT:                  product.VAT,
			Stock:                0,
			MinThreshold:         product.MinThreshold,
			MaxThreshold:         product.MaxThreshold,
			Image:                product.Image,
			Prescription:         product.Prescription,
			SyncStatus:           "pending",
		}

		if err := db.Create(&newItem).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"status": false,
				"error":  "Create failed",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": true, "message": "Product created"})
		return
	}

	if err == nil {

		db.Model(&existing).Updates(models.ERPSyncMedicine{
			BrandName:            product.BrandName,
			GenericName:          product.GenericName,
			Power:                product.Power,
			DosageForm:           product.DosageForm,
			Packaging:            product.Packaging,
			Description:          product.Description,
			Category:             product.Category,
			SubCategory:          product.SubCategory,
			DirectionsForUse:     product.DirectionsForUse,
			Manufacturer:         product.Manufacturer,
			Distributor:          product.Distributor,
			OriginCountry:        product.OriginCountry,
			UnitOfMeasurement:    product.UnitOfMeasurement,
			NumberOfPiecesPerBox: product.NumberOfPiecesPerBox,
			SellingPricePerPiece: product.SellingPricePerPiece,
			CostPricePerBox:      product.CostPricePerBox,
			Discount:             product.Discount,
			VATClassification:    product.VATClassification,
			VAT:                  product.VAT,
			MinThreshold:         product.MinThreshold,
			MaxThreshold:         product.MaxThreshold,
			Image:                product.Image,
			Prescription:         product.Prescription,
		})

		//  If mapped → Push to KloudPX
		if existing.IsMapped {
			go maperpandkp.SyncToKloudPX(existing, db)
		}

		c.JSON(http.StatusOK, gin.H{"status": true, "message": "Product updated"})
		return
	}

	c.JSON(http.StatusInternalServerError, gin.H{"status": false})
}

func ERPStockWebhookHandler(c *gin.Context, db *gorm.DB) {

	var payload cf.ERPStockWebhookRequest

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  "Invalid payload",
		})
		return
	}

	var existing models.ERPSyncMedicine

	if err := db.Where("item_code = ?", payload.ItemCode).
		First(&existing).Error; err != nil {

		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "Item not found",
		})
		return
	}

	// Update only stock
	db.Model(&existing).Update("stock", payload.Stock)

	// If mapped → Push only stock update
	if existing.IsMapped {
		go maperpandkp.SyncToKloudPX(existing, db)
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Stock updated",
	})
}
