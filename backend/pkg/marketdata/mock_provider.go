package marketdata

import (
	"context"
	"fmt"
	"math"
	"time"
)

type mockProvider struct{}

func NewMockProvider() Provider {
	return &mockProvider{}
}

func (p *mockProvider) GetQuotes(ctx context.Context, symbols []string) ([]Quote, error) {
	_ = ctx
	now := time.Now().Truncate(time.Minute)
	quotes := make([]Quote, 0, len(symbols))
	for i, symbol := range symbols {
		base := 3000.0 + float64(i)*420.0
		wave := math.Sin(float64(now.Unix()/60)+float64(i)) * 12
		lastPrice := base + wave
		prevClose := lastPrice - (float64(i+1)*3.2 - 5)
		changeAmount := lastPrice - prevClose
		changePercent := 0.0
		if prevClose != 0 {
			changePercent = changeAmount / prevClose * 100
		}

		quotes = append(quotes, Quote{
			Symbol:        symbol,
			Name:          DefaultName(symbol),
			Market:        "cn_index",
			SnapshotTime:  now,
			LastPrice:     Round(lastPrice),
			ChangeAmount:  Round(changeAmount),
			ChangePercent: Round(changePercent),
			OpenPrice:     Round(lastPrice - 6.4),
			HighPrice:     Round(lastPrice + 8.2),
			LowPrice:      Round(lastPrice - 10.1),
			PrevClose:     Round(prevClose),
			Volume:        100000000 + float64((i+1)*28000000),
			Turnover:      200000000000 + float64((i+1)*50000000000),
			Source:        "mock",
		})
	}
	return quotes, nil
}

func (p *mockProvider) GetStockDetail(ctx context.Context, symbol string) (*StockDetail, error) {
	_ = ctx
	normalized := normalizeProviderSymbol(symbol)
	if normalized == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	now := time.Now().Truncate(time.Minute)
	base := 88.35 + math.Sin(float64(now.Unix()/300))*3
	prevClose := base - 1.2
	changeAmount := base - prevClose
	changePercent := 0.0
	if prevClose != 0 {
		changePercent = changeAmount / prevClose * 100
	}
	return &StockDetail{
		Symbol:         normalized,
		Name:           DefaultName(normalized),
		Market:         "cn_stock",
		LastPrice:      Round(base),
		OpenPrice:      Round(base - 0.5),
		HighPrice:      Round(base + 1.4),
		LowPrice:       Round(base - 1.1),
		PrevClose:      Round(prevClose),
		ChangeAmount:   Round(changeAmount),
		ChangePercent:  Round(changePercent),
		Volume:         253000,
		Turnover:       2100000000,
		VolumeRatio:    0.82,
		TurnoverRate:   1.63,
		Amplitude:      2.54,
		LimitUp:        Round(prevClose * 1.1),
		LimitDown:      Round(prevClose * 0.9),
		AveragePrice:   Round(base - 0.18),
		TotalShares:    3881608005,
		FloatShares:    3881513391,
		TotalMarketCap: 326171520660.15,
		FloatMarketCap: 326163570245.73,
		Industry:       "食品饮料",
		Region:         "四川",
		Concepts:       []string{"白酒", "消费", "价值蓝筹"},
		Source:         "mock",
		FetchedAt:      now,
	}, nil
}

func (p *mockProvider) GetKlines(ctx context.Context, symbol, period, adjust string, limit int) ([]KlineBar, error) {
	_ = ctx
	normalized := normalizeProviderSymbol(symbol)
	if normalized == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if limit <= 0 {
		limit = 60
	}
	if period == "" {
		period = "day"
	}
	if adjust == "" {
		adjust = "qfq"
	}

	bars := make([]KlineBar, 0, limit)
	start := time.Now().AddDate(0, 0, -limit+1)
	base := 82.0
	for i := 0; i < limit; i++ {
		wave := math.Sin(float64(i)/3.5) * 2.2
		open := base + wave + float64(i)*0.08
		close := open + math.Sin(float64(i)/2.1)*1.1
		high := math.Max(open, close) + 0.8
		low := math.Min(open, close) - 0.7
		changeAmount := close - open
		changePercent := 0.0
		if open != 0 {
			changePercent = changeAmount / open * 100
		}
		bars = append(bars, KlineBar{
			Symbol:        normalized,
			Period:        period,
			AdjustType:    adjust,
			BarTime:       start.AddDate(0, 0, i),
			Open:          Round(open),
			Close:         Round(close),
			High:          Round(high),
			Low:           Round(low),
			Volume:        180000 + float64(i*1300),
			Amount:        150000000 + float64(i*2400000),
			Amplitude:     Round((high-low)/open*100),
			ChangePercent: Round(changePercent),
			ChangeAmount:  Round(changeAmount),
			TurnoverRate:  Round(0.6 + float64(i%7)*0.1),
			Source:        "mock",
		})
	}
	return bars, nil
}
