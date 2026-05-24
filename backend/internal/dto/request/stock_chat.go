package request

type StockChatMessageRequest struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type StockChatRequest struct {
	Symbol        string                    `json:"symbol" binding:"required"`
	Question      string                    `json:"question" binding:"required"`
	Messages      []StockChatMessageRequest `json:"messages"`
	RefreshMarket bool                      `json:"refresh_market"`
}
