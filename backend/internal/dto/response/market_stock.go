package response

import (
	"sort"
	"strings"

	"stock-analysis-backend/internal/model"
)

type MarketStockDetailResponse struct {
	Symbol           string   `json:"symbol"`
	Name             string   `json:"name"`
	Market           string   `json:"market"`
	LastPrice        string   `json:"last_price"`
	OpenPrice        string   `json:"open_price"`
	HighPrice        string   `json:"high_price"`
	LowPrice         string   `json:"low_price"`
	PrevClose        string   `json:"prev_close"`
	ChangeAmount     string   `json:"change_amount"`
	ChangePercent    string   `json:"change_percent"`
	Volume           string   `json:"volume"`
	Turnover         string   `json:"turnover"`
	VolumeRatio      string   `json:"volume_ratio"`
	TurnoverRate     string   `json:"turnover_rate"`
	Amplitude        string   `json:"amplitude"`
	LimitUp          string   `json:"limit_up"`
	LimitDown        string   `json:"limit_down"`
	AveragePrice     string   `json:"average_price"`
	TotalShares      string   `json:"total_shares"`
	FloatShares      string   `json:"float_shares"`
	TotalMarketCap   string   `json:"total_market_cap"`
	FloatMarketCap   string   `json:"float_market_cap"`
	Industry         string   `json:"industry"`
	Region           string   `json:"region"`
	Concepts         []string `json:"concepts"`
	Source           string   `json:"source"`
	FetchedAt        string   `json:"fetched_at"`
	IsStale          bool     `json:"is_stale"`
	RefreshTriggered bool     `json:"refresh_triggered"`
}

type MarketKlineBarResponse struct {
	BarTime       string `json:"bar_time"`
	OpenPrice     string `json:"open_price"`
	ClosePrice    string `json:"close_price"`
	HighPrice     string `json:"high_price"`
	LowPrice      string `json:"low_price"`
	Volume        string `json:"volume"`
	Turnover      string `json:"turnover"`
	Amplitude     string `json:"amplitude"`
	ChangePercent string `json:"change_percent"`
	ChangeAmount  string `json:"change_amount"`
	TurnoverRate  string `json:"turnover_rate"`
}

type MarketStockKlineResponse struct {
	Symbol           string                   `json:"symbol"`
	Period           string                   `json:"period"`
	AdjustType       string                   `json:"adjust_type"`
	Source           string                   `json:"source"`
	FetchedAt        string                   `json:"fetched_at"`
	IsStale          bool                     `json:"is_stale"`
	RefreshTriggered bool                     `json:"refresh_triggered"`
	Items            []MarketKlineBarResponse `json:"items"`
}

func NewMarketStockDetailResponse(detail *model.StockQuoteDetail, isStale, refreshTriggered bool) *MarketStockDetailResponse {
	concepts := model.NormalizeConceptList(splitConceptString(detail.Concepts))
	return &MarketStockDetailResponse{
		Symbol:           detail.Symbol,
		Name:             detail.Name,
		Market:           detail.Market,
		LastPrice:        detail.LastPrice.String(),
		OpenPrice:        detail.OpenPrice.String(),
		HighPrice:        detail.HighPrice.String(),
		LowPrice:         detail.LowPrice.String(),
		PrevClose:        detail.PrevClose.String(),
		ChangeAmount:     detail.ChangeAmount.String(),
		ChangePercent:    detail.ChangePercent.String(),
		Volume:           detail.Volume.String(),
		Turnover:         detail.Turnover.String(),
		VolumeRatio:      detail.VolumeRatio.String(),
		TurnoverRate:     detail.TurnoverRate.String(),
		Amplitude:        detail.Amplitude.String(),
		LimitUp:          detail.LimitUp.String(),
		LimitDown:        detail.LimitDown.String(),
		AveragePrice:     detail.AveragePrice.String(),
		TotalShares:      detail.TotalShares.String(),
		FloatShares:      detail.FloatShares.String(),
		TotalMarketCap:   detail.TotalMarketCap.String(),
		FloatMarketCap:   detail.FloatMarketCap.String(),
		Industry:         model.NormalizeIndustryLabel(detail.Industry),
		Region:           model.NormalizeRegionLabel(detail.Region),
		Concepts:         concepts,
		Source:           detail.Source,
		FetchedAt:        detail.FetchedAt.Format("2006-01-02 15:04:05"),
		IsStale:          isStale,
		RefreshTriggered: refreshTriggered,
	}
}

func NewMarketStockKlineResponse(symbol, period, adjust string, bars []model.StockKlineBar, isStale, refreshTriggered bool) *MarketStockKlineResponse {
	items := make([]MarketKlineBarResponse, 0, len(bars))
	sortedBars := append([]model.StockKlineBar(nil), bars...)
	sort.Slice(sortedBars, func(i, j int) bool {
		return sortedBars[i].BarTime.Before(sortedBars[j].BarTime)
	})
	latestUpdatedAt := model.StockKlineBar{}
	source := ""
	for _, bar := range sortedBars {
		if bar.UpdatedAt.After(latestUpdatedAt.UpdatedAt) {
			latestUpdatedAt = bar
		}
		if source == "" {
			source = bar.Source
		}
		items = append(items, MarketKlineBarResponse{
			BarTime:       bar.BarTime.Format("2006-01-02 15:04:05"),
			OpenPrice:     bar.OpenPrice.String(),
			ClosePrice:    bar.ClosePrice.String(),
			HighPrice:     bar.HighPrice.String(),
			LowPrice:      bar.LowPrice.String(),
			Volume:        bar.Volume.String(),
			Turnover:      bar.Turnover.String(),
			Amplitude:     bar.Amplitude.String(),
			ChangePercent: bar.ChangePercent.String(),
			ChangeAmount:  bar.ChangeAmount.String(),
			TurnoverRate:  bar.TurnoverRate.String(),
		})
	}
	fetchedAt := ""
	if !latestUpdatedAt.UpdatedAt.IsZero() {
		fetchedAt = latestUpdatedAt.UpdatedAt.Format("2006-01-02 15:04:05")
	}
	return &MarketStockKlineResponse{
		Symbol:           symbol,
		Period:           period,
		AdjustType:       adjust,
		Source:           source,
		FetchedAt:        fetchedAt,
		IsStale:          isStale,
		RefreshTriggered: refreshTriggered,
		Items:            items,
	}
}

func splitConceptString(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}
