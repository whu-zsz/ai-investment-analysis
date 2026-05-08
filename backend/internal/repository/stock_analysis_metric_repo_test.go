package repository_test

import (
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// InMemoryStockAnalysisMetricRepository 内存股票分析指标仓储用于测试
type InMemoryStockAnalysisMetricRepository struct {
	metrics map[uint64]*model.StockAnalysisMetric
	nextID  uint64
}

func NewInMemoryStockAnalysisMetricRepository() *InMemoryStockAnalysisMetricRepository {
	return &InMemoryStockAnalysisMetricRepository{
		metrics: make(map[uint64]*model.StockAnalysisMetric),
		nextID:  1,
	}
}

func (r *InMemoryStockAnalysisMetricRepository) Upsert(metric *model.StockAnalysisMetric) error {
	// 查找是否存在相同 user_id + symbol + period_start + period_end 的记录
	for id, m := range r.metrics {
		if m.UserID == metric.UserID && m.Symbol == metric.Symbol &&
			m.PeriodStart.Format("2006-01-02") == metric.PeriodStart.Format("2006-01-02") &&
			m.PeriodEnd.Format("2006-01-02") == metric.PeriodEnd.Format("2006-01-02") {
			// 更新现有记录
			metric.ID = id
			r.metrics[id] = metric
			return nil
		}
	}
	// 创建新记录
	metric.ID = r.nextID
	r.metrics[r.nextID] = metric
	r.nextID++
	return nil
}

func (r *InMemoryStockAnalysisMetricRepository) BatchUpsert(metrics []model.StockAnalysisMetric) error {
	for i := range metrics {
		if err := r.Upsert(&metrics[i]); err != nil {
			return err
		}
	}
	return nil
}

func (r *InMemoryStockAnalysisMetricRepository) FindByUserPeriod(userID uint64, start, end time.Time, symbols []string) ([]model.StockAnalysisMetric, error) {
	var result []model.StockAnalysisMetric
	symbolSet := make(map[string]bool)
	for _, s := range symbols {
		symbolSet[s] = true
	}

	for _, m := range r.metrics {
		if m.UserID != userID {
			continue
		}
		if m.PeriodStart.Format("2006-01-02") != start.Format("2006-01-02") {
			continue
		}
		if m.PeriodEnd.Format("2006-01-02") != end.Format("2006-01-02") {
			continue
		}
		if len(symbolSet) > 0 && !symbolSet[m.Symbol] {
			continue
		}
		result = append(result, *m)
	}
	return result, nil
}

func (r *InMemoryStockAnalysisMetricRepository) FindByUserSymbolPeriod(userID uint64, symbol string, start, end time.Time) (*model.StockAnalysisMetric, error) {
	for _, m := range r.metrics {
		if m.UserID == userID && m.Symbol == symbol &&
			m.PeriodStart.Format("2006-01-02") == start.Format("2006-01-02") &&
			m.PeriodEnd.Format("2006-01-02") == end.Format("2006-01-02") {
			return m, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// TestStockAnalysisMetricRepository_Upsert 测试创建指标
func TestStockAnalysisMetricRepository_Upsert(t *testing.T) {
	repo := NewInMemoryStockAnalysisMetricRepository()

	now := time.Now()
	metric := &model.StockAnalysisMetric{
		UserID:               1,
		Symbol:               "600519.SH",
		AssetName:            "贵州茅台",
		PeriodStart:          now.AddDate(0, -1, 0),
		PeriodEnd:            now,
		TradeCount:           5,
		BuyCount:             3,
		SellCount:            2,
		BuyQuantity:          decimal.NewFromInt(300),
		SellQuantity:         decimal.NewFromInt(100),
		BuyAmount:            decimal.NewFromInt(500000),
		SellAmount:           decimal.NewFromInt(180000),
		NetQuantity:          decimal.NewFromInt(200),
		RealizedProfit:       decimal.NewFromInt(10000),
		RealizedProfitRate:   decimal.NewFromFloat(0.05),
		EndingPositionQty:    decimal.NewFromInt(200),
		EndingAvgCost:        decimal.NewFromInt(1800),
		LatestPrice:          decimal.NewFromInt(1900),
		LatestMarketValue:    decimal.NewFromInt(380000),
		UnrealizedProfit:     decimal.NewFromInt(20000),
		UnrealizedProfitRate: decimal.NewFromFloat(0.10),
		TotalProfit:          decimal.NewFromInt(30000),
		TotalProfitRate:      decimal.NewFromFloat(0.15),
		ComputedAt:           now,
	}

	err := repo.Upsert(metric)
	if err != nil {
		t.Errorf("Upsert() error = %v", err)
	}

	if len(repo.metrics) != 1 {
		t.Errorf("Upsert() should create 1 metric, got %d", len(repo.metrics))
	}
}

// TestStockAnalysisMetricRepository_Upsert_Update 测试更新指标
func TestStockAnalysisMetricRepository_Upsert_Update(t *testing.T) {
	repo := NewInMemoryStockAnalysisMetricRepository()

	now := time.Now()
	periodStart := now.AddDate(0, -1, 0)
	periodEnd := now

	// 先创建
	metric1 := &model.StockAnalysisMetric{
		UserID:            1,
		Symbol:            "600519.SH",
		AssetName:         "贵州茅台",
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		TradeCount:        5,
		EndingPositionQty: decimal.NewFromInt(200),
		ComputedAt:        now,
	}
	repo.Upsert(metric1)

	// 再更新 - 相同的 user_id + symbol + period
	metric2 := &model.StockAnalysisMetric{
		UserID:            1,
		Symbol:            "600519.SH",
		AssetName:         "贵州茅台",
		PeriodStart:       periodStart,
		PeriodEnd:         periodEnd,
		TradeCount:        10, // 更新交易次数
		EndingPositionQty: decimal.NewFromInt(300),
		ComputedAt:        now,
	}
	err := repo.Upsert(metric2)
	if err != nil {
		t.Errorf("Upsert() update error = %v", err)
	}

	// 应该还是只有一条记录
	if len(repo.metrics) != 1 {
		t.Errorf("Upsert() should still have 1 metric after update, got %d", len(repo.metrics))
	}

	// 验证数据已更新
	for _, m := range repo.metrics {
		if m.TradeCount != 10 {
			t.Errorf("Upsert() should update TradeCount to 10, got %d", m.TradeCount)
		}
	}
}

// TestStockAnalysisMetricRepository_BatchUpsert 测试批量创建
func TestStockAnalysisMetricRepository_BatchUpsert(t *testing.T) {
	repo := NewInMemoryStockAnalysisMetricRepository()

	now := time.Now()
	periodStart := now.AddDate(0, -1, 0)
	periodEnd := now

	metrics := []model.StockAnalysisMetric{
		{
			UserID:            1,
			Symbol:            "600519.SH",
			AssetName:         "贵州茅台",
			PeriodStart:       periodStart,
			PeriodEnd:         periodEnd,
			EndingPositionQty: decimal.NewFromInt(100),
			ComputedAt:        now,
		},
		{
			UserID:            1,
			Symbol:            "000858.SZ",
			AssetName:         "五粮液",
			PeriodStart:       periodStart,
			PeriodEnd:         periodEnd,
			EndingPositionQty: decimal.NewFromInt(200),
			ComputedAt:        now,
		},
	}

	err := repo.BatchUpsert(metrics)
	if err != nil {
		t.Errorf("BatchUpsert() error = %v", err)
	}

	if len(repo.metrics) != 2 {
		t.Errorf("BatchUpsert() should create 2 metrics, got %d", len(repo.metrics))
	}
}

// TestStockAnalysisMetricRepository_BatchUpsert_Empty 测试空批量
func TestStockAnalysisMetricRepository_BatchUpsert_Empty(t *testing.T) {
	repo := NewInMemoryStockAnalysisMetricRepository()

	err := repo.BatchUpsert([]model.StockAnalysisMetric{})
	if err != nil {
		t.Errorf("BatchUpsert() with empty slice should not error, got %v", err)
	}
}

// TestStockAnalysisMetricRepository_FindByUserPeriod 测试按用户和时间查找
func TestStockAnalysisMetricRepository_FindByUserPeriod(t *testing.T) {
	repo := NewInMemoryStockAnalysisMetricRepository()

	now := time.Now()
	periodStart := now.AddDate(0, -1, 0)
	periodEnd := now

	repo.BatchUpsert([]model.StockAnalysisMetric{
		{
			UserID:            1,
			Symbol:            "600519.SH",
			AssetName:         "贵州茅台",
			PeriodStart:       periodStart,
			PeriodEnd:         periodEnd,
			EndingPositionQty: decimal.Zero,
			ComputedAt:        now,
		},
		{
			UserID:            1,
			Symbol:            "000858.SZ",
			AssetName:         "五粮液",
			PeriodStart:       periodStart,
			PeriodEnd:         periodEnd,
			EndingPositionQty: decimal.Zero,
			ComputedAt:        now,
		},
		{
			UserID:            2,
			Symbol:            "600519.SH",
			AssetName:         "贵州茅台",
			PeriodStart:       periodStart,
			PeriodEnd:         periodEnd,
			EndingPositionQty: decimal.Zero,
			ComputedAt:        now,
		},
	})

	metrics, err := repo.FindByUserPeriod(1, periodStart, periodEnd, nil)
	if err != nil {
		t.Errorf("FindByUserPeriod() error = %v", err)
	}

	if len(metrics) != 2 {
		t.Errorf("FindByUserPeriod() returned %d metrics, want 2", len(metrics))
	}
}

// TestStockAnalysisMetricRepository_FindByUserPeriod_WithSymbols 测试按股票代码过滤
func TestStockAnalysisMetricRepository_FindByUserPeriod_WithSymbols(t *testing.T) {
	repo := NewInMemoryStockAnalysisMetricRepository()

	now := time.Now()
	periodStart := now.AddDate(0, -1, 0)
	periodEnd := now

	repo.BatchUpsert([]model.StockAnalysisMetric{
		{
			UserID:            1,
			Symbol:            "600519.SH",
			AssetName:         "贵州茅台",
			PeriodStart:       periodStart,
			PeriodEnd:         periodEnd,
			EndingPositionQty: decimal.Zero,
			ComputedAt:        now,
		},
		{
			UserID:            1,
			Symbol:            "000858.SZ",
			AssetName:         "五粮液",
			PeriodStart:       periodStart,
			PeriodEnd:         periodEnd,
			EndingPositionQty: decimal.Zero,
			ComputedAt:        now,
		},
	})

	metrics, err := repo.FindByUserPeriod(1, periodStart, periodEnd, []string{"600519.SH"})
	if err != nil {
		t.Errorf("FindByUserPeriod() error = %v", err)
	}

	if len(metrics) != 1 {
		t.Errorf("FindByUserPeriod() returned %d metrics, want 1", len(metrics))
	}

	if len(metrics) > 0 && metrics[0].Symbol != "600519.SH" {
		t.Errorf("FindByUserPeriod() should return only 600519.SH")
	}
}

// TestStockAnalysisMetricRepository_FindByUserPeriod_Empty 测试空结果
func TestStockAnalysisMetricRepository_FindByUserPeriod_Empty(t *testing.T) {
	repo := NewInMemoryStockAnalysisMetricRepository()

	now := time.Now()
	periodStart := now.AddDate(0, -1, 0)
	periodEnd := now

	metrics, err := repo.FindByUserPeriod(999, periodStart, periodEnd, nil)
	if err != nil {
		t.Errorf("FindByUserPeriod() error = %v", err)
	}

	if len(metrics) != 0 {
		t.Errorf("FindByUserPeriod() should return empty for non-existent user")
	}
}

// TestStockAnalysisMetricRepository_FindByUserSymbolPeriod 测试按用户股票时间查找
func TestStockAnalysisMetricRepository_FindByUserSymbolPeriod(t *testing.T) {
	repo := NewInMemoryStockAnalysisMetricRepository()

	now := time.Now()
	periodStart := now.AddDate(0, -1, 0)
	periodEnd := now

	repo.BatchUpsert([]model.StockAnalysisMetric{
		{
			UserID:            1,
			Symbol:            "600519.SH",
			AssetName:         "贵州茅台",
			PeriodStart:       periodStart,
			PeriodEnd:         periodEnd,
			EndingPositionQty: decimal.NewFromInt(100),
			ComputedAt:        now,
		},
	})

	metric, err := repo.FindByUserSymbolPeriod(1, "600519.SH", periodStart, periodEnd)
	if err != nil {
		t.Errorf("FindByUserSymbolPeriod() error = %v", err)
	}

	if metric.Symbol != "600519.SH" {
		t.Errorf("FindByUserSymbolPeriod() returned symbol = %s, want 600519.SH", metric.Symbol)
	}
}

// TestStockAnalysisMetricRepository_FindByUserSymbolPeriod_NotFound 测试找不到
func TestStockAnalysisMetricRepository_FindByUserSymbolPeriod_NotFound(t *testing.T) {
	repo := NewInMemoryStockAnalysisMetricRepository()

	now := time.Now()
	periodStart := now.AddDate(0, -1, 0)
	periodEnd := now

	_, err := repo.FindByUserSymbolPeriod(999, "999999.SH", periodStart, periodEnd)
	if err != gorm.ErrRecordNotFound {
		t.Error("FindByUserSymbolPeriod() should return ErrRecordNotFound for non-existent metric")
	}
}

// 确保 InMemoryStockAnalysisMetricRepository 实现了 StockAnalysisMetricRepository 接口
var _ repository.StockAnalysisMetricRepository = (*InMemoryStockAnalysisMetricRepository)(nil)

// TestStockAnalysisMetricRepository_Interface 测试接口实现
func TestStockAnalysisMetricRepository_Interface(t *testing.T) {
	var repo repository.StockAnalysisMetricRepository = NewInMemoryStockAnalysisMetricRepository()

	now := time.Now()
	_ = repo.Upsert(&model.StockAnalysisMetric{
		UserID:            1,
		Symbol:            "600519.SH",
		AssetName:         "贵州茅台",
		PeriodStart:       now.AddDate(0, -1, 0),
		PeriodEnd:         now,
		EndingPositionQty: decimal.Zero,
		ComputedAt:        now,
	})
	_ = repo.BatchUpsert([]model.StockAnalysisMetric{})
	_, _ = repo.FindByUserPeriod(1, now, now, nil)
	_, _ = repo.FindByUserSymbolPeriod(1, "600519.SH", now, now)
}
