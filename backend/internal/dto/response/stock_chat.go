package response

type StockChatMessageResponse struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type StockChatNewsItemResponse struct {
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Source      string `json:"source"`
	URL         string `json:"url"`
	PublishedAt string `json:"published_at"`
	Provider    string `json:"provider"`
}

type StockChatSnapshotResponse struct {
	LastPrice     string `json:"last_price"`
	ChangePercent string `json:"change_percent"`
	HighPrice     string `json:"high_price"`
	LowPrice      string `json:"low_price"`
	Volume        string `json:"volume"`
	Turnover      string `json:"turnover"`
	Source        string `json:"source"`
	FetchedAt     string `json:"fetched_at"`
	Period        string `json:"period"`
	TrendSummary  string `json:"trend_summary"`
}

type StockChatResponse struct {
	Symbol       string                      `json:"symbol"`
	AssetName    string                      `json:"asset_name"`
	Market       string                      `json:"market"`
	Reply        string                      `json:"reply"`
	AIModel      string                      `json:"ai_model"`
	GeneratedAt  string                      `json:"generated_at"`
	NewsStatus   string                      `json:"news_status"`
	NewsSummary  string                      `json:"news_summary"`
	NewsCoverage string                      `json:"news_coverage"`
	NewsItems    []StockChatNewsItemResponse `json:"news_items"`
	Snapshot     StockChatSnapshotResponse   `json:"snapshot"`
	Messages     []StockChatMessageResponse  `json:"messages"`
}

type BoardChatResponse struct {
	BoardType    string                      `json:"board_type"`
	Code         string                      `json:"code"`
	AssetName    string                      `json:"asset_name"`
	Market       string                      `json:"market"`
	Reply        string                      `json:"reply"`
	AIModel      string                      `json:"ai_model"`
	GeneratedAt  string                      `json:"generated_at"`
	NewsStatus   string                      `json:"news_status"`
	NewsSummary  string                      `json:"news_summary"`
	NewsCoverage string                      `json:"news_coverage"`
	NewsItems    []StockChatNewsItemResponse `json:"news_items"`
	Snapshot     StockChatSnapshotResponse   `json:"snapshot"`
	Messages     []StockChatMessageResponse  `json:"messages"`
}

type StockChatStreamEvent struct {
	Type    string      `json:"type"`
	Stage   string      `json:"stage,omitempty"`
	Message string      `json:"message,omitempty"`
	Token   string      `json:"token,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
