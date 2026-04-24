package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"stock-analysis-backend/internal/dto/request"
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/handler"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/service"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// MockTransactionService 实现 TransactionService 接口用于测试
type MockTransactionService struct {
	CreateTransactionFunc  func(userID uint64, req *request.CreateTransactionRequest) error
	GetTransactionsFunc    func(userID uint64, page, pageSize int) (*response.TransactionListResponse, error)
	GetTransactionByIDFunc func(userID uint64, id uint64) (*model.Transaction, error)
	UpdateTransactionFunc  func(userID uint64, id uint64, req *request.UpdateTransactionRequest) (*model.Transaction, error)
	DeleteTransactionFunc  func(userID uint64, id uint64) error
	GetTransactionStatsFunc func(userID uint64) (*response.TransactionStats, error)
}

func (m *MockTransactionService) CreateTransaction(userID uint64, req *request.CreateTransactionRequest) error {
	if m.CreateTransactionFunc != nil {
		return m.CreateTransactionFunc(userID, req)
	}
	return nil
}

func (m *MockTransactionService) GetTransactions(userID uint64, page, pageSize int) (*response.TransactionListResponse, error) {
	if m.GetTransactionsFunc != nil {
		return m.GetTransactionsFunc(userID, page, pageSize)
	}
	return &response.TransactionListResponse{}, nil
}

func (m *MockTransactionService) GetTransactionByID(userID uint64, id uint64) (*model.Transaction, error) {
	if m.GetTransactionByIDFunc != nil {
		return m.GetTransactionByIDFunc(userID, id)
	}
	return nil, service.ErrTransactionNotFound
}

func (m *MockTransactionService) UpdateTransaction(userID uint64, id uint64, req *request.UpdateTransactionRequest) (*model.Transaction, error) {
	if m.UpdateTransactionFunc != nil {
		return m.UpdateTransactionFunc(userID, id, req)
	}
	return nil, service.ErrTransactionNotFound
}

func (m *MockTransactionService) DeleteTransaction(userID uint64, id uint64) error {
	if m.DeleteTransactionFunc != nil {
		return m.DeleteTransactionFunc(userID, id)
	}
	return nil
}

func (m *MockTransactionService) GetTransactionStats(userID uint64) (*response.TransactionStats, error) {
	if m.GetTransactionStatsFunc != nil {
		return m.GetTransactionStatsFunc(userID)
	}
	return &response.TransactionStats{}, nil
}

func init() {
	gin.SetMode(gin.TestMode)
}

