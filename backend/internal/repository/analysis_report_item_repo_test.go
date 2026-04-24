package repository_test

import (
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

// InMemoryAnalysisReportItemRepository 内存分析报告项仓储用于测试
type InMemoryAnalysisReportItemRepository struct {
	items  map[uint64]*model.AnalysisReportItem
	nextID uint64
}

func NewInMemoryAnalysisReportItemRepository() *InMemoryAnalysisReportItemRepository {
	return &InMemoryAnalysisReportItemRepository{
		items:  make(map[uint64]*model.AnalysisReportItem),
		nextID: 1,
	}
}

func (r *InMemoryAnalysisReportItemRepository) BatchCreate(items []model.AnalysisReportItem) error {
	for i := range items {
		items[i].ID = r.nextID
		r.items[r.nextID] = &items[i]
		r.nextID++
	}
	return nil
}

func (r *InMemoryAnalysisReportItemRepository) FindByReportID(reportID uint64) ([]model.AnalysisReportItem, error) {
	var result []model.AnalysisReportItem
	for _, item := range r.items {
		if item.ReportID == reportID {
			result = append(result, *item)
		}
	}
	return result, nil
}

// TestAnalysisReportItemRepository_BatchCreate 测试批量创建
func TestAnalysisReportItemRepository_BatchCreate(t *testing.T) {
	repo := NewInMemoryAnalysisReportItemRepository()

	items := []model.AnalysisReportItem{
		{
			ReportID:         1,
			UserID:           1,
			Symbol:           "600519.SH",
			AssetName:        "贵州茅台",
			TradeCount:       1,
			BuyCount:         1,
			SellCount:        0,
			BuyAmount:        decimal.NewFromInt(180000),
			EndingPositionQty: decimal.NewFromInt(100),
			EndingAvgCost:    decimal.NewFromInt(1800),
			LatestPrice:      decimal.NewFromInt(1900),
			UnrealizedProfit: decimal.NewFromInt(10000),
			TotalProfit:      decimal.NewFromInt(10000),
			RiskLevel:        "medium",
			AnalysisText:     "测试分析",
			Recommendation:   "hold",
		},
		{
			ReportID:         1,
			UserID:           1,
			Symbol:           "000858.SZ",
			AssetName:        "五粮液",
			TradeCount:       1,
			BuyCount:         1,
			SellCount:        0,
			BuyAmount:        decimal.NewFromInt(30000),
			EndingPositionQty: decimal.NewFromInt(200),
			EndingAvgCost:    decimal.NewFromInt(150),
			LatestPrice:      decimal.NewFromInt(160),
			UnrealizedProfit: decimal.NewFromInt(2000),
			TotalProfit:      decimal.NewFromInt(2000),
			RiskLevel:        "low",
			AnalysisText:     "测试分析",
			Recommendation:   "buy",
		},
	}

	err := repo.BatchCreate(items)
	if err != nil {
		t.Errorf("BatchCreate() error = %v", err)
	}

	if len(repo.items) != 2 {
		t.Errorf("BatchCreate() should create 2 items, got %d", len(repo.items))
	}
}

// TestAnalysisReportItemRepository_BatchCreate_Empty 测试空批量创建
func TestAnalysisReportItemRepository_BatchCreate_Empty(t *testing.T) {
	repo := NewInMemoryAnalysisReportItemRepository()

	err := repo.BatchCreate([]model.AnalysisReportItem{})
	if err != nil {
		t.Errorf("BatchCreate() with empty slice should not error, got %v", err)
	}
}

// TestAnalysisReportItemRepository_FindByReportID 测试按报告ID查找
func TestAnalysisReportItemRepository_FindByReportID(t *testing.T) {
	repo := NewInMemoryAnalysisReportItemRepository()

	now := time.Now()
	repo.BatchCreate([]model.AnalysisReportItem{
		{
			ReportID:         1,
			UserID:           1,
			Symbol:           "600519.SH",
			AssetName:        "贵州茅台",
			EndingPositionQty: decimal.Zero,
			EndingAvgCost:    decimal.Zero,
			LatestPrice:      decimal.Zero,
			UnrealizedProfit: decimal.Zero,
			TotalProfit:      decimal.Zero,
			RiskLevel:        "medium",
			AnalysisText:     "分析1",
			Recommendation:   "hold",
			CreatedAt:        now,
		},
		{
			ReportID:         1,
			UserID:           1,
			Symbol:           "000858.SZ",
			AssetName:        "五粮液",
			EndingPositionQty: decimal.Zero,
			EndingAvgCost:    decimal.Zero,
			LatestPrice:      decimal.Zero,
			UnrealizedProfit: decimal.Zero,
			TotalProfit:      decimal.Zero,
			RiskLevel:        "low",
			AnalysisText:     "分析2",
			Recommendation:   "buy",
			CreatedAt:        now,
		},
		{
			ReportID:         2,
			UserID:           1,
			Symbol:           "600519.SH",
			AssetName:        "贵州茅台",
			EndingPositionQty: decimal.Zero,
			EndingAvgCost:    decimal.Zero,
			LatestPrice:      decimal.Zero,
			UnrealizedProfit: decimal.Zero,
			TotalProfit:      decimal.Zero,
			RiskLevel:        "high",
			AnalysisText:     "分析3",
			Recommendation:   "sell",
			CreatedAt:        now,
		},
	})

	items, err := repo.FindByReportID(1)
	if err != nil {
		t.Errorf("FindByReportID() error = %v", err)
	}

	if len(items) != 2 {
		t.Errorf("FindByReportID() returned %d items, want 2", len(items))
	}
}

// TestAnalysisReportItemRepository_FindByReportID_Empty 测试空结果
func TestAnalysisReportItemRepository_FindByReportID_Empty(t *testing.T) {
	repo := NewInMemoryAnalysisReportItemRepository()

	items, err := repo.FindByReportID(999)
	if err != nil {
		t.Errorf("FindByReportID() error = %v", err)
	}

	if len(items) != 0 {
		t.Errorf("FindByReportID() should return empty slice for non-existent report")
	}
}

// 确保 InMemoryAnalysisReportItemRepository 实现了 AnalysisReportItemRepository 接口
var _ repository.AnalysisReportItemRepository = (*InMemoryAnalysisReportItemRepository)(nil)

// TestAnalysisReportItemRepository_Interface 测试接口实现
func TestAnalysisReportItemRepository_Interface(t *testing.T) {
	var repo repository.AnalysisReportItemRepository = NewInMemoryAnalysisReportItemRepository()

	_ = repo.BatchCreate([]model.AnalysisReportItem{{
		ReportID:         1,
		UserID:           1,
		Symbol:           "600519.SH",
		AssetName:        "贵州茅台",
		EndingPositionQty: decimal.Zero,
		EndingAvgCost:    decimal.Zero,
		LatestPrice:      decimal.Zero,
		UnrealizedProfit: decimal.Zero,
		TotalProfit:      decimal.Zero,
		RiskLevel:        "medium",
		AnalysisText:     "测试",
		Recommendation:   "hold",
	}})
	_, _ = repo.FindByReportID(1)
}
