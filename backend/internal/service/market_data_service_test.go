package service

import (
	"context"
	"errors"
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
	Quotes          []marketdata.Quote
	Klines          []marketdata.KlineBar
	Detail          *marketdata.StockDetail
	Err             error
	DetailCalls     int
	KlineCalls      int
}

func (p *MockMarketDataProvider) GetQuotes(ctx context.Context, symbols []string) ([]marketdata.Quote, error) {
	if p.Err != nil {
		return nil, p.Err
	}
	return p.Quotes, nil
}

func (p *MockMarketDataProvider) GetStockDetail(ctx context.Context, symbol string) (*marketdata.StockDetail, error) {
	p.DetailCalls++
	if p.Err != nil {
		return nil, p.Err
	}
	if p.Detail != nil {
		return p.Detail, nil
	}
	return &marketdata.StockDetail{Symbol: symbol, Name: symbol, Market: "cn_stock", FetchedAt: time.Now(), Source: "mock"}, nil
}

func (p *MockMarketDataProvider) GetKlines(ctx context.Context, symbol, period, adjust string, limit int) ([]marketdata.KlineBar, error) {
	p.KlineCalls++
	if p.Err != nil {
		return nil, p.Err
	}
	if len(p.Klines) > 0 {
		return p.Klines, nil
	}
	return []marketdata.KlineBar{}, nil
}

// MockMarketDataSnapshotRepo 模拟市场快照仓储
type MockMarketDataSnapshotRepo struct {
	Snapshots []model.MarketSnapshot
	Err       error
}

type MockStockQuoteDetailRepo struct {
	Detail       *model.StockQuoteDetail
	FindErr      error
	Upserted     *model.StockQuoteDetail
	UpsertErr    error
}

func (r *MockStockQuoteDetailRepo) Upsert(detail *model.StockQuoteDetail) error {
	r.Upserted = detail
	return r.UpsertErr
}

func (r *MockStockQuoteDetailRepo) FindBySymbol(symbol string) (*model.StockQuoteDetail, error) {
	if r.FindErr != nil {
		return nil, r.FindErr
	}
	if r.Detail == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return r.Detail, nil
}

type MockStockKlineRepo struct {
	Bars           []model.StockKlineBar
	LatestBars     map[string]*model.StockKlineBar
	UpsertedBars   []model.StockKlineBar
	FindLatestErr  error
	UpsertErr      error
}

func (r *MockStockKlineRepo) UpsertBars(bars []model.StockKlineBar) error {
	r.UpsertedBars = append(r.UpsertedBars, bars...)
	return r.UpsertErr
}

func (r *MockStockKlineRepo) FindBars(symbol, period, adjust string, limit int) ([]model.StockKlineBar, error) {
	return r.Bars, nil
}

func (r *MockStockKlineRepo) FindLatestBar(symbol, period, adjust string) (*model.StockKlineBar, error) {
	if r.LatestBars != nil {
		if bar, ok := r.LatestBars[symbol+"|"+period+"|"+adjust]; ok {
			return bar, nil
		}
	}
	if r.FindLatestErr != nil {
		return nil, r.FindLatestErr
	}
	return nil, gorm.ErrRecordNotFound
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
		nil,
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

func TestMarketDataService_EnsureTrackedIndexHistory_FetchMissing(t *testing.T) {
	mockProvider := &MockMarketDataProvider{
		Klines: []marketdata.KlineBar{
			{
				Symbol:     "000001.SZ",
				Period:     "day",
				AdjustType: "none",
				BarTime:    time.Now(),
				Open:       3000,
				Close:      3050,
				High:       3060,
				Low:        2990,
				Source:     "mock",
			},
		},
	}
	mockRepo := &MockStockKlineRepo{}

	svc := &marketDataService{
		marketConfig: config.MarketConfig{Symbols: "000001.SH"},
		provider:     mockProvider,
		klineRepo:    mockRepo,
	}

	if err := svc.EnsureTrackedIndexHistory(context.Background()); err != nil {
		t.Fatalf("EnsureTrackedIndexHistory() error = %v", err)
	}

	if len(mockRepo.UpsertedBars) != 1 {
		t.Fatalf("UpsertedBars = %d, want 1", len(mockRepo.UpsertedBars))
	}
	if mockRepo.UpsertedBars[0].Symbol != "000001.SZ" {
		t.Fatalf("Symbol = %s, want 000001.SZ", mockRepo.UpsertedBars[0].Symbol)
	}
	if mockRepo.UpsertedBars[0].Period != "day" {
		t.Fatalf("Period = %s, want day", mockRepo.UpsertedBars[0].Period)
	}
	if mockRepo.UpsertedBars[0].AdjustType != "none" {
		t.Fatalf("AdjustType = %s, want none", mockRepo.UpsertedBars[0].AdjustType)
	}
}

func TestMarketDataService_EnsureTrackedIndexHistory_SkipExisting(t *testing.T) {
	mockProvider := &MockMarketDataProvider{
		Klines: []marketdata.KlineBar{{Symbol: "000001.SH", Period: "day", AdjustType: "none", BarTime: time.Now(), Source: "mock"}},
	}
	mockRepo := &MockStockKlineRepo{
		LatestBars: map[string]*model.StockKlineBar{
			"000001.SH|day|none": {Symbol: "000001.SH", Period: "day", AdjustType: "none", BarTime: time.Now()},
		},
	}

	svc := &marketDataService{
		marketConfig: config.MarketConfig{Symbols: "000001.SH"},
		provider:     mockProvider,
		klineRepo:    mockRepo,
	}

	if err := svc.EnsureTrackedIndexHistory(context.Background()); err != nil {
		t.Fatalf("EnsureTrackedIndexHistory() error = %v", err)
	}
	if len(mockRepo.UpsertedBars) != 0 {
		t.Fatalf("UpsertedBars = %d, want 0", len(mockRepo.UpsertedBars))
	}
}

func TestMarketDataService_EnsureTrackedIndexHistory_ContinueOnError(t *testing.T) {
	expectedErr := errors.New("provider failed")
	mockProvider := &MockMarketDataProvider{Err: expectedErr}
	mockRepo := &MockStockKlineRepo{}

	svc := &marketDataService{
		marketConfig: config.MarketConfig{Symbols: "000001.SH,399001.SZ"},
		provider:     mockProvider,
		klineRepo:    mockRepo,
	}

	err := svc.EnsureTrackedIndexHistory(context.Background())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("EnsureTrackedIndexHistory() error = %v, want %v", err, expectedErr)
	}
}

