package service

import (
	"errors"
	"sort"
	"time"

	marketResponse "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"

	"gorm.io/gorm"
)

type MarketSnapshotService interface {
	GetLatestSnapshots() ([]marketResponse.MarketSnapshotResponse, error)
	GetHistory(symbol string, limit int, startTime, endTime *time.Time) ([]marketResponse.MarketSnapshotResponse, error)
	GetDashboardSnapshot() (*marketResponse.DashboardMarketSnapshotResponse, error)
}

type marketSnapshotService struct {
	snapshotRepo repository.MarketSnapshotRepository
}

func NewMarketSnapshotService(snapshotRepo repository.MarketSnapshotRepository) MarketSnapshotService {
	return &marketSnapshotService{snapshotRepo: snapshotRepo}
}

func (s *marketSnapshotService) GetLatestSnapshots() ([]marketResponse.MarketSnapshotResponse, error) {
	batchNo, err := s.snapshotRepo.FindLatestBatchNo()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []marketResponse.MarketSnapshotResponse{}, nil
		}
		return nil, err
	}

	snapshots, err := s.snapshotRepo.FindByBatchNo(batchNo)
	if err != nil {
		return nil, err
	}

	return convertSnapshots(snapshots), nil
}

func (s *marketSnapshotService) GetHistory(symbol string, limit int, startTime, endTime *time.Time) ([]marketResponse.MarketSnapshotResponse, error) {
	var (
		snapshots []model.MarketSnapshot
		err       error
	)

	if symbol == "" {
		snapshots, err = s.snapshotRepo.FindHistory(limit, startTime, endTime)
	} else {
		snapshots, err = s.snapshotRepo.FindHistoryBySymbol(symbol, limit, startTime, endTime)
	}
	if err != nil {
		return nil, err
	}
	return convertSnapshots(snapshots), nil
}

func (s *marketSnapshotService) GetDashboardSnapshot() (*marketResponse.DashboardMarketSnapshotResponse, error) {
	batchNo, err := s.snapshotRepo.FindLatestBatchNo()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &marketResponse.DashboardMarketSnapshotResponse{
				Indices: []marketResponse.MarketIndexItemResponse{},
				Stats:   []marketResponse.DashboardStatResponse{},
			}, nil
		}
		return nil, err
	}

	snapshots, err := s.snapshotRepo.FindByBatchNo(batchNo)
	if err != nil {
		return nil, err
	}
	if len(snapshots) == 0 {
		return &marketResponse.DashboardMarketSnapshotResponse{
			Indices: []marketResponse.MarketIndexItemResponse{},
			Stats:   []marketResponse.DashboardStatResponse{},
		}, nil
	}

	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].SnapshotTime.Before(snapshots[j].SnapshotTime)
	})

	latest := snapshots[len(snapshots)-1]
	indices := make([]marketResponse.MarketIndexItemResponse, 0, len(snapshots))
	for _, snapshot := range snapshots {
		indices = append(indices, marketResponse.MarketIndexItemResponse{
			Symbol:        snapshot.Symbol,
			Name:          snapshot.Name,
			LastPrice:     snapshot.LastPrice.String(),
			ChangeAmount:  snapshot.ChangeAmount.String(),
			ChangePercent: snapshot.ChangePercent.String(),
		})
	}

	chartSeries := make([]marketResponse.MarketChartPoint, 0, len(snapshots))
	for _, snapshot := range snapshots {
		chartSeries = append(chartSeries, marketResponse.MarketChartPoint{
			Label: snapshot.Name,
			Value: snapshot.LastPrice.String(),
		})
	}

	stats := []marketResponse.DashboardStatResponse{
		{Label: "指数数量", Value: formatInt(len(snapshots))},
		{Label: "上涨数", Value: formatInt(countPositive(snapshots))},
		{Label: "下跌数", Value: formatInt(countNegative(snapshots))},
		{Label: "平均涨跌幅", Value: averageChangePercent(snapshots)},
		{Label: "总成交额", Value: totalTurnover(snapshots)},
	}

	return &marketResponse.DashboardMarketSnapshotResponse{
		SnapshotTime: latest.SnapshotTime.Format("2006-01-02 15:04:05"),
		IsStale:      time.Since(latest.SnapshotTime) > 2*time.Minute,
		Source:       latest.Source,
		Indices:      indices,
		MainChart: marketResponse.MarketChartResponse{
			IndexName: snapshots[0].Name,
			Series:    chartSeries,
		},
		Stats: stats,
	}, nil
}

func convertSnapshots(snapshots []model.MarketSnapshot) []marketResponse.MarketSnapshotResponse {
	responses := make([]marketResponse.MarketSnapshotResponse, 0, len(snapshots))
	for _, snapshot := range snapshots {
		responses = append(responses, marketResponse.MarketSnapshotResponse{
			Symbol:        snapshot.Symbol,
			Name:          snapshot.Name,
			Market:        snapshot.Market,
			SnapshotTime:  snapshot.SnapshotTime.Format("2006-01-02 15:04:05"),
			LastPrice:     snapshot.LastPrice.String(),
			ChangeAmount:  snapshot.ChangeAmount.String(),
			ChangePercent: snapshot.ChangePercent.String(),
			OpenPrice:     snapshot.OpenPrice.String(),
			HighPrice:     snapshot.HighPrice.String(),
			LowPrice:      snapshot.LowPrice.String(),
			PrevClose:     snapshot.PrevClose.String(),
			Volume:        snapshot.Volume.String(),
			Turnover:      snapshot.Turnover.String(),
			Source:        snapshot.Source,
			BatchNo:       snapshot.BatchNo,
		})
	}
	return responses
}

func countPositive(snapshots []model.MarketSnapshot) int {
	count := 0
	for _, snapshot := range snapshots {
		if snapshot.ChangeAmount.GreaterThanOrEqual(modelDecimalZero()) {
			count++
		}
	}
	return count
}

func countNegative(snapshots []model.MarketSnapshot) int {
	count := 0
	for _, snapshot := range snapshots {
		if snapshot.ChangeAmount.LessThan(modelDecimalZero()) {
			count++
		}
	}
	return count
}

func averageChangePercent(snapshots []model.MarketSnapshot) string {
	if len(snapshots) == 0 {
		return "0%"
	}
	total := modelDecimalZero()
	for _, snapshot := range snapshots {
		total = total.Add(snapshot.ChangePercent)
	}
	return total.Div(modelDecimalFromInt(len(snapshots))).StringFixed(2) + "%"
}

func totalTurnover(snapshots []model.MarketSnapshot) string {
	total := modelDecimalZero()
	for _, snapshot := range snapshots {
		total = total.Add(snapshot.Turnover)
	}
	return total.Div(modelDecimalFromInt(100000000)).StringFixed(2) + "亿"
}

func formatInt(value int) string {
	return modelDecimalFromInt(value).String()
}
