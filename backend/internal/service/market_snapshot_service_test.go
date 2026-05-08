package service_test

import (
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/service"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// MockMarketSnapshotRepositoryForService 模拟市场快照仓储
type MockMarketSnapshotRepositoryForService struct {
	Snapshots   []model.MarketSnapshot
	LatestBatch string
	Err         error
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

	svc := service.NewMarketSnapshotService(mockRepo)
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

	svc := service.NewMarketSnapshotService(mockRepo)
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

	svc := service.NewMarketSnapshotService(mockRepo)
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

	svc := service.NewMarketSnapshotService(mockRepo)
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
				SnapshotTime:  now,
				BatchNo:       "batch001",
				Source:        "mock",
			},
			{
				Symbol:        "399001.SZ",
				Name:          "深证成指",
				Market:        "cn_index",
				LastPrice:     decimal.NewFromFloat(10000.00),
				ChangeAmount:  decimal.NewFromFloat(-50.0),
				ChangePercent: decimal.NewFromFloat(-0.5),
				Turnover:      decimal.NewFromInt(300000000000),
				SnapshotTime:  now,
				BatchNo:       "batch001",
				Source:        "mock",
			},
		},
	}

	svc := service.NewMarketSnapshotService(mockRepo)
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
}

// TestMarketSnapshotService_GetDashboardSnapshot_Empty 测试空仪表盘
func TestMarketSnapshotService_GetDashboardSnapshot_Empty(t *testing.T) {
	mockRepo := &MockMarketSnapshotRepositoryForService{
		Err: gorm.ErrRecordNotFound,
	}

	svc := service.NewMarketSnapshotService(mockRepo)
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
				LastPrice:     decimal.Zero,
				ChangeAmount:  decimal.NewFromInt(10),  // 上涨
				ChangePercent: decimal.NewFromFloat(0.5),
				Turnover:      decimal.NewFromInt(100000000000),
				SnapshotTime:  now,
				BatchNo:       "batch001",
				Source:        "mock",
			},
			{
				Symbol:        "399001.SZ",
				Name:          "深证成指",
				LastPrice:     decimal.Zero,
				ChangeAmount:  decimal.NewFromInt(-10), // 下跌
				ChangePercent: decimal.NewFromFloat(-0.5),
				Turnover:      decimal.NewFromInt(200000000000),
				SnapshotTime:  now,
				BatchNo:       "batch001",
				Source:        "mock",
			},
			{
				Symbol:        "399006.SZ",
				Name:          "创业板指",
				LastPrice:     decimal.Zero,
				ChangeAmount:  decimal.Zero, // 不涨不跌
				ChangePercent: decimal.Zero,
				Turnover:      decimal.NewFromInt(150000000000),
				SnapshotTime:  now,
				BatchNo:       "batch001",
				Source:        "mock",
			},
		},
	}

	svc := service.NewMarketSnapshotService(mockRepo)
	dashboard, err := svc.GetDashboardSnapshot()
	if err != nil {
		t.Errorf("GetDashboardSnapshot() error = %v", err)
	}

	// 验证统计
	// 指数数量: 3
	// 上涨数: 2 (包含0)
	// 下跌数: 1
	if dashboard.Stats[0].Value != "3" {
		t.Errorf("Stats[0] 指数数量 = %v, want 3", dashboard.Stats[0].Value)
	}
}

// TestMarketSnapshotService_Interface 测试接口实现
func TestMarketSnapshotService_Interface(t *testing.T) {
	mockRepo := &MockMarketSnapshotRepositoryForService{}
	var _ service.MarketSnapshotService = service.NewMarketSnapshotService(mockRepo)
}
