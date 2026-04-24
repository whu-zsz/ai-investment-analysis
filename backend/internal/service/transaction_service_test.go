package service_test

import (
	"errors"
	"stock-analysis-backend/internal/dto/request"
	dtoResponse "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/service"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// MockTransactionRepository 实现 TransactionRepository 接口
type MockTransactionRepository struct {
	transactions    map[uint64]*model.Transaction
	nextID          uint64
	errOnCreate     error
	errOnFindByID   error
	statsResult     *dtoResponse.TransactionStats
}

func NewMockTransactionRepository() *MockTransactionRepository {
	return &MockTransactionRepository{
		transactions: make(map[uint64]*model.Transaction),
		nextID:       1,
	}
}

func (m *MockTransactionRepository) Create(transaction *model.Transaction) error {
	if m.errOnCreate != nil {
		return m.errOnCreate
	}
	transaction.ID = m.nextID
	m.transactions[m.nextID] = transaction
	m.nextID++
	return nil
}

func (m *MockTransactionRepository) BatchCreate(transactions []model.Transaction) error {
	for i := range transactions {
		transactions[i].ID = m.nextID
		m.transactions[m.nextID] = &transactions[i]
		m.nextID++
	}
	return nil
}

func (m *MockTransactionRepository) FindByID(id uint64) (*model.Transaction, error) {
	if m.errOnFindByID != nil {
		return nil, m.errOnFindByID
	}
	transaction, ok := m.transactions[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return transaction, nil
}

func (m *MockTransactionRepository) FindByUserID(userID uint64, limit, offset int) ([]model.Transaction, int64, error) {
	var result []model.Transaction
	for _, t := range m.transactions {
		if t.UserID == userID {
			result = append(result, *t)
		}
	}
	total := int64(len(result))
	// 简单分页
	if offset < len(result) {
		end := offset + limit
		if end > len(result) {
			end = len(result)
		}
		return result[offset:end], total, nil
	}
	return []model.Transaction{}, total, nil
}

func (m *MockTransactionRepository) FindByAssetCode(userID uint64, assetCode string) ([]model.Transaction, error) {
	var result []model.Transaction
	for _, t := range m.transactions {
		if t.UserID == userID && t.AssetCode == assetCode {
			result = append(result, *t)
		}
	}
	return result, nil
}

func (m *MockTransactionRepository) FindByDateRange(userID uint64, startDate, endDate string) ([]model.Transaction, error) {
	var result []model.Transaction
	// 简化实现，不过滤日期
	for _, t := range m.transactions {
		if t.UserID == userID {
			result = append(result, *t)
		}
	}
	return result, nil
}

func (m *MockTransactionRepository) Update(transaction *model.Transaction) error {
	m.transactions[transaction.ID] = transaction
	return nil
}

func (m *MockTransactionRepository) Delete(id uint64) error {
	delete(m.transactions, id)
	return nil
}

func (m *MockTransactionRepository) GetTransactionStats(userID uint64) (*dtoResponse.TransactionStats, error) {
	if m.statsResult != nil {
		return m.statsResult, nil
	}

	var buyCount, sellCount int64
	var totalInvestment decimal.Decimal

	for _, t := range m.transactions {
		if t.UserID == userID {
			switch t.TransactionType {
			case "buy":
				buyCount++
				totalInvestment = totalInvestment.Add(t.TotalAmount)
			case "sell":
				sellCount++
			}
		}
	}

	return &dtoResponse.TransactionStats{
		TotalTransactions: int64(len(m.transactions)),
		BuyCount:         buyCount,
		SellCount:        sellCount,
		TotalInvestment:  totalInvestment.String(),
		TotalProfit:      "0.00",
	}, nil
}

// MockPortfolioService 实现 PortfolioService 接口
type MockPortfolioService struct {
	errOnUpdate bool
}

func (m *MockPortfolioService) UpdatePortfolioFromTransaction(userID uint64, transaction *model.Transaction) error {
	if m.errOnUpdate {
		return errors.New("portfolio update failed")
	}
	return nil
}

func (m *MockPortfolioService) RecalculatePortfolio(userID uint64, assetCode string) error {
	if m.errOnUpdate {
		return errors.New("portfolio recalculate failed")
	}
	return nil
}

func (m *MockPortfolioService) GetPortfolios(userID uint64) ([]model.Portfolio, error) {
	return []model.Portfolio{}, nil
}

// 测试辅助函数
func createTestTransaction(userID uint64, transactionType string, assetCode string, quantity int, price float64) *model.Transaction {
	return &model.Transaction{
		UserID:          userID,
		TransactionDate: time.Now(),
		TransactionType: transactionType,
		AssetType:       "stock",
		AssetCode:       assetCode,
		AssetName:       "测试股票",
		Quantity:        decimal.NewFromInt(int64(quantity)),
		PricePerUnit:    decimal.NewFromFloat(price),
		TotalAmount:     decimal.NewFromFloat(float64(quantity) * price),
		Commission:      decimal.Zero,
	}
}

// TestTransactionService_CreateTransaction_Success 测试创建交易成功
func TestTransactionService_CreateTransaction_Success(t *testing.T) {
	repo := NewMockTransactionRepository()
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	req := &request.CreateTransactionRequest{
		TransactionDate: "2024-01-15",
		TransactionType: "buy",
		AssetType:       "stock",
		AssetCode:       "600519",
		AssetName:       "贵州茅台",
		Quantity:        "100",
		PricePerUnit:    "1850.00",
		Commission:      "18.50",
	}

	err := svc.CreateTransaction(1, req)

	if err != nil {
		t.Fatalf("CreateTransaction() error = %v", err)
	}
}

// TestTransactionService_CreateTransaction_InvalidDate 测试无效日期格式
func TestTransactionService_CreateTransaction_InvalidDate(t *testing.T) {
	repo := NewMockTransactionRepository()
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	req := &request.CreateTransactionRequest{
		TransactionDate: "2024/01/15", // 错误格式
		TransactionType: "buy",
		AssetType:       "stock",
		AssetCode:       "600519",
		AssetName:       "贵州茅台",
		Quantity:        "100",
		PricePerUnit:    "1850.00",
	}

	err := svc.CreateTransaction(1, req)

	if err == nil {
		t.Error("CreateTransaction() should return error for invalid date format")
	}
}

// TestTransactionService_CreateTransaction_InvalidQuantity 测试无效数量
func TestTransactionService_CreateTransaction_InvalidQuantity(t *testing.T) {
	repo := NewMockTransactionRepository()
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	req := &request.CreateTransactionRequest{
		TransactionDate: "2024-01-15",
		TransactionType: "buy",
		AssetType:       "stock",
		AssetCode:       "600519",
		AssetName:       "贵州茅台",
		Quantity:        "invalid",
		PricePerUnit:    "1850.00",
	}

	err := svc.CreateTransaction(1, req)

	if err == nil {
		t.Error("CreateTransaction() should return error for invalid quantity")
	}
}

// TestTransactionService_CreateTransaction_InvalidPrice 测试无效价格
func TestTransactionService_CreateTransaction_InvalidPrice(t *testing.T) {
	repo := NewMockTransactionRepository()
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	req := &request.CreateTransactionRequest{
		TransactionDate: "2024-01-15",
		TransactionType: "buy",
		AssetType:       "stock",
		AssetCode:       "600519",
		AssetName:       "贵州茅台",
		Quantity:        "100",
		PricePerUnit:    "invalid",
	}

	err := svc.CreateTransaction(1, req)

	if err == nil {
		t.Error("CreateTransaction() should return error for invalid price")
	}
}

// TestTransactionService_GetTransactions_Success 测试获取交易列表成功
func TestTransactionService_GetTransactions_Success(t *testing.T) {
	repo := NewMockTransactionRepository()
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	// 创建测试数据
	repo.Create(createTestTransaction(1, "buy", "600519", 100, 1850.00))
	repo.Create(createTestTransaction(1, "sell", "600519", 50, 1900.00))

	result, err := svc.GetTransactions(1, 1, 10)

	if err != nil {
		t.Fatalf("GetTransactions() error = %v", err)
	}

	if result.Total != 2 {
		t.Errorf("Total = %v, want 2", result.Total)
	}

	if len(result.Transactions) != 2 {
		t.Errorf("Transactions count = %v, want 2", len(result.Transactions))
	}
}

// TestTransactionService_GetTransactions_Pagination 测试分页
func TestTransactionService_GetTransactions_Pagination(t *testing.T) {
	repo := NewMockTransactionRepository()
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	// 创建 25 条记录
	for i := 0; i < 25; i++ {
		repo.Create(createTestTransaction(1, "buy", "600519", 100, 1850.00))
	}

	// 第一页
	result, err := svc.GetTransactions(1, 1, 10)
	if err != nil {
		t.Fatalf("GetTransactions() error = %v", err)
	}
	if result.Page != 1 {
		t.Errorf("Page = %v, want 1", result.Page)
	}
	if result.PageSize != 10 {
		t.Errorf("PageSize = %v, want 10", result.PageSize)
	}
	if result.Total != 25 {
		t.Errorf("Total = %v, want 25", result.Total)
	}

	// 第三页
	result, _ = svc.GetTransactions(1, 3, 10)
	if len(result.Transactions) != 5 {
		t.Errorf("Page 3 should have 5 items, got %d", len(result.Transactions))
	}
}

// TestTransactionService_GetTransactions_DefaultPage 测试默认分页
func TestTransactionService_GetTransactions_DefaultPage(t *testing.T) {
	repo := NewMockTransactionRepository()
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	// 传入无效分页参数
	result, err := svc.GetTransactions(1, 0, 0)

	if err != nil {
		t.Fatalf("GetTransactions() error = %v", err)
	}

	// 应该使用默认值
	if result.Page != 1 {
		t.Errorf("Default page should be 1, got %d", result.Page)
	}
}

// TestTransactionService_GetTransactionByID_Success 测试获取交易详情成功
func TestTransactionService_GetTransactionByID_Success(t *testing.T) {
	repo := NewMockTransactionRepository()
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	// 创建交易
	transaction := createTestTransaction(1, "buy", "600519", 100, 1850.00)
	repo.Create(transaction)

	result, err := svc.GetTransactionByID(1, transaction.ID)

	if err != nil {
		t.Fatalf("GetTransactionByID() error = %v", err)
	}

	if result.AssetCode != "600519" {
		t.Errorf("AssetCode = %v, want 600519", result.AssetCode)
	}
}

// TestTransactionService_GetTransactionByID_NotFound 测试交易不存在
func TestTransactionService_GetTransactionByID_NotFound(t *testing.T) {
	repo := NewMockTransactionRepository()
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	_, err := svc.GetTransactionByID(1, 999)

	if err != service.ErrTransactionNotFound {
		t.Errorf("Error = %v, want ErrTransactionNotFound", err)
	}
}

// TestTransactionService_GetTransactionByID_WrongUser 测试其他用户的交易
func TestTransactionService_GetTransactionByID_WrongUser(t *testing.T) {
	repo := NewMockTransactionRepository()
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	// 用户 1 创建交易
	transaction := createTestTransaction(1, "buy", "600519", 100, 1850.00)
	repo.Create(transaction)

	// 用户 2 尝试访问
	_, err := svc.GetTransactionByID(2, transaction.ID)

	if err != service.ErrTransactionNotFound {
		t.Errorf("Error = %v, want ErrTransactionNotFound", err)
	}
}

// TestTransactionService_UpdateTransaction_Success 测试更新交易成功
func TestTransactionService_UpdateTransaction_Success(t *testing.T) {
	repo := NewMockTransactionRepository()
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	// 创建交易
	transaction := createTestTransaction(1, "buy", "600519", 100, 1850.00)
	repo.Create(transaction)

	req := &request.UpdateTransactionRequest{
		TransactionDate: "2024-01-15",
		TransactionType: "buy",
		AssetType:       "stock",
		AssetCode:       "600519",
		AssetName:       "贵州茅台",
		Quantity:        "150", // 修改数量
		PricePerUnit:    "1900.00",
	}

	result, err := svc.UpdateTransaction(1, transaction.ID, req)

	if err != nil {
		t.Fatalf("UpdateTransaction() error = %v", err)
	}

	if result.Quantity.String() != "150" {
		t.Errorf("Quantity = %v, want 150", result.Quantity)
	}
}

// TestTransactionService_UpdateTransaction_NotFound 测试更新不存在的交易
func TestTransactionService_UpdateTransaction_NotFound(t *testing.T) {
	repo := NewMockTransactionRepository()
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	req := &request.UpdateTransactionRequest{
		TransactionDate: "2024-01-15",
		TransactionType: "buy",
		AssetType:       "stock",
		AssetCode:       "600519",
		AssetName:       "贵州茅台",
		Quantity:        "100",
		PricePerUnit:    "1850.00",
	}

	_, err := svc.UpdateTransaction(1, 999, req)

	if err != service.ErrTransactionNotFound {
		t.Errorf("Error = %v, want ErrTransactionNotFound", err)
	}
}

// TestTransactionService_DeleteTransaction_Success 测试删除交易成功
func TestTransactionService_DeleteTransaction_Success(t *testing.T) {
	repo := NewMockTransactionRepository()
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	// 创建交易
	transaction := createTestTransaction(1, "buy", "600519", 100, 1850.00)
	repo.Create(transaction)

	err := svc.DeleteTransaction(1, transaction.ID)

	if err != nil {
		t.Fatalf("DeleteTransaction() error = %v", err)
	}

	// 验证已删除
	_, err = svc.GetTransactionByID(1, transaction.ID)
	if err != service.ErrTransactionNotFound {
		t.Error("Transaction should be deleted")
	}
}

// TestTransactionService_DeleteTransaction_NotFound 测试删除不存在的交易
func TestTransactionService_DeleteTransaction_NotFound(t *testing.T) {
	repo := NewMockTransactionRepository()
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	err := svc.DeleteTransaction(1, 999)

	if err != service.ErrTransactionNotFound {
		t.Errorf("Error = %v, want ErrTransactionNotFound", err)
	}
}

// TestTransactionService_GetTransactionStats_Success 测试获取交易统计成功
func TestTransactionService_GetTransactionStats_Success(t *testing.T) {
	repo := NewMockTransactionRepository()
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	// 创建交易
	repo.Create(createTestTransaction(1, "buy", "600519", 100, 1850.00))
	repo.Create(createTestTransaction(1, "buy", "000858", 200, 180.00))
	repo.Create(createTestTransaction(1, "sell", "600519", 50, 1900.00))

	stats, err := svc.GetTransactionStats(1)

	if err != nil {
		t.Fatalf("GetTransactionStats() error = %v", err)
	}

	if stats.TotalTransactions != 3 {
		t.Errorf("TotalTransactions = %v, want 3", stats.TotalTransactions)
	}

	if stats.BuyCount != 2 {
		t.Errorf("BuyCount = %v, want 2", stats.BuyCount)
	}

	if stats.SellCount != 1 {
		t.Errorf("SellCount = %v, want 1", stats.SellCount)
	}
}

// TestTransactionService_CreateTransaction_RepositoryError 测试仓储层错误
func TestTransactionService_CreateTransaction_RepositoryError(t *testing.T) {
	repo := NewMockTransactionRepository()
	repo.errOnCreate = errors.New("database error")
	portfolio := &MockPortfolioService{}
	svc := service.NewTransactionService(repo, portfolio)

	req := &request.CreateTransactionRequest{
		TransactionDate: "2024-01-15",
		TransactionType: "buy",
		AssetType:       "stock",
		AssetCode:       "600519",
		AssetName:       "贵州茅台",
		Quantity:        "100",
		PricePerUnit:    "1850.00",
	}

	err := svc.CreateTransaction(1, req)

	if err == nil {
		t.Error("CreateTransaction() should return error on repository error")
	}
}
