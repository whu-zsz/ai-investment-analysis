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
		s.warmup(ctx)
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

func (s *marketScheduler) warmup(ctx context.Context) {
	if err := s.marketDataService.EnsureTrackedIndexHistory(ctx); err != nil {
		s.logger.Warn("market index history warmup failed", zap.Error(err))
		return
	}
	s.logger.Info("market index history warmup finished")
}

func (s *marketScheduler) runOnce(ctx context.Context) {
	batchNo, count, err := s.marketDataService.FetchAndStoreMarketSnapshots(ctx)
	if err != nil {
		s.logger.Warn("tracked market snapshot fetch failed", zap.Error(err))
	} else {
		s.logger.Info("tracked market snapshot fetch succeeded", zap.String("batch_no", batchNo), zap.Int("count", count))
	}

	fullBatchNo, fullCount, fullErr := s.marketDataService.FetchAndStoreFullMarketSnapshots(ctx)
	if fullErr != nil {
		s.logger.Warn("full market snapshot fetch failed", zap.Error(fullErr))
	} else {
		s.logger.Info("full market snapshot fetch succeeded", zap.String("batch_no", fullBatchNo), zap.Int("count", fullCount))
	}

	boardBatchNo, boardCount, boardErr := s.marketDataService.FetchAndStoreMarketBoardSnapshots(ctx)
	if boardErr != nil {
		s.logger.Warn("market board snapshot fetch failed", zap.Error(boardErr))
		return
	}

	s.logger.Info("market board snapshot fetch succeeded", zap.String("batch_no", boardBatchNo), zap.Int("count", boardCount))
}
