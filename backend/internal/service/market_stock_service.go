package service

import (
	"context"
	"errors"
	"strings"

	marketResponse "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/pkg/marketdata"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type MarketStockService interface {
	GetStockDetail(symbol string, forceRefresh bool) (*marketResponse.MarketStockDetailResponse, error)
	GetStockKlines(symbol, period, adjust string, limit int, forceRefresh bool) (*marketResponse.MarketStockKlineResponse, error)
}

type marketStockService struct {
	provider    marketdata.Provider
	detailRepo  repository.StockQuoteDetailRepository
	klineRepo   repository.StockKlineRepository
}

func NewMarketStockService(
	provider marketdata.Provider,
	detailRepo repository.StockQuoteDetailRepository,
	klineRepo repository.StockKlineRepository,
) MarketStockService {
	return &marketStockService{
		provider:   provider,
		detailRepo: detailRepo,
		klineRepo:  klineRepo,
	}
}

func (s *marketStockService) GetStockDetail(symbol string, forceRefresh bool) (*marketResponse.MarketStockDetailResponse, error) {
	normalized := normalizeSymbol(symbol)
	if normalized == "" {
		return nil, errors.New("symbol is required")
	}

	cached, err := s.detailRepo.FindBySymbol(normalized)
	if err == nil && cached != nil && !forceRefresh {
		return marketResponse.NewMarketStockDetailResponse(cached, false, false), nil
	}

	if s.provider == nil {
		if err == nil && cached != nil {
			return marketResponse.NewMarketStockDetailResponse(cached, true, false), nil
		}
		return nil, errors.New("market provider is unavailable")
	}

	fetched, fetchErr := s.provider.GetStockDetail(context.Background(), normalized)
	if fetchErr != nil {
		if err == nil && cached != nil {
			return marketResponse.NewMarketStockDetailResponse(cached, true, false), nil
		}
		return nil, fetchErr
	}

	entity := convertStockDetailToModel(fetched)
	if saveErr := s.detailRepo.Upsert(entity); saveErr != nil {
		return nil, saveErr
	}
	return marketResponse.NewMarketStockDetailResponse(entity, false, true), nil
}

func (s *marketStockService) GetStockKlines(symbol, period, adjust string, limit int, forceRefresh bool) (*marketResponse.MarketStockKlineResponse, error) {
	normalized := normalizeSymbol(symbol)
	if normalized == "" {
		return nil, errors.New("symbol is required")
	}
	normalizedPeriod := normalizeKlinePeriod(period)
	normalizedAdjust := normalizeAdjustType(adjust)
	if limit <= 0 {
		limit = 120
	}

	cachedBars, cachedErr := s.klineRepo.FindBars(normalized, normalizedPeriod, normalizedAdjust, limit)
	if cachedErr == nil && !forceRefresh && len(cachedBars) >= limit {
		return marketResponse.NewMarketStockKlineResponse(normalized, normalizedPeriod, normalizedAdjust, cachedBars, false, false), nil
	}

	if s.provider == nil {
		if cachedErr == nil && len(cachedBars) > 0 {
			return marketResponse.NewMarketStockKlineResponse(normalized, normalizedPeriod, normalizedAdjust, cachedBars, true, false), nil
		}
		return nil, errors.New("market provider is unavailable")
	}

	fetchedBars, fetchErr := s.provider.GetKlines(context.Background(), normalized, normalizedPeriod, normalizedAdjust, limit)
	if fetchErr != nil {
		if cachedErr == nil && len(cachedBars) > 0 {
			return marketResponse.NewMarketStockKlineResponse(normalized, normalizedPeriod, normalizedAdjust, cachedBars, true, false), nil
		}
		return nil, fetchErr
	}

	entities := convertKlinesToModels(fetchedBars)
	if saveErr := s.klineRepo.UpsertBars(entities); saveErr != nil {
		return nil, saveErr
	}
	return marketResponse.NewMarketStockKlineResponse(normalized, normalizedPeriod, normalizedAdjust, entities, false, true), nil
}

func convertStockDetailToModel(detail *marketdata.StockDetail) *model.StockQuoteDetail {
	changeAmount := detail.ChangeAmount
	changePercent := detail.ChangePercent
	if changeAmount == 0 && detail.LastPrice != 0 && detail.PrevClose != 0 {
		changeAmount = detail.LastPrice - detail.PrevClose
	}
	if changePercent == 0 && detail.PrevClose != 0 {
		changePercent = changeAmount / detail.PrevClose * 100
	}
	return &model.StockQuoteDetail{
		Symbol:         detail.Symbol,
		Name:           fallbackString(detail.Name, detail.Symbol),
		Market:         fallbackString(detail.Market, marketFromNormalizedSymbol(detail.Symbol)),
		LastPrice:      decimal.NewFromFloat(detail.LastPrice),
		OpenPrice:      decimal.NewFromFloat(detail.OpenPrice),
		HighPrice:      decimal.NewFromFloat(detail.HighPrice),
		LowPrice:       decimal.NewFromFloat(detail.LowPrice),
		PrevClose:      decimal.NewFromFloat(detail.PrevClose),
		ChangeAmount:   decimal.NewFromFloat(changeAmount),
		ChangePercent:  decimal.NewFromFloat(changePercent),
		Volume:         decimal.NewFromFloat(detail.Volume),
		Turnover:       decimal.NewFromFloat(detail.Turnover),
		VolumeRatio:    decimal.NewFromFloat(detail.VolumeRatio),
		TurnoverRate:   decimal.NewFromFloat(detail.TurnoverRate),
		Amplitude:      decimal.NewFromFloat(detail.Amplitude),
		LimitUp:        decimal.NewFromFloat(detail.LimitUp),
		LimitDown:      decimal.NewFromFloat(detail.LimitDown),
		AveragePrice:   decimal.NewFromFloat(detail.AveragePrice),
		TotalShares:    decimal.NewFromFloat(detail.TotalShares),
		FloatShares:    decimal.NewFromFloat(detail.FloatShares),
		TotalMarketCap: decimal.NewFromFloat(detail.TotalMarketCap),
		FloatMarketCap: decimal.NewFromFloat(detail.FloatMarketCap),
		Industry:       model.NormalizeIndustryLabel(detail.Industry),
		Region:         model.NormalizeRegionLabel(detail.Region),
		Concepts:       strings.Join(model.NormalizeConceptList(detail.Concepts), ","),
		Source:         detail.Source,
		FetchedAt:      detail.FetchedAt,
	}
}

func convertKlinesToModels(bars []marketdata.KlineBar) []model.StockKlineBar {
	entities := make([]model.StockKlineBar, 0, len(bars))
	for _, bar := range bars {
		entities = append(entities, model.StockKlineBar{
			Symbol:        bar.Symbol,
			Period:        normalizeKlinePeriod(bar.Period),
			AdjustType:    normalizeAdjustType(bar.AdjustType),
			BarTime:       bar.BarTime,
			OpenPrice:     decimal.NewFromFloat(bar.Open),
			ClosePrice:    decimal.NewFromFloat(bar.Close),
			HighPrice:     decimal.NewFromFloat(bar.High),
			LowPrice:      decimal.NewFromFloat(bar.Low),
			Volume:        decimal.NewFromFloat(bar.Volume),
			Turnover:      decimal.NewFromFloat(bar.Amount),
			Amplitude:     decimal.NewFromFloat(bar.Amplitude),
			ChangePercent: decimal.NewFromFloat(bar.ChangePercent),
			ChangeAmount:  decimal.NewFromFloat(bar.ChangeAmount),
			TurnoverRate:  decimal.NewFromFloat(bar.TurnoverRate),
			Source:        bar.Source,
		})
	}
	return entities
}

func normalizeKlinePeriod(period string) string {
	switch strings.ToLower(strings.TrimSpace(period)) {
	case "1m", "5m", "15m", "30m", "60m", "week", "month":
		return strings.ToLower(strings.TrimSpace(period))
	case "", "day", "daily", "d":
		return "day"
	default:
		return strings.ToLower(strings.TrimSpace(period))
	}
}

func normalizeAdjustType(adjust string) string {
	switch strings.ToLower(strings.TrimSpace(adjust)) {
	case "", "qfq", "forward":
		return "qfq"
	case "hfq", "backward":
		return "hfq"
	case "none", "raw":
		return "none"
	default:
		return strings.ToLower(strings.TrimSpace(adjust))
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func marketFromNormalizedSymbol(symbol string) string {
	if strings.HasPrefix(symbol, "399") || strings.HasPrefix(symbol, "000300") || strings.HasPrefix(symbol, "000001") {
		return "cn_index"
	}
	return "cn_stock"
}

var _ = gorm.ErrRecordNotFound
