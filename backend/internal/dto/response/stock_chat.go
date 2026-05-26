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
	ContextID    uint64                         `json:"context_id"`
	Symbol       string                         `json:"symbol"`
	AssetName    string                         `json:"asset_name"`
	Market       string                         `json:"market"`
	Reply        string                         `json:"reply"`
	AIModel      string                         `json:"ai_model"`
	GeneratedAt  string                         `json:"generated_at"`
	NewsStatus   string                         `json:"news_status"`
	NewsSummary  string                         `json:"news_summary"`
	NewsCoverage string                         `json:"news_coverage"`
	NewsItems    []StockChatNewsItemResponse    `json:"news_items"`
	Snapshot     StockChatSnapshotResponse      `json:"snapshot"`
	Messages     []StockChatMessageResponse     `json:"messages"`
	ToolTrace    []ChatToolTraceStepResponse    `json:"tool_trace,omitempty"`
	ToolResults  []ChatToolResultSnapshotResponse `json:"tool_results,omitempty"`
}

type BoardChatResponse struct {
	ContextID    uint64                         `json:"context_id"`
	BoardType    string                         `json:"board_type"`
	Code         string                         `json:"code"`
	AssetName    string                         `json:"asset_name"`
	Market       string                         `json:"market"`
	Reply        string                         `json:"reply"`
	AIModel      string                         `json:"ai_model"`
	GeneratedAt  string                         `json:"generated_at"`
	NewsStatus   string                         `json:"news_status"`
	NewsSummary  string                         `json:"news_summary"`
	NewsCoverage string                         `json:"news_coverage"`
	NewsItems    []StockChatNewsItemResponse    `json:"news_items"`
	Snapshot     StockChatSnapshotResponse      `json:"snapshot"`
	Messages     []StockChatMessageResponse     `json:"messages"`
	ToolTrace    []ChatToolTraceStepResponse    `json:"tool_trace,omitempty"`
	ToolResults  []ChatToolResultSnapshotResponse `json:"tool_results,omitempty"`
}

type RecommendationChatContextResponse struct {
	StepKey          string                       `json:"step_key"`
	StepLabel        string                       `json:"step_label"`
	ProfileSummary   RecommendationProfileSummary `json:"profile_summary"`
	CandidateCount   int                          `json:"candidate_count"`
	DiscoveryCount   int                          `json:"discovery_count"`
	HeldCount        int                          `json:"held_count"`
	FocusSummary     string                       `json:"focus_summary"`
}

type RecommendationChatResponse struct {
	ContextID    uint64                           `json:"context_id"`
	AssetName    string                           `json:"asset_name"`
	Market       string                           `json:"market"`
	Reply        string                           `json:"reply"`
	AIModel      string                           `json:"ai_model"`
	GeneratedAt  string                           `json:"generated_at"`
	NewsStatus   string                           `json:"news_status"`
	NewsSummary  string                           `json:"news_summary"`
	NewsCoverage string                           `json:"news_coverage"`
	NewsItems    []StockChatNewsItemResponse      `json:"news_items"`
	Snapshot     StockChatSnapshotResponse        `json:"snapshot"`
	Messages     []StockChatMessageResponse       `json:"messages"`
	Context      RecommendationChatContextResponse `json:"context"`
	ReportID     uint64                           `json:"report_id"`
	ReportTitle  string                           `json:"report_title"`
	Candidates   []RecommendationItemResponse     `json:"candidates"`
	ToolTrace    []ChatToolTraceStepResponse      `json:"tool_trace,omitempty"`
	ToolResults  []ChatToolResultSnapshotResponse `json:"tool_results,omitempty"`
	StepContext  *ChatStepContextResponse         `json:"step_context,omitempty"`
}


type RecommendationChatContextSnapshotResponse struct {
	Messages    []StockChatMessageResponse       `json:"messages"`
	ToolTrace   []ChatToolTraceStepResponse      `json:"tool_trace,omitempty"`
	ToolResults []ChatToolResultSnapshotResponse `json:"tool_results,omitempty"`
	StepContext *ChatStepContextResponse         `json:"step_context,omitempty"`
	NewsItems   []StockChatNewsItemResponse      `json:"news_items,omitempty"`
	Reply       string                           `json:"reply"`
	GeneratedAt string                           `json:"generated_at"`
	ReportID    uint64                           `json:"report_id"`
	ReportTitle string                           `json:"report_title"`
}


type ChatContextSnapshotResponse struct {
	ContextID   uint64                         `json:"context_id"`
	ContextType string                         `json:"context_type"`
	TargetKey   string                         `json:"target_key"`
	Title       string                         `json:"title"`
	ReportID    uint64                         `json:"report_id"`
	Messages    []StockChatMessageResponse     `json:"messages"`
	ToolTrace   []ChatToolTraceStepResponse    `json:"tool_trace,omitempty"`
	ToolResults []ChatToolResultSnapshotResponse `json:"tool_results,omitempty"`
	StepContext *ChatStepContextResponse       `json:"step_context,omitempty"`
	NewsItems   []StockChatNewsItemResponse    `json:"news_items,omitempty"`
	Reply       string                         `json:"reply"`
	GeneratedAt string                         `json:"generated_at"`
	ReportTitle string                         `json:"report_title,omitempty"`
}

type ChatToolTraceStepResponse struct {
	Stage      string `json:"stage"`
	Label      string `json:"label"`
	Status     string `json:"status"`
	Summary    string `json:"summary,omitempty"`
	ToolName   string `json:"tool_name,omitempty"`
	StartedAt  string `json:"started_at,omitempty"`
	FinishedAt string `json:"finished_at,omitempty"`
}

type ChatToolResultSnapshotResponse struct {
	ToolName string `json:"tool_name"`
	Status   string `json:"status,omitempty"`
	Summary  string `json:"summary,omitempty"`
	Payload  any    `json:"payload,omitempty"`
	Error    string `json:"error,omitempty"`
}

type ChatStepContextResponse struct {
	Stage             string                         `json:"stage"`
	Label             string                         `json:"label"`
	Summary           string                         `json:"summary,omitempty"`
	ToolName          string                         `json:"tool_name,omitempty"`
	NewsItems         []StockChatNewsItemResponse    `json:"news_items,omitempty"`
	ProfileSummary    *RecommendationProfileSummary  `json:"profile_summary,omitempty"`
	CandidateCount    int                            `json:"candidate_count,omitempty"`
	HeldCount         int                            `json:"held_count,omitempty"`
	DiscoveryCount    int                            `json:"discovery_count,omitempty"`
	FocusSummary      string                         `json:"focus_summary,omitempty"`
	ReferenceSymbols  []string                       `json:"reference_symbols,omitempty"`
	ReferenceBoards   []string                       `json:"reference_boards,omitempty"`
}

type StockChatStreamEvent struct {
	Type    string      `json:"type"`
	Stage   string      `json:"stage,omitempty"`
	Message string      `json:"message,omitempty"`
	Token   string      `json:"token,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}
