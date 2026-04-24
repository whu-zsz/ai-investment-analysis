package service_test

import (
	
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/service"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// MockPortfolioRepository 模拟持仓仓储
type MockPortfolioRepository struct {
	Portfolios map[uint64]*model.Portfolio
	NextID     uint64
}

func NewMockPortfolioRepository() *MockPortfolioRepository {
	return &MockPortfolioRepository{
		Portfolios: make(map[uint64]*model.Portfolio),
		NextID:     1,
	}
}

func (r *MockPortfolioRepository) Create(portfolio *model.Portfolio) error {
	portfolio.ID = r.NextID
	r.Portfolios[r.NextID] = portfolio
	r.NextID++
	return nil
}

func (r *MockPortfolioRepository) FindByID(id uint64) (*model.Portfolio, error) {
	portfolio, ok := r.Portfolios[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return portfolio, nil
}

func (r *MockPortfolioRepository) FindByUserID(userID uint64) ([]model.Portfolio, error) {
	var result []model.Portfolio
	for _, p := range r.Portfolios {
		if p.UserID == userID {
			result = append(result, *p)
		}
	}
	return result, nil
}

func (r *MockPortfolioRepository) FindByUserAndAsset(userID uint64, assetCode string) (*model.Portfolio, error) {
	for _, p := range r.Portfolios {
		if p.UserID == userID && p.AssetCode == assetCode {
			return p, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *MockPortfolioRepository) Update(portfolio *model.Portfolio) error {
	r.Portfolios[portfolio.ID] = portfolio
	return nil
}

func (r *MockPortfolioRepository) Delete(id uint64) error {
	delete(r.Portfolios, id)
	return nil
}

func (r *MockPortfolioRepository) UpdateCurrentPrice(assetCode string, currentPrice decimal.Decimal) error {
	return nil
}

// MockTransactionRepositoryForPortfolio 模拟交易仓储
type MockTransactionRepositoryForPortfolio struct {
	Transactions []model.Transaction
}

func (r *MockTransactionRepositoryForPortfolio) Create(transaction *model.Transaction) error {
	return nil
}

func (r *MockTransactionRepositoryForPortfolio) BatchCreate(transactions []model.Transaction) error {
	return nil
}

func (r *MockTransactionRepositoryForPortfolio) FindByID(id uint64) (*model.Transaction, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *MockTransactionRepositoryForPortfolio) FindByUserID(userID uint64, limit, offset int) ([]model.Transaction, int64, error) {
	return []model.Transaction{}, 0, nil
}

func (r *MockTransactionRepositoryForPortfolio) FindByAssetCode(userID uint64, assetCode string) ([]model.Transaction, error) {
	var result []model.Transaction
	for _, t := range r.Transactions {
		if t.AssetCode == assetCode {
			result = append(result, t)
		}
	}
	return result, nil
}

func (r *MockTransactionRepositoryForPortfolio) FindByDateRange(userID uint64, startDate, endDate string) ([]model.Transaction, error) {
	return []model.Transaction{}, nil
}

func (r *MockTransactionRepositoryForPortfolio) Update(transaction *model.Transaction) error {
	return nil
}

func (r *MockTransactionRepositoryForPortfolio) Delete(id uint64) error {
	return nil
}

func (r *MockTransactionRepositoryForPortfolio) GetTransactionStats(userID uint64) (*response.TransactionStats, error) {
	return &response.TransactionStats{}, nil
}

// TestPortfolioService_GetPortfolios 测试获取持仓列表
func TestPortfolioService_GetPortfolios(t *testing.T) {
	portfolioRepo := NewMockPortfolioRepository()
	portfolioRepo.Create(&model.Portfolio{
		UserID:        1,
		AssetCode:     "600519",
		AssetName:     "贵州茅台",
		TotalQuantity: decimal.NewFromInt(100),
	})

	txRepo := &MockTransactionRepositoryForPortfolio{}
	portfolioService := service.NewPortfolioService(portfolioRepo, txRepo)

	portfolios, err := portfolioService.GetPortfolios(1)
	if err != nil {
		t.Fatalf("GetPortfolios() error = %v", err)
	}

	if len(portfolios) != 1 {
		t.Errorf("Expected 1 portfolio, got %d", len(portfolios))
	}
}

// TestPortfolioService_GetPortfolios_Empty 测试空持仓列表
func TestPortfolioService_GetPortfolios_Empty(t *testing.T) {
	portfolioRepo := NewMockPortfolioRepository()
	txRepo := &MockTransactionRepositoryForPortfolio{}
	portfolioService := service.NewPortfolioService(portfolioRepo, txRepo)

	portfolios, err := portfolioService.GetPortfolios(1)
	if err != nil {
		t.Fatalf("GetPortfolios() error = %v", err)
	}

	if len(portfolios) != 0 {
		t.Errorf("Expected 0 portfolios, got %d", len(portfolios))
	}
}

// TestPortfolioService_UpdatePortfolioFromTransaction_BuyNew 测试买入新股票
func TestPortfolioService_UpdatePortfolioFromTransaction_BuyNew(t *testing.T) {
	portfolioRepo := NewMockPortfolioRepository()
	txRepo := &MockTransactionRepositoryForPortfolio{}
	portfolioService := service.NewPortfolioService(portfolioRepo, txRepo)

	tx := &model.Transaction{
		UserID:          1,
		AssetCode:       "600519",
		AssetName:       "贵州茅台",
		AssetType:       "stock",
		TransactionType: "buy",
		Quantity:        decimal.NewFromInt(100),
		PricePerUnit:    decimal.NewFromFloat(1850.00),
		TransactionDate: time.Now(),
	}

	err := portfolioService.UpdatePortfolioFromTransaction(1, tx)
	if err != nil {
		t.Fatalf("UpdatePortfolioFromTransaction() error = %v", err)
	}

	portfolio, _ := portfolioRepo.FindByUserAndAsset(1, "600519")
	if portfolio.TotalQuantity.Cmp(decimal.NewFromInt(100)) != 0 {
		t.Errorf("Expected quantity 100, got %s", portfolio.TotalQuantity.String())
	}
}

// TestPortfolioService_UpdatePortfolioFromTransaction_BuyExisting 测试追加买入
func TestPortfolioService_UpdatePortfolioFromTransaction_BuyExisting(t *testing.T) {
	portfolioRepo := NewMockPortfolioRepository()
	portfolioRepo.Create(&model.Portfolio{
		UserID:        1,
		AssetCode:     "600519",
		AssetName:     "贵州茅台",
		AssetType:     "stock",
		TotalQuantity: decimal.NewFromInt(100),
		AvailableQuantity: decimal.NewFromInt(100),
		AverageCost:   decimal.NewFromFloat(1800.00),
	})

	txRepo := &MockTransactionRepositoryForPortfolio{}
	portfolioService := service.NewPortfolioService(portfolioRepo, txRepo)

	tx := &model.Transaction{
		UserID:          1,
		AssetCode:       "600519",
		AssetName:       "贵州茅台",
		AssetType:       "stock",
		TransactionType: "buy",
		Quantity:        decimal.NewFromInt(100),
		PricePerUnit:    decimal.NewFromFloat(1900.00),
		TransactionDate: time.Now(),
	}

	err := portfolioService.UpdatePortfolioFromTransaction(1, tx)
	if err != nil {
		t.Fatalf("UpdatePortfolioFromTransaction() error = %v", err)
	}

	portfolio, _ := portfolioRepo.FindByUserAndAsset(1, "600519")

	// 总数量应该是 200
	if portfolio.TotalQuantity.Cmp(decimal.NewFromInt(200)) != 0 {
		t.Errorf("Expected quantity 200, got %s", portfolio.TotalQuantity.String())
	}

	// 平均成本应该是 (1800*100 + 1900*100) / 200 = 1850
	expectedAvgCost := decimal.NewFromFloat(1850.00)
	if !portfolio.AverageCost.Equal(expectedAvgCost) {
		t.Errorf("Expected average cost %s, got %s", expectedAvgCost.String(), portfolio.AverageCost.String())
	}
}

// TestPortfolioService_UpdatePortfolioFromTransaction_Sell 测试卖出
func TestPortfolioService_UpdatePortfolioFromTransaction_Sell(t *testing.T) {
	portfolioRepo := NewMockPortfolioRepository()
	portfolioRepo.Create(&model.Portfolio{
		UserID:            1,
		AssetCode:         "600519",
		AssetName:         "贵州茅台",
		AssetType:         "stock",
		TotalQuantity:     decimal.NewFromInt(100),
		AvailableQuantity: decimal.NewFromInt(100),
		AverageCost:       decimal.NewFromFloat(1850.00),
	})

	txRepo := &MockTransactionRepositoryForPortfolio{}
	portfolioService := service.NewPortfolioService(portfolioRepo, txRepo)

	tx := &model.Transaction{
		UserID:          1,
		AssetCode:       "600519",
		TransactionType: "sell",
		Quantity:        decimal.NewFromInt(50),
		TransactionDate: time.Now(),
	}

	err := portfolioService.UpdatePortfolioFromTransaction(1, tx)
	if err != nil {
		t.Fatalf("UpdatePortfolioFromTransaction() error = %v", err)
	}

	portfolio, _ := portfolioRepo.FindByUserAndAsset(1, "600519")
	if portfolio.TotalQuantity.Cmp(decimal.NewFromInt(50)) != 0 {
		t.Errorf("Expected quantity 50, got %s", portfolio.TotalQuantity.String())
	}
}

// TestPortfolioService_UpdatePortfolioFromTransaction_SellAll 测试清仓卖出
func TestPortfolioService_UpdatePortfolioFromTransaction_SellAll(t *testing.T) {
	portfolioRepo := NewMockPortfolioRepository()
	portfolioRepo.Create(&model.Portfolio{
		UserID:            1,
		AssetCode:         "600519",
		AssetName:         "贵州茅台",
		AssetType:         "stock",
		TotalQuantity:     decimal.NewFromInt(100),
		AvailableQuantity: decimal.NewFromInt(100),
		AverageCost:       decimal.NewFromFloat(1850.00),
	})

	txRepo := &MockTransactionRepositoryForPortfolio{}
	portfolioService := service.NewPortfolioService(portfolioRepo, txRepo)

	tx := &model.Transaction{
		UserID:          1,
		AssetCode:       "600519",
		TransactionType: "sell",
		Quantity:        decimal.NewFromInt(100),
		TransactionDate: time.Now(),
	}

	err := portfolioService.UpdatePortfolioFromTransaction(1, tx)
	if err != nil {
		t.Fatalf("UpdatePortfolioFromTransaction() error = %v", err)
	}

	// 清仓后记录应被删除
	_, err = portfolioRepo.FindByUserAndAsset(1, "600519")
	if err == nil {
		t.Error("Expected portfolio to be deleted after selling all")
	}
}

// TestPortfolioService_UpdatePortfolioFromTransaction_SellInsufficient 测试卖出数量不足
func TestPortfolioService_UpdatePortfolioFromTransaction_SellInsufficient(t *testing.T) {
	portfolioRepo := NewMockPortfolioRepository()
	portfolioRepo.Create(&model.Portfolio{
		UserID:            1,
		AssetCode:         "600519",
		AssetName:         "贵州茅台",
		TotalQuantity:     decimal.NewFromInt(100),
		AvailableQuantity: decimal.NewFromInt(50), // 可用只有50
		AverageCost:       decimal.NewFromFloat(1850.00),
	})

	txRepo := &MockTransactionRepositoryForPortfolio{}
	portfolioService := service.NewPortfolioService(portfolioRepo, txRepo)

	tx := &model.Transaction{
		UserID:          1,
		AssetCode:       "600519",
		TransactionType: "sell",
		Quantity:        decimal.NewFromInt(100), // 尝试卖出100
		TransactionDate: time.Now(),
	}

	err := portfolioService.UpdatePortfolioFromTransaction(1, tx)
	if err == nil {
		t.Error("Expected error for insufficient quantity")
	}
}

// TestPortfolioService_UpdatePortfolioFromTransaction_SellNonExistent 测试卖出不存在的股票
func TestPortfolioService_UpdatePortfolioFromTransaction_SellNonExistent(t *testing.T) {
	portfolioRepo := NewMockPortfolioRepository()
	txRepo := &MockTransactionRepositoryForPortfolio{}
	portfolioService := service.NewPortfolioService(portfolioRepo, txRepo)

	tx := &model.Transaction{
		UserID:          1,
		AssetCode:       "600519",
		TransactionType: "sell",
		Quantity:        decimal.NewFromInt(100),
		TransactionDate: time.Now(),
	}

	err := portfolioService.UpdatePortfolioFromTransaction(1, tx)
	if err == nil {
		t.Error("Expected error for selling non-existent asset")
	}
}

// TestPortfolioService_UpdatePortfolioFromTransaction_Dividend 测试分红
func TestPortfolioService_UpdatePortfolioFromTransaction_Dividend(t *testing.T) {
	portfolioRepo := NewMockPortfolioRepository()
	portfolioRepo.Create(&model.Portfolio{
		UserID:            1,
		AssetCode:         "600519",
		AssetName:         "贵州茅台",
		TotalQuantity:     decimal.NewFromInt(100),
		AvailableQuantity: decimal.NewFromInt(100),
		AverageCost:       decimal.NewFromFloat(1850.00),
	})

	txRepo := &MockTransactionRepositoryForPortfolio{}
	portfolioService := service.NewPortfolioService(portfolioRepo, txRepo)

	tx := &model.Transaction{
		UserID:          1,
		AssetCode:       "600519",
		TransactionType: "dividend",
		TransactionDate: time.Now(),
	}

	err := portfolioService.UpdatePortfolioFromTransaction(1, tx)
	if err != nil {
		t.Fatalf("UpdatePortfolioFromTransaction() error = %v", err)
	}

	// 分红不影响持仓数量
	portfolio, _ := portfolioRepo.FindByUserAndAsset(1, "600519")
	if portfolio.TotalQuantity.Cmp(decimal.NewFromInt(100)) != 0 {
		t.Errorf("Dividend should not change quantity, got %s", portfolio.TotalQuantity.String())
	}
}

// TestPortfolioService_UpdatePortfolioFromTransaction_InvalidType 测试无效交易类型
func TestPortfolioService_UpdatePortfolioFromTransaction_InvalidType(t *testing.T) {
	portfolioRepo := NewMockPortfolioRepository()
	portfolioRepo.Create(&model.Portfolio{
		UserID:            1,
		AssetCode:         "600519",
		TotalQuantity:     decimal.NewFromInt(100),
		AvailableQuantity: decimal.NewFromInt(100),
	})

	txRepo := &MockTransactionRepositoryForPortfolio{}
	portfolioService := service.NewPortfolioService(portfolioRepo, txRepo)

	tx := &model.Transaction{
		UserID:          1,
		AssetCode:       "600519",
		TransactionType: "invalid",
		Quantity:        decimal.NewFromInt(10),
		TransactionDate: time.Now(),
	}

	err := portfolioService.UpdatePortfolioFromTransaction(1, tx)
	if err == nil {
		t.Error("Expected error for invalid transaction type")
	}
}

// TestPortfolioService_RecalculatePortfolio 测试重新计算持仓
func TestPortfolioService_RecalculatePortfolio(t *testing.T) {
	portfolioRepo := NewMockPortfolioRepository()
	txRepo := &MockTransactionRepositoryForPortfolio{
		Transactions: []model.Transaction{
			{
				UserID:          1,
				AssetCode:       "600519",
				AssetName:       "贵州茅台",
				AssetType:       "stock",
				TransactionType: "buy",
				Quantity:        decimal.NewFromInt(100),
				PricePerUnit:    decimal.NewFromFloat(1800.00),
			},
			{
				UserID:          1,
				AssetCode:       "600519",
				AssetName:       "贵州茅台",
				AssetType:       "stock",
				TransactionType: "buy",
				Quantity:        decimal.NewFromInt(100),
				PricePerUnit:    decimal.NewFromFloat(1900.00),
			},
		},
	}
	portfolioService := service.NewPortfolioService(portfolioRepo, txRepo)

	err := portfolioService.RecalculatePortfolio(1, "600519")
	if err != nil {
		t.Fatalf("RecalculatePortfolio() error = %v", err)
	}

	portfolio, _ := portfolioRepo.FindByUserAndAsset(1, "600519")
	if portfolio.TotalQuantity.Cmp(decimal.NewFromInt(200)) != 0 {
		t.Errorf("Expected quantity 200, got %s", portfolio.TotalQuantity.String())
	}
}

// TestPortfolioService_RecalculatePortfolio_WithSell 测试重新计算持仓（含卖出）
func TestPortfolioService_RecalculatePortfolio_WithSell(t *testing.T) {
	portfolioRepo := NewMockPortfolioRepository()
	txRepo := &MockTransactionRepositoryForPortfolio{
		Transactions: []model.Transaction{
			{
				UserID:          1,
				AssetCode:       "600519",
				AssetName:       "贵州茅台",
				AssetType:       "stock",
				TransactionType: "buy",
				Quantity:        decimal.NewFromInt(100),
				PricePerUnit:    decimal.NewFromFloat(1800.00),
			},
			{
				UserID:          1,
				AssetCode:       "600519",
				AssetName:       "贵州茅台",
				AssetType:       "stock",
				TransactionType: "sell",
				Quantity:        decimal.NewFromInt(50),
			},
		},
	}
	portfolioService := service.NewPortfolioService(portfolioRepo, txRepo)

	err := portfolioService.RecalculatePortfolio(1, "600519")
	if err != nil {
		t.Fatalf("RecalculatePortfolio() error = %v", err)
	}

	portfolio, _ := portfolioRepo.FindByUserAndAsset(1, "600519")
	if portfolio.TotalQuantity.Cmp(decimal.NewFromInt(50)) != 0 {
		t.Errorf("Expected quantity 50, got %s", portfolio.TotalQuantity.String())
	}
}

// TestPortfolioService_RecalculatePortfolio_Empty 测试重新计算空持仓
func TestPortfolioService_RecalculatePortfolio_Empty(t *testing.T) {
	portfolioRepo := NewMockPortfolioRepository()
	txRepo := &MockTransactionRepositoryForPortfolio{}
	portfolioService := service.NewPortfolioService(portfolioRepo, txRepo)

	err := portfolioService.RecalculatePortfolio(1, "600519")
	if err != nil {
		t.Fatalf("RecalculatePortfolio() error = %v", err)
	}

	// 无交易记录，不应创建持仓
	portfolios, _ := portfolioRepo.FindByUserID(1)
	if len(portfolios) != 0 {
		t.Errorf("Expected 0 portfolios, got %d", len(portfolios))
	}
}

// TestPortfolioService_RecalculatePortfolio_AllSold 测试重新计算已清仓的持仓
func TestPortfolioService_RecalculatePortfolio_AllSold(t *testing.T) {
	portfolioRepo := NewMockPortfolioRepository()
	portfolioRepo.Create(&model.Portfolio{
		UserID:            1,
		AssetCode:         "600519",
		TotalQuantity:     decimal.NewFromInt(100),
		AvailableQuantity: decimal.NewFromInt(100),
	})

	txRepo := &MockTransactionRepositoryForPortfolio{
		Transactions: []model.Transaction{
			{
				UserID:          1,
				AssetCode:       "600519",
				TransactionType: "buy",
				Quantity:        decimal.NewFromInt(100),
			},
			{
				UserID:          1,
				AssetCode:       "600519",
				TransactionType: "sell",
				Quantity:        decimal.NewFromInt(100),
			},
		},
	}
	portfolioService := service.NewPortfolioService(portfolioRepo, txRepo)

	err := portfolioService.RecalculatePortfolio(1, "600519")
	if err != nil {
		t.Fatalf("RecalculatePortfolio() error = %v", err)
	}

	// 全部卖出后，持仓应被删除
	_, err = portfolioRepo.FindByUserAndAsset(1, "600519")
	if err == nil {
		t.Error("Expected portfolio to be deleted after all sold")
	}
}
