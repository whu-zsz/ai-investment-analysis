package marketdata

import "context"

type Provider interface {
	GetQuotes(ctx context.Context, symbols []string) ([]Quote, error)
}

