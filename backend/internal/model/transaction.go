package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID              uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint64          `gorm:"not null;index" json:"user_id"`
	TransactionDate time.Time       `gorm:"not null;type:date" json:"transaction_date"`
	TransactionType string          `gorm:"size:20;not null" json:"transaction_type"` // buy, sell, dividend
	AssetType       string          `gorm:"size:20;not null" json:"asset_type"`       // stock, fund, bond
	AssetCode       string          `gorm:"size:20;not null;index" json:"asset_code"`
	AssetName       string          `gorm:"size:100;not null" json:"asset_name"`
	Quantity        decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"quantity"`
	PricePerUnit    decimal.Decimal `gorm:"type:decimal(10,2);not null" json:"price_per_unit"`
	TotalAmount     decimal.Decimal `gorm:"type:decimal(15,2);not null" json:"total_amount"`
	Commission      decimal.Decimal `gorm:"type:decimal(10,2);default:0.00" json:"commission"`
	Profit          *decimal.Decimal `gorm:"type:decimal(15,2)" json:"profit"`
	Notes           *string         `gorm:"type:text" json:"notes"`
	SourceFile      *string         `gorm:"size:255" json:"source_file"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func (Transaction) TableName() string {
	return "transactions"
}
