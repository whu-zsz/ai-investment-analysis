package response

type PortfolioResponse struct {
	ID                uint64 `json:"id"`
	AssetCode         string `json:"asset_code"`
	AssetName         string `json:"asset_name"`
	AssetType         string `json:"asset_type"`
	TotalQuantity     string `json:"total_quantity"`
	AvailableQuantity string `json:"available_quantity"`
	AverageCost       string `json:"average_cost"`
	CurrentPrice      string `json:"current_price"`
	MarketValue       string `json:"market_value"`
	ProfitLoss        string `json:"profit_loss"`
	ProfitLossPercent string `json:"profit_loss_percent"`
	LastUpdated       string `json:"last_updated"`
}
