package request

type StockChatMessageRequest struct {
	Role    string `json:"role" binding:"required"`
	Content string `json:"content" binding:"required"`
}

type StockChatRequest struct {
	Symbol        string                    `json:"symbol" binding:"required"`
	Question      string                    `json:"question" binding:"required"`
	Messages      []StockChatMessageRequest `json:"messages"`
	ContextID     uint64                    `json:"context_id"`
	RefreshMarket bool                      `json:"refresh_market"`
}

type BoardChatRequest struct {
	BoardType string                    `json:"board_type" binding:"required"`
	Code      string                    `json:"code" binding:"required"`
	Question  string                    `json:"question" binding:"required"`
	Messages  []StockChatMessageRequest `json:"messages"`
	ContextID uint64                    `json:"context_id"`
}

type RecommendationChatRequest struct {
	Question string                    `json:"question" binding:"required"`
	Messages []StockChatMessageRequest `json:"messages"`
	ReportID uint64                    `json:"report_id"`
	ContextID uint64                   `json:"context_id"`
}
