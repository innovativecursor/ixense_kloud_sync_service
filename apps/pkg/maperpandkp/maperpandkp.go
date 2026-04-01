package maperpandkp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	cfg "github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/config"
	"github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/maperpandkp/config"
	"github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/models"
	"gorm.io/gorm"
)

func generateSyncCode() string {
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)
	random := rand.Intn(1000)
	return fmt.Sprintf("SYNC-%d-%03d", timestamp, random)
}

// func UpdateERPMappingHandler(c *gin.Context, db *gorm.DB) {

// 	var payload config.UpdateMappingRequest
// 	if err := c.ShouldBindJSON(&payload); err != nil {
// 		c.JSON(http.StatusBadRequest, gin.H{
// 			"status": false,
// 			"error":  err.Error(),
// 		})
// 		return
// 	}

// 	tx := db.Begin()

// 	var erpItem models.ERPSyncMedicine

// 	if err := tx.Where("item_code = ?", payload.ERPItemCode).
// 		First(&erpItem).Error; err != nil {

// 		tx.Rollback()

// 		c.JSON(http.StatusNotFound, gin.H{
// 			"status":  false,
// 			"message": "ERP item not found",
// 		})
// 		return
// 	}

// 	klCode := payload.KloudpxItemCode
// 	syncCode := generateSyncCode()

// 	erpItem.KloudpxItemCode = &klCode
// 	erpItem.SyncID = &syncCode
// 	erpItem.IsMapped = true
// 	// Save mapping
// 	if err := tx.Save(&erpItem).Error; err != nil {
// 		tx.Rollback()

// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"status": false,
// 			"error":  "Failed to save mapping",
// 		})
// 		return
// 	}

// 	if err := SyncToKloudPX(erpItem); err != nil {
// 		tx.Rollback()

// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"status":  false,
// 			"message": "Sync failed, mapping reverted",
// 			"error":   err.Error(),
// 		})
// 		return
// 	}

// 	tx.Commit()

// 	c.JSON(http.StatusOK, gin.H{
// 		"status":  true,
// 		"message": "Mapping and sync successful",
// 		"data":    erpItem,
// 	})
// }

func UpdateERPMappingHandler(c *gin.Context, db *gorm.DB) {

	var payload config.UpdateMappingRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"status": false,
			"error":  err.Error(),
		})
		return
	}

	tx := db.Begin()

	var erpItem models.ERPSyncMedicine

	//  Check ERP item exists
	if err := tx.Where("item_code = ?", payload.ERPItemCode).
		First(&erpItem).Error; err != nil {

		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{
			"status":  false,
			"message": "ERP item not found",
		})
		return
	}

	//  RULE 1: Check if this ERP item already mapped
	if erpItem.IsMapped {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "This ERP item is already mapped",
		})
		return
	}

	// RULE 2: Check if KloudPX item already used by another ERP item
	var existingMapping models.ERPSyncMedicine
	err := tx.Where("kloudpx_item_code = ?", payload.KloudpxItemCode).
		First(&existingMapping).Error

	if err == nil {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  false,
			"message": "This KloudPX item code is already mapped to another ERP item",
		})
		return
	}

	klCode := payload.KloudpxItemCode
	syncCode := generateSyncCode()

	erpItem.KloudpxItemCode = &klCode
	erpItem.SyncID = &syncCode
	erpItem.IsMapped = true

	if err := tx.Save(&erpItem).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": false,
			"error":  "Failed to save mapping",
		})
		return
	}

	if err := SyncToKloudPX(erpItem, db); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  false,
			"message": "Sync failed, mapping reverted",
			"error":   err.Error(),
		})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"status":  true,
		"message": "Mapping and sync successful",
		"data":    erpItem,
	})
}
func SyncToKloudPX(erp models.ERPSyncMedicine, db *gorm.DB) error {

	cfgData, err := cfg.Env()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/internal/sync-medicine", cfgData.KloudPX.BaseURL)

	// Handle nullable pointer
	var klCode string
	if erp.KloudpxItemCode != nil {
		klCode = *erp.KloudpxItemCode
	}

	payload := map[string]interface{}{
		"erp_item_code":            erp.ItemCode,
		"kloudpx_item_code":        klCode,
		"brand_name":               erp.BrandName,
		"generic_name":             erp.GenericName,
		"category":                 erp.Category,
		"power":                    erp.Power,
		"dosage_form":              erp.DosageForm,
		"packaging":                erp.Packaging,
		"description":              erp.Description,
		"unit_of_measurement":      "per piece",
		"number_of_pieces_per_box": erp.NumberOfPiecesPerBox,
		"selling_price_per_piece":  erp.SellingPricePerPiece,
		"cost_price_per_box":       erp.CostPricePerBox,
		"vat_classification":       erp.VATClassification,
		"prescription":             erp.Prescription,
		"manufacturer":             erp.Manufacturer,
		"stock":                    erp.Stock,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-INTERNAL-KEY", cfgData.KloudPX.ServiceKey)

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		erp.SyncStatus = "failed"
		db.Save(&erp)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		erp.SyncStatus = "failed"
		db.Save(&erp)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		erp.SyncStatus = "failed"
		db.Save(&erp)
		return fmt.Errorf("sync failed: %s", string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err == nil {

		if code, ok := result["kloudpx_item_code"].(string); ok && code != "" {

			erp.KloudpxItemCode = &code
		}
	}

	now := time.Now()
	erp.SyncStatus = "success"
	erp.LastSyncedAt = &now

	if err := db.Save(&erp).Error; err != nil {
		return err
	}

	return nil
}
func RecoverySyncHandler(c *gin.Context, db *gorm.DB) {

	var items []models.ERPSyncMedicine

	// Get all mapped medicines
	err := db.Where("is_mapped = ?", true).
		Find(&items).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"status": false,
			"error":  "Failed to fetch items",
		})
		return
	}

	if len(items) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status":  true,
			"message": "No mapped items found",
		})
		return
	}

	successCount := 0
	failedCount := 0

	for _, item := range items {

		err := SyncToKloudPX(item, db)
		if err != nil {
			failedCount++
		} else {
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":        true,
		"total_items":   len(items),
		"success_count": successCount,
		"failed_count":  failedCount,
		"message":       "Recovery sync completed",
	})
}
