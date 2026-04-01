package service

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type MarketScheduler interface {
	Start(ctx context.Context)
}

type marketScheduler struct {
	interval          time.Duration
	marketDataService MarketDataService
	logger            *zap.Logger
}

func NewMarketScheduler(interval time.Duration, marketDataService MarketDataService, logger *zap.Logger) MarketScheduler {
	if interval <= 0 {
		interval = time.Minute
	}
	return &marketScheduler{
		interval:          interval,
		marketDataService: marketDataService,
		logger:            logger,
	}
}

func (s *marketScheduler) Start(ctx context.Context) {
	go func() {
		s.runOnce(ctx)
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				s.logger.Info("market scheduler stopped")
				return
			case <-ticker.C:
				s.runOnce(ctx)
			}
		}
	}()
}

func (s *marketScheduler) runOnce(ctx context.Context) {
	batchNo, count, err := s.marketDataService.FetchAndStoreMarketSnapshots(ctx)
	if err != nil {
		s.logger.Warn("market snapshot fetch failed", zap.Error(err))
		return
	}

	s.logger.Info("market snapshot fetch succeeded", zap.String("batch_no", batchNo), zap.Int("count", count))
}
