package marketdata

import (
	"context"
	"errors"
	"fmt"
)

type hybridProvider struct {
	realtime Provider
	detail   Provider
	history  Provider
}

func NewHybridProvider(realtime Provider, history Provider) Provider {
	detail := history
	if detail == nil {
		detail = realtime
	}
	return &hybridProvider{realtime: realtime, detail: detail, history: history}
}

func (p *hybridProvider) GetQuotes(ctx context.Context, symbols []string) ([]Quote, error) {
	if p.realtime == nil {
		return nil, fmt.Errorf("realtime market provider is unavailable")
	}
	return p.realtime.GetQuotes(ctx, symbols)
}

func (p *hybridProvider) GetStockDetail(ctx context.Context, symbol string) (*StockDetail, error) {
	provider := p.detail
	if provider == nil {
		provider = p.realtime
	}
	if provider == nil {
		return nil, fmt.Errorf("realtime market provider is unavailable")
	}
	return provider.GetStockDetail(ctx, symbol)
}

func (p *hybridProvider) GetKlines(ctx context.Context, symbol, period, adjust string, limit int) ([]KlineBar, error) {
	if p.history == nil {
		if p.realtime == nil {
			return nil, fmt.Errorf("historical kline provider is unavailable")
		}
		return p.realtime.GetKlines(ctx, symbol, period, adjust, limit)
	}
	bars, err := p.history.GetKlines(ctx, symbol, period, adjust, limit)
	if err == nil {
		return bars, nil
	}
	if p.realtime == nil || !shouldFallbackToRealtime(err, period) {
		return nil, err
	}
	return p.realtime.GetKlines(ctx, symbol, period, adjust, limit)
}

func shouldFallbackToRealtime(err error, period string) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	normalizedPeriod := normalizePeriod(period)
	switch normalizedPeriod {
	case "1m", "5m", "15m", "30m", "60m":
		return true
	default:
		return false
	}
}
