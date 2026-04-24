package repository_test

import (
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// InMemoryPortfolioRepository 内存持仓仓储用于测试
type InMemoryPortfolioRepository struct {
	portfolios map[uint64]*model.Portfolio
	nextID     uint64
}

func NewInMemoryPortfolioRepository() *InMemoryPortfolioRepository {
	return &InMemoryPortfolioRepository{
		portfolios: make(map[uint64]*model.Portfolio),
		nextID:     1,
	}
}

func (r *InMemoryPortfolioRepository) Create(portfolio *model.Portfolio) error {
	portfolio.ID = r.nextID
	r.portfolios[r.nextID] = portfolio
	r.nextID++
	return nil
}

func (r *InMemoryPortfolioRepository) FindByID(id uint64) (*model.Portfolio, error) {
	portfolio, ok := r.portfolios[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return portfolio, nil
}

func (r *InMemoryPortfolioRepository) FindByUserID(userID uint64) ([]model.Portfolio, error) {
	var result []model.Portfolio
	for _, p := range r.portfolios {
		if p.UserID == userID {
			result = append(result, *p)
		}
	}
	return result, nil
}

func (r *InMemoryPortfolioRepository) FindByUserAndAsset(userID uint64, assetCode string) (*model.Portfolio, error) {
	for _, p := range r.portfolios {
		if p.UserID == userID && p.AssetCode == assetCode {
			return p, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *InMemoryPortfolioRepository) Update(portfolio *model.Portfolio) error {
	r.portfolios[portfolio.ID] = portfolio
	return nil
}

func (r *InMemoryPortfolioRepository) Delete(id uint64) error {
	delete(r.portfolios, id)
	return nil
}

func (r *InMemoryPortfolioRepository) UpdateCurrentPrice(assetCode string, currentPrice decimal.Decimal) error {
	for _, p := range r.portfolios {
		if p.AssetCode == assetCode {
			p.CurrentPrice = &currentPrice
		}
	}
	return nil
}

// TestPortfolioRepository_Create 测试创建持仓
func TestPortfolioRepository_Create(t *testing.T) {
	repo := NewInMemoryPortfolioRepository()

	portfolio := &model.Portfolio{
		UserID:            1,
		AssetCode:         "600519",
		AssetName:         "贵州茅台",
		AssetType:         "stock",
		TotalQuantity:     decimal.NewFromInt(100),
		AvailableQuantity: decimal.NewFromInt(100),
		AverageCost:       decimal.NewFromFloat(1800.00),
		LastUpdated:       time.Now(),
	}

	err := repo.Create(portfolio)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	if portfolio.ID == 0 {
		t.Error("Create() should set portfolio ID")
	}
}

// TestPortfolioRepository_FindByID 测试通过ID查找
func TestPortfolioRepository_FindByID(t *testing.T) {
	repo := NewInMemoryPortfolioRepository()

	portfolio := &model.Portfolio{
		UserID:            1,
		AssetCode:         "600519",
		AssetName:         "贵州茅台",
		TotalQuantity:     decimal.NewFromInt(100),
		AvailableQuantity: decimal.NewFromInt(100),
		AverageCost:       decimal.NewFromFloat(1800.00),
		LastUpdated:       time.Now(),
	}
	repo.Create(portfolio)

	found, err := repo.FindByID(portfolio.ID)
	if err != nil {
		t.Errorf("FindByID() error = %v", err)
	}

	if found.AssetCode != "600519" {
		t.Errorf("FindByID() AssetCode = %v, want 600519", found.AssetCode)
	}
}

// TestPortfolioRepository_FindByID_NotFound 测试查找不存在的持仓
func TestPortfolioRepository_FindByID_NotFound(t *testing.T) {
	repo := NewInMemoryPortfolioRepository()

	_, err := repo.FindByID(999)
	if err != gorm.ErrRecordNotFound {
		t.Error("FindByID() should return ErrRecordNotFound for non-existent ID")
	}
}

// TestPortfolioRepository_FindByUserID 测试通过用户ID查找
func TestPortfolioRepository_FindByUserID(t *testing.T) {
	repo := NewInMemoryPortfolioRepository()

	now := time.Now()
	repo.Create(&model.Portfolio{
		UserID:            1,
		AssetCode:         "600519",
		AssetName:         "贵州茅台",
		TotalQuantity:     decimal.NewFromInt(100),
		AvailableQuantity: decimal.NewFromInt(100),
		AverageCost:       decimal.NewFromFloat(1800.00),
		LastUpdated:       now,
	})
	repo.Create(&model.Portfolio{
		UserID:            1,
		AssetCode:         "000858",
		AssetName:         "五粮液",
		TotalQuantity:     decimal.NewFromInt(200),
		AvailableQuantity: decimal.NewFromInt(200),
		AverageCost:       decimal.NewFromFloat(150.00),
		LastUpdated:       now,
	})
	repo.Create(&model.Portfolio{
		UserID:            2,
		AssetCode:         "000001",
		AssetName:         "平安银行",
		TotalQuantity:     decimal.NewFromInt(500),
		AvailableQuantity: decimal.NewFromInt(500),
		AverageCost:       decimal.NewFromFloat(12.50),
		LastUpdated:       now,
	})

	portfolios, err := repo.FindByUserID(1)
	if err != nil {
		t.Errorf("FindByUserID() error = %v", err)
	}

	if len(portfolios) != 2 {
		t.Errorf("FindByUserID() returned %d portfolios, want 2", len(portfolios))
	}
}

// TestPortfolioRepository_FindByUserAndAsset 测试通过用户和资产代码查找
func TestPortfolioRepository_FindByUserAndAsset(t *testing.T) {
	repo := NewInMemoryPortfolioRepository()

	repo.Create(&model.Portfolio{
		UserID:            1,
		AssetCode:         "600519",
		AssetName:         "贵州茅台",
		TotalQuantity:     decimal.NewFromInt(100),
		AvailableQuantity: decimal.NewFromInt(100),
		AverageCost:       decimal.NewFromFloat(1800.00),
		LastUpdated:       time.Now(),
	})

	portfolio, err := repo.FindByUserAndAsset(1, "600519")
	if err != nil {
		t.Errorf("FindByUserAndAsset() error = %v", err)
	}

	if portfolio.AssetName != "贵州茅台" {
		t.Errorf("FindByUserAndAsset() AssetName = %v, want 贵州茅台", portfolio.AssetName)
	}
}

// TestPortfolioRepository_FindByUserAndAsset_NotFound 测试查找不存在的资产
func TestPortfolioRepository_FindByUserAndAsset_NotFound(t *testing.T) {
	repo := NewInMemoryPortfolioRepository()

	_, err := repo.FindByUserAndAsset(1, "999999")
	if err != gorm.ErrRecordNotFound {
		t.Error("FindByUserAndAsset() should return ErrRecordNotFound for non-existent asset")
	}
}

// TestPortfolioRepository_Update 测试更新持仓
func TestPortfolioRepository_Update(t *testing.T) {
	repo := NewInMemoryPortfolioRepository()

	portfolio := &model.Portfolio{
		UserID:            1,
		AssetCode:         "600519",
		AssetName:         "贵州茅台",
		TotalQuantity:     decimal.NewFromInt(100),
		AvailableQuantity: decimal.NewFromInt(100),
		AverageCost:       decimal.NewFromFloat(1800.00),
		LastUpdated:       time.Now(),
	}
	repo.Create(portfolio)

	portfolio.TotalQuantity = decimal.NewFromInt(200)
	portfolio.AverageCost = decimal.NewFromFloat(1850.00)
	err := repo.Update(portfolio)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	found, _ := repo.FindByID(portfolio.ID)
	if found.TotalQuantity.Cmp(decimal.NewFromInt(200)) != 0 {
		t.Errorf("Update() TotalQuantity not updated correctly")
	}
}

// TestPortfolioRepository_Delete 测试删除持仓
func TestPortfolioRepository_Delete(t *testing.T) {
	repo := NewInMemoryPortfolioRepository()

	portfolio := &model.Portfolio{
		UserID:            1,
		AssetCode:         "600519",
		AssetName:         "贵州茅台",
		TotalQuantity:     decimal.NewFromInt(100),
		AvailableQuantity: decimal.NewFromInt(100),
		AverageCost:       decimal.NewFromFloat(1800.00),
		LastUpdated:       time.Now(),
	}
	repo.Create(portfolio)

	err := repo.Delete(portfolio.ID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	_, err = repo.FindByID(portfolio.ID)
	if err != gorm.ErrRecordNotFound {
		t.Error("Delete() should remove portfolio")
	}
}

// TestPortfolioRepository_UpdateCurrentPrice 测试更新当前价格
func TestPortfolioRepository_UpdateCurrentPrice(t *testing.T) {
	repo := NewInMemoryPortfolioRepository()

	repo.Create(&model.Portfolio{
		UserID:            1,
		AssetCode:         "600519",
		AssetName:         "贵州茅台",
		TotalQuantity:     decimal.NewFromInt(100),
		AvailableQuantity: decimal.NewFromInt(100),
		AverageCost:       decimal.NewFromFloat(1800.00),
		LastUpdated:       time.Now(),
	})

	newPrice := decimal.NewFromFloat(1900.00)
	err := repo.UpdateCurrentPrice("600519", newPrice)
	if err != nil {
		t.Errorf("UpdateCurrentPrice() error = %v", err)
	}

	portfolio, _ := repo.FindByUserAndAsset(1, "600519")
	if portfolio.CurrentPrice == nil || !portfolio.CurrentPrice.Equal(newPrice) {
		t.Errorf("UpdateCurrentPrice() price not updated correctly")
	}
}

// 确保 InMemoryPortfolioRepository 实现了 PortfolioRepository 接口
var _ repository.PortfolioRepository = (*InMemoryPortfolioRepository)(nil)

// TestPortfolioRepository_Interface 测试接口实现
func TestPortfolioRepository_Interface(t *testing.T) {
	var repo repository.PortfolioRepository = NewInMemoryPortfolioRepository()

	portfolio := &model.Portfolio{
		UserID:            1,
		AssetCode:         "600519",
		AssetName:         "贵州茅台",
		TotalQuantity:     decimal.NewFromInt(100),
		AvailableQuantity: decimal.NewFromInt(100),
		AverageCost:       decimal.NewFromFloat(1800.00),
		LastUpdated:       time.Now(),
	}

	_ = repo.Create(portfolio)
	_, _ = repo.FindByID(portfolio.ID)
	_, _ = repo.FindByUserID(1)
	_, _ = repo.FindByUserAndAsset(1, "600519")
	_ = repo.Update(portfolio)
	_ = repo.UpdateCurrentPrice("600519", decimal.NewFromInt(1900))
	_ = repo.Delete(portfolio.ID)
}
