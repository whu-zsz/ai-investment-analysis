package model

import (
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Portfolio struct {
	ID                uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID            uint64          `gorm:"not null;uniqueIndex:idx_user_asset" json:"user_id"`
	AssetCode         string          `gorm:"size:20;not null;uniqueIndex:idx_user_asset" json:"asset_code"`
	AssetName         string          `gorm:"size:100;not null" json:"asset_name"`
	AssetType         string          `gorm:"size:20;not null" json:"asset_type"`
	TotalQuantity     decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"total_quantity"`
	AvailableQuantity decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"available_quantity"`
	AverageCost       decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"average_cost"`
	CurrentPrice      *decimal.Decimal `gorm:"type:decimal(10,2)" json:"current_price"`
	MarketValue       decimal.Decimal `gorm:"type:decimal(15,2);" json:"market_value"`
	ProfitLoss        decimal.Decimal `gorm:"type:decimal(15,2);" json:"profit_loss"`
	ProfitLossPercent decimal.Decimal `gorm:"type:decimal(5,2);" json:"profit_loss_percent"`
	LastUpdated       time.Time       `json:"last_updated"`
}

func (Portfolio) TableName() string {
	return "portfolios"
}

// BeforeSave 钩子函数，在保存前计算市值和盈亏
func (p *Portfolio) BeforeSave(tx *gorm.DB) error {
	if p.CurrentPrice != nil {
		p.MarketValue = p.TotalQuantity.Mul(*p.CurrentPrice)
		p.ProfitLoss = p.CurrentPrice.Sub(p.AverageCost).Mul(p.TotalQuantity)

		if !p.AverageCost.IsZero() {
			percent := p.CurrentPrice.Sub(p.AverageCost).Div(p.AverageCost).Mul(decimal.NewFromInt(100))
			p.ProfitLossPercent = percent
		}
	}
	return nil
}