func TestMarketStockService_GetStockDetailUsesDatabaseCache(t *testing.T) {
	detailRepo := &MockStockQuoteDetailRepo{
		Detail: &model.StockQuoteDetail{
			Symbol:    "000858.SZ",
			Name:      "五粮液",
			Market:    "cn_stock",
			LastPrice: decimal.NewFromFloat(128.88),
			Industry:  "GP-A-CYB",
			Region:    "—",
			Concepts:  "腾讯行情,白酒,白酒",
			Source:    "eastmoney",
			FetchedAt: time.Now().Add(-24 * time.Hour),
		},
	}
	provider := &MockMarketDataProvider{}
	svc := &marketStockService{provider: provider, detailRepo: detailRepo, klineRepo: &MockStockKlineRepo{}}

	res, err := svc.GetStockDetail("000858.SZ", false)
	if err != nil {
		t.Fatalf("GetStockDetail() error = %v", err)
	}
	if provider.DetailCalls != 0 {
		t.Fatalf("DetailCalls = %d, want 0", provider.DetailCalls)
	}
	if res.Symbol != "000858.SZ" {
		t.Fatalf("Symbol = %s, want 000858.SZ", res.Symbol)
	}
	if res.RefreshTriggered {
		t.Fatal("RefreshTriggered = true, want false")
	}
	if res.Industry != "" {
		t.Fatalf("Industry = %q, want empty", res.Industry)
	}
	if res.Region != "" {
		t.Fatalf("Region = %q, want empty", res.Region)
	}
	if len(res.Concepts) != 1 || res.Concepts[0] != "白酒" {
		t.Fatalf("Concepts = %#v, want [白酒]", res.Concepts)
	}
}

func TestMarketStockService_GetStockDetailForceRefreshAndUpsert(t *testing.T) {
	detailRepo := &MockStockQuoteDetailRepo{
		Detail: &model.StockQuoteDetail{
			Symbol:    "000858.SZ",
			Name:      "旧详情",
			Market:    "cn_stock",
			FetchedAt: time.Now().Add(-24 * time.Hour),
		},
	}
	provider := &MockMarketDataProvider{
		Detail: &marketdata.StockDetail{
			Symbol:    "000858.SZ",
			Name:      "五粮液",
			Market:    "cn_stock",
			LastPrice: 129.66,
			PrevClose: 128.00,
			FetchedAt: time.Now(),
			Source:    "eastmoney",
		},
	}
	svc := &marketStockService{provider: provider, detailRepo: detailRepo, klineRepo: &MockStockKlineRepo{}}

	res, err := svc.GetStockDetail("000858.SZ", true)
	if err != nil {
		t.Fatalf("GetStockDetail() error = %v", err)
	}
	if provider.DetailCalls != 1 {
		t.Fatalf("DetailCalls = %d, want 1", provider.DetailCalls)
	}
	if detailRepo.Upserted == nil {
		t.Fatal("expected detail upsert")
	}
	if detailRepo.Upserted.Industry != "" {
		t.Fatalf("Upserted Industry = %q, want empty", detailRepo.Upserted.Industry)
	}
	if detailRepo.Upserted.Region != "" {
		t.Fatalf("Upserted Region = %q, want empty", detailRepo.Upserted.Region)
	}
	if detailRepo.Upserted.Concepts != "" {
		t.Fatalf("Upserted Concepts = %q, want empty", detailRepo.Upserted.Concepts)
	}
	if !res.RefreshTriggered {
		t.Fatal("RefreshTriggered = false, want true")
	}
}

