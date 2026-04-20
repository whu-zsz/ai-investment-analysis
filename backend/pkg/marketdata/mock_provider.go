package marketdata

import (
	"context"
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
