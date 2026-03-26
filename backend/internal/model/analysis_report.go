package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type AnalysisReport struct {
	ID                  uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID              uint64          `gorm:"not null;index" json:"user_id"`
	ReportType          string          `gorm:"size:20;not null" json:"report_type"` // summary, risk, prediction, pattern
	ReportTitle         string          `gorm:"size:200;not null" json:"report_title"`
	AnalysisPeriodStart time.Time       `gorm:"not null;type:date" json:"analysis_period_start"`
	AnalysisPeriodEnd   time.Time       `gorm:"not null;type:date" json:"analysis_period_end"`
	TotalInvestment     decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"total_investment"`
	TotalProfit         decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"total_profit"`
	ProfitRate          decimal.Decimal `gorm:"type:decimal(5,2);not null" json:"profit_rate"`
	RiskLevel           string          `gorm:"size:10;not null" json:"risk_level"` // low, medium, high
	InvestmentStyle     *string         `gorm:"size:50" json:"investment_style"`
	SummaryText         string          `gorm:"type:text;not null" json:"summary_text"`
	RiskAnalysis        *string         `gorm:"type:text" json:"risk_analysis"`
	PatternInsights     *string         `gorm:"type:text" json:"pattern_insights"`
	PredictionText      *string         `gorm:"type:text" json:"prediction_text"`
	ChartData           *string         `gorm:"type:json" json:"chart_data"`
	Recommendations     *string         `gorm:"type:text" json:"recommendations"`
	AIModel             string          `gorm:"size:50;default:'deepseek'" json:"ai_model"`
	CreatedAt           time.Time       `json:"created_at"`
}

func (AnalysisReport) TableName() string {
	return "ai_analysis_reports"
}