func TestMarketStockService_GetStockKlinesUsesDatabaseCacheWhenEnoughBars(t *testing.T) {
	now := time.Now()
	bars := make([]model.StockKlineBar, 0, 120)
	for i := 0; i < 120; i++ {
		bars = append(bars, model.StockKlineBar{
			Symbol:     "000858.SZ",
			Period:     "day",
			AdjustType: "qfq",
			BarTime:    now.Add(-time.Duration(i) * 24 * time.Hour),
			ClosePrice: decimal.NewFromFloat(100 + float64(i)),
			Source:     "tencent",
		})
	}
	provider := &MockMarketDataProvider{}
	svc := &marketStockService{
		provider:   provider,
		detailRepo: &MockStockQuoteDetailRepo{},
		klineRepo:  &MockStockKlineRepo{Bars: bars},
	}

	res, err := svc.GetStockKlines("000858.SZ", "day", "qfq", 60, false)
	if err != nil {
		t.Fatalf("GetStockKlines() error = %v", err)
	}
	if provider.KlineCalls != 0 {
		t.Fatalf("KlineCalls = %d, want 0", provider.KlineCalls)
	}
	if len(res.Items) != 120 {
		t.Fatalf("Items len = %d, want 120", len(res.Items))
	}
	if res.RefreshTriggered {
		t.Fatal("RefreshTriggered = true, want false")
	}
}

func TestMarketStockService_GetStockKlinesFetchesAndUpsertsWhenCacheInsufficient(t *testing.T) {
	now := time.Now()
	cached := []model.StockKlineBar{{
		Symbol:     "000858.SZ",
		Period:     "day",
		AdjustType: "qfq",
		BarTime:    now.Add(-24 * time.Hour),
		ClosePrice: decimal.NewFromFloat(120),
		Source:     "tencent",
	}}
	fetched := []marketdata.KlineBar{
		{Symbol: "000858.SZ", Period: "day", AdjustType: "qfq", BarTime: now.Add(-24 * time.Hour), Close: 120, Source: "tencent"},
		{Symbol: "000858.SZ", Period: "day", AdjustType: "qfq", BarTime: now, Close: 121, Source: "tencent"},
	}
	provider := &MockMarketDataProvider{Klines: fetched}
	repo := &MockStockKlineRepo{Bars: cached}
	svc := &marketStockService{
		provider:   provider,
		detailRepo: &MockStockQuoteDetailRepo{},
		klineRepo:  repo,
	}

	res, err := svc.GetStockKlines("000858.SZ", "day", "qfq", 2, false)
	if err != nil {
		t.Fatalf("GetStockKlines() error = %v", err)
	}
	if provider.KlineCalls != 1 {
		t.Fatalf("KlineCalls = %d, want 1", provider.KlineCalls)
	}
	if len(repo.UpsertedBars) != 2 {
		t.Fatalf("UpsertedBars len = %d, want 2", len(repo.UpsertedBars))
	}
	if !res.RefreshTriggered {
		t.Fatal("RefreshTriggered = false, want true")
	}
}

func TestMarketDataService_FetchAndStoreMarketSnapshotsWithTencentStyleQuotes(t *testing.T) {
	now := time.Now()
	mockProvider := &MockMarketDataProvider{
		Quotes: []marketdata.Quote{
			{Symbol: "000001.SH", Name: "上证指数", Market: "cn_index", LastPrice: 4112.9, ChangeAmount: 35.62, ChangePercent: 0.87, SnapshotTime: now, Source: "tencent"},
			{Symbol: "399001.SZ", Name: "深证成指", Market: "cn_index", LastPrice: 15597.3, ChangeAmount: 350.03, ChangePercent: 2.30, SnapshotTime: now, Source: "tencent"},
		},
	}
	mockRepo := &MockMarketDataSnapshotRepo{}

	svc := NewMarketDataService(
		config.MarketConfig{Symbols: "000001.SH,399001.SZ"},
		mockProvider,
		mockRepo,
		nil,
	)

	batchNo, count, err := svc.FetchAndStoreMarketSnapshots(context.Background())
	if err != nil {
		t.Fatalf("FetchAndStoreMarketSnapshots() error = %v", err)
	}
	if batchNo == "" {
		t.Fatal("BatchNo should not be empty")
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}
	if len(mockRepo.Snapshots) != 2 {
		t.Fatalf("Snapshots len = %d, want 2", len(mockRepo.Snapshots))
	}
	if mockRepo.Snapshots[0].Source != "tencent" {
		t.Fatalf("Source = %s, want tencent", mockRepo.Snapshots[0].Source)
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
