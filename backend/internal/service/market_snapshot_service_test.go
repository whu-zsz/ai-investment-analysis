package service_test

import (
	"context"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/service"
	"strconv"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// MockMarketSnapshotRepositoryForService 模拟市场快照仓储
type MockMarketSnapshotRepositoryForService struct {
	Snapshots     []model.MarketSnapshot
	LatestBatch   string
	RecentBatches []string
	Err           error
}

type MockStockQuoteDetailRepositoryForService struct {
	Details []model.StockQuoteDetail
	Err     error
}

type MockMarketBoardSnapshotRepositoryForService struct {
	Boards []model.MarketBoardSnapshot
	Err    error
}

type MockMarketBoardConstituentRepositoryForService struct {
	Items []model.MarketBoardConstituent
	Err   error
}

type MockMarketDataServiceForSnapshot struct {
	FetchFullMarketSnapshotsFunc func(ctx context.Context) (string, int, error)
}

func (r *MockMarketBoardSnapshotRepositoryForService) BatchCreate(boards []model.MarketBoardSnapshot) error {
	r.Boards = append(r.Boards, boards...)
	return r.Err
}

func (r *MockMarketBoardSnapshotRepositoryForService) FindLatestBatchNo(boardType string) (string, error) {
	return "", r.Err
}

func (r *MockMarketBoardSnapshotRepositoryForService) FindByBatchNo(boardType, batchNo string, limit int) ([]model.MarketBoardSnapshot, error) {
	return nil, r.Err
}

func (r *MockMarketBoardSnapshotRepositoryForService) FindLatest(limit int) ([]model.MarketBoardSnapshot, error) {
	return nil, r.Err
}

func (r *MockMarketBoardConstituentRepositoryForService) ReplaceAll(items []model.MarketBoardConstituent) error {
	r.Items = append([]model.MarketBoardConstituent(nil), items...)
	return r.Err
}

func (r *MockMarketBoardConstituentRepositoryForService) FindAll() ([]model.MarketBoardConstituent, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	return append([]model.MarketBoardConstituent(nil), r.Items...), nil
}

func (r *MockMarketBoardConstituentRepositoryForService) FindLatestSyncedAt() (time.Time, error) {
	if r.Err != nil {
		return time.Time{}, r.Err
	}
	if len(r.Items) == 0 {
		return time.Time{}, gorm.ErrRecordNotFound
	}
	return r.Items[0].SyncedAt, nil
}

func (m *MockMarketDataServiceForSnapshot) FetchAndStoreMarketSnapshots(ctx context.Context) (string, int, error) {
	return "", 0, nil
}

func (m *MockMarketDataServiceForSnapshot) FetchAndStoreQuotesBySymbols(ctx context.Context, symbols []string) ([]model.MarketSnapshot, error) {
	return nil, nil
}

func (m *MockMarketDataServiceForSnapshot) FetchAndStoreFullMarketSnapshots(ctx context.Context) (string, int, error) {
	if m.FetchFullMarketSnapshotsFunc != nil {
		return m.FetchFullMarketSnapshotsFunc(ctx)
	}
	return "", 0, nil
}

func (m *MockMarketDataServiceForSnapshot) FetchAndStoreMarketBoardSnapshots(ctx context.Context) (string, int, error) {
	return "", 0, nil
}

func (m *MockMarketDataServiceForSnapshot) EnsureTrackedIndexHistory(ctx context.Context) error {
	return nil
}

func (r *MockStockQuoteDetailRepositoryForService) Upsert(detail *model.StockQuoteDetail) error {
	return nil
}

func (r *MockStockQuoteDetailRepositoryForService) FindBySymbol(symbol string) (*model.StockQuoteDetail, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	for _, detail := range r.Details {
		if detail.Symbol == symbol {
			copied := detail
			return &copied, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *MockStockQuoteDetailRepositoryForService) FindBySymbols(symbols []string) ([]model.StockQuoteDetail, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	if len(symbols) == 0 {
		return []model.StockQuoteDetail{}, nil
	}
	wanted := make(map[string]struct{}, len(symbols))
	for _, symbol := range symbols {
		wanted[symbol] = struct{}{}
	}
	result := make([]model.StockQuoteDetail, 0, len(r.Details))
	for _, detail := range r.Details {
		if _, ok := wanted[detail.Symbol]; ok {
			result = append(result, detail)
		}
	}
	return result, nil
}

func (r *MockMarketSnapshotRepositoryForService) BatchCreate(snapshots []model.MarketSnapshot) error {
	return nil
}

func (r *MockMarketSnapshotRepositoryForService) FindLatestBatchNo() (string, error) {
	if r.Err != nil {
		return "", r.Err
	}
	return r.LatestBatch, nil
}

func (r *MockMarketSnapshotRepositoryForService) FindLatestBatchNoBySource(source string) (string, error) {
	if r.Err != nil {
		return "", r.Err
	}
	return r.LatestBatch, nil
}

func (r *MockMarketSnapshotRepositoryForService) FindRecentBatchNos(limit int) ([]string, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	if len(r.RecentBatches) == 0 {
		return []string{r.LatestBatch}, nil
	}
	return r.RecentBatches, nil
}

func (r *MockMarketSnapshotRepositoryForService) FindByBatchNo(batchNo string) ([]model.MarketSnapshot, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	var result []model.MarketSnapshot
	for _, s := range r.Snapshots {
		if s.BatchNo == batchNo {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *MockMarketSnapshotRepositoryForService) FindLatestBySymbol(symbol string) (*model.MarketSnapshot, error) {
	for i := len(r.Snapshots) - 1; i >= 0; i-- {
		if r.Snapshots[i].Symbol == symbol {
			return &r.Snapshots[i], nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *MockMarketSnapshotRepositoryForService) FindHistory(limit int, startTime, endTime *time.Time) ([]model.MarketSnapshot, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	return r.Snapshots, nil
}

func (r *MockMarketSnapshotRepositoryForService) FindHistoryBySymbol(symbol string, limit int, startTime, endTime *time.Time) ([]model.MarketSnapshot, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	var result []model.MarketSnapshot
	for _, s := range r.Snapshots {
		if s.Symbol == symbol {
			result = append(result, s)
		}
	}
	return result, nil
}

func (r *MockMarketSnapshotRepositoryForService) SearchStocks(query string, limit int) ([]model.MarketSnapshot, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	return r.Snapshots, nil
}

// TestMarketSnapshotService_GetLatestSnapshots 测试获取最新快照
func TestMarketSnapshotService_GetLatestSnapshots(t *testing.T) {
	now := time.Now()
	mockRepo := &MockMarketSnapshotRepositoryForService{
		LatestBatch: "batch001",
		Snapshots: []model.MarketSnapshot{
			{
				Symbol:        "000001.SH",
				Name:          "上证指数",
				LastPrice:     decimal.NewFromFloat(3000.50),
				ChangeAmount:  decimal.NewFromFloat(10.5),
				ChangePercent: decimal.NewFromFloat(0.35),
				SnapshotTime:  now,
				BatchNo:       "batch001",
				Source:        "mock",
			},
		},
	}

	svc := service.NewMarketSnapshotService(mockRepo, nil, nil, nil, nil)
	snapshots, err := svc.GetLatestSnapshots()
	if err != nil {
		t.Errorf("GetLatestSnapshots() error = %v", err)
	}

	if len(snapshots) != 1 {
		t.Errorf("GetLatestSnapshots() returned %d snapshots, want 1", len(snapshots))
	}
}

// TestMarketSnapshotService_GetLatestSnapshots_Empty 测试空快照
func TestMarketSnapshotService_GetLatestSnapshots_Empty(t *testing.T) {
	mockRepo := &MockMarketSnapshotRepositoryForService{
		Err: gorm.ErrRecordNotFound,
	}

	svc := service.NewMarketSnapshotService(mockRepo, nil, nil, nil, nil)
	snapshots, err := svc.GetLatestSnapshots()
	if err != nil {
		t.Errorf("GetLatestSnapshots() error = %v", err)
	}

	if len(snapshots) != 0 {
		t.Errorf("GetLatestSnapshots() should return empty slice for no data")
	}
}

// TestMarketSnapshotService_GetHistory 测试获取历史
func TestMarketSnapshotService_GetHistory(t *testing.T) {
	now := time.Now()
	mockRepo := &MockMarketSnapshotRepositoryForService{
		Snapshots: []model.MarketSnapshot{
			{
				Symbol:        "000001.SH",
				Name:          "上证指数",
				LastPrice:     decimal.NewFromFloat(3000.50),
				ChangeAmount:  decimal.Zero,
				ChangePercent: decimal.Zero,
				SnapshotTime:  now,
				BatchNo:       "batch001",
				Source:        "mock",
			},
		},
	}

	svc := service.NewMarketSnapshotService(mockRepo, nil, nil, nil, nil)
	snapshots, err := svc.GetHistory("", 10, nil, nil)
	if err != nil {
		t.Errorf("GetHistory() error = %v", err)
	}

	if len(snapshots) != 1 {
		t.Errorf("GetHistory() returned %d snapshots, want 1", len(snapshots))
	}
}

// TestMarketSnapshotService_GetHistory_BySymbol 测试按代码获取历史
func TestMarketSnapshotService_GetHistory_BySymbol(t *testing.T) {
	now := time.Now()
	mockRepo := &MockMarketSnapshotRepositoryForService{
		Snapshots: []model.MarketSnapshot{
			{
				Symbol:        "000001.SH",
				Name:          "上证指数",
				LastPrice:     decimal.Zero,
				ChangeAmount:  decimal.Zero,
				ChangePercent: decimal.Zero,
				SnapshotTime:  now,
				BatchNo:       "batch001",
				Source:        "mock",
			},
			{
				Symbol:        "399001.SZ",
				Name:          "深证成指",
				LastPrice:     decimal.Zero,
				ChangeAmount:  decimal.Zero,
				ChangePercent: decimal.Zero,
				SnapshotTime:  now,
				BatchNo:       "batch001",
				Source:        "mock",
			},
		},
	}

	svc := service.NewMarketSnapshotService(mockRepo, nil, nil, nil, nil)
	snapshots, err := svc.GetHistory("000001.SH", 10, nil, nil)
	if err != nil {
		t.Errorf("GetHistory() error = %v", err)
	}

	for _, s := range snapshots {
		if s.Symbol != "000001.SH" {
			t.Errorf("GetHistory() should only return snapshots for specified symbol")
			break
		}
	}
}

// TestMarketSnapshotService_GetDashboardSnapshot 测试获取仪表盘快照
func TestMarketSnapshotService_GetDashboardSnapshot(t *testing.T) {
	now := time.Now()
	mockRepo := &MockMarketSnapshotRepositoryForService{
		LatestBatch: "batch001",
		Snapshots: []model.MarketSnapshot{
			{
				Symbol:        "000001.SH",
				Name:          "上证指数",
				Market:        "cn_index",
				LastPrice:     decimal.NewFromFloat(3000.50),
				ChangeAmount:  decimal.NewFromFloat(10.5),
				ChangePercent: decimal.NewFromFloat(0.35),
				Turnover:      decimal.NewFromInt(250000000000),
				SnapshotTime:  now.Add(-48 * time.Hour),
				BatchNo:       "batch001",
				Source:        "mock",
				CreatedAt:     now,
			},
			{
				Symbol:        "399001.SZ",
				Name:          "深证成指",
				Market:        "cn_index",
				LastPrice:     decimal.NewFromFloat(10000.00),
				ChangeAmount:  decimal.NewFromFloat(-50.0),
				ChangePercent: decimal.NewFromFloat(-0.5),
				Turnover:      decimal.NewFromInt(300000000000),
				SnapshotTime:  now.Add(-48 * time.Hour),
				BatchNo:       "batch001",
				Source:        "mock",
				CreatedAt:     now,
			},
		},
	}

	svc := service.NewMarketSnapshotService(mockRepo, nil, nil, nil, nil)
	dashboard, err := svc.GetDashboardSnapshot()
	if err != nil {
		t.Errorf("GetDashboardSnapshot() error = %v", err)
	}

	if len(dashboard.Indices) != 2 {
		t.Errorf("GetDashboardSnapshot() returned %d indices, want 2", len(dashboard.Indices))
	}

	if len(dashboard.Stats) != 5 {
		t.Errorf("GetDashboardSnapshot() returned %d stats, want 5", len(dashboard.Stats))
	}
	if dashboard.IsStale {
		t.Error("GetDashboardSnapshot() IsStale = true, want false when created_at is recent")
	}
	if dashboard.RefreshedAt == "" {
		t.Error("GetDashboardSnapshot() RefreshedAt should not be empty")
	}
}

// TestMarketSnapshotService_GetDashboardSnapshot_Empty 测试空仪表盘
func TestMarketSnapshotService_GetDashboardSnapshot_Empty(t *testing.T) {
	mockRepo := &MockMarketSnapshotRepositoryForService{
		Err: gorm.ErrRecordNotFound,
	}

	svc := service.NewMarketSnapshotService(mockRepo, nil, nil, nil, nil)
	dashboard, err := svc.GetDashboardSnapshot()
	if err != nil {
		t.Errorf("GetDashboardSnapshot() error = %v", err)
	}

	if len(dashboard.Indices) != 0 {
		t.Errorf("GetDashboardSnapshot() should return empty indices for no data")
	}
}

// TestMarketSnapshotService_GetDashboardSnapshot_Stats 测试统计计算
func TestMarketSnapshotService_GetDashboardSnapshot_Stats(t *testing.T) {
	now := time.Now()
	mockRepo := &MockMarketSnapshotRepositoryForService{
		LatestBatch: "batch001",
		Snapshots: []model.MarketSnapshot{
			{
				Symbol:        "000001.SH",
				Name:          "上证指数",
				Market:        "cn_index",
				LastPrice:     decimal.Zero,
				ChangeAmount:  decimal.NewFromInt(10),
				ChangePercent: decimal.NewFromFloat(0.5),
				Turnover:      decimal.NewFromInt(100000000000),
				SnapshotTime:  now,
				BatchNo:       "batch001",
				Source:        "mock",
			},
			{
				Symbol:        "399001.SZ",
				Name:          "深证成指",
				Market:        "cn_index",
				LastPrice:     decimal.Zero,
				ChangeAmount:  decimal.NewFromInt(-10),
				ChangePercent: decimal.NewFromFloat(-0.5),
				Turnover:      decimal.NewFromInt(200000000000),
				SnapshotTime:  now,
				BatchNo:       "batch001",
				Source:        "mock",
			},
			{
				Symbol:        "399006.SZ",
				Name:          "创业板指",
				Market:        "cn_index",
				LastPrice:     decimal.Zero,
				ChangeAmount:  decimal.Zero,
				ChangePercent: decimal.Zero,
				Turnover:      decimal.NewFromInt(150000000000),
				SnapshotTime:  now,
				BatchNo:       "batch001",
				Source:        "mock",
			},
		},
	}

	svc := service.NewMarketSnapshotService(mockRepo, nil, nil, nil, nil)
	dashboard, err := svc.GetDashboardSnapshot()
	if err != nil {
		t.Errorf("GetDashboardSnapshot() error = %v", err)
	}

	if dashboard.Stats[0].Value != "3" {
		t.Errorf("Stats[0] 指数数量 = %v, want 3", dashboard.Stats[0].Value)
	}
}

func TestMarketSnapshotService_GetDashboardSnapshot_UsesTrackedSymbolsInsteadOfFullMarketBatch(t *testing.T) {
	now := time.Now()
	mockRepo := &MockMarketSnapshotRepositoryForService{
		LatestBatch: "full-market-batch",
		Snapshots: []model.MarketSnapshot{
			{
				Symbol:        "600519.SH",
				Name:          "贵州茅台",
				Market:        "cn_stock",
				LastPrice:     decimal.NewFromFloat(1728.50),
				ChangeAmount:  decimal.NewFromFloat(18.20),
				ChangePercent: decimal.NewFromFloat(1.06),
				Turnover:      decimal.NewFromInt(12300000000),
				SnapshotTime:  now,
				BatchNo:       "full-market-batch",
				Source:        "sina",
				CreatedAt:     now,
			},
			{
				Symbol:        "000001.SH",
				Name:          "上证指数",
				Market:        "cn_index",
				LastPrice:     decimal.NewFromFloat(3200.12),
				ChangeAmount:  decimal.NewFromFloat(8.33),
				ChangePercent: decimal.NewFromFloat(0.26),
				Turnover:      decimal.NewFromInt(250000000000),
				SnapshotTime:  now,
				BatchNo:       "tracked-batch",
				Source:        "sina",
				CreatedAt:     now,
			},
			{
				Symbol:        "399001.SZ",
				Name:          "深证成指",
				Market:        "cn_index",
				LastPrice:     decimal.NewFromFloat(10320.88),
				ChangeAmount:  decimal.NewFromFloat(-21.50),
				ChangePercent: decimal.NewFromFloat(-0.21),
				Turnover:      decimal.NewFromInt(310000000000),
				SnapshotTime:  now,
				BatchNo:       "tracked-batch",
				Source:        "sina",
				CreatedAt:     now,
			},
			{
				Symbol:        "399006.SZ",
				Name:          "创业板指",
				Market:        "cn_index",
				LastPrice:     decimal.NewFromFloat(1988.66),
				ChangeAmount:  decimal.NewFromFloat(6.88),
				ChangePercent: decimal.NewFromFloat(0.35),
				Turnover:      decimal.NewFromInt(166000000000),
				SnapshotTime:  now,
				BatchNo:       "tracked-batch",
				Source:        "sina",
				CreatedAt:     now,
			},
			{
				Symbol:        "000300.SH",
				Name:          "沪深300",
				Market:        "cn_index",
				LastPrice:     decimal.NewFromFloat(3720.10),
				ChangeAmount:  decimal.NewFromFloat(10.02),
				ChangePercent: decimal.NewFromFloat(0.27),
				Turnover:      decimal.NewFromInt(280000000000),
				SnapshotTime:  now,
				BatchNo:       "tracked-batch",
				Source:        "sina",
				CreatedAt:     now,
			},
		},
	}

	svc := service.NewMarketSnapshotService(mockRepo, nil, nil, nil, nil)
	dashboard, err := svc.GetDashboardSnapshot()
	if err != nil {
		t.Fatalf("GetDashboardSnapshot() error = %v", err)
	}

	if len(dashboard.Indices) != 4 {
		t.Fatalf("GetDashboardSnapshot() returned %d tracked indices, want 4", len(dashboard.Indices))
	}
	for _, item := range dashboard.Indices {
		if item.Symbol == "600519.SH" {
			t.Fatalf("GetDashboardSnapshot() should ignore full market stock snapshots, got %s", item.Symbol)
		}
	}
}

func TestMarketSnapshotService_GetDashboardSnapshot_PrefersIndexSnapshotWhenSymbolCollides(t *testing.T) {
	now := time.Now()
	mockRepo := &MockMarketSnapshotRepositoryForService{
		Snapshots: []model.MarketSnapshot{
			{
				Symbol:        "000001.SH",
				Name:          "平安银行",
				Market:        "cn_index",
				LastPrice:     decimal.NewFromFloat(10.68),
				ChangeAmount:  decimal.Zero,
				ChangePercent: decimal.Zero,
				Turnover:      decimal.NewFromFloat(709614127.77),
				SnapshotTime:  now.Add(2 * time.Minute),
				BatchNo:       "full-market-batch",
				Source:        "akshare_sina",
				CreatedAt:     now.Add(2 * time.Minute),
			},
			{
				Symbol:        "000001.SH",
				Name:          "上证指数",
				Market:        "cn_index",
				LastPrice:     decimal.NewFromFloat(4152.57),
				ChangeAmount:  decimal.NewFromFloat(39.67),
				ChangePercent: decimal.NewFromFloat(0.96),
				Turnover:      decimal.NewFromInt(1445655380000),
				SnapshotTime:  now,
				BatchNo:       "tracked-batch",
				Source:        "tencent",
				CreatedAt:     now,
			},
			{
				Symbol:        "399001.SZ",
				Name:          "深证成指",
				Market:        "cn_index",
				LastPrice:     decimal.NewFromFloat(15856.61),
				ChangeAmount:  decimal.NewFromFloat(259.31),
				ChangePercent: decimal.NewFromFloat(1.66),
				Turnover:      decimal.NewFromInt(1760074860000),
				SnapshotTime:  now,
				BatchNo:       "tracked-batch",
				Source:        "tencent",
				CreatedAt:     now,
			},
			{
				Symbol:        "399006.SZ",
				Name:          "创业板指",
				Market:        "cn_index",
				LastPrice:     decimal.NewFromFloat(4021.16),
				ChangeAmount:  decimal.NewFromFloat(82.66),
				ChangePercent: decimal.NewFromFloat(2.10),
				Turnover:      decimal.NewFromInt(878510330000),
				SnapshotTime:  now,
				BatchNo:       "tracked-batch",
				Source:        "tencent",
				CreatedAt:     now,
			},
			{
				Symbol:        "000300.SH",
				Name:          "沪深300",
				Market:        "cn_index",
				LastPrice:     decimal.NewFromFloat(4921.60),
				ChangeAmount:  decimal.NewFromFloat(76.50),
				ChangePercent: decimal.NewFromFloat(1.58),
				Turnover:      decimal.NewFromInt(861179420000),
				SnapshotTime:  now,
				BatchNo:       "tracked-batch",
				Source:        "tencent",
				CreatedAt:     now,
			},
		},
	}

	svc := service.NewMarketSnapshotService(mockRepo, nil, nil, nil, nil)
	dashboard, err := svc.GetDashboardSnapshot()
	if err != nil {
		t.Fatalf("GetDashboardSnapshot() error = %v", err)
	}
	if len(dashboard.Indices) != 4 {
		t.Fatalf("GetDashboardSnapshot() returned %d tracked indices, want 4", len(dashboard.Indices))
	}
	if dashboard.Indices[0].Symbol != "000001.SH" || dashboard.Indices[0].Name != "上证指数" {
		t.Fatalf("first tracked index = %+v, want 上证指数", dashboard.Indices[0])
	}
	for _, item := range dashboard.Indices {
		if item.Name == "平安银行" {
			t.Fatalf("GetDashboardSnapshot() should ignore stock snapshot collision, got %+v", item)
		}
	}
}

// TestMarketSnapshotService_Interface 测试接口实现
func TestMarketSnapshotService_Interface(t *testing.T) {
	mockRepo := &MockMarketSnapshotRepositoryForService{}
	var _ service.MarketSnapshotService = service.NewMarketSnapshotService(mockRepo, nil, nil, nil, nil)
}

func TestMarketSnapshotService_GetDashboardMarketBreadth_FallbackBoards(t *testing.T) {
	now := time.Now()
	mockRepo := &MockMarketSnapshotRepositoryForService{
		LatestBatch: "batch001",
		Snapshots: []model.MarketSnapshot{
			{
				Symbol:        "000858.SZ",
				Name:          "五粮液",
				LastPrice:     decimal.NewFromFloat(128.6),
				ChangeAmount:  decimal.NewFromFloat(1.8),
				ChangePercent: decimal.NewFromFloat(1.42),
				Turnover:      decimal.NewFromInt(8_500_000_000),
				Volume:        decimal.NewFromInt(42_000_000),
				SnapshotTime:  now,
				BatchNo:       "batch001",
				Source:        "sina",
				CreatedAt:     now,
			},
			{
				Symbol:        "600519.SH",
				Name:          "贵州茅台",
				LastPrice:     decimal.NewFromFloat(1688.0),
				ChangeAmount:  decimal.NewFromFloat(-6.2),
				ChangePercent: decimal.NewFromFloat(-0.37),
				Turnover:      decimal.NewFromInt(6_300_000_000),
				Volume:        decimal.NewFromInt(5_200_000),
				SnapshotTime:  now,
				BatchNo:       "batch001",
				Source:        "sina",
				CreatedAt:     now,
			},
			{
				Symbol:        "300750.SZ",
				Name:          "宁德时代",
				LastPrice:     decimal.NewFromFloat(205.1),
				ChangeAmount:  decimal.NewFromFloat(5.4),
				ChangePercent: decimal.NewFromFloat(2.7),
				Turnover:      decimal.NewFromInt(12_600_000_000),
				Volume:        decimal.NewFromInt(38_000_000),
				SnapshotTime:  now,
				BatchNo:       "batch001",
				Source:        "sina",
				CreatedAt:     now,
			},
		},
	}
	boardRepo := &MockMarketBoardSnapshotRepositoryForService{Err: gorm.ErrRecordNotFound}
	detailRepo := &MockStockQuoteDetailRepositoryForService{Details: []model.StockQuoteDetail{
		{Symbol: "000858.SZ", Industry: "食品饮料", Concepts: "白酒,新零售"},
		{Symbol: "600519.SH", Industry: "食品饮料", Concepts: "白酒,央企改革"},
		{Symbol: "300750.SZ", Industry: "电力设备", Concepts: "新能源车,锂电池概念"},
	}}

	svc := service.NewMarketSnapshotService(mockRepo, boardRepo, nil, detailRepo, nil)
	breadth, err := svc.GetDashboardMarketBreadth(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetDashboardMarketBreadth() error = %v", err)
	}
	if len(breadth.Sectors) == 0 {
		t.Fatalf("expected fallback sectors, got none")
	}
	if len(breadth.Concepts) == 0 {
		t.Fatalf("expected fallback concepts, got none")
	}
	if !breadth.IsPartial {
		t.Fatalf("expected partial response when fallback boards are used")
	}
	if breadth.Message == "" {
		t.Fatalf("expected fallback message when fallback boards are used")
	}
	if breadth.Sectors[0].Name != "电力设备" && breadth.Sectors[0].Name != "食品饮料" {
		t.Fatalf("unexpected sector fallback result: %s", breadth.Sectors[0].Name)
	}

	foundConcept := false
	for _, concept := range breadth.Concepts {
		if concept.Name == "白酒" || concept.Name == "新能源车" {
			foundConcept = true
			break
		}
	}
	if !foundConcept {
		t.Fatalf("expected known concepts in fallback results")
	}
}

func TestMarketSnapshotService_GetDashboardMarketBreadth_UsesRecentCompleteBatch(t *testing.T) {
	now := time.Now()
	buildBatch := func(batchNo, source string, count int, createdAt time.Time) []model.MarketSnapshot {
		items := make([]model.MarketSnapshot, 0, count)
		for i := 0; i < count; i++ {
			items = append(items, model.MarketSnapshot{
				Symbol:        "SZ" + strconv.Itoa(100000+i),
				Name:          "样本",
				LastPrice:     decimal.NewFromInt(10),
				ChangeAmount:  decimal.NewFromFloat(0.1),
				ChangePercent: decimal.NewFromFloat(1.2),
				Turnover:      decimal.NewFromInt(1000000),
				Volume:        decimal.NewFromInt(100000),
				SnapshotTime:  createdAt,
				BatchNo:       batchNo,
				Source:        source,
				CreatedAt:     createdAt,
			})
		}
		return items
	}

	snapshots := append(buildBatch("batch-new", "sina", 4200, now), buildBatch("batch-old", "eastmoney", 5200, now.Add(-5*time.Minute))...)
	mockRepo := &MockMarketSnapshotRepositoryForService{
		LatestBatch:   "batch-new",
		RecentBatches: []string{"batch-new", "batch-old"},
		Snapshots:     snapshots,
	}
	boardRepo := &MockMarketBoardSnapshotRepositoryForService{Err: gorm.ErrRecordNotFound}
	svc := service.NewMarketSnapshotService(mockRepo, boardRepo, nil, &MockStockQuoteDetailRepositoryForService{}, nil)

	breadth, err := svc.GetDashboardMarketBreadth(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetDashboardMarketBreadth() error = %v", err)
	}
	if breadth.Source != "eastmoney" {
		t.Fatalf("GetDashboardMarketBreadth() source = %s, want eastmoney", breadth.Source)
	}
	if !breadth.IsPartial {
		t.Fatalf("expected partial response when latest batch is incomplete")
	}
	if breadth.Message == "" {
		t.Fatalf("expected fallback message for recent complete batch")
	}
}

func TestMarketSnapshotService_GetDashboardMarketBreadth_ClearsFallbackAfterRefresh(t *testing.T) {
	now := time.Now()
	buildBatch := func(batchNo, source string, count int, createdAt time.Time) []model.MarketSnapshot {
		items := make([]model.MarketSnapshot, 0, count)
		for i := 0; i < count; i++ {
			items = append(items, model.MarketSnapshot{
				Symbol:        "SZ" + strconv.Itoa(200000+i),
				Name:          "样本",
				LastPrice:     decimal.NewFromInt(10),
				ChangeAmount:  decimal.NewFromFloat(0.1),
				ChangePercent: decimal.NewFromFloat(1.2),
				Turnover:      decimal.NewFromInt(1000000),
				Volume:        decimal.NewFromInt(100000),
				SnapshotTime:  createdAt,
				BatchNo:       batchNo,
				Source:        source,
				CreatedAt:     createdAt,
			})
		}
		return items
	}

	mockRepo := &MockMarketSnapshotRepositoryForService{
		LatestBatch:   "batch-stale",
		RecentBatches: []string{"batch-stale", "batch-old"},
		Snapshots: append(
			buildBatch("batch-stale", "sina", 4200, now.Add(-31*time.Minute)),
			buildBatch("batch-old", "eastmoney", 5200, now.Add(-35*time.Minute))...,
		),
	}
	boardRepo := &MockMarketBoardSnapshotRepositoryForService{Err: gorm.ErrRecordNotFound}
	marketDataSvc := &MockMarketDataServiceForSnapshot{
		FetchFullMarketSnapshotsFunc: func(ctx context.Context) (string, int, error) {
			fresh := buildBatch("batch-fresh", "akshare_sina", 5521, now)
			mockRepo.LatestBatch = "batch-fresh"
			mockRepo.RecentBatches = []string{"batch-fresh", "batch-stale", "batch-old"}
			mockRepo.Snapshots = append(fresh, mockRepo.Snapshots...)
			return "batch-fresh", len(fresh), nil
		},
	}
	svc := service.NewMarketSnapshotService(mockRepo, boardRepo, nil, &MockStockQuoteDetailRepositoryForService{}, marketDataSvc)

	breadth, err := svc.GetDashboardMarketBreadth(context.Background(), 5)
	if err != nil {
		t.Fatalf("GetDashboardMarketBreadth() error = %v", err)
	}
	if breadth.Source != "akshare_sina" {
		t.Fatalf("GetDashboardMarketBreadth() source = %s, want akshare_sina", breadth.Source)
	}
	if breadth.IsPartial {
		t.Fatalf("expected refreshed breadth response to be non-partial")
	}
	if breadth.Message != "" {
		t.Fatalf("expected refreshed breadth response to clear fallback message, got %q", breadth.Message)
	}
	if len(breadth.Coverage) == 0 || breadth.Coverage[0].Value != "5521" {
		t.Fatalf("expected refreshed coverage to reflect latest batch, got %+v", breadth.Coverage)
	}
}
