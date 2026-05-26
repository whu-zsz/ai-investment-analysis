package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type MarketBoardConstituent struct {
	ID             uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	BoardType      string          `gorm:"size:32;not null;index:idx_board_constituent_board;uniqueIndex:uniq_board_constituent_symbol" json:"board_type"`
	BoardCode      string          `gorm:"size:64;not null;index:idx_board_constituent_board;uniqueIndex:uniq_board_constituent_symbol" json:"board_code"`
	BoardName      string          `gorm:"size:128;not null" json:"board_name"`
	Symbol         string          `gorm:"size:32;not null;uniqueIndex:uniq_board_constituent_symbol;index" json:"symbol"`
	StockName      string          `gorm:"size:128" json:"stock_name"`
	TotalMarketCap decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"total_market_cap"`
	FloatMarketCap decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"float_market_cap"`
	Source         string          `gorm:"size:32;not null" json:"source"`
	SyncedAt       time.Time       `gorm:"not null;index" json:"synced_at"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func (MarketBoardConstituent) TableName() string {
	return "market_board_constituents"
}
