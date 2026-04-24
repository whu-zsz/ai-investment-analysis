package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/handler"
	"stock-analysis-backend/internal/service"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// MockMarketSnapshotService 模拟市场快照服务
type MockMarketSnapshotService struct {
	Snapshots        []response.MarketSnapshotResponse
	DashboardSnapshot *response.DashboardMarketSnapshotResponse
	Err              error
}

func (m *MockMarketSnapshotService) GetLatestSnapshots() ([]response.MarketSnapshotResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Snapshots, nil
}

func (m *MockMarketSnapshotService) GetHistory(symbol string, limit int, startTime, endTime *time.Time) ([]response.MarketSnapshotResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Snapshots, nil
}

func (m *MockMarketSnapshotService) GetDashboardSnapshot() (*response.DashboardMarketSnapshotResponse, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.DashboardSnapshot, nil
}

// TestMarketHandler_GetLatestSnapshots 测试获取最新快照
func TestMarketHandler_GetLatestSnapshots(t *testing.T) {
	mockService := &MockMarketSnapshotService{
		Snapshots: []response.MarketSnapshotResponse{
			{Symbol: "600519.SH", Name: "贵州茅台", LastPrice: "1850.00", ChangePercent: "2.5"},
			{Symbol: "000858.SZ", Name: "五粮液", LastPrice: "180.00", ChangePercent: "-1.2"},
		},
	}

	h := handler.NewMarketHandler(mockService)
	router := gin.New()
	router.GET("/market/snapshots", h.GetLatestSnapshots)

	req := httptest.NewRequest("GET", "/market/snapshots", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp["data"].([]interface{})
	if len(data) != 2 {
		t.Errorf("Expected 2 snapshots, got %d", len(data))
	}
}

// TestMarketHandler_GetLatestSnapshots_Empty 测试空快照列表
func TestMarketHandler_GetLatestSnapshots_Empty(t *testing.T) {
	mockService := &MockMarketSnapshotService{
		Snapshots: []response.MarketSnapshotResponse{},
	}

	h := handler.NewMarketHandler(mockService)
	router := gin.New()
	router.GET("/market/snapshots", h.GetLatestSnapshots)

	req := httptest.NewRequest("GET", "/market/snapshots", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestMarketHandler_GetLatestSnapshots_Error 测试服务错误
func TestMarketHandler_GetLatestSnapshots_Error(t *testing.T) {
	mockService := &MockMarketSnapshotService{
		Err: service.ErrTransactionNotFound,
	}

	h := handler.NewMarketHandler(mockService)
	router := gin.New()
	router.GET("/market/snapshots", h.GetLatestSnapshots)

	req := httptest.NewRequest("GET", "/market/snapshots", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// TestMarketHandler_GetSnapshotHistory 测试获取快照历史
func TestMarketHandler_GetSnapshotHistory(t *testing.T) {
	mockService := &MockMarketSnapshotService{
		Snapshots: []response.MarketSnapshotResponse{
			{Symbol: "600519.SH", Name: "贵州茅台", LastPrice: "1850.00", ChangePercent: "2.5"},
		},
	}

	h := handler.NewMarketHandler(mockService)
	router := gin.New()
	router.GET("/market/history", h.GetSnapshotHistory)

	req := httptest.NewRequest("GET", "/market/history?symbol=600519.SH&limit=30", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// TestMarketHandler_GetSnapshotHistory_WithTimeRange 测试带时间范围的历史
func TestMarketHandler_GetSnapshotHistory_WithTimeRange(t *testing.T) {
	mockService := &MockMarketSnapshotService{
		Snapshots: []response.MarketSnapshotResponse{
			{Symbol: "600519.SH", Name: "贵州茅台", LastPrice: "1850.00", ChangePercent: "2.5"},
		},
	}

	h := handler.NewMarketHandler(mockService)
	router := gin.New()
	router.GET("/market/history", h.GetSnapshotHistory)

	req := httptest.NewRequest("GET", "/market/history?symbol=600519.SH&start_time=2024-01-01&end_time=2024-12-31", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// TestMarketHandler_GetSnapshotHistory_InvalidStartTime 测试无效开始时间
func TestMarketHandler_GetSnapshotHistory_InvalidStartTime(t *testing.T) {
	mockService := &MockMarketSnapshotService{}

	h := handler.NewMarketHandler(mockService)
	router := gin.New()
	router.GET("/market/history", h.GetSnapshotHistory)

	req := httptest.NewRequest("GET", "/market/history?start_time=invalid-time", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestMarketHandler_GetSnapshotHistory_InvalidEndTime 测试无效结束时间
func TestMarketHandler_GetSnapshotHistory_InvalidEndTime(t *testing.T) {
	mockService := &MockMarketSnapshotService{}

	h := handler.NewMarketHandler(mockService)
	router := gin.New()
	router.GET("/market/history", h.GetSnapshotHistory)

	req := httptest.NewRequest("GET", "/market/history?end_time=invalid-time", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestMarketHandler_GetDashboardSnapshot 测试获取仪表盘快照
func TestMarketHandler_GetDashboardSnapshot(t *testing.T) {
	mockService := &MockMarketSnapshotService{
		DashboardSnapshot: &response.DashboardMarketSnapshotResponse{
			SnapshotTime: "2024-01-01 15:00:00",
			IsStale:      false,
			Source:       "test",
		},
	}

	h := handler.NewMarketHandler(mockService)
	router := gin.New()
	router.GET("/market/dashboard", h.GetDashboardSnapshot)

	req := httptest.NewRequest("GET", "/market/dashboard", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp["data"].(map[string]interface{})
	if data["source"].(string) != "test" {
		t.Errorf("Expected source test, got %v", data["source"])
	}
}

// TestMarketHandler_GetDashboardSnapshot_Error 测试仪表盘服务错误
func TestMarketHandler_GetDashboardSnapshot_Error(t *testing.T) {
	mockService := &MockMarketSnapshotService{
		Err: service.ErrTransactionNotFound,
	}

	h := handler.NewMarketHandler(mockService)
	router := gin.New()
	router.GET("/market/dashboard", h.GetDashboardSnapshot)

	req := httptest.NewRequest("GET", "/market/dashboard", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// TestParseOptionalTime 测试时间解析
func TestParseOptionalTime(t *testing.T) {
	// 空字符串
	result, err := parseOptionalTime("")
	if err != nil || result != nil {
		t.Error("Expected nil for empty string")
	}

	// 日期格式
	result, err = parseOptionalTime("2024-01-15")
	if err != nil {
		t.Fatalf("Failed to parse date: %v", err)
	}
	if result.Year() != 2024 {
		t.Errorf("Expected year 2024, got %d", result.Year())
	}

	// 日期时间格式
	result, err = parseOptionalTime("2024-01-15 10:30:00")
	if err != nil {
		t.Fatalf("Failed to parse datetime: %v", err)
	}
	if result.Hour() != 10 {
		t.Errorf("Expected hour 10, got %d", result.Hour())
	}

	// 无效格式
	_, err = parseOptionalTime("invalid")
	if err == nil {
		t.Error("Expected error for invalid format")
	}
}

// TestStrconvAtoi 测试字符串转整数
func TestStrconvAtoi(t *testing.T) {
	// 正常数字
	result, err := strconvAtoi("123")
	if err != nil || result != 123 {
		t.Errorf("Expected 123, got %d", result)
	}

	// 零
	result, err = strconvAtoi("0")
	if err != nil || result != 0 {
		t.Errorf("Expected 0, got %d", result)
	}

	// 无效字符
	_, err = strconvAtoi("12a3")
	if err == nil {
		t.Error("Expected error for invalid string")
	}

	// 负数（不支持）
	_, err = strconvAtoi("-123")
	if err == nil {
		t.Error("Expected error for negative number")
	}
}

// 辅助函数（从 market.go 复制用于测试）
func parseOptionalTime(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return &parsed, nil
		}
	}
	return nil, http.ErrHandlerTimeout
}

func strconvAtoi(value string) (int, error) {
	var result int
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return 0, http.ErrHandlerTimeout
		}
		result = result*10 + int(ch-'0')
	}
	return result, nil
}

// 禁用未使用变量警告
var _ = decimal.NewFromInt(0)
var _ service.MarketSnapshotService = (*MockMarketSnapshotService)(nil)