// TestCreateTransaction_Success 测试创建交易成功
func TestCreateTransaction_Success(t *testing.T) {
	mockService := &MockTransactionService{
		CreateTransactionFunc: func(userID uint64, req *request.CreateTransactionRequest) error {
			return nil
		},
	}

	h := handler.NewTransactionHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/transactions", h.CreateTransaction)

	body := `{
		"transaction_date": "2024-01-15",
		"transaction_type": "buy",
		"asset_type": "stock",
		"asset_code": "600519",
		"asset_name": "贵州茅台",
		"quantity": "100",
		"price_per_unit": "1850.00",
		"commission": "18.50"
	}`

	req := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// TestCreateTransaction_InvalidRequest 测试无效创建请求
func TestCreateTransaction_InvalidRequest(t *testing.T) {
	testCases := []struct {
		name string
		body string
	}{
		{"Empty body", ""},
		{"Missing asset_code", `{"transaction_date":"2024-01-15","transaction_type":"buy","asset_type":"stock","asset_name":"测试","quantity":"100","price_per_unit":"10.00"}`},
		{"Invalid transaction_type", `{"transaction_date":"2024-01-15","transaction_type":"invalid","asset_type":"stock","asset_code":"600519","asset_name":"测试","quantity":"100","price_per_unit":"10.00"}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService := &MockTransactionService{}
			h := handler.NewTransactionHandler(mockService)
			router := gin.New()
			router.Use(func(c *gin.Context) {
				c.Set("user_id", uint64(1))
				c.Next()
			})
			router.POST("/transactions", h.CreateTransaction)

			req := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code == http.StatusOK {
				t.Errorf("Expected error status, got %d", w.Code)
			}
		})
	}
}

// TestCreateTransaction_ServiceError 测试服务层错误
func TestCreateTransaction_ServiceError(t *testing.T) {
	mockService := &MockTransactionService{
		CreateTransactionFunc: func(userID uint64, req *request.CreateTransactionRequest) error {
			return errors.New("insufficient available quantity")
		},
	}

	h := handler.NewTransactionHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/transactions", h.CreateTransaction)

	body := `{
		"transaction_date": "2024-01-15",
		"transaction_type": "sell",
		"asset_type": "stock",
		"asset_code": "600519",
		"asset_name": "贵州茅台",
		"quantity": "100",
		"price_per_unit": "1850.00"
	}`

	req := httptest.NewRequest("POST", "/transactions", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestGetTransactions_Success 测试获取交易列表成功
func TestGetTransactions_Success(t *testing.T) {
	mockService := &MockTransactionService{
		GetTransactionsFunc: func(userID uint64, page, pageSize int) (*response.TransactionListResponse, error) {
			return &response.TransactionListResponse{
				Transactions: []response.TransactionResponse{
					{ID: 1, AssetCode: "600519", AssetName: "贵州茅台"},
				},
				Total:    1,
				Page:     1,
				PageSize: 20,
			}, nil
		},
	}

	h := handler.NewTransactionHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/transactions", h.GetTransactions)

	req := httptest.NewRequest("GET", "/transactions?page=1&page_size=20", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 1 {
		t.Errorf("Expected total 1, got %v", data["total"])
	}
}

// TestGetTransactions_DefaultPagination 测试默认分页
func TestGetTransactions_DefaultPagination(t *testing.T) {
	var capturedPage, capturedPageSize int

	mockService := &MockTransactionService{
		GetTransactionsFunc: func(userID uint64, page, pageSize int) (*response.TransactionListResponse, error) {
			capturedPage = page
			capturedPageSize = pageSize
			return &response.TransactionListResponse{}, nil
		},
	}

	h := handler.NewTransactionHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/transactions", h.GetTransactions)

	req := httptest.NewRequest("GET", "/transactions", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if capturedPage != 1 {
		t.Errorf("Expected page 1, got %d", capturedPage)
	}
	if capturedPageSize != 20 {
		t.Errorf("Expected pageSize 20, got %d", capturedPageSize)
	}
}

// TestGetTransaction_Success 测试获取交易详情成功
func TestGetTransaction_Success(t *testing.T) {
	quantity := decimal.NewFromInt(100)
	price := decimal.NewFromFloat(1850.00)
	total := decimal.NewFromFloat(185000.00)

	mockService := &MockTransactionService{
		GetTransactionByIDFunc: func(userID uint64, id uint64) (*model.Transaction, error) {
			return &model.Transaction{
				ID:              id,
				UserID:          userID,
				TransactionDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				TransactionType: "buy",
				AssetCode:       "600519",
				AssetName:       "贵州茅台",
				Quantity:        quantity,
				PricePerUnit:    price,
				TotalAmount:     total,
			}, nil
		},
	}

	h := handler.NewTransactionHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/transactions/:id", h.GetTransaction)

	req := httptest.NewRequest("GET", "/transactions/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestGetTransaction_NotFound 测试交易不存在
func TestGetTransaction_NotFound(t *testing.T) {
	mockService := &MockTransactionService{
		GetTransactionByIDFunc: func(userID uint64, id uint64) (*model.Transaction, error) {
			return nil, service.ErrTransactionNotFound
		},
	}

	h := handler.NewTransactionHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/transactions/:id", h.GetTransaction)

	req := httptest.NewRequest("GET", "/transactions/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// TestGetTransaction_InvalidID 测试无效 ID
func TestGetTransaction_InvalidID(t *testing.T) {
	mockService := &MockTransactionService{}
	h := handler.NewTransactionHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/transactions/:id", h.GetTransaction)

	req := httptest.NewRequest("GET", "/transactions/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestUpdateTransaction_Success 测试更新交易成功
func TestUpdateTransaction_Success(t *testing.T) {
	quantity := decimal.NewFromInt(100)
	price := decimal.NewFromFloat(1900.00)
	total := decimal.NewFromFloat(190000.00)

	mockService := &MockTransactionService{
		UpdateTransactionFunc: func(userID uint64, id uint64, req *request.UpdateTransactionRequest) (*model.Transaction, error) {
			return &model.Transaction{
				ID:              id,
				UserID:          userID,
				TransactionDate: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				TransactionType: "sell",
				AssetCode:       "600519",
				AssetName:       "贵州茅台",
				Quantity:        quantity,
				PricePerUnit:    price,
				TotalAmount:     total,
			}, nil
		},
	}

	h := handler.NewTransactionHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.PUT("/transactions/:id", h.UpdateTransaction)

	body := `{
		"transaction_date": "2024-02-20",
		"transaction_type": "sell",
		"asset_type": "stock",
		"asset_code": "600519",
		"asset_name": "贵州茅台",
		"quantity": "50",
		"price_per_unit": "1900.00"
	}`

	req := httptest.NewRequest("PUT", "/transactions/1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// TestUpdateTransaction_NotFound 测试更新不存在的交易
func TestUpdateTransaction_NotFound(t *testing.T) {
	mockService := &MockTransactionService{
		UpdateTransactionFunc: func(userID uint64, id uint64, req *request.UpdateTransactionRequest) (*model.Transaction, error) {
			return nil, service.ErrTransactionNotFound
		},
	}

	h := handler.NewTransactionHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.PUT("/transactions/:id", h.UpdateTransaction)

	body := `{
		"transaction_date": "2024-02-20",
		"transaction_type": "sell",
		"asset_type": "stock",
		"asset_code": "600519",
		"asset_name": "贵州茅台",
		"quantity": "50",
		"price_per_unit": "1900.00"
	}`

	req := httptest.NewRequest("PUT", "/transactions/999", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// TestDeleteTransaction_Success 测试删除交易成功
func TestDeleteTransaction_Success(t *testing.T) {
	mockService := &MockTransactionService{
		DeleteTransactionFunc: func(userID uint64, id uint64) error {
			return nil
		},
	}

	h := handler.NewTransactionHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.DELETE("/transactions/:id", h.DeleteTransaction)

	req := httptest.NewRequest("DELETE", "/transactions/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestDeleteTransaction_NotFound 测试删除不存在的交易
func TestDeleteTransaction_NotFound(t *testing.T) {
	mockService := &MockTransactionService{
		DeleteTransactionFunc: func(userID uint64, id uint64) error {
			return service.ErrTransactionNotFound
		},
	}

	h := handler.NewTransactionHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.DELETE("/transactions/:id", h.DeleteTransaction)

	req := httptest.NewRequest("DELETE", "/transactions/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// TestGetTransactionStats_Success 测试获取交易统计成功
func TestGetTransactionStats_Success(t *testing.T) {
	mockService := &MockTransactionService{
		GetTransactionStatsFunc: func(userID uint64) (*response.TransactionStats, error) {
			return &response.TransactionStats{
				TotalTransactions: 10,
				BuyCount:         6,
				SellCount:        4,
				TotalInvestment:  "100000.00",
				TotalProfit:      "5000.00",
			}, nil
		},
	}

	h := handler.NewTransactionHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/transactions/stats", h.GetTransactionStats)

	req := httptest.NewRequest("GET", "/transactions/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp["data"].(map[string]interface{})
	if data["total_transactions"].(float64) != 10 {
		t.Errorf("Expected total_transactions 10, got %v", data["total_transactions"])
	}
}

// TestGetTransactionStats_Error 测试获取交易统计错误
func TestGetTransactionStats_Error(t *testing.T) {
	mockService := &MockTransactionService{
		GetTransactionStatsFunc: func(userID uint64) (*response.TransactionStats, error) {
			return nil, errors.New("database error")
		},
	}

	h := handler.NewTransactionHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/transactions/stats", h.GetTransactionStats)

	req := httptest.NewRequest("GET", "/transactions/stats", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}
