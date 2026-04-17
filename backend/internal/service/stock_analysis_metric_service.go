package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"

	"github.com/shopspring/decimal"
)

type StockAnalysisMetricService interface {
	PrepareMetrics(ctx context.Context, userID uint64, taskID *uint64, start, end time.Time, symbols []string, forceRefreshMarket bool, forceRefreshMetrics bool) ([]model.StockAnalysisMetric, error)
}

type stockAnalysisMetricService struct {
	metricRepo         repository.StockAnalysisMetricRepository
	transactionRepo    repository.TransactionRepository
	marketSnapshotRepo repository.MarketSnapshotRepository
	marketDataService  MarketDataService
}

type metricAggregate struct {
	Symbol               string
	AssetName            string
	TradeCount           int
	BuyCount             int
	SellCount            int
	BuyQuantity          decimal.Decimal
	SellQuantity         decimal.Decimal
	BuyAmount            decimal.Decimal
	SellAmount           decimal.Decimal
	NetQuantity          decimal.Decimal
	RealizedProfit       decimal.Decimal
	RealizedProfitRate   decimal.Decimal
	EndingPositionQty    decimal.Decimal
	EndingAvgCost        decimal.Decimal
	LatestPrice          decimal.Decimal
	LatestMarketValue    decimal.Decimal
	UnrealizedProfit     decimal.Decimal
	UnrealizedProfitRate decimal.Decimal
	TotalProfit          decimal.Decimal
	TotalProfitRate      decimal.Decimal
	PeriodStartPrice     decimal.Decimal
	PeriodEndPrice       decimal.Decimal
	PeriodPriceChangePct decimal.Decimal
	PeriodHighPrice      decimal.Decimal
	PeriodLowPrice       decimal.Decimal
	Market               string
	MarketDataStatus     string
}

func NewStockAnalysisMetricService(
	metricRepo repository.StockAnalysisMetricRepository,
	transactionRepo repository.TransactionRepository,
	marketSnapshotRepo repository.MarketSnapshotRepository,
	marketDataService MarketDataService,
) StockAnalysisMetricService {
	return &stockAnalysisMetricService{
		metricRepo:         metricRepo,
		transactionRepo:    transactionRepo,
		marketSnapshotRepo: marketSnapshotRepo,
		marketDataService:  marketDataService,
	}
}

