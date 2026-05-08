package service

import (
	"context"
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// MockMetricRepository 模拟股票分析指标仓储
type MockMetricRepository struct {
	Metrics []model.StockAnalysisMetric
	Err     error
}

func (r *MockMetricRepository) Upsert(metric *model.StockAnalysisMetric) error {
	r.Metrics = append(r.Metrics, *metric)
	return r.Err
}

func (r *MockMetricRepository) BatchUpsert(metrics []model.StockAnalysisMetric) error {
	r.Metrics = append(r.Metrics, metrics...)
	return r.Err
}

func (r *MockMetricRepository) FindByUserPeriod(userID uint64, start, end time.Time, symbols []string) ([]model.StockAnalysisMetric, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	var result []model.StockAnalysisMetric
	symbolSet := make(map[string]bool)
	for _, s := range symbols {
		symbolSet[s] = true
	}
	for _, m := range r.Metrics {
		if m.UserID != userID {
			continue
		}
		if len(symbolSet) > 0 && !symbolSet[m.Symbol] {
			continue
		}
		result = append(result, m)
	}
	return result, nil
}

func (r *MockMetricRepository) FindByUserSymbolPeriod(userID uint64, symbol string, start, end time.Time) (*model.StockAnalysisMetric, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	for _, m := range r.Metrics {
		if m.UserID == userID && m.Symbol == symbol {
			return &m, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

// MockMetricTransactionRepository 模拟交易仓储
type MockMetricTransactionRepository struct {
	Transactions []model.Transaction
	Err          error
}

func (r *MockMetricTransactionRepository) Create(tx *model.Transaction) error {
	return r.Err
}

func (r *MockMetricTransactionRepository) FindByUserID(userID uint64, limit, offset int) ([]model.Transaction, int64, error) {
	return nil, 0, r.Err
}

func (r *MockMetricTransactionRepository) FindByDateRange(userID uint64, startDate, endDate string) ([]model.Transaction, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	return r.Transactions, nil
}

func (r *MockMetricTransactionRepository) FindByAssetCode(userID uint64, assetCode string) ([]model.Transaction, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	var result []model.Transaction
	for _, tx := range r.Transactions {
		if tx.AssetCode == assetCode {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (r *MockMetricTransactionRepository) Update(tx *model.Transaction) error {
	return r.Err
}

func (r *MockMetricTransactionRepository) Delete(id uint64) error {
	return r.Err
}

func (r *MockMetricTransactionRepository) BatchCreate(transactions []model.Transaction) error {
	return r.Err
}

func (r *MockMetricTransactionRepository) GetTransactionStats(userID uint64) (*response.TransactionStats, error) {
	return nil, r.Err
}

func (r *MockMetricTransactionRepository) FindByID(id uint64) (*model.Transaction, error) {
	return nil, r.Err
}

// MockMetricMarketSnapshotRepository 模拟市场快照仓储
type MockMetricMarketSnapshotRepository struct {
	Snapshots []model.MarketSnapshot
	Err       error
}

func (r *MockMetricMarketSnapshotRepository) BatchCreate(snapshots []model.MarketSnapshot) error {
	return r.Err
}

func (r *MockMetricMarketSnapshotRepository) FindLatestBatchNo() (string, error) {
	return "", r.Err
}

func (r *MockMetricMarketSnapshotRepository) FindByBatchNo(batchNo string) ([]model.MarketSnapshot, error) {
	return nil, r.Err
}

func (r *MockMetricMarketSnapshotRepository) FindLatestBySymbol(symbol string) (*model.MarketSnapshot, error) {
	return nil, r.Err
}

func (r *MockMetricMarketSnapshotRepository) FindHistory(limit int, startTime, endTime *time.Time) ([]model.MarketSnapshot, error) {
	return nil, r.Err
}

func (r *MockMetricMarketSnapshotRepository) FindHistoryBySymbol(symbol string, limit int, startTime, endTime *time.Time) ([]model.MarketSnapshot, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	return r.Snapshots, nil
}

// MockMetricMarketDataService 模拟市场数据服务
type MockMetricMarketDataService struct {
	Snapshots []model.MarketSnapshot
	Err       error
}

func (s *MockMetricMarketDataService) FetchAndStoreMarketSnapshots(ctx context.Context) (string, int, error) {
	return "batch001", len(s.Snapshots), s.Err
}

func (s *MockMetricMarketDataService) FetchAndStoreQuotesBySymbols(ctx context.Context, symbols []string) ([]model.MarketSnapshot, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	return s.Snapshots, nil
}

// TestNormalizeSymbols 测试股票代码标准化
func TestNormalizeSymbols(t *testing.T) {
	tests := []struct {
		input    []string
		expected int
	}{
		{[]string{"600519.SH", "000858.SZ"}, 2},
		{[]string{" 600519.SH ", "600519.SH"}, 1}, // 去重和去空格
		{[]string{"", "   "}, 0},
		{[]string{}, 0},
	}

	for _, tt := range tests {
		result := normalizeSymbols(tt.input)
		if len(result) != tt.expected {
			t.Errorf("normalizeSymbols(%v) returned %d symbols, want %d", tt.input, len(result), tt.expected)
		}
	}
}

// TestAggregateMetricTransactions 测试交易聚合
func TestAggregateMetricTransactions(t *testing.T) {
	profit := decimal.NewFromInt(1000)
	transactions := []model.Transaction{
		{
			AssetCode:       "600519.SH",
			AssetName:       "贵州茅台",
			TransactionType: "buy",
			Quantity:        decimal.NewFromInt(100),
			TotalAmount:     decimal.NewFromInt(180000),
		},
		{
			AssetCode:       "600519.SH",
			AssetName:       "贵州茅台",
			TransactionType: "sell",
			Quantity:        decimal.NewFromInt(50),
			TotalAmount:     decimal.NewFromInt(100000),
			Profit:          &profit,
		},
		{
			AssetCode:       "000858.SZ",
			AssetName:       "五粮液",
			TransactionType: "buy",
			Quantity:        decimal.NewFromInt(200),
			TotalAmount:     decimal.NewFromInt(30000),
		},
	}

	result := aggregateMetricTransactions(transactions)

	if len(result) != 2 {
		t.Errorf("aggregateMetricTransactions() returned %d symbols, want 2", len(result))
	}

	agg, ok := result["600519.SH"]
	if !ok {
		t.Fatal("Expected 600519.SH in result")
	}

	if agg.TradeCount != 2 {
		t.Errorf("TradeCount = %d, want 2", agg.TradeCount)
	}

	if agg.BuyCount != 1 {
		t.Errorf("BuyCount = %d, want 1", agg.BuyCount)
	}

	if agg.SellCount != 1 {
		t.Errorf("SellCount = %d, want 1", agg.SellCount)
	}

	if !agg.NetQuantity.Equal(decimal.NewFromInt(50)) {
		t.Errorf("NetQuantity = %v, want 50", agg.NetQuantity)
	}
}

// TestAggregateMetricTransactions_Empty 测试空交易列表
func TestAggregateMetricTransactions_Empty(t *testing.T) {
	result := aggregateMetricTransactions([]model.Transaction{})
	if len(result) != 0 {
		t.Errorf("aggregateMetricTransactions() should return empty map for empty input")
	}
}

// TestApplyMetricMarketHistory 测试应用市场历史数据
func TestApplyMetricMarketHistory(t *testing.T) {
	now := time.Now()
	agg := &metricAggregate{
		Symbol:           "600519.SH",
		EndingPositionQty: decimal.NewFromInt(100),
	}

	history := []model.MarketSnapshot{
		{
			Symbol:     "600519.SH",
			LastPrice:  decimal.NewFromInt(1800),
			HighPrice:  decimal.NewFromInt(1850),
			LowPrice:   decimal.NewFromInt(1780),
			Market:     "cn_stock",
			SnapshotTime: now.Add(-time.Hour),
		},
		{
			Symbol:     "600519.SH",
			LastPrice:  decimal.NewFromInt(1900),
			HighPrice:  decimal.NewFromInt(1920),
			LowPrice:   decimal.NewFromInt(1790),
			Market:     "cn_stock",
			SnapshotTime: now,
		},
	}

	applyMetricMarketHistory(agg, history, marketDataStatusComplete)

	if !agg.PeriodStartPrice.Equal(decimal.NewFromInt(1800)) {
		t.Errorf("PeriodStartPrice = %v, want 1800", agg.PeriodStartPrice)
	}

	if !agg.PeriodEndPrice.Equal(decimal.NewFromInt(1900)) {
		t.Errorf("PeriodEndPrice = %v, want 1900", agg.PeriodEndPrice)
	}

	if !agg.LatestPrice.Equal(decimal.NewFromInt(1900)) {
		t.Errorf("LatestPrice = %v, want 1900", agg.LatestPrice)
	}

	if agg.Market != "cn_stock" {
		t.Errorf("Market = %s, want cn_stock", agg.Market)
	}

	if agg.MarketDataStatus != marketDataStatusComplete {
		t.Errorf("MarketDataStatus = %s, want %s", agg.MarketDataStatus, marketDataStatusComplete)
	}
}

// TestApplyMetricMarketHistory_HighLowPrice 测试最高最低价计算
func TestApplyMetricMarketHistory_HighLowPrice(t *testing.T) {
	now := time.Now()
	agg := &metricAggregate{
		Symbol: "600519.SH",
	}

	history := []model.MarketSnapshot{
		{
			LastPrice:     decimal.Zero,
			HighPrice:     decimal.NewFromInt(1850),
			LowPrice:      decimal.NewFromInt(1780),
			SnapshotTime:  now.Add(-time.Hour),
		},
		{
			LastPrice:     decimal.Zero,
			HighPrice:     decimal.NewFromInt(1920),
			LowPrice:      decimal.NewFromInt(1750),
			SnapshotTime:  now,
		},
	}

	applyMetricMarketHistory(agg, history, marketDataStatusComplete)

	if !agg.PeriodHighPrice.Equal(decimal.NewFromInt(1920)) {
		t.Errorf("PeriodHighPrice = %v, want 1920", agg.PeriodHighPrice)
	}

	if !agg.PeriodLowPrice.Equal(decimal.NewFromInt(1750)) {
		t.Errorf("PeriodLowPrice = %v, want 1750", agg.PeriodLowPrice)
	}
}

// TestStockAnalysisMetricService_PrepareMetrics 测试准备指标
func TestStockAnalysisMetricService_PrepareMetrics(t *testing.T) {
	now := time.Now()
	start := now.AddDate(0, -1, 0)
	end := now

	mockMetricRepo := &MockMetricRepository{}
	mockTxRepo := &MockMetricTransactionRepository{
		Transactions: []model.Transaction{
			{
				AssetCode:       "600519.SH",
				AssetName:       "贵州茅台",
				TransactionType: "buy",
				Quantity:        decimal.NewFromInt(100),
				TotalAmount:     decimal.NewFromInt(180000),
			},
		},
	}
	mockSnapshotRepo := &MockMetricMarketSnapshotRepository{}
	mockMarketSvc := &MockMetricMarketDataService{}

	svc := NewStockAnalysisMetricService(mockMetricRepo, mockTxRepo, mockSnapshotRepo, mockMarketSvc)

	metrics, err := svc.PrepareMetrics(context.Background(), 1, nil, start, end, nil, true, true)
	if err != nil {
		t.Errorf("PrepareMetrics() error = %v", err)
	}

	if len(metrics) != 1 {
		t.Errorf("PrepareMetrics() returned %d metrics, want 1", len(metrics))
	}
}

// TestStockAnalysisMetricService_PrepareMetrics_Empty 测试空交易
func TestStockAnalysisMetricService_PrepareMetrics_Empty(t *testing.T) {
	now := time.Now()
	start := now.AddDate(0, -1, 0)
	end := now

	mockMetricRepo := &MockMetricRepository{}
	mockTxRepo := &MockMetricTransactionRepository{
		Transactions: []model.Transaction{},
	}
	mockSnapshotRepo := &MockMetricMarketSnapshotRepository{}
	mockMarketSvc := &MockMetricMarketDataService{}

	svc := NewStockAnalysisMetricService(mockMetricRepo, mockTxRepo, mockSnapshotRepo, mockMarketSvc)

	metrics, err := svc.PrepareMetrics(context.Background(), 1, nil, start, end, []string{"600519.SH"}, true, true)
	if err != nil {
		t.Errorf("PrepareMetrics() error = %v", err)
	}

	if len(metrics) != 0 {
		t.Errorf("PrepareMetrics() should return empty for no transactions")
	}
}

// TestStockAnalysisMetricService_PrepareMetrics_Cached 测试缓存
func TestStockAnalysisMetricService_PrepareMetrics_Cached(t *testing.T) {
	now := time.Now()
	start := now.AddDate(0, -1, 0)
	end := now

	mockMetricRepo := &MockMetricRepository{
		Metrics: []model.StockAnalysisMetric{
			{
				UserID:         1,
				Symbol:         "600519.SH",
				PeriodStart:    start,
				PeriodEnd:      end,
				TradeCount:     5,
				ComputedAt:     now,
			},
		},
	}
	mockTxRepo := &MockMetricTransactionRepository{}
	mockSnapshotRepo := &MockMetricMarketSnapshotRepository{}
	mockMarketSvc := &MockMetricMarketDataService{}

	svc := NewStockAnalysisMetricService(mockMetricRepo, mockTxRepo, mockSnapshotRepo, mockMarketSvc)

	// 不强制刷新，应该返回缓存的指标
	metrics, err := svc.PrepareMetrics(context.Background(), 1, nil, start, end, []string{"600519.SH"}, false, false)
	if err != nil {
		t.Errorf("PrepareMetrics() error = %v", err)
	}

	if len(metrics) != 1 {
		t.Errorf("PrepareMetrics() should return cached metric")
	}
}

// TestStockAnalysisMetricService_Interface 测试接口实现
func TestStockAnalysisMetricService_Interface(t *testing.T) {
	mockMetricRepo := &MockMetricRepository{}
	mockTxRepo := &MockMetricTransactionRepository{}
	mockSnapshotRepo := &MockMetricMarketSnapshotRepository{}
	mockMarketSvc := &MockMetricMarketDataService{}

	var _ StockAnalysisMetricService = NewStockAnalysisMetricService(mockMetricRepo, mockTxRepo, mockSnapshotRepo, mockMarketSvc)
}
