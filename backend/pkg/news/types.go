package news

import (
	"context"
	"time"
)

type Item struct {
	Title       string
	Summary     string
	Source      string
	URL         string
	PublishedAt time.Time
	Provider    string
	Score       int
}

type Provider interface {
	Name() string
	FetchByStock(ctx context.Context, symbol, assetName string) ([]Item, error)
}