func (s *stockAnalysisMetricService) PrepareMetrics(ctx context.Context, userID uint64, taskID *uint64, start, end time.Time, symbols []string, forceRefreshMarket bool, forceRefreshMetrics bool) ([]model.StockAnalysisMetric, error) {
	normalizedSymbols := normalizeSymbols(symbols)
	if !forceRefreshMetrics {
		metrics, err := s.metricRepo.FindByUserPeriod(userID, start, end, normalizedSymbols)
		if err == nil {
			if len(normalizedSymbols) == 0 && len(metrics) > 0 {
				return metrics, nil
			}
			if len(normalizedSymbols) > 0 && len(metrics) == len(normalizedSymbols) {
				return metrics, nil
			}
		}
	}

	transactions, err := s.transactionRepo.FindByDateRange(userID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	aggregates := aggregateMetricTransactions(transactions)
	if len(normalizedSymbols) > 0 {
		filter := make(map[string]struct{}, len(normalizedSymbols))
		for _, symbol := range normalizedSymbols {
			filter[symbol] = struct{}{}
		}
		for symbol := range aggregates {
			if _, ok := filter[symbol]; !ok {
				delete(aggregates, symbol)
			}
		}
	}
	if len(aggregates) == 0 {
		return []model.StockAnalysisMetric{}, nil
	}

	if err := s.fillMetricMarketData(ctx, aggregates, start, end, forceRefreshMarket); err != nil {
		return nil, err
	}

	metrics := make([]model.StockAnalysisMetric, 0, len(aggregates))
	symbolKeys := make([]string, 0, len(aggregates))
	for symbol := range aggregates {
		symbolKeys = append(symbolKeys, symbol)
	}
	sort.Strings(symbolKeys)
	now := time.Now()
	for _, symbol := range symbolKeys {
		agg := aggregates[symbol]
		metric := model.StockAnalysisMetric{
			TaskID:               taskID,
			UserID:               userID,
			Symbol:               symbol,
			AssetName:            agg.AssetName,
			Market:               agg.Market,
			PeriodStart:          start,
			PeriodEnd:            end,
			TradeCount:           agg.TradeCount,
			BuyCount:             agg.BuyCount,
			SellCount:            agg.SellCount,
			BuyQuantity:          agg.BuyQuantity,
			SellQuantity:         agg.SellQuantity,
			BuyAmount:            agg.BuyAmount,
			SellAmount:           agg.SellAmount,
			NetQuantity:          agg.NetQuantity,
			RealizedProfit:       agg.RealizedProfit,
			RealizedProfitRate:   agg.RealizedProfitRate,
			EndingPositionQty:    agg.EndingPositionQty,
			EndingAvgCost:        agg.EndingAvgCost,
			LatestPrice:          agg.LatestPrice,
			LatestMarketValue:    agg.LatestMarketValue,
			UnrealizedProfit:     agg.UnrealizedProfit,
			UnrealizedProfitRate: agg.UnrealizedProfitRate,
			TotalProfit:          agg.TotalProfit,
			TotalProfitRate:      agg.TotalProfitRate,
			PeriodStartPrice:     agg.PeriodStartPrice,
			PeriodEndPrice:       agg.PeriodEndPrice,
			PeriodPriceChangePct: agg.PeriodPriceChangePct,
			PeriodHighPrice:      agg.PeriodHighPrice,
			PeriodLowPrice:       agg.PeriodLowPrice,
			MarketDataStatus:     agg.MarketDataStatus,
			SourceType:           "on_demand",
			ComputedAt:           now,
		}
		metrics = append(metrics, metric)
	}
	if err := s.metricRepo.BatchUpsert(metrics); err != nil {
		return nil, err
	}
	return s.metricRepo.FindByUserPeriod(userID, start, end, normalizedSymbols)
}

func aggregateMetricTransactions(transactions []model.Transaction) map[string]*metricAggregate {
	result := make(map[string]*metricAggregate)
	for _, tx := range transactions {
		symbol := normalizeSymbol(tx.AssetCode)
		if symbol == "" {
			continue
		}
		agg, ok := result[symbol]
		if !ok {
			agg = &metricAggregate{
				Symbol:               symbol,
				AssetName:            tx.AssetName,
				BuyQuantity:          modelDecimalZero(),
				SellQuantity:         modelDecimalZero(),
				BuyAmount:            modelDecimalZero(),
				SellAmount:           modelDecimalZero(),
				NetQuantity:          modelDecimalZero(),
				RealizedProfit:       modelDecimalZero(),
				RealizedProfitRate:   modelDecimalZero(),
				EndingPositionQty:    modelDecimalZero(),
				EndingAvgCost:        modelDecimalZero(),
				LatestPrice:          modelDecimalZero(),
				LatestMarketValue:    modelDecimalZero(),
				UnrealizedProfit:     modelDecimalZero(),
				UnrealizedProfitRate: modelDecimalZero(),
				TotalProfit:          modelDecimalZero(),
				TotalProfitRate:      modelDecimalZero(),
				PeriodStartPrice:     modelDecimalZero(),
				PeriodEndPrice:       modelDecimalZero(),
				PeriodPriceChangePct: modelDecimalZero(),
				PeriodHighPrice:      modelDecimalZero(),
				PeriodLowPrice:       modelDecimalZero(),
				MarketDataStatus:     marketDataStatusUnavailable,
			}
			result[symbol] = agg
		}
		agg.TradeCount++
		switch tx.TransactionType {
		case "buy":
			agg.BuyCount++
			agg.BuyQuantity = agg.BuyQuantity.Add(tx.Quantity)
			agg.BuyAmount = agg.BuyAmount.Add(tx.TotalAmount)
			agg.NetQuantity = agg.NetQuantity.Add(tx.Quantity)
		case "sell":
			agg.SellCount++
			agg.SellQuantity = agg.SellQuantity.Add(tx.Quantity)
			agg.SellAmount = agg.SellAmount.Add(tx.TotalAmount)
			agg.NetQuantity = agg.NetQuantity.Sub(tx.Quantity)
		}
		if tx.Profit != nil {
			agg.RealizedProfit = agg.RealizedProfit.Add(*tx.Profit)
		}
	}
	for _, agg := range result {
		agg.EndingPositionQty = agg.NetQuantity
		if !agg.BuyQuantity.IsZero() {
			agg.EndingAvgCost = agg.BuyAmount.Div(agg.BuyQuantity)
		}
		if !agg.BuyAmount.IsZero() {
			agg.RealizedProfitRate = agg.RealizedProfit.Div(agg.BuyAmount).Mul(decimal.NewFromInt(100))
		}
	}
	return result
}

func (s *stockAnalysisMetricService) fillMetricMarketData(ctx context.Context, aggregates map[string]*metricAggregate, startTime, endTime time.Time, forceRefreshMarket bool) error {
	start := startTime
	end := endTime.Add(24*time.Hour - time.Second)
	missing := make([]string, 0)

	for symbol, aggregate := range aggregates {
		history, err := s.marketSnapshotRepo.FindHistoryBySymbol(symbol, 300, &start, &end)
		if err != nil && !errorsIsRecordNotFound(err) {
			return err
		}
		if len(history) == 0 || forceRefreshMarket {
			missing = append(missing, symbol)
			if len(history) == 0 {
				aggregate.MarketDataStatus = marketDataStatusUnavailable
			}
			continue
		}
		applyMetricMarketHistory(aggregate, history, marketDataStatusComplete)
	}

	if len(missing) > 0 {
		quotes, err := s.marketDataService.FetchAndStoreQuotesBySymbols(ctx, missing)
		if err == nil {
			quoteMap := make(map[string]model.MarketSnapshot, len(quotes))
			for _, snapshot := range quotes {
				quoteMap[strings.ToUpper(snapshot.Symbol)] = snapshot
			}
			for _, symbol := range missing {
				agg := aggregates[symbol]
				if snapshot, ok := quoteMap[strings.ToUpper(symbol)]; ok {
					if agg.PeriodStartPrice.IsZero() {
						agg.PeriodStartPrice = snapshot.LastPrice
					}
					agg.PeriodEndPrice = snapshot.LastPrice
					agg.PeriodHighPrice = snapshot.HighPrice
					if agg.PeriodLowPrice.IsZero() || snapshot.LowPrice.LessThan(agg.PeriodLowPrice) {
						agg.PeriodLowPrice = snapshot.LowPrice
					}
					agg.LatestPrice = snapshot.LastPrice
					agg.Market = snapshot.Market
					agg.MarketDataStatus = marketDataStatusFetchedLive
				}
			}
		}
	}

	for _, agg := range aggregates {
		if agg.MarketDataStatus == "" {
			agg.MarketDataStatus = marketDataStatusUnavailable
		}
		if !agg.LatestPrice.IsZero() {
			agg.LatestMarketValue = agg.EndingPositionQty.Mul(agg.LatestPrice)
			if !agg.EndingAvgCost.IsZero() {
				agg.UnrealizedProfit = agg.LatestPrice.Sub(agg.EndingAvgCost).Mul(agg.EndingPositionQty)
				agg.UnrealizedProfitRate = agg.LatestPrice.Sub(agg.EndingAvgCost).Div(agg.EndingAvgCost).Mul(decimal.NewFromInt(100))
			}
		}
		agg.TotalProfit = agg.RealizedProfit.Add(agg.UnrealizedProfit)
		if !agg.BuyAmount.IsZero() {
			agg.TotalProfitRate = agg.TotalProfit.Div(agg.BuyAmount).Mul(decimal.NewFromInt(100))
		}
		if !agg.PeriodStartPrice.IsZero() && !agg.PeriodEndPrice.IsZero() {
			agg.PeriodPriceChangePct = agg.PeriodEndPrice.Sub(agg.PeriodStartPrice).Div(agg.PeriodStartPrice).Mul(decimal.NewFromInt(100))
		}
	}
	return nil
}

func applyMetricMarketHistory(aggregate *metricAggregate, history []model.MarketSnapshot, status string) {
	if len(history) == 0 {
		return
	}
	sort.Slice(history, func(i, j int) bool {
		return history[i].SnapshotTime.Before(history[j].SnapshotTime)
	})
	first := history[0]
	last := history[len(history)-1]
	aggregate.PeriodStartPrice = first.LastPrice
	aggregate.PeriodEndPrice = last.LastPrice
	aggregate.LatestPrice = last.LastPrice
	aggregate.Market = last.Market
	aggregate.MarketDataStatus = status
	aggregate.PeriodHighPrice = first.HighPrice
	aggregate.PeriodLowPrice = first.LowPrice
	for _, snapshot := range history {
		if aggregate.PeriodHighPrice.IsZero() || snapshot.HighPrice.GreaterThan(aggregate.PeriodHighPrice) {
			aggregate.PeriodHighPrice = snapshot.HighPrice
		}
		if aggregate.PeriodLowPrice.IsZero() || snapshot.LowPrice.LessThan(aggregate.PeriodLowPrice) {
			aggregate.PeriodLowPrice = snapshot.LowPrice
		}
	}
}
