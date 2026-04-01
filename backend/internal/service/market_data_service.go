package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/pkg/marketdata"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MarketDataService interface {
	FetchAndStoreMarketSnapshots(ctx context.Context) (string, int, error)
}

type marketDataService struct {
	marketConfig config.MarketConfig
	provider     marketdata.Provider
	snapshotRepo repository.MarketSnapshotRepository
}

func NewMarketDataService(
	marketConfig config.MarketConfig,
	provider marketdata.Provider,
	snapshotRepo repository.MarketSnapshotRepository,
) MarketDataService {
	return &marketDataService{
		marketConfig: marketConfig,
		provider:     provider,
		snapshotRepo: snapshotRepo,
	}
}

func (s *marketDataService) FetchAndStoreMarketSnapshots(ctx context.Context) (string, int, error) {
	symbols := s.symbols()
	if len(symbols) == 0 {
		return "", 0, fmt.Errorf("market symbols are empty")
	}

	quotes, err := s.provider.GetQuotes(ctx, symbols)
	if err != nil {
		return "", 0, err
	}
	if len(quotes) == 0 {
		return "", 0, fmt.Errorf("no quotes returned")
	}

	batchNo := time.Now().Format("20060102150405") + "-" + uuid.NewString()
	snapshots := make([]model.MarketSnapshot, 0, len(quotes))
	for _, quote := range quotes {
		snapshots = append(snapshots, model.MarketSnapshot{
			Symbol:        quote.Symbol,
			Name:          quote.Name,
			Market:        quote.Market,
			SnapshotTime:  quote.SnapshotTime,
			LastPrice:     decimal.NewFromFloat(quote.LastPrice),
			ChangeAmount:  decimal.NewFromFloat(quote.ChangeAmount),
			ChangePercent: decimal.NewFromFloat(quote.ChangePercent),
			OpenPrice:     decimal.NewFromFloat(quote.OpenPrice),
			HighPrice:     decimal.NewFromFloat(quote.HighPrice),
			LowPrice:      decimal.NewFromFloat(quote.LowPrice),
			PrevClose:     decimal.NewFromFloat(quote.PrevClose),
			Volume:        decimal.NewFromFloat(quote.Volume),
			Turnover:      decimal.NewFromFloat(quote.Turnover),
			Source:        quote.Source,
			BatchNo:       batchNo,
		})
	}

	if err := s.snapshotRepo.BatchCreate(snapshots); err != nil {
		return "", 0, err
	}

	return batchNo, len(snapshots), nil
}

func (s *marketDataService) symbols() []string {
	parts := strings.Split(s.marketConfig.Symbols, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		symbol := strings.TrimSpace(part)
		if symbol != "" {
			result = append(result, symbol)
		}
	}
	return result
}
