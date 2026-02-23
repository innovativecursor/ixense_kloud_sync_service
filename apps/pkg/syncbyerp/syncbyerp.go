package syncbyerp

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/config"
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

		// 🔥 Sync Products
		for _, product := range erpResponse.Products {

			var existing models.ERPSyncMedicine

			err := db.Where("erp_product_id = ?", product.ID).
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
					Stock:                product.Stock,
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
					Stock:                product.Stock,
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
