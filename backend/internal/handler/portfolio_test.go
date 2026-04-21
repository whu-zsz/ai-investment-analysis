package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"stock-analysis-backend/internal/handler"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/service"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// MockPortfolioService 实现 PortfolioService 接口用于测试
type MockPortfolioService struct {
	GetPortfoliosFunc func(userID uint64) ([]model.Portfolio, error)
}

func (m *MockPortfolioService) GetPortfolios(userID uint64) ([]model.Portfolio, error) {
	if m.GetPortfoliosFunc != nil {
		return m.GetPortfoliosFunc(userID)
	}
	return []model.Portfolio{}, nil
}

func (m *MockPortfolioService) UpdatePortfolioFromTransaction(userID uint64, transaction *model.Transaction) error {
	return nil
}

func (m *MockPortfolioService) RecalculatePortfolio(userID uint64, assetCode string) error {
	return nil
}

// TestGetPortfolios_Success 测试获取持仓列表成功
func TestGetPortfolios_Success(t *testing.T) {
	currentPrice := decimal.NewFromFloat(1900.00)
	mockService := &MockPortfolioService{
		GetPortfoliosFunc: func(userID uint64) ([]model.Portfolio, error) {
			return []model.Portfolio{
				{
					ID:                1,
					UserID:            userID,
					AssetCode:         "600519",
					AssetName:         "贵州茅台",
					AssetType:         "stock",
					TotalQuantity:     decimal.NewFromInt(100),
					AvailableQuantity: decimal.NewFromInt(100),
					AverageCost:       decimal.NewFromFloat(1850.00),
					CurrentPrice:      &currentPrice,
					MarketValue:       decimal.NewFromFloat(190000.00),
					ProfitLoss:        decimal.NewFromFloat(5000.00),
					ProfitLossPercent: decimal.NewFromFloat(2.7),
					LastUpdated:       time.Now(),
				},
			}, nil
		},
	}

	h := handler.NewPortfolioHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/portfolios", h.GetPortfolios)

	req := httptest.NewRequest("GET", "/portfolios", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp["data"].([]interface{})
	if len(data) != 1 {
		t.Errorf("Expected 1 portfolio, got %d", len(data))
	}

	portfolio := data[0].(map[string]interface{})
	if portfolio["asset_code"] != "600519" {
		t.Errorf("AssetCode = %v, want 600519", portfolio["asset_code"])
	}
}

// TestGetPortfolios_Empty 测试空持仓列表
func TestGetPortfolios_Empty(t *testing.T) {
	mockService := &MockPortfolioService{
		GetPortfoliosFunc: func(userID uint64) ([]model.Portfolio, error) {
			return []model.Portfolio{}, nil
		},
	}

	h := handler.NewPortfolioHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/portfolios", h.GetPortfolios)

	req := httptest.NewRequest("GET", "/portfolios", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	// 空数组可能返回为 nil 或空数组
	data, ok := resp["data"].([]interface{})
	if !ok {
		// 如果为 nil，也是正常的
		if resp["data"] != nil {
			t.Errorf("Expected nil or empty array, got %v", resp["data"])
		}
	} else if len(data) != 0 {
		t.Errorf("Expected 0 portfolios, got %d", len(data))
	}
}

// TestGetPortfolios_MultipleAssets 测试多个持仓
func TestGetPortfolios_MultipleAssets(t *testing.T) {
	mockService := &MockPortfolioService{
		GetPortfoliosFunc: func(userID uint64) ([]model.Portfolio, error) {
			return []model.Portfolio{
				{
					ID:                1,
					UserID:            userID,
					AssetCode:         "600519",
					AssetName:         "贵州茅台",
					AssetType:         "stock",
					TotalQuantity:     decimal.NewFromInt(100),
					AvailableQuantity: decimal.NewFromInt(100),
					AverageCost:       decimal.NewFromFloat(1850.00),
					MarketValue:       decimal.NewFromFloat(185000.00),
					ProfitLoss:        decimal.Zero,
					ProfitLossPercent: decimal.Zero,
					LastUpdated:       time.Now(),
				},
				{
					ID:                2,
					UserID:            userID,
					AssetCode:         "000858",
					AssetName:         "五粮液",
					AssetType:         "stock",
					TotalQuantity:     decimal.NewFromInt(200),
					AvailableQuantity: decimal.NewFromInt(200),
					AverageCost:       decimal.NewFromFloat(180.00),
					MarketValue:       decimal.NewFromFloat(36000.00),
					ProfitLoss:        decimal.Zero,
					ProfitLossPercent: decimal.Zero,
					LastUpdated:       time.Now(),
				},
			}, nil
		},
	}

	h := handler.NewPortfolioHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/portfolios", h.GetPortfolios)

	req := httptest.NewRequest("GET", "/portfolios", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp["data"].([]interface{})
	if len(data) != 2 {
		t.Errorf("Expected 2 portfolios, got %d", len(data))
	}
}

// TestGetPortfolios_ServiceError 测试服务层错误
func TestGetPortfolios_ServiceError(t *testing.T) {
	mockService := &MockPortfolioService{
		GetPortfoliosFunc: func(userID uint64) ([]model.Portfolio, error) {
			return nil, service.ErrTransactionNotFound
		},
	}

	h := handler.NewPortfolioHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/portfolios", h.GetPortfolios)

	req := httptest.NewRequest("GET", "/portfolios", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// TestGetPortfolios_WithProfit 测试带盈亏的持仓
func TestGetPortfolios_WithProfit(t *testing.T) {
	currentPrice := decimal.NewFromFloat(2000.00)
	mockService := &MockPortfolioService{
		GetPortfoliosFunc: func(userID uint64) ([]model.Portfolio, error) {
			return []model.Portfolio{
				{
					ID:                1,
					UserID:            userID,
					AssetCode:         "600519",
					AssetName:         "贵州茅台",
					AssetType:         "stock",
					TotalQuantity:     decimal.NewFromInt(100),
					AvailableQuantity: decimal.NewFromInt(100),
					AverageCost:       decimal.NewFromFloat(1850.00),
					CurrentPrice:      &currentPrice,
					MarketValue:       decimal.NewFromFloat(200000.00),
					ProfitLoss:        decimal.NewFromFloat(15000.00),
					ProfitLossPercent: decimal.NewFromFloat(8.1),
					LastUpdated:       time.Now(),
				},
			}, nil
		},
	}

	h := handler.NewPortfolioHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/portfolios", h.GetPortfolios)

	req := httptest.NewRequest("GET", "/portfolios", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp["data"].([]interface{})
	portfolio := data[0].(map[string]interface{})

	// 验证盈亏数据（decimal.String() 可能不带 .00 后缀）
	profitLoss := portfolio["profit_loss"].(string)
	if profitLoss != "15000" && profitLoss != "15000.00" {
		t.Errorf("ProfitLoss = %v, want 15000 or 15000.00", profitLoss)
	}

	currentPriceStr := portfolio["current_price"].(string)
	if currentPriceStr != "2000" && currentPriceStr != "2000.00" {
		t.Errorf("CurrentPrice = %v, want 2000 or 2000.00", currentPriceStr)
	}
}
