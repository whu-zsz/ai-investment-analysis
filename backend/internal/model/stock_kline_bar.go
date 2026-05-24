package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type StockKlineBar struct {
	ID            uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol        string          `gorm:"size:32;not null;uniqueIndex:idx_symbol_period_adjust_time;index" json:"symbol"`
	Period        string          `gorm:"size:16;not null;uniqueIndex:idx_symbol_period_adjust_time;index" json:"period"`
	AdjustType    string          `gorm:"size:16;not null;uniqueIndex:idx_symbol_period_adjust_time;index" json:"adjust_type"`
	BarTime       time.Time       `gorm:"not null;uniqueIndex:idx_symbol_period_adjust_time;index" json:"bar_time"`
	OpenPrice     decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"open_price"`
	ClosePrice    decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"close_price"`
	HighPrice     decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"high_price"`
	LowPrice      decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"low_price"`
	Volume        decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"volume"`
	Turnover      decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"turnover"`
	Amplitude     decimal.Decimal `gorm:"type:decimal(10,4);default:0" json:"amplitude"`
	ChangePercent decimal.Decimal `gorm:"type:decimal(10,4);default:0" json:"change_percent"`
	ChangeAmount  decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"change_amount"`
	TurnoverRate  decimal.Decimal `gorm:"type:decimal(10,4);default:0" json:"turnover_rate"`
	Source        string          `gorm:"size:32;not null" json:"source"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func (StockKlineBar) TableName() string {
	return "stock_kline_bars"
}
