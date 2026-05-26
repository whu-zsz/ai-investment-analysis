package main

import (
	"context"
	"fmt"

	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/internal/service"
	"stock-analysis-backend/pkg/marketdata"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}

	db, err := config.InitDB(&cfg.Database)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = config.CloseDB(db)
	}()

	snapshotRepo := repository.NewMarketSnapshotRepository(db)
	boardRepo := repository.NewMarketBoardSnapshotRepository(db)
	constituentRepo := repository.NewMarketBoardConstituentRepository(db)
	klineRepo := repository.NewStockKlineRepository(db)
	rankingProvider := marketdata.NewRankingProvider(cfg.Market)

	var provider marketdata.Provider
	if cfg.Market.Enabled {
		provider, err = marketdata.NewProvider(cfg.Market)
		if err != nil {
			panic(err)
		}
	}

	svc := service.NewMarketDataService(cfg.Market, provider, rankingProvider, snapshotRepo, boardRepo, constituentRepo, klineRepo)
	batchNo, count, err := svc.FetchAndStoreMarketBoardSnapshots(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Printf("batch=%s count=%d\n", batchNo, count)
}
