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

	if err := SyncToKloudPX(erpItem); err != nil {
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
func SyncToKloudPX(erp models.ERPSyncMedicine) error {

	cfg, _ := cfg.Env()

	url := fmt.Sprintf("%s/internal/sync-medicine", cfg.KloudPX.BaseURL)

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
		"unit_of_measurement":      erp.UnitOfMeasurement,
		"number_of_pieces_per_box": erp.NumberOfPiecesPerBox,
		"selling_price_per_piece":  erp.SellingPricePerPiece,
		"cost_price_per_box":       erp.CostPricePerBox,
		"vat_classification":       erp.VATClassification,
		"prescription":             erp.Prescription,
		"manufacturer":             erp.Manufacturer,
		"stock":                    erp.Stock,
	}

	jsonData, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-INTERNAL-KEY", cfg.KloudPX.ServiceKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 🔥 ADD THIS
	body, _ := io.ReadAll(resp.Body)
	fmt.Println("KLOUDPX RESPONSE:", string(body))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to sync to KloudPX")
	}

	return nil
}
