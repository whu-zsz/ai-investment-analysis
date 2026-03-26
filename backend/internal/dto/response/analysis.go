package response

type AnalysisReportResponse struct {
	ID                  uint64             `json:"id"`
	ReportType          string             `json:"report_type"`
	ReportTitle         string             `json:"report_title"`
	AnalysisPeriodStart string             `json:"analysis_period_start"`
	AnalysisPeriodEnd   string             `json:"analysis_period_end"`
	TotalInvestment     string             `json:"total_investment"`
	TotalProfit         string             `json:"total_profit"`
	ProfitRate          string             `json:"profit_rate"`
	RiskLevel           string             `json:"risk_level"`
	InvestmentStyle     string             `json:"investment_style"`
	SummaryText         string             `json:"summary_text"`
	RiskAnalysis        string             `json:"risk_analysis"`
	PatternInsights     string             `json:"pattern_insights"`
	PredictionText      string             `json:"prediction_text"`
	ChartData           string             `json:"chart_data"`
	Recommendations     string             `json:"recommendations"`
	AIModel             string             `json:"ai_model"`
	CreatedAt           string             `json:"created_at"`
}

type RiskAnalysisResponse struct {
	RiskLevel       string            `json:"risk_level"`
	RiskScore       int               `json:"risk_score"`
	RiskFactors     []string          `json:"risk_factors"`
	Recommendations []string          `json:"recommendations"`
}

type PredictionResponse struct {
	Prediction string   `json:"prediction"`
	Trend      string   `json:"trend"`
	Confidence int      `json:"confidence"`
	Tips       []string `json:"tips"`
}
