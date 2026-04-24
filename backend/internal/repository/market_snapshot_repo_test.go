package repository_test

import (
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// InMemoryMarketSnapshotRepository 内存市场快照仓储用于测试
type InMemoryMarketSnapshotRepository struct {
	snapshots map[uint64]*model.MarketSnapshot
	nextID    uint64
}

func NewInMemoryMarketSnapshotRepository() *InMemoryMarketSnapshotRepository {
	return &InMemoryMarketSnapshotRepository{
		snapshots: make(map[uint64]*model.MarketSnapshot),
		nextID:    1,
	}
}

func (r *InMemoryMarketSnapshotRepository) BatchCreate(snapshots []model.MarketSnapshot) error {
	for i := range snapshots {
		snapshots[i].ID = r.nextID
		r.snapshots[r.nextID] = &snapshots[i]
		r.nextID++
	}
	return nil
}

func (r *InMemoryMarketSnapshotRepository) FindLatestBatchNo() (string, error) {
	if len(r.snapshots) == 0 {
		return "", gorm.ErrRecordNotFound
	}

	var latest *model.MarketSnapshot
	for _, s := range r.snapshots {
		if latest == nil || s.SnapshotTime.After(latest.SnapshotTime) {
			latest = s
		}
	}
	return latest.BatchNo, nil
}

func (r *InMemoryMarketSnapshotRepository) FindByBatchNo(batchNo string) ([]model.MarketSnapshot, error) {
	var result []model.MarketSnapshot
	for _, s := range r.snapshots {
		if s.BatchNo == batchNo {
			result = append(result, *s)
		}
	}
	return result, nil
}

func (r *InMemoryMarketSnapshotRepository) FindLatestBySymbol(symbol string) (*model.MarketSnapshot, error) {
	var latest *model.MarketSnapshot
	for _, s := range r.snapshots {
		if s.Symbol == symbol {
			if latest == nil || s.SnapshotTime.After(latest.SnapshotTime) {
				latest = s
			}
		}
	}
	if latest == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return latest, nil
}

func (r *InMemoryMarketSnapshotRepository) FindHistory(limit int, startTime, endTime *time.Time) ([]model.MarketSnapshot, error) {
	if limit <= 0 {
		limit = 60
	}

	var result []model.MarketSnapshot
	for _, s := range r.snapshots {
		if startTime != nil && s.SnapshotTime.Before(*startTime) {
			continue
		}
		if endTime != nil && s.SnapshotTime.After(*endTime) {
			continue
		}
		result = append(result, *s)
	}

	// 按时间降序排序并限制数量
	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

func (r *InMemoryMarketSnapshotRepository) FindHistoryBySymbol(symbol string, limit int, startTime, endTime *time.Time) ([]model.MarketSnapshot, error) {
	if limit <= 0 {
		limit = 60
	}

	var result []model.MarketSnapshot
	for _, s := range r.snapshots {
		if s.Symbol != symbol {
			continue
		}
		if startTime != nil && s.SnapshotTime.Before(*startTime) {
			continue
		}
		if endTime != nil && s.SnapshotTime.After(*endTime) {
			continue
		}
		result = append(result, *s)
	}

	if len(result) > limit {
		result = result[:limit]
	}
	return result, nil
}

// TestMarketSnapshotRepository_BatchCreate 测试批量创建
func TestMarketSnapshotRepository_BatchCreate(t *testing.T) {
	repo := NewInMemoryMarketSnapshotRepository()

	now := time.Now()
	snapshots := []model.MarketSnapshot{
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
		{
			Symbol:        "399001.SZ",
			Name:          "深证成指",
			LastPrice:     decimal.NewFromFloat(10000.00),
			ChangeAmount:  decimal.NewFromFloat(-50.0),
			ChangePercent: decimal.NewFromFloat(-0.5),
			SnapshotTime:  now,
			BatchNo:       "batch001",
			Source:        "mock",
		},
	}

	err := repo.BatchCreate(snapshots)
	if err != nil {
		t.Errorf("BatchCreate() error = %v", err)
	}

	if len(repo.snapshots) != 2 {
		t.Errorf("BatchCreate() should create 2 snapshots, got %d", len(repo.snapshots))
	}
}

// TestMarketSnapshotRepository_BatchCreate_Empty 测试空批量创建
func TestMarketSnapshotRepository_BatchCreate_Empty(t *testing.T) {
	repo := NewInMemoryMarketSnapshotRepository()

	err := repo.BatchCreate([]model.MarketSnapshot{})
	if err != nil {
		t.Errorf("BatchCreate() with empty slice should not error, got %v", err)
	}
}

// TestMarketSnapshotRepository_FindLatestBatchNo 测试查找最新批次号
func TestMarketSnapshotRepository_FindLatestBatchNo(t *testing.T) {
	repo := NewInMemoryMarketSnapshotRepository()

	now := time.Now()
	repo.BatchCreate([]model.MarketSnapshot{{
		Symbol:        "000001.SH",
		Name:          "上证指数",
		LastPrice:     decimal.Zero,
		ChangeAmount:  decimal.Zero,
		ChangePercent: decimal.Zero,
		SnapshotTime:  now.Add(-time.Hour),
		BatchNo:       "batch001",
		Source:        "mock",
	}})
	repo.BatchCreate([]model.MarketSnapshot{{
		Symbol:        "000001.SH",
		Name:          "上证指数",
		LastPrice:     decimal.Zero,
		ChangeAmount:  decimal.Zero,
		ChangePercent: decimal.Zero,
		SnapshotTime:  now,
		BatchNo:       "batch002",
		Source:        "mock",
	}})

	batchNo, err := repo.FindLatestBatchNo()
	if err != nil {
		t.Errorf("FindLatestBatchNo() error = %v", err)
	}

	if batchNo != "batch002" {
		t.Errorf("FindLatestBatchNo() = %v, want batch002", batchNo)
	}
}

// TestMarketSnapshotRepository_FindLatestBatchNo_Empty 测试空仓储
func TestMarketSnapshotRepository_FindLatestBatchNo_Empty(t *testing.T) {
	repo := NewInMemoryMarketSnapshotRepository()

	_, err := repo.FindLatestBatchNo()
	if err != gorm.ErrRecordNotFound {
		t.Error("FindLatestBatchNo() should return ErrRecordNotFound for empty repository")
	}
}

// TestMarketSnapshotRepository_FindByBatchNo 测试按批次号查找
func TestMarketSnapshotRepository_FindByBatchNo(t *testing.T) {
	repo := NewInMemoryMarketSnapshotRepository()

	now := time.Now()
	repo.BatchCreate([]model.MarketSnapshot{
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
		{
			Symbol:        "000001.SH",
			Name:          "上证指数",
			LastPrice:     decimal.Zero,
			ChangeAmount:  decimal.Zero,
			ChangePercent: decimal.Zero,
			SnapshotTime:  now.Add(time.Hour),
			BatchNo:       "batch002",
			Source:        "mock",
		},
	})

	snapshots, err := repo.FindByBatchNo("batch001")
	if err != nil {
		t.Errorf("FindByBatchNo() error = %v", err)
	}

	if len(snapshots) != 2 {
		t.Errorf("FindByBatchNo() returned %d snapshots, want 2", len(snapshots))
	}
}

// TestMarketSnapshotRepository_FindLatestBySymbol 测试按代码查找最新
func TestMarketSnapshotRepository_FindLatestBySymbol(t *testing.T) {
	repo := NewInMemoryMarketSnapshotRepository()

	now := time.Now()
	repo.BatchCreate([]model.MarketSnapshot{
		{
			Symbol:        "000001.SH",
			Name:          "上证指数",
			LastPrice:     decimal.NewFromInt(3000),
			SnapshotTime:  now.Add(-time.Hour),
			BatchNo:       "batch001",
			Source:        "mock",
		},
		{
			Symbol:        "000001.SH",
			Name:          "上证指数",
			LastPrice:     decimal.NewFromInt(3100),
			SnapshotTime:  now,
			BatchNo:       "batch002",
			Source:        "mock",
		},
	})

	snapshot, err := repo.FindLatestBySymbol("000001.SH")
	if err != nil {
		t.Errorf("FindLatestBySymbol() error = %v", err)
	}

	if snapshot.LastPrice.Cmp(decimal.NewFromInt(3100)) != 0 {
		t.Errorf("FindLatestBySymbol() should return latest snapshot")
	}
}

// TestMarketSnapshotRepository_FindLatestBySymbol_NotFound 测试找不到
func TestMarketSnapshotRepository_FindLatestBySymbol_NotFound(t *testing.T) {
	repo := NewInMemoryMarketSnapshotRepository()

	_, err := repo.FindLatestBySymbol("999999")
	if err != gorm.ErrRecordNotFound {
		t.Error("FindLatestBySymbol() should return ErrRecordNotFound for non-existent symbol")
	}
}

// TestMarketSnapshotRepository_FindHistory 测试查找历史
func TestMarketSnapshotRepository_FindHistory(t *testing.T) {
	repo := NewInMemoryMarketSnapshotRepository()

	now := time.Now()
	for i := 0; i < 10; i++ {
		repo.BatchCreate([]model.MarketSnapshot{{
			Symbol:        "000001.SH",
			Name:          "上证指数",
			LastPrice:     decimal.NewFromInt(int64(3000 + i)),
			SnapshotTime:  now.Add(time.Duration(i) * time.Minute),
			BatchNo:       "batch001",
			Source:        "mock",
		}})
	}

	snapshots, err := repo.FindHistory(5, nil, nil)
	if err != nil {
		t.Errorf("FindHistory() error = %v", err)
	}

	if len(snapshots) > 5 {
		t.Errorf("FindHistory() should limit to 5, got %d", len(snapshots))
	}
}

// TestMarketSnapshotRepository_FindHistoryBySymbol 测试按代码查找历史
func TestMarketSnapshotRepository_FindHistoryBySymbol(t *testing.T) {
	repo := NewInMemoryMarketSnapshotRepository()

	now := time.Now()
	repo.BatchCreate([]model.MarketSnapshot{
		{
			Symbol:       "000001.SH",
			Name:         "上证指数",
			LastPrice:    decimal.Zero,
			SnapshotTime: now.Add(-time.Hour),
			BatchNo:      "batch001",
			Source:       "mock",
		},
		{
			Symbol:       "399001.SZ",
			Name:         "深证成指",
			LastPrice:    decimal.Zero,
			SnapshotTime: now,
			BatchNo:      "batch001",
			Source:       "mock",
		},
	})

	snapshots, err := repo.FindHistoryBySymbol("000001.SH", 10, nil, nil)
	if err != nil {
		t.Errorf("FindHistoryBySymbol() error = %v", err)
	}

	for _, s := range snapshots {
		if s.Symbol != "000001.SH" {
			t.Errorf("FindHistoryBySymbol() should only return snapshots for the specified symbol")
			break
		}
	}
}

// 确保 InMemoryMarketSnapshotRepository 实现了 MarketSnapshotRepository 接口
var _ repository.MarketSnapshotRepository = (*InMemoryMarketSnapshotRepository)(nil)

// TestMarketSnapshotRepository_Interface 测试接口实现
func TestMarketSnapshotRepository_Interface(t *testing.T) {
	var repo repository.MarketSnapshotRepository = NewInMemoryMarketSnapshotRepository()

	now := time.Now()
	_ = repo.BatchCreate([]model.MarketSnapshot{{
		Symbol:        "000001.SH",
		Name:          "上证指数",
		LastPrice:     decimal.Zero,
		ChangeAmount:  decimal.Zero,
		ChangePercent: decimal.Zero,
		SnapshotTime:  now,
		BatchNo:       "batch001",
		Source:        "mock",
	}})
	_, _ = repo.FindLatestBatchNo()
	_, _ = repo.FindByBatchNo("batch001")
	_, _ = repo.FindLatestBySymbol("000001.SH")
	_, _ = repo.FindHistory(10, nil, nil)
	_, _ = repo.FindHistoryBySymbol("000001.SH", 10, nil, nil)
}
