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

func TestPortfolioRepository_Integration(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(t, db)

	repo := repository.NewPortfolioRepository(db)

	// 创建测试用户
	user := &model.User{
		Username:     "portuser",
		Email:        "port@example.com",
		PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxyz123456",
		IsActive:     true,
	}
	db.Create(user)

	t.Run("Create", func(t *testing.T) {
		currentPrice := decimal.NewFromFloat(1900.00)
		portfolio := &model.Portfolio{
			UserID:            user.ID,
			AssetCode:         "600519",
			AssetName:         "贵州茅台",
			AssetType:         "stock",
			TotalQuantity:     decimal.NewFromInt(100),
			AvailableQuantity: decimal.NewFromInt(100),
			AverageCost:       decimal.NewFromFloat(1850.00),
			CurrentPrice:      &currentPrice,
			LastUpdated:       time.Now(),
		}

		err := repo.Create(portfolio)
		require.NoError(t, err)
		assert.NotZero(t, portfolio.ID)
	})

	t.Run("FindByID", func(t *testing.T) {
		currentPrice := decimal.NewFromFloat(180.00)
		portfolio := &model.Portfolio{
			UserID:            user.ID,
			AssetCode:         "000858",
			AssetName:         "五粮液",
			AssetType:         "stock",
			TotalQuantity:     decimal.NewFromInt(200),
			AvailableQuantity: decimal.NewFromInt(200),
			AverageCost:       decimal.NewFromFloat(175.00),
			CurrentPrice:      &currentPrice,
			LastUpdated:       time.Now(),
		}
		repo.Create(portfolio)

		found, err := repo.FindByID(portfolio.ID)
		require.NoError(t, err)
		assert.Equal(t, portfolio.AssetCode, found.AssetCode)
		assert.Equal(t, portfolio.AssetName, found.AssetName)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		_, err := repo.FindByID(999999)
		assert.Error(t, err)
	})

	t.Run("FindByUserID", func(t *testing.T) {
		portfolios, err := repo.FindByUserID(user.ID)
		require.NoError(t, err)
		assert.Greater(t, len(portfolios), 0)

		for _, p := range portfolios {
			assert.Equal(t, user.ID, p.UserID)
		}
	})

	t.Run("FindByUserAndAsset", func(t *testing.T) {
		found, err := repo.FindByUserAndAsset(user.ID, "600519")
		require.NoError(t, err)
		assert.Equal(t, "600519", found.AssetCode)
		assert.Equal(t, "贵州茅台", found.AssetName)
	})

	t.Run("FindByUserAndAsset_NotFound", func(t *testing.T) {
		_, err := repo.FindByUserAndAsset(user.ID, "999999")
		assert.Error(t, err)
	})

	t.Run("Update", func(t *testing.T) {
		portfolio, err := repo.FindByUserAndAsset(user.ID, "600519")
		require.NoError(t, err)

		// 更新持仓
		portfolio.TotalQuantity = decimal.NewFromInt(150)
		portfolio.AvailableQuantity = decimal.NewFromInt(150)
		err = repo.Update(portfolio)
		require.NoError(t, err)

		// 验证更新
		found, err := repo.FindByID(portfolio.ID)
		require.NoError(t, err)
		assert.True(t, found.TotalQuantity.Equal(decimal.NewFromInt(150)))
	})

	t.Run("Delete", func(t *testing.T) {
		currentPrice := decimal.NewFromFloat(380.00)
		portfolio := &model.Portfolio{
			UserID:            user.ID,
			AssetCode:         "300750",
			AssetName:         "宁德时代",
			AssetType:         "stock",
			TotalQuantity:     decimal.NewFromInt(50),
			AvailableQuantity: decimal.NewFromInt(50),
			AverageCost:       decimal.NewFromFloat(370.00),
			CurrentPrice:      &currentPrice,
			LastUpdated:       time.Now(),
		}
		repo.Create(portfolio)

		// 删除持仓
		err := repo.Delete(portfolio.ID)
		require.NoError(t, err)

		// 验证删除
		_, err = repo.FindByID(portfolio.ID)
		assert.Error(t, err)
	})

	t.Run("UpdateCurrentPrice", func(t *testing.T) {
		newPrice := decimal.NewFromFloat(2000.00)
		err := repo.UpdateCurrentPrice("600519", newPrice)
		require.NoError(t, err)

		// 验证价格更新
		found, err := repo.FindByUserAndAsset(user.ID, "600519")
		require.NoError(t, err)
		assert.True(t, found.CurrentPrice.Equal(newPrice))
	})
}
