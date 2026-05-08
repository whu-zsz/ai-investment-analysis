//go:build integration
// +build integration

package repository_test

import (
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository_Integration(t *testing.T) {
	db := SetupTestDB(t)
	defer CleanupDB(t, db)

	repo := repository.NewUserRepository(db)

	t.Run("Create", func(t *testing.T) {
		user := &model.User{
			Username:            "testuser1",
			Email:               "test1@example.com",
			PasswordHash:        "$2a$10$abcdefghijklmnopqrstuvwxyz123456",
			Phone:               "13800138000",
			AvatarURL:           "https://example.com/avatar.jpg",
			InvestmentPreference: "balanced",
			RiskTolerance:       "medium",
			TotalProfit:         decimal.NewFromFloat(1000.50),
			IsActive:            true,
			LastLoginAt:         time.Now(),
		}

		err := repo.Create(user)
		require.NoError(t, err)
		assert.NotZero(t, user.ID)
	})

	t.Run("FindByID", func(t *testing.T) {
		// 先创建用户
		user := &model.User{
			Username:     "testuser2",
			Email:        "test2@example.com",
			PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxyz123456",
			IsActive:     true,
		}
		err := repo.Create(user)
		require.NoError(t, err)

		// 查找用户
		found, err := repo.FindByID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, user.Username, found.Username)
		assert.Equal(t, user.Email, found.Email)
	})

	t.Run("FindByID_NotFound", func(t *testing.T) {
		_, err := repo.FindByID(999999)
		assert.Error(t, err)
	})

	t.Run("FindByUsername", func(t *testing.T) {
		user := &model.User{
			Username:     "uniqueuser",
			Email:        "unique@example.com",
			PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxyz123456",
			IsActive:     true,
		}
		err := repo.Create(user)
		require.NoError(t, err)

		found, err := repo.FindByUsername("uniqueuser")
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
	})

	t.Run("FindByUsername_NotFound", func(t *testing.T) {
		_, err := repo.FindByUsername("nonexistent")
		assert.Error(t, err)
	})

	t.Run("FindByEmail", func(t *testing.T) {
		user := &model.User{
			Username:     "emailuser",
			Email:        "emailtest@example.com",
			PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxyz123456",
			IsActive:     true,
		}
		err := repo.Create(user)
		require.NoError(t, err)

		found, err := repo.FindByEmail("emailtest@example.com")
		require.NoError(t, err)
		assert.Equal(t, user.ID, found.ID)
	})

	t.Run("FindByEmail_NotFound", func(t *testing.T) {
		_, err := repo.FindByEmail("nonexistent@example.com")
		assert.Error(t, err)
	})

	t.Run("Update", func(t *testing.T) {
		user := &model.User{
			Username:     "updateuser",
			Email:        "update@example.com",
			PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxyz123456",
			IsActive:     true,
		}
		err := repo.Create(user)
		require.NoError(t, err)

		// 更新用户信息
		user.Phone = "13900139000"
		user.InvestmentPreference = "aggressive"
		err = repo.Update(user)
		require.NoError(t, err)

		// 验证更新
		found, err := repo.FindByID(user.ID)
		require.NoError(t, err)
		assert.Equal(t, "13900139000", found.Phone)
		assert.Equal(t, "aggressive", found.InvestmentPreference)
	})

	t.Run("Delete", func(t *testing.T) {
		user := &model.User{
			Username:     "deleteuser",
			Email:        "delete@example.com",
			PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxyz123456",
			IsActive:     true,
		}
		err := repo.Create(user)
		require.NoError(t, err)

		// 删除用户
		err = repo.Delete(user.ID)
		require.NoError(t, err)

		// 验证删除
		_, err = repo.FindByID(user.ID)
		assert.Error(t, err)
	})

	t.Run("UpdateLastLogin", func(t *testing.T) {
		user := &model.User{
			Username:     "loginuser",
			Email:        "login@example.com",
			PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxyz123456",
			IsActive:     true,
		}
		err := repo.Create(user)
		require.NoError(t, err)

		// 更新最后登录时间
		err = repo.UpdateLastLogin(user.ID)
		require.NoError(t, err)

		// 验证更新
		found, err := repo.FindByID(user.ID)
		require.NoError(t, err)
		assert.False(t, found.LastLoginAt.IsZero())
	})

	t.Run("UpdateTotalProfit", func(t *testing.T) {
		user := &model.User{
			Username:     "profituser",
			Email:        "profit@example.com",
			PasswordHash: "$2a$10$abcdefghijklmnopqrstuvwxyz123456",
			IsActive:     true,
			TotalProfit:  decimal.Zero,
		}
		err := repo.Create(user)
		require.NoError(t, err)

		// 更新总收益
		newProfit := decimal.NewFromFloat(5000.75)
		err = repo.UpdateTotalProfit(user.ID, newProfit)
		require.NoError(t, err)

		// 验证更新
		found, err := repo.FindByID(user.ID)
		require.NoError(t, err)
		assert.True(t, found.TotalProfit.Equal(newProfit))
	})
}
