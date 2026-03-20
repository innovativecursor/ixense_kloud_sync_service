package deduct

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/config"
	"github.com/innovativecursor/ixense_kloud_sync_service/apps/pkg/models"
	"gorm.io/gorm"
)

type DeductStockRequest struct {
	KloudpxItemCode string `json:"kloudpx_item_code"`
	Quantity        int    `json:"quantity"` // in PIECES
}

func DeductStockHandler(c *gin.Context, db *gorm.DB) {
	var req DeductStockRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid payload"})
		return
	}

	var item models.ERPSyncMedicine

	if err := db.Where("kloudpx_item_code = ? AND is_mapped = ?", req.KloudpxItemCode, true).
		First(&item).Error; err != nil {

		c.JSON(404, gin.H{"error": "Mapped ERP item not found"})
		return
	}

	// Convert to pieces
	available := item.Stock
	if item.UnitOfMeasurement == "per box" {
		available = item.Stock * item.NumberOfPiecesPerBox
	}

	if req.Quantity > available {
		c.JSON(400, gin.H{"error": "Not enough stock in ERP Sync"})
		return
	}

	//  STEP 1: CALL ACTUAL ERP FIRST
	if err := DeductStockFromERP(item.ItemCode, req.Quantity); err != nil {
		c.JSON(500, gin.H{
			"error": "ERP deduction failed",
		})
		return
	}

	// STEP 2: UPDATE SYNC DB ONLY IF ERP SUCCESS
	newStock := available - req.Quantity

	if item.UnitOfMeasurement == "per box" {
		item.Stock = newStock / item.NumberOfPiecesPerBox
	} else {
		item.Stock = newStock
	}

	if err := db.Save(&item).Error; err != nil {
		c.JSON(500, gin.H{"error": "Failed to update ERP sync stock"})
		return
	}

	c.JSON(200, gin.H{"status": true})
}
func DeductStockFromERP(itemCode string, qty int) error {
	cfgData, err := config.Env()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/deductStock", cfgData.ERP.BaseURL)

	payload := map[string]interface{}{
		"item_code": itemCode,
		"qty":       qty,
	}

	jsonData, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("ERP-API-KEY", cfgData.ERP.APIKey)

	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ERP deduction failed: %s", string(body))
	}

	return nil
}
