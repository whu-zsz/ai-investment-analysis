package response

type AnalysisCandidateSource struct {
	Type string `json:"type"`
}

type AnalysisCandidateResponse struct {
	Symbol       string                    `json:"symbol"`
	AssetName    string                    `json:"asset_name"`
	AssetType    string                    `json:"asset_type"`
	Market       string                    `json:"market"`
	Sources      []AnalysisCandidateSource `json:"sources"`
	IsHeld       bool                      `json:"is_held"`
	TradeCount   int                       `json:"trade_count"`
	LastPrice    string                    `json:"last_price"`
	ChangePercent string                   `json:"change_percent"`
}

type AnalysisCandidatesResponse struct {
	DefaultSymbol string                      `json:"default_symbol"`
	Candidates    []AnalysisCandidateResponse `json:"candidates"`
}

type RecommendationProfileSummary struct {
	InvestmentPreference string `json:"investment_preference"`
	RiskTolerance        string `json:"risk_tolerance"`
	TotalProfit          string `json:"total_profit"`
	HeldPositions        int    `json:"held_positions"`
	CandidateCount       int    `json:"candidate_count"`
}

type RecommendationItemResponse struct {
	Symbol         string `json:"symbol"`
	AssetName      string `json:"asset_name"`
	AssetType      string `json:"asset_type"`
	Market         string `json:"market"`
	SourceTags     []string `json:"source_tags"`
	Action         string `json:"action"`
	Score          string `json:"score"`
	LatestPrice    string `json:"latest_price"`
	ChangePercent  string `json:"change_percent"`
	MatchReason    string `json:"match_reason"`
	RiskNote       string `json:"risk_note"`
	DataStatus     string `json:"data_status"`
	IsHeld         bool   `json:"is_held"`
	TradeCount     int    `json:"trade_count"`
}

type AnalysisRecommendationsResponse struct {
	GeneratedAt    string                        `json:"generated_at"`
	ReportID       uint64                        `json:"report_id"`
	ProfileSummary RecommendationProfileSummary  `json:"profile_summary"`
	SummaryText    string                        `json:"summary_text"`
	Candidates     []RecommendationItemResponse  `json:"candidates"`
}
