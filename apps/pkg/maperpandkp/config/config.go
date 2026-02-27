package config

type UpdateMappingRequest struct {
	ERPItemCode     string `json:"erp_item_code" binding:"required"`
	KloudpxItemCode string `json:"kloudpx_item_code"`
}
