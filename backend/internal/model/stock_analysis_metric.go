package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type StockAnalysisMetric struct {
	ID                   uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	TaskID               *uint64         `gorm:"index" json:"task_id"`
	UserID               uint64          `gorm:"not null;index;uniqueIndex:idx_metric_scope" json:"user_id"`
	Symbol               string          `gorm:"size:32;not null;index;uniqueIndex:idx_metric_scope" json:"symbol"`
	AssetName            string          `gorm:"size:100;not null" json:"asset_name"`
	Market               string          `gorm:"size:32;default:''" json:"market"`
	PeriodStart          time.Time       `gorm:"not null;type:date;uniqueIndex:idx_metric_scope" json:"period_start"`
	PeriodEnd            time.Time       `gorm:"not null;type:date;uniqueIndex:idx_metric_scope" json:"period_end"`
	TradeCount           int             `gorm:"default:0" json:"trade_count"`
	BuyCount             int             `gorm:"default:0" json:"buy_count"`
	SellCount            int             `gorm:"default:0" json:"sell_count"`
	BuyQuantity          decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"buy_quantity"`
	SellQuantity         decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"sell_quantity"`
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
	UnrealizedProfitRate decimal.Decimal `gorm:"type:decimal(10,4);not null" json:"unrealized_profit_rate"`
	TotalProfit          decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"total_profit"`
	TotalProfitRate      decimal.Decimal `gorm:"type:decimal(10,4);not null" json:"total_profit_rate"`
	PeriodStartPrice     decimal.Decimal `gorm:"type:decimal(15,4);not null" json:"period_start_price"`
	PeriodEndPrice       decimal.Decimal `gorm:"type:decimal(15,4);not null" json:"period_end_price"`
	PeriodPriceChangePct decimal.Decimal `gorm:"type:decimal(10,4);not null" json:"period_price_change_pct"`
	PeriodHighPrice      decimal.Decimal `gorm:"type:decimal(15,4);not null" json:"period_high_price"`
	PeriodLowPrice       decimal.Decimal `gorm:"type:decimal(15,4);not null" json:"period_low_price"`
	MarketDataStatus     string          `gorm:"size:20;default:'complete'" json:"market_data_status"`
	SourceType           string          `gorm:"size:20;default:'on_demand'" json:"source_type"`
	ComputedAt           time.Time       `json:"computed_at"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

func (StockAnalysisMetric) TableName() string {
	return "stock_analysis_metrics"
}
