package service

import (
	"context"
	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/pkg/marketdata"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// MockMarketDataProvider 模拟市场数据提供者
type MockMarketDataProvider struct {
	Quotes []marketdata.Quote
	Err    error
}

func (p *MockMarketDataProvider) GetQuotes(ctx context.Context, symbols []string) ([]marketdata.Quote, error) {
	if p.Err != nil {
		return nil, p.Err
	}
	return p.Quotes, nil
}

// MockMarketDataSnapshotRepo 模拟市场快照仓储
type MockMarketDataSnapshotRepo struct {
	Snapshots []model.MarketSnapshot
	Err       error
}

func (r *MockMarketDataSnapshotRepo) BatchCreate(snapshots []model.MarketSnapshot) error {
	r.Snapshots = append(r.Snapshots, snapshots...)
	return r.Err
}

func (r *MockMarketDataSnapshotRepo) FindLatestBatchNo() (string, error) {
	return "", r.Err
}

func (r *MockMarketDataSnapshotRepo) FindByBatchNo(batchNo string) ([]model.MarketSnapshot, error) {
	return nil, r.Err
}

func (r *MockMarketDataSnapshotRepo) FindLatestBySymbol(symbol string) (*model.MarketSnapshot, error) {
	return nil, r.Err
}

func (r *MockMarketDataSnapshotRepo) FindHistory(limit int, startTime, endTime *time.Time) ([]model.MarketSnapshot, error) {
	return nil, r.Err
}

func (r *MockMarketDataSnapshotRepo) FindHistoryBySymbol(symbol string, limit int, startTime, endTime *time.Time) ([]model.MarketSnapshot, error) {
	return nil, r.Err
}

// TestNormalizeSymbol 测试股票代码标准化
func TestNormalizeSymbol(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"600519.SH", "600519.SH"},
		{" 600519.SH ", "600519.SH"},
		{"", ""},
		{"   ", ""},
		{"000858.SZ", "000858.SZ"},
	}

	for _, tt := range tests {
		result := normalizeSymbol(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeSymbol(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// TestNormalizeSymbols_Multiple 测试多个股票代码标准化
func TestNormalizeSymbols_Multiple(t *testing.T) {
	input := []string{"600519.SH", " 000858.SZ ", "", "600519.SH", "000858.SZ"}
	result := normalizeSymbols(input)

	// 应该去重和去空
	if len(result) != 2 {
		t.Errorf("normalizeSymbols() returned %d symbols, want 2", len(result))
	}

	// 验证结果包含预期的股票代码
	seen := make(map[string]bool)
	for _, s := range result {
		seen[s] = true
	}
	if !seen["600519.SH"] || !seen["000858.SZ"] {
		t.Errorf("normalizeSymbols() missing expected symbols")
	}
}

// TestMarketDataService_FetchAndStoreQuotesBySymbols 测试按代码获取行情
func TestMarketDataService_FetchAndStoreQuotesBySymbols(t *testing.T) {
	now := time.Now()
	mockProvider := &MockMarketDataProvider{
		Quotes: []marketdata.Quote{
			{
				Symbol:       "600519.SH",
				Name:         "贵州茅台",
				Market:       "cn_stock",
				LastPrice:    1900.00,
				ChangeAmount: 10.5,
				ChangePercent: 0.55,
				SnapshotTime: now,
				Source:       "mock",
			},
		},
	}
	mockRepo := &MockMarketDataSnapshotRepo{}

	// 使用简化的配置
	svc := &marketDataService{
		provider:     mockProvider,
		snapshotRepo: mockRepo,
	}

	snapshots, err := svc.FetchAndStoreQuotesBySymbols(context.Background(), []string{"600519.SH"})
	if err != nil {
		t.Errorf("FetchAndStoreQuotesBySymbols() error = %v", err)
	}

	if len(snapshots) != 1 {
		t.Errorf("FetchAndStoreQuotesBySymbols() returned %d snapshots, want 1", len(snapshots))
	}

	if snapshots[0].Symbol != "600519.SH" {
		t.Errorf("Symbol = %s, want 600519.SH", snapshots[0].Symbol)
	}

	if !snapshots[0].LastPrice.Equal(decimal.NewFromFloat(1900.00)) {
		t.Errorf("LastPrice = %v, want 1900.00", snapshots[0].LastPrice)
	}
}

// TestMarketDataService_FetchAndStoreQuotesBySymbols_Empty 测试空代码列表
func TestMarketDataService_FetchAndStoreQuotesBySymbols_Empty(t *testing.T) {
	mockProvider := &MockMarketDataProvider{}
	mockRepo := &MockMarketDataSnapshotRepo{}

	svc := &marketDataService{
		provider:     mockProvider,
		snapshotRepo: mockRepo,
	}

	snapshots, err := svc.FetchAndStoreQuotesBySymbols(context.Background(), []string{})
	if err != nil {
		t.Errorf("FetchAndStoreQuotesBySymbols() error = %v", err)
	}

	if len(snapshots) != 0 {
		t.Errorf("FetchAndStoreQuotesBySymbols() should return empty for empty symbols")
	}
}

// TestMarketDataService_FetchAndStoreQuotesBySymbols_ProviderError 测试提供者错误
func TestMarketDataService_FetchAndStoreQuotesBySymbols_ProviderError(t *testing.T) {
	mockProvider := &MockMarketDataProvider{
		Err: gorm.ErrRecordNotFound,
	}
	mockRepo := &MockMarketDataSnapshotRepo{}

	svc := &marketDataService{
		provider:     mockProvider,
		snapshotRepo: mockRepo,
	}

	_, err := svc.FetchAndStoreQuotesBySymbols(context.Background(), []string{"600519.SH"})
	if err == nil {
		t.Error("FetchAndStoreQuotesBySymbols() should return error when provider fails")
	}
}

// TestMarketDataService_FetchAndStoreQuotesBySymbols_NoQuotes 测试无行情返回
func TestMarketDataService_FetchAndStoreQuotesBySymbols_NoQuotes(t *testing.T) {
	mockProvider := &MockMarketDataProvider{
		Quotes: []marketdata.Quote{},
	}
	mockRepo := &MockMarketDataSnapshotRepo{}

	svc := &marketDataService{
		provider:     mockProvider,
		snapshotRepo: mockRepo,
	}

	_, err := svc.FetchAndStoreQuotesBySymbols(context.Background(), []string{"600519.SH"})
	if err == nil {
		t.Error("FetchAndStoreQuotesBySymbols() should return error when no quotes returned")
	}
}

// TestMarketDataService_FetchAndStoreMarketSnapshots 测试获取市场快照
func TestMarketDataService_FetchAndStoreMarketSnapshots(t *testing.T) {
	now := time.Now()
	mockProvider := &MockMarketDataProvider{
		Quotes: []marketdata.Quote{
			{
				Symbol:       "000001.SH",
				Name:         "上证指数",
				Market:       "cn_index",
				LastPrice:    3000.50,
				SnapshotTime: now,
				Source:       "mock",
			},
		},
	}
	mockRepo := &MockMarketDataSnapshotRepo{}

	svc := NewMarketDataService(
		config.MarketConfig{Symbols: "000001.SH"},
		mockProvider,
		mockRepo,
	)

	batchNo, count, err := svc.FetchAndStoreMarketSnapshots(context.Background())
	if err != nil {
		t.Errorf("FetchAndStoreMarketSnapshots() error = %v", err)
	}

	if batchNo == "" {
		t.Error("FetchAndStoreMarketSnapshots() should return non-empty batch number")
	}

	if count != 1 {
		t.Errorf("FetchAndStoreMarketSnapshots() count = %d, want 1", count)
	}
}

// TestMarketDataService_FetchAndStoreMarketSnapshots_EmptySymbols 测试空配置
func TestMarketDataService_FetchAndStoreMarketSnapshots_EmptySymbols(t *testing.T) {
	mockProvider := &MockMarketDataProvider{}
	mockRepo := &MockMarketDataSnapshotRepo{}

	svc := &marketDataService{
		provider:     mockProvider,
		snapshotRepo: mockRepo,
		marketConfig: config.MarketConfig{Symbols: ""},
	}

	_, _, err := svc.FetchAndStoreMarketSnapshots(context.Background())
	if err == nil {
		t.Error("FetchAndStoreMarketSnapshots() should return error for empty symbols config")
	}
}

// TestMarketDataService_BatchNo 测试批次号生成
func TestMarketDataService_BatchNo(t *testing.T) {
	now := time.Now()
	mockProvider := &MockMarketDataProvider{
		Quotes: []marketdata.Quote{
			{
				Symbol:       "600519.SH",
				Name:         "贵州茅台",
				LastPrice:    1900.00,
				SnapshotTime: now,
				Source:       "mock",
			},
		},
	}
	mockRepo := &MockMarketDataSnapshotRepo{}

	svc := &marketDataService{
		provider:     mockProvider,
		snapshotRepo: mockRepo,
	}

	snapshots, _ := svc.FetchAndStoreQuotesBySymbols(context.Background(), []string{"600519.SH"})

	// 批次号应该包含时间戳
	if len(snapshots) > 0 && len(snapshots[0].BatchNo) < 10 {
		t.Errorf("BatchNo = %s, should contain timestamp", snapshots[0].BatchNo)
	}
}

// TestMarketDataService_QuoteConversion 测试行情转换
func TestMarketDataService_QuoteConversion(t *testing.T) {
	now := time.Now()
	quote := marketdata.Quote{
		Symbol:        "600519.SH",
		Name:          "贵州茅台",
		Market:        "cn_stock",
		SnapshotTime:  now,
		LastPrice:     1900.50,
		ChangeAmount:  10.5,
		ChangePercent: 0.55,
		OpenPrice:     1890.00,
		HighPrice:     1910.00,
		LowPrice:      1880.00,
		Volume:        1000000,
		Turnover:      1900000000,
		Source:        "mock",
	}

	mockProvider := &MockMarketDataProvider{
		Quotes: []marketdata.Quote{quote},
	}
	mockRepo := &MockMarketDataSnapshotRepo{}

	svc := &marketDataService{
		provider:     mockProvider,
		snapshotRepo: mockRepo,
	}

	snapshots, _ := svc.FetchAndStoreQuotesBySymbols(context.Background(), []string{"600519.SH"})

	if len(snapshots) != 1 {
		t.Fatalf("Expected 1 snapshot, got %d", len(snapshots))
	}

	s := snapshots[0]
	if s.Symbol != quote.Symbol {
		t.Errorf("Symbol = %s, want %s", s.Symbol, quote.Symbol)
	}
	if s.Name != quote.Name {
		t.Errorf("Name = %s, want %s", s.Name, quote.Name)
	}
	if !s.LastPrice.Equal(decimal.NewFromFloat(quote.LastPrice)) {
		t.Errorf("LastPrice = %v, want %v", s.LastPrice, quote.LastPrice)
	}
	if !s.Volume.Equal(decimal.NewFromFloat(quote.Volume)) {
		t.Errorf("Volume = %v, want %v", s.Volume, quote.Volume)
	}
}

// TestMarketDataService_Interface 测试接口实现
func TestMarketDataService_Interface(t *testing.T) {
	mockProvider := &MockMarketDataProvider{}
	mockRepo := &MockMarketDataSnapshotRepo{}

	var _ MarketDataService = &marketDataService{
		provider:     mockProvider,
		snapshotRepo: mockRepo,
	}
}
