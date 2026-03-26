package request

type CreateTransactionRequest struct {
	TransactionDate string  `json:"transaction_date" binding:"required"`
	TransactionType string  `json:"transaction_type" binding:"required,oneof=buy sell dividend"`
	AssetType       string  `json:"asset_type" binding:"required"`
	AssetCode       string  `json:"asset_code" binding:"required"`
	AssetName       string  `json:"asset_name" binding:"required"`
	Quantity        string  `json:"quantity" binding:"required"`
	PricePerUnit    string  `json:"price_per_unit" binding:"required"`
	Commission      string  `json:"commission"`
	Notes           *string `json:"notes"`
}
