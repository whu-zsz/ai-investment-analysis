package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type MarketSnapshot struct {
	ID            uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol        string          `gorm:"size:32;not null;index:idx_symbol_snapshot" json:"symbol"`
	Name          string          `gorm:"size:64;not null" json:"name"`
	Market        string          `gorm:"size:32;not null;index" json:"market"`
	SnapshotTime  time.Time       `gorm:"not null;index:idx_symbol_snapshot;index" json:"snapshot_time"`
	LastPrice     decimal.Decimal `gorm:"type:decimal(18,4);not null" json:"last_price"`
	ChangeAmount  decimal.Decimal `gorm:"type:decimal(18,4);not null" json:"change_amount"`
	ChangePercent decimal.Decimal `gorm:"type:decimal(10,4);not null" json:"change_percent"`
	OpenPrice     decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"open_price"`
	HighPrice     decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"high_price"`
	LowPrice      decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"low_price"`
	PrevClose     decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"prev_close"`
	Volume        decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"volume"`
	Turnover      decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"turnover"`
	Source        string          `gorm:"size:32;not null" json:"source"`
	BatchNo       string          `gorm:"size:64;not null;index" json:"batch_no"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

func (MarketSnapshot) TableName() string {
	return "market_snapshots"
}
