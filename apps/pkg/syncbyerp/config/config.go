package config

type ERPItemsResponse struct {
	Status   bool         `json:"status"`
	Products []ERPProduct `json:"products"`
	Meta     ERPMeta      `json:"meta"`
}

type ERPMeta struct {
	CurrentPage int `json:"current_page"`
	LastPage    int `json:"last_page"`
	Total       int `json:"total"`
}

type ERPProduct struct {
	ID                   uint    `json:"id"`
	ItemCode             string  `json:"itemCode"`
	Prescription         bool    `json:"prescription"`
	BrandName            string  `json:"brandName"`
	GenericName          string  `json:"genericName"`
	Power                string  `json:"power"`
	DosageForm           string  `json:"dosageForm"`
	Packaging            string  `json:"packaging"`
	Description          string  `json:"description"`
	Category             string  `json:"category"`
	SubCategory          string  `json:"subCategory"`
	DirectionsForUse     string  `json:"directionsForUse"`
	Manufacturer         string  `json:"manufacturer"`
	Distributor          string  `json:"distributor"`
	OriginCountry        string  `json:"originCountry"`
	UnitOfMeasurement    string  `json:"unitOfMeasurement"`
	Stock                int     `json:"stock"`
	NumberOfPiecesPerBox int     `json:"numberOfPiecesPerBox"`
	SellingPricePerPiece float64 `json:"sellingPricePerPiece"`
	CostPricePerBox      float64 `json:"costPricePerBox"`
	Discount             float64 `json:"discount"`
	VATClassification    string  `json:"vatClassification"`
	VAT                  float64 `json:"vat"`
	MinThreshold         int     `json:"minThreshold"`
	MaxThreshold         int     `json:"maxThreshold"`
	Image                string  `json:"image"`
}

type ERPProductWebhookRequest struct {
	Product ERPProduct `json:"product" binding:"required"`
}

type ERPStockWebhookRequest struct {
	ItemCode string `json:"itemCode" binding:"required"`
	Stock    int    `json:"stock" binding:"required"`
}
