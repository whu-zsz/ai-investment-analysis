//go:build integration
// +build integration

package integration

import (
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTransactionRepository_Integration(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(t, db)

	repo := repository.NewTransactionRepository(db)

	// 创建测试用户
	user := &model.User{
		Username:     "txuser",
		Email:        "tx@example.com",
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxyz123456",
		IsActive:     true,
	}
	db.Create(user)

	t.Run("Create", func(t *testing.T) {
		tx := &model.Transaction{
			UserID:          user.ID,
			TransactionDate: time.Now(),
			TransactionType: "buy",
			AssetCode:       "600519",
			AssetName:       "贵州茅台",
			AssetType:       "stock",
			Quantity:        decimal.NewFromInt(100),
			PricePerUnit:    decimal.NewFromFloat(1850.00),
			TotalAmount:     decimal.NewFromFloat(185000.00),
		}

		err := repo.Create(tx)
		require.NoError(t, err)
		assert.NotZero(t, tx.ID)
	})

	t.Run("BatchCreate", func(t *testing.T) {
		transactions := []model.Transaction{
			{
				UserID:          user.ID,
				TransactionDate: time.Now().AddDate(0, 0, -2),
				TransactionType: "buy",
				AssetCode:       "000858",
				AssetName:       "五粮液",
				AssetType:       "stock",
				Quantity:        decimal.NewFromInt(200),
				PricePerUnit:    decimal.NewFromFloat(180.00),
				TotalAmount:     decimal.NewFromFloat(36000.00),
			},
			{
				UserID:          user.ID,
				TransactionDate: time.Now().AddDate(0, 0, -1),
				TransactionType: "sell",
				AssetCode:       "600519",
				AssetName:       "贵州茅台",
				AssetType:       "stock",
				Quantity:        decimal.NewFromInt(50),
				PricePerUnit:    decimal.NewFromFloat(1900.00),
				TotalAmount:     decimal.NewFromFloat(95000.00),
			},
		}

		err := repo.BatchCreate(transactions)
		require.NoError(t, err)
		assert.NotZero(t, transactions[0].ID)
		assert.NotZero(t, transactions[1].ID)
	})

	t.Run("FindByID", func(t *testing.T) {
		tx := &model.Transaction{
			UserID:          user.ID,
			TransactionDate: time.Now(),
			TransactionType: "buy",
			AssetCode:       "300750",
			AssetName:       "宁德时代",
			AssetType:       "stock",
			Quantity:        decimal.NewFromInt(50),
			PricePerUnit:    decimal.NewFromFloat(380.00),
			TotalAmount:     decimal.NewFromFloat(19000.00),
		}
		repo.Create(tx)

		found, err := repo.FindByID(tx.ID)
		require.NoError(t, err)
		assert.Equal(t, tx.AssetCode, found.AssetCode)
		assert.Equal(t, tx.AssetName, found.AssetName)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		_, err := repo.FindByID(999999)
		assert.Error(t, err)
	})

	t.Run("FindByUserID", func(t *testing.T) {
		// 创建多个交易记录
		for i := 0; i < 5; i++ {
			tx := &model.Transaction{
				UserID:          user.ID,
				TransactionDate: time.Now().AddDate(0, 0, -i),
				TransactionType: "buy",
				AssetCode:       "600519",
				AssetName:       "贵州茅台",
				AssetType:       "stock",
				Quantity:        decimal.NewFromInt(10),
				PricePerUnit:    decimal.NewFromFloat(1850.00),
				TotalAmount:     decimal.NewFromFloat(18500.00),
			}
			repo.Create(tx)
		}

		// 测试分页查询
		transactions, total, err := repo.FindByUserID(user.ID, 3, 0)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(transactions), 3)
		assert.Greater(t, total, int64(0))

		// 测试第二页
		transactions2, _, err := repo.FindByUserID(user.ID, 3, 3)
		require.NoError(t, err)
		assert.LessOrEqual(t, len(transactions2), 3)
	})

	t.Run("FindByAssetCode", func(t *testing.T) {
		transactions, err := repo.FindByAssetCode(user.ID, "600519")
		require.NoError(t, err)
		assert.Greater(t, len(transactions), 0)

		for _, tx := range transactions {
			assert.Equal(t, "600519", tx.AssetCode)
		}
	})

	t.Run("FindByDateRange", func(t *testing.T) {
		startDate := time.Now().AddDate(0, 0, -30).Format("2006-01-02")
		endDate := time.Now().Format("2006-01-02")

		transactions, err := repo.FindByDateRange(user.ID, startDate, endDate)
		require.NoError(t, err)
		assert.Greater(t, len(transactions), 0)
	})

	t.Run("Update", func(t *testing.T) {
		tx := &model.Transaction{
			UserID:          user.ID,
			TransactionDate: time.Now(),
			TransactionType: "buy",
			AssetCode:       "000001",
			AssetName:       "平安银行",
			AssetType:       "stock",
			Quantity:        decimal.NewFromInt(100),
			PricePerUnit:    decimal.NewFromFloat(10.00),
			TotalAmount:     decimal.NewFromFloat(1000.00),
		}
		repo.Create(tx)

		// 更新交易记录
		tx.Quantity = decimal.NewFromInt(200)
		tx.TotalAmount = decimal.NewFromFloat(2000.00)
		err := repo.Update(tx)
		require.NoError(t, err)

		// 验证更新
		found, err := repo.FindByID(tx.ID)
		require.NoError(t, err)
		assert.True(t, found.Quantity.Equal(decimal.NewFromInt(200)))
	})

	t.Run("Delete", func(t *testing.T) {
		tx := &model.Transaction{
			UserID:          user.ID,
			TransactionDate: time.Now(),
			TransactionType: "buy",
			AssetCode:       "000002",
			AssetName:       "万科A",
			AssetType:       "stock",
			Quantity:        decimal.NewFromInt(100),
			PricePerUnit:    decimal.NewFromFloat(15.00),
			TotalAmount:     decimal.NewFromFloat(1500.00),
		}
		repo.Create(tx)

		// 删除交易记录
		err := repo.Delete(tx.ID)
		require.NoError(t, err)

		// 验证删除
		_, err = repo.FindByID(tx.ID)
		assert.Error(t, err)
	})

	t.Run("GetTransactionStats", func(t *testing.T) {
		stats, err := repo.GetTransactionStats(user.ID)
		require.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Greater(t, stats.TotalTransactions, int64(0))
		assert.Greater(t, stats.BuyCount, int64(0))
	})
}
