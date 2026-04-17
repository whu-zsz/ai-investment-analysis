package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type AnalysisReportItem struct {
	ID                   uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	ReportID             uint64          `gorm:"not null;index;uniqueIndex:idx_report_symbol" json:"report_id"`
	UserID               uint64          `gorm:"not null;index" json:"user_id"`
	Symbol               string          `gorm:"size:32;not null;uniqueIndex:idx_report_symbol;index" json:"symbol"`
	AssetName            string          `gorm:"size:100;not null" json:"asset_name"`
	Market               string          `gorm:"size:32;default:''" json:"market"`
	TradeCount           int             `gorm:"default:0" json:"trade_count"`
	BuyCount             int             `gorm:"default:0" json:"buy_count"`
	SellCount            int             `gorm:"default:0" json:"sell_count"`
	BuyAmount            decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"buy_amount"`
	SellAmount           decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"sell_amount"`
	NetQuantity          decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"net_quantity"`
	RealizedProfit       decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"realized_profit"`
	RealizedProfitRate   decimal.Decimal `gorm:"type:decimal(10,4);not null" json:"realized_profit_rate"`
	EndingPositionQty    decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"ending_position_qty"`
	EndingAvgCost        decimal.Decimal `gorm:"type:decimal(15,4);not null" json:"ending_avg_cost"`
	LatestPrice          decimal.Decimal `gorm:"type:decimal(15,4);not null" json:"latest_price"`
	LatestMarketValue    decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"latest_market_value"`
	UnrealizedProfit     decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"unrealized_profit"`
	TotalProfit          decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"total_profit"`
	ChangePercent7D      decimal.Decimal `gorm:"type:decimal(10,4);not null" json:"change_percent_7d"`
	PeriodPriceChangePct decimal.Decimal `gorm:"type:decimal(10,4);not null" json:"period_price_change_pct"`
	MarketDataStatus     string          `gorm:"size:20;default:'complete'" json:"market_data_status"`
	RiskLevel            string          `gorm:"size:10;not null" json:"risk_level"`
	InvestmentStyle      *string         `gorm:"size:50" json:"investment_style"`
	AnalysisText         string          `gorm:"type:text;not null" json:"analysis_text"`
	Recommendation       string          `gorm:"size:20;not null" json:"recommendation"`
	KeyPoints            *string         `gorm:"type:json" json:"key_points"`
	RawAIOutput          *string         `gorm:"type:longtext" json:"raw_ai_output"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

func (AnalysisReportItem) TableName() string {
	return "ai_analysis_report_items"
}
