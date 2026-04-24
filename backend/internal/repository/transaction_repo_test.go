package repository_test

import (
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// InMemoryTransactionRepository 内存交易仓储用于测试
type InMemoryTransactionRepository struct {
	transactions map[uint64]*model.Transaction
	nextID       uint64
}

func NewInMemoryTransactionRepository() *InMemoryTransactionRepository {
	return &InMemoryTransactionRepository{
		transactions: make(map[uint64]*model.Transaction),
		nextID:       1,
	}
}

func (r *InMemoryTransactionRepository) Create(transaction *model.Transaction) error {
	transaction.ID = r.nextID
	r.transactions[r.nextID] = transaction
	r.nextID++
	return nil
}

func (r *InMemoryTransactionRepository) BatchCreate(transactions []model.Transaction) error {
	for i := range transactions {
		transactions[i].ID = r.nextID
		r.transactions[r.nextID] = &transactions[i]
		r.nextID++
	}
	return nil
}

func (r *InMemoryTransactionRepository) FindByID(id uint64) (*model.Transaction, error) {
	transaction, ok := r.transactions[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return transaction, nil
}

func (r *InMemoryTransactionRepository) FindByUserID(userID uint64, limit, offset int) ([]model.Transaction, int64, error) {
	var result []model.Transaction
	for _, t := range r.transactions {
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

func (r *InMemoryTransactionRepository) FindByAssetCode(userID uint64, assetCode string) ([]model.Transaction, error) {
	var result []model.Transaction
	for _, t := range r.transactions {
		if t.UserID == userID && t.AssetCode == assetCode {
			result = append(result, *t)
		}
	}
	return result, nil
}

func (r *InMemoryTransactionRepository) FindByDateRange(userID uint64, startDate, endDate string) ([]model.Transaction, error) {
	var result []model.Transaction
	for _, t := range r.transactions {
		if t.UserID == userID {
			result = append(result, *t)
		}
	}
	return result, nil
}

func (r *InMemoryTransactionRepository) Update(transaction *model.Transaction) error {
	r.transactions[transaction.ID] = transaction
	return nil
}

func (r *InMemoryTransactionRepository) Delete(id uint64) error {
	_, ok := r.transactions[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	delete(r.transactions, id)
	return nil
}

func (r *InMemoryTransactionRepository) GetTransactionStats(userID uint64) (*response.TransactionStats, error) {
	var buyCount, sellCount int64
	var totalInvestment decimal.Decimal

	for _, t := range r.transactions {
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

	return &response.TransactionStats{
		TotalTransactions: int64(len(r.transactions)),
		BuyCount:         buyCount,
		SellCount:        sellCount,
		TotalInvestment:  totalInvestment.String(),
		TotalProfit:      "0.00",
	}, nil
}

// 确保 InMemoryTransactionRepository 实现了 TransactionRepository 接口
var _ repository.TransactionRepository = (*InMemoryTransactionRepository)(nil)

// 辅助函数：创建测试交易
func createTestTx(userID uint64, txType string, assetCode string, qty int, price float64) *model.Transaction {
	return &model.Transaction{
		UserID:          userID,
		TransactionDate: time.Now(),
		TransactionType: txType,
		AssetType:       "stock",
		AssetCode:       assetCode,
		AssetName:       "测试股票",
		Quantity:        decimal.NewFromInt(int64(qty)),
		PricePerUnit:    decimal.NewFromFloat(price),
		TotalAmount:     decimal.NewFromFloat(float64(qty) * price),
		Commission:      decimal.Zero,
	}
}

// TestTransactionRepository_Create 测试创建交易
func TestTransactionRepository_Create(t *testing.T) {
	repo := NewInMemoryTransactionRepository()

	tx := createTestTx(1, "buy", "600519", 100, 1850.00)

	err := repo.Create(tx)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if tx.ID == 0 {
		t.Error("Create() should assign ID")
	}
}

// TestTransactionRepository_BatchCreate 测试批量创建
func TestTransactionRepository_BatchCreate(t *testing.T) {
	repo := NewInMemoryTransactionRepository()

	transactions := []model.Transaction{
		*createTestTx(1, "buy", "600519", 100, 1850.00),
		*createTestTx(1, "buy", "000858", 200, 180.00),
		*createTestTx(1, "sell", "600519", 50, 1900.00),
	}

	err := repo.BatchCreate(transactions)
	if err != nil {
		t.Fatalf("BatchCreate() error = %v", err)
	}

	// 验证所有交易都有 ID
	for _, tx := range transactions {
		if tx.ID == 0 {
			t.Error("BatchCreate() should assign ID to all transactions")
		}
	}
}

// TestTransactionRepository_FindByID 测试通过 ID 查找
func TestTransactionRepository_FindByID(t *testing.T) {
	repo := NewInMemoryTransactionRepository()

	tx := createTestTx(1, "buy", "600519", 100, 1850.00)
	repo.Create(tx)

	// 查找存在的交易
	found, err := repo.FindByID(tx.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if found.AssetCode != "600519" {
		t.Errorf("AssetCode = %v, want 600519", found.AssetCode)
	}

	// 查找不存在的交易
	_, err = repo.FindByID(999)
	if err != gorm.ErrRecordNotFound {
		t.Error("FindByID() should return ErrRecordNotFound for non-existent transaction")
	}
}

// TestTransactionRepository_FindByUserID 测试通过用户 ID 查找
func TestTransactionRepository_FindByUserID(t *testing.T) {
	repo := NewInMemoryTransactionRepository()

	// 用户 1 的交易
	repo.Create(createTestTx(1, "buy", "600519", 100, 1850.00))
	repo.Create(createTestTx(1, "sell", "600519", 50, 1900.00))
	// 用户 2 的交易
	repo.Create(createTestTx(2, "buy", "000858", 200, 180.00))

	// 查找用户 1 的交易
	transactions, total, err := repo.FindByUserID(1, 10, 0)
	if err != nil {
		t.Fatalf("FindByUserID() error = %v", err)
	}
	if total != 2 {
		t.Errorf("Total = %v, want 2", total)
	}
	if len(transactions) != 2 {
		t.Errorf("Transactions count = %v, want 2", len(transactions))
	}
}

// TestTransactionRepository_FindByUserID_Pagination 测试分页
func TestTransactionRepository_FindByUserID_Pagination(t *testing.T) {
	repo := NewInMemoryTransactionRepository()

	// 创建 25 条交易
	for i := 0; i < 25; i++ {
		repo.Create(createTestTx(1, "buy", "600519", 100, 1850.00))
	}

	// 第一页
	txs, total, _ := repo.FindByUserID(1, 10, 0)
	if total != 25 {
		t.Errorf("Total = %v, want 25", total)
	}
	if len(txs) != 10 {
		t.Errorf("Page 1 should have 10 items, got %d", len(txs))
	}

	// 第二页
	txs, _, _ = repo.FindByUserID(1, 10, 10)
	if len(txs) != 10 {
		t.Errorf("Page 2 should have 10 items, got %d", len(txs))
	}

	// 第三页
	txs, _, _ = repo.FindByUserID(1, 10, 20)
	if len(txs) != 5 {
		t.Errorf("Page 3 should have 5 items, got %d", len(txs))
	}
}

// TestTransactionRepository_FindByAssetCode 测试通过资产代码查找
func TestTransactionRepository_FindByAssetCode(t *testing.T) {
	repo := NewInMemoryTransactionRepository()

	repo.Create(createTestTx(1, "buy", "600519", 100, 1850.00))
	repo.Create(createTestTx(1, "buy", "000858", 200, 180.00))
	repo.Create(createTestTx(1, "sell", "600519", 50, 1900.00))

	// 查找 600519 的交易
	transactions, err := repo.FindByAssetCode(1, "600519")
	if err != nil {
		t.Fatalf("FindByAssetCode() error = %v", err)
	}
	if len(transactions) != 2 {
		t.Errorf("600519 should have 2 transactions, got %d", len(transactions))
	}

	// 查找 000858 的交易
	transactions, _ = repo.FindByAssetCode(1, "000858")
	if len(transactions) != 1 {
		t.Errorf("000858 should have 1 transaction, got %d", len(transactions))
	}
}

// TestTransactionRepository_Update 测试更新交易
func TestTransactionRepository_Update(t *testing.T) {
	repo := NewInMemoryTransactionRepository()

	tx := createTestTx(1, "buy", "600519", 100, 1850.00)
	repo.Create(tx)

	// 更新数量
	tx.Quantity = decimal.NewFromInt(150)
	err := repo.Update(tx)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// 验证更新
	found, _ := repo.FindByID(tx.ID)
	if !found.Quantity.Equals(decimal.NewFromInt(150)) {
		t.Errorf("Quantity = %v, want 150", found.Quantity)
	}
}

// TestTransactionRepository_Delete 测试删除交易
func TestTransactionRepository_Delete(t *testing.T) {
	repo := NewInMemoryTransactionRepository()

	tx := createTestTx(1, "buy", "600519", 100, 1850.00)
	repo.Create(tx)

	// 删除交易
	err := repo.Delete(tx.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// 验证删除
	_, err = repo.FindByID(tx.ID)
	if err != gorm.ErrRecordNotFound {
		t.Error("Delete() should remove transaction")
	}

	// 删除不存在的交易
	err = repo.Delete(999)
	if err != gorm.ErrRecordNotFound {
		t.Error("Delete() should return error for non-existent transaction")
	}
}

// TestTransactionRepository_GetTransactionStats 测试获取交易统计
func TestTransactionRepository_GetTransactionStats(t *testing.T) {
	repo := NewInMemoryTransactionRepository()

	repo.Create(createTestTx(1, "buy", "600519", 100, 1850.00))
	repo.Create(createTestTx(1, "buy", "000858", 200, 180.00))
	repo.Create(createTestTx(1, "sell", "600519", 50, 1900.00))
	repo.Create(createTestTx(1, "dividend", "600519", 0, 0))

	stats, err := repo.GetTransactionStats(1)
	if err != nil {
		t.Fatalf("GetTransactionStats() error = %v", err)
	}

	if stats.TotalTransactions != 4 {
		t.Errorf("TotalTransactions = %v, want 4", stats.TotalTransactions)
	}
	if stats.BuyCount != 2 {
		t.Errorf("BuyCount = %v, want 2", stats.BuyCount)
	}
	if stats.SellCount != 1 {
		t.Errorf("SellCount = %v, want 1", stats.SellCount)
	}
}

// TestTransactionRepository_Interface 测试接口实现
func TestTransactionRepository_Interface(t *testing.T) {
	var repo repository.TransactionRepository = NewInMemoryTransactionRepository()

	tx := createTestTx(1, "buy", "600519", 100, 1850.00)

	// 测试所有接口方法
	_ = repo.Create(tx)
	_ = repo.BatchCreate([]model.Transaction{*createTestTx(1, "buy", "000858", 100, 100.00)})
	_, _ = repo.FindByID(tx.ID)
	_, _, _ = repo.FindByUserID(1, 10, 0)
	_, _ = repo.FindByAssetCode(1, "600519")
	_, _ = repo.FindByDateRange(1, "2024-01-01", "2024-12-31")
	_ = repo.Update(tx)
	_, _ = repo.GetTransactionStats(1)
	_ = repo.Delete(tx.ID)
}
