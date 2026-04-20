package response

import "stock-analysis-backend/internal/model"

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

func NewTransactionResponse(transaction *model.Transaction) TransactionResponse {
	var profit *string
	if transaction.Profit != nil {
		profitStr := transaction.Profit.String()
		profit = &profitStr
	}

	return TransactionResponse{
		ID:              transaction.ID,
		TransactionDate: transaction.TransactionDate.Format("2006-01-02"),
		TransactionType: transaction.TransactionType,
		AssetType:       transaction.AssetType,
		AssetCode:       transaction.AssetCode,
		AssetName:       transaction.AssetName,
		Quantity:        transaction.Quantity.String(),
		PricePerUnit:    transaction.PricePerUnit.String(),
		TotalAmount:     transaction.TotalAmount.String(),
		Commission:      transaction.Commission.String(),
		Profit:          profit,
		Notes:           transaction.Notes,
		CreatedAt:       transaction.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}
