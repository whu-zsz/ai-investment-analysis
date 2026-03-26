package response

type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Total        int64                 `json:"total"`
	Page         int                   `json:"page"`
	PageSize     int                   `json:"page_size"`
}

type TransactionResponse struct {
	ID              uint64  `json:"id"`
	TransactionDate string  `json:"transaction_date"`
	TransactionType string  `json:"transaction_type"`
	AssetType       string  `json:"asset_type"`
	AssetCode       string  `json:"asset_code"`
	AssetName       string  `json:"asset_name"`
	Quantity        string  `json:"quantity"`
	PricePerUnit    string  `json:"price_per_unit"`
	TotalAmount     string  `json:"total_amount"`
	Commission      string  `json:"commission"`
	Profit          *string `json:"profit"`
	Notes           *string `json:"notes"`
	CreatedAt       string  `json:"created_at"`
}

type TransactionStats struct {
	TotalTransactions int64  `json:"total_transactions"`
	BuyCount          int64  `json:"buy_count"`
	SellCount         int64  `json:"sell_count"`
	TotalInvestment   string `json:"total_investment"`
	TotalProfit       string `json:"total_profit"`
}
