package model

import (
	"strings"
	"time"
	"unicode"

	"github.com/shopspring/decimal"
)

type StockQuoteDetail struct {
	ID             uint64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol         string          `gorm:"size:32;not null;uniqueIndex" json:"symbol"`
	Name           string          `gorm:"size:64;not null" json:"name"`
	Market         string          `gorm:"size:32;not null;index" json:"market"`
	LastPrice      decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"last_price"`
	OpenPrice      decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"open_price"`
	HighPrice      decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"high_price"`
	LowPrice       decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"low_price"`
	PrevClose      decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"prev_close"`
	ChangeAmount   decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"change_amount"`
	ChangePercent  decimal.Decimal `gorm:"type:decimal(10,4);default:0" json:"change_percent"`
	Volume         decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"volume"`
	Turnover       decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"turnover"`
	VolumeRatio    decimal.Decimal `gorm:"type:decimal(10,4);default:0" json:"volume_ratio"`
	TurnoverRate   decimal.Decimal `gorm:"type:decimal(10,4);default:0" json:"turnover_rate"`
	Amplitude      decimal.Decimal `gorm:"type:decimal(10,4);default:0" json:"amplitude"`
	LimitUp        decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"limit_up"`
	LimitDown      decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"limit_down"`
	AveragePrice   decimal.Decimal `gorm:"type:decimal(18,4);default:0" json:"average_price"`
	TotalShares    decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"total_shares"`
	FloatShares    decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"float_shares"`
	TotalMarketCap decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"total_market_cap"`
	FloatMarketCap decimal.Decimal `gorm:"type:decimal(24,4);default:0" json:"float_market_cap"`
	Industry       string          `gorm:"size:128" json:"industry"`
	Region         string          `gorm:"size:64" json:"region"`
	Concepts       string          `gorm:"type:text" json:"concepts"`
	Source         string          `gorm:"size:32;not null" json:"source"`
	FetchedAt      time.Time       `gorm:"not null;index" json:"fetched_at"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func (StockQuoteDetail) TableName() string {
	return "stock_quote_details"
}

func NormalizeIndustryLabel(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	upper := strings.ToUpper(trimmed)
	if strings.HasPrefix(upper, "GP-") {
		return ""
	}
	if looksLikeMarketCode(trimmed) {
		return ""
	}
	return trimmed
}

func NormalizeRegionLabel(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || trimmed == "-" || trimmed == "—" {
		return ""
	}
	return trimmed
}

func NormalizeConceptList(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if trimmed == "腾讯行情" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func looksLikeMarketCode(value string) bool {
	hasLetter := false
	for _, r := range value {
		if r == '-' || r == '_' || unicode.IsDigit(r) {
			continue
		}
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				return false
			}
			continue
		}
		return false
	}
	return hasLetter
}
