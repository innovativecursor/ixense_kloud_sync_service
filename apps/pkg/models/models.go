package models

import (
	"time"

	"gorm.io/gorm"
)

type ERPSyncMedicine struct {
	gorm.Model

	ERPProductID         uint   `gorm:"uniqueIndex"`
	ItemCode             string `gorm:"size:100;uniqueIndex"`
	BrandName            string
	GenericName          string
	Power                string
	DosageForm           string
	Packaging            string
	Description          string
	Category             string
	SubCategory          string
	DirectionsForUse     string
	Manufacturer         string
	Distributor          string
	OriginCountry        string
	UnitOfMeasurement    string
	Stock                int
	NumberOfPiecesPerBox int
	SellingPricePerPiece float64
	CostPricePerBox      float64
	Discount             float64
	VATClassification    string
	VAT                  float64
	MinThreshold         int
	MaxThreshold         int
	Image                string
	Prescription         bool

	LastSyncedAt    *time.Time
	SyncStatus      string  `gorm:"size:20;default:'pending'"`
	KloudpxItemCode *string `gorm:"size:100;index"`
	SyncID          *string `gorm:"size:100;uniqueIndex"`
	IsMapped        bool    `gorm:"default:false"`
}
