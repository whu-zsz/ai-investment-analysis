package response

type AnalysisReportResponse struct {
	ID                  uint64 `json:"id"`
	ReportType          string `json:"report_type"`
	ReportTitle         string `json:"report_title"`
	AnalysisPeriodStart string `json:"analysis_period_start"`
	AnalysisPeriodEnd   string `json:"analysis_period_end"`
	TotalInvestment     string `json:"total_investment"`
	TotalProfit         string `json:"total_profit"`
	ProfitRate          string `json:"profit_rate"`
	RiskLevel           string `json:"risk_level"`
	MarketDataStatus    string `json:"market_data_status"`
	InvestmentStyle     string `json:"investment_style"`
	SummaryText         string `json:"summary_text"`
	RiskAnalysis        string `json:"risk_analysis"`
	PatternInsights     string `json:"pattern_insights"`
	PredictionText      string `json:"prediction_text"`
	ChartData           string `json:"chart_data"`
	Recommendations     string `json:"recommendations"`
	AIModel             string `json:"ai_model"`
	CreatedAt           string `json:"created_at"`
}

type AnalysisTaskResponse struct {
	ID            uint64 `json:"id"`
	Status        string `json:"status"`
	ProgressStage string `json:"progress_stage"`
	CreatedAt     string `json:"created_at"`
}

type AnalysisTaskDetailResponse struct {
	ID                  uint64 `json:"id"`
	TaskType            string `json:"task_type"`
	Status              string `json:"status"`
	ProgressStage       string `json:"progress_stage"`
	AnalysisPeriodStart string `json:"analysis_period_start"`
	AnalysisPeriodEnd   string `json:"analysis_period_end"`
	ResultReportID      uint64 `json:"result_report_id,omitempty"`
	ErrorMessage        string `json:"error_message"`
	StartedAt           string `json:"started_at"`
	FinishedAt          string `json:"finished_at"`
	CreatedAt           string `json:"created_at"`
	UpdatedAt           string `json:"updated_at"`
}

type AnalysisTaskListResponse struct {
	Items    []AnalysisTaskDetailResponse `json:"items"`
	Total    int64                        `json:"total"`
	Page     int                          `json:"page"`
	PageSize int                          `json:"page_size"`
}

type AnalysisReportItemResponse struct {
	ID                   uint64   `json:"id"`
	Symbol               string   `json:"symbol"`
	AssetName            string   `json:"asset_name"`
	Market               string   `json:"market"`
	TradeCount           int      `json:"trade_count"`
	BuyCount             int      `json:"buy_count"`
	SellCount            int      `json:"sell_count"`
	BuyAmount            string   `json:"buy_amount"`
	SellAmount           string   `json:"sell_amount"`
	NetQuantity          string   `json:"net_quantity"`
	RealizedProfit       string   `json:"realized_profit"`
	RealizedProfitRate   string   `json:"realized_profit_rate"`
	EndingPositionQty    string   `json:"ending_position_qty"`
	EndingAvgCost        string   `json:"ending_avg_cost"`
	LatestPrice          string   `json:"latest_price"`
	LatestMarketValue    string   `json:"latest_market_value"`
	UnrealizedProfit     string   `json:"unrealized_profit"`
	TotalProfit          string   `json:"total_profit"`
	ChangePercent7D      string   `json:"change_percent_7d"`
	PeriodPriceChangePct string   `json:"period_price_change_pct"`
	MarketDataStatus     string   `json:"market_data_status"`
	RiskLevel            string   `json:"risk_level"`
	InvestmentStyle      string   `json:"investment_style"`
	AnalysisText         string   `json:"analysis_text"`
	Recommendation       string   `json:"recommendation"`
	KeyPoints            []string `json:"key_points"`
	CreatedAt            string   `json:"created_at"`
}

type AnalysisReportDetailResponse struct {
	ID                  uint64                       `json:"id"`
	TaskID              uint64                       `json:"task_id,omitempty"`
	ReportType          string                       `json:"report_type"`
	ReportTitle         string                       `json:"report_title"`
	AnalysisPeriodStart string                       `json:"analysis_period_start"`
	AnalysisPeriodEnd   string                       `json:"analysis_period_end"`
	SymbolsCount        int                          `json:"symbols_count"`
	WinningTrades       int                          `json:"winning_trades"`
	LosingTrades        int                          `json:"losing_trades"`
	TotalInvestment     string                       `json:"total_investment"`
	TotalProfit         string                       `json:"total_profit"`
	ProfitRate          string                       `json:"profit_rate"`
	RiskLevel           string                       `json:"risk_level"`
	MarketDataStatus    string                       `json:"market_data_status"`
	InvestmentStyle     string                       `json:"investment_style"`
	SummaryText         string                       `json:"summary_text"`
	RiskAnalysis        string                       `json:"risk_analysis"`
	PatternInsights     string                       `json:"pattern_insights"`
	PredictionText      string                       `json:"prediction_text"`
	ChartData           string                       `json:"chart_data"`
	Recommendations     []string                     `json:"recommendations"`
	AIModel             string                       `json:"ai_model"`
	CreatedAt           string                       `json:"created_at"`
	Items               []AnalysisReportItemResponse `json:"items"`
}

type RiskAnalysisResponse struct {
	RiskLevel       string   `json:"risk_level"`
	RiskScore       int      `json:"risk_score"`
	RiskFactors     []string `json:"risk_factors"`
	Recommendations []string `json:"recommendations"`
}

type PredictionResponse struct {
	Prediction string   `json:"prediction"`
	Trend      string   `json:"trend"`
	Confidence int      `json:"confidence"`
	Tips       []string `json:"tips"`
}
