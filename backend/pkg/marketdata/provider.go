package marketdata

import "context"

type Provider interface {
	GetQuotes(ctx context.Context, symbols []string) ([]Quote, error)
	GetStockDetail(ctx context.Context, symbol string) (*StockDetail, error)
	GetKlines(ctx context.Context, symbol, period, adjust string, limit int) ([]KlineBar, error)
}
