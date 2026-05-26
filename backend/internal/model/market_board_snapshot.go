package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type MarketBoardSnapshot struct {
	ID              uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	BoardType       string          `gorm:"size:32;not null;index:idx_board_batch;index:idx_board_code_time" json:"board_type"`
	Code            string          `gorm:"size:32;not null;index:idx_board_code_time" json:"code"`
	Name            string          `gorm:"size:128;not null" json:"name"`
	LastPrice       decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"last_price"`
	ChangeAmount    decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"change_amount"`
	ChangePercent   decimal.Decimal `gorm:"type:decimal(10,4);default:0" json:"change_percent"`
	Volume          decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"volume"`
	Turnover        decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"turnover"`
	TotalMarketCap  decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"total_market_cap"`
	FloatMarketCap  decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"float_market_cap"`
	StockCount      int             `gorm:"default:0" json:"stock_count"`
	RiseCount       int             `gorm:"default:0" json:"rise_count"`
	FallCount       int             `gorm:"default:0" json:"fall_count"`
	FlatCount       int             `gorm:"default:0" json:"flat_count"`
	Source          string          `gorm:"size:32;not null" json:"source"`
	BatchNo         string          `gorm:"size:64;not null;index:idx_board_batch" json:"batch_no"`
	SnapshotTime    time.Time       `gorm:"not null;index:idx_board_code_time;index" json:"snapshot_time"`
	ConstituentNode string          `gorm:"size:64" json:"constituent_node"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func (MarketBoardSnapshot) TableName() string {
	return "market_board_snapshots"
}
