package service

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"stock-analysis-backend/pkg/news"
)

type StockNewsContext struct {
	Status          string
	Summary         string
	Coverage        string
	Items           []news.Item
	SuccessProviders []string
	FailedProviders  []string
}

type NewsService interface {
	GetStockNews(ctx context.Context, symbol, assetName string) (*StockNewsContext, error)
}

type newsService struct {
	providers []news.Provider
}

func NewNewsService(providers ...news.Provider) NewsService {
	filtered := make([]news.Provider, 0, len(providers))
	for _, provider := range providers {
		if provider != nil {
			filtered = append(filtered, provider)
		}
	}
	return &newsService{providers: filtered}
}

func (s *newsService) GetStockNews(ctx context.Context, symbol, assetName string) (*StockNewsContext, error) {
	if len(s.providers) == 0 {
		return nil, fmt.Errorf("news providers are unavailable")
	}
	keywords := news.BuildKeywords(symbol, assetName)
	type providerResult struct {
		name  string
		items []news.Item
		err   error
	}
	results := make(chan providerResult, len(s.providers))
	var wg sync.WaitGroup
	for _, provider := range s.providers {
		wg.Add(1)
		go func(provider news.Provider) {
			defer wg.Done()
			items, err := provider.FetchByStock(ctx, symbol, assetName)
			results <- providerResult{name: provider.Name(), items: items, err: err}
		}(provider)
	}
	wg.Wait()
	close(results)

	merged := make([]news.Item, 0, 24)
	successProviders := make([]string, 0, len(s.providers))
	failedProviders := make([]string, 0, len(s.providers))
	for result := range results {
		if result.err != nil || len(result.items) == 0 {
			failedProviders = append(failedProviders, result.name)
			continue
		}
		successProviders = append(successProviders, result.name)
		merged = append(merged, result.items...)
	}
	merged = news.MergeAndRankItems(merged, keywords, 8)
	if len(merged) == 0 {
		return nil, fmt.Errorf("failed to fetch relevant news for %s", strings.TrimSpace(symbol))
	}
	status := "complete"
	if len(successProviders) < len(s.providers) {
		status = "partial"
	}
	return &StockNewsContext{
		Status:           status,
		Summary:          news.BuildSummary(merged),
		Coverage:         news.BuildCoverage(successProviders, failedProviders, len(merged)),
		Items:            merged,
		SuccessProviders: successProviders,
		FailedProviders:  failedProviders,
	}, nil
}
