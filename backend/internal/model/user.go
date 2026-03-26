package model

import (
	"time"

	"github.com/shopspring/decimal"
)

type User struct {
	ID                   uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Username             string          `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Email                string          `gorm:"uniqueIndex;size:100;not null" json:"email"`
	PasswordHash         string          `gorm:"size:255;not null" json:"-"`
	Phone                *string         `gorm:"size:20" json:"phone"`
	AvatarURL            *string         `gorm:"size:500" json:"avatar_url"`
	InvestmentPreference string          `gorm:"size:20;default:'balanced'" json:"investment_preference"`
	TotalProfit          decimal.Decimal `gorm:"type:decimal(15,2);default:0.00" json:"total_profit"`
	RiskTolerance        string          `gorm:"size:10;default:'medium'" json:"risk_tolerance"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
	LastLoginAt          *time.Time      `json:"last_login_at"`
	IsActive             bool            `gorm:"default:true" json:"is_active"`
}

func (User) TableName() string {
	return "users"
}
