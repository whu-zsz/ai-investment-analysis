package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"stock-analysis-backend/internal/dto/request"
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/handler"
	"stock-analysis-backend/internal/service"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// MockAIService 模拟 AI 服务
type MockAIService struct {
	CreateTaskResult          *response.AnalysisTaskResponse
	CreateTaskErr             error
	GetTaskResult             *response.AnalysisTaskDetailResponse
	GetTaskErr                error
	GetTasksResult            *response.AnalysisTaskListResponse
	GetTasksErr               error
	GetReportDetailResult     *response.AnalysisReportDetailResponse
	GetReportDetailErr        error
	GenerateSummaryResult     *response.AnalysisReportResponse
	GenerateSummaryErr        error
	GetReportsResult          []response.AnalysisReportResponse
	GetReportsErr             error
}

func (m *MockAIService) CreateStockAnalysisTask(userID uint64, req *request.CreateAnalysisTaskRequest) (*response.AnalysisTaskResponse, error) {
	if m.CreateTaskErr != nil {
		return nil, m.CreateTaskErr
	}
	return m.CreateTaskResult, nil
}

func (m *MockAIService) GetAnalysisTask(userID, taskID uint64) (*response.AnalysisTaskDetailResponse, error) {
	if m.GetTaskErr != nil {
		return nil, m.GetTaskErr
	}
	return m.GetTaskResult, nil
}

func (m *MockAIService) GetAnalysisTasks(userID uint64, status string, page, pageSize int) (*response.AnalysisTaskListResponse, error) {
	if m.GetTasksErr != nil {
		return nil, m.GetTasksErr
	}
	return m.GetTasksResult, nil
}

func (m *MockAIService) GetAnalysisReportDetail(userID, reportID uint64) (*response.AnalysisReportDetailResponse, error) {
	if m.GetReportDetailErr != nil {
		return nil, m.GetReportDetailErr
	}
	return m.GetReportDetailResult, nil
}

func (m *MockAIService) GenerateInvestmentSummary(userID uint64, startDate, endDate string) (*response.AnalysisReportResponse, error) {
	if m.GenerateSummaryErr != nil {
		return nil, m.GenerateSummaryErr
	}
	return m.GenerateSummaryResult, nil
}

func (m *MockAIService) GetReports(userID uint64, reportType string, limit int) ([]response.AnalysisReportResponse, error) {
	if m.GetReportsErr != nil {
		return nil, m.GetReportsErr
	}
	return m.GetReportsResult, nil
}

// TestAnalysisHandler_CreateTask 测试创建分析任务
func TestAnalysisHandler_CreateTask(t *testing.T) {
	mockService := &MockAIService{
		CreateTaskResult: &response.AnalysisTaskResponse{
			ID:            1,
			Status:        "pending",
			ProgressStage: "pending",
			CreatedAt:     "2024-01-01 10:00:00",
		},
	}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/analysis/tasks", h.CreateTask)

	body := `{"start_date":"2024-01-01","end_date":"2024-12-31"}`
	req := httptest.NewRequest("POST", "/analysis/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp["data"].(map[string]interface{})
	if data["status"].(string) != "pending" {
		t.Errorf("Expected status pending, got %v", data["status"])
	}
}

// TestAnalysisHandler_CreateTask_InvalidJSON 测试无效 JSON
func TestAnalysisHandler_CreateTask_InvalidJSON(t *testing.T) {
	mockService := &MockAIService{}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/analysis/tasks", h.CreateTask)

	req := httptest.NewRequest("POST", "/analysis/tasks", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestAnalysisHandler_CreateTask_ServiceError 测试服务层错误
func TestAnalysisHandler_CreateTask_ServiceError(t *testing.T) {
	mockService := &MockAIService{
		CreateTaskErr: service.ErrTransactionNotFound,
	}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/analysis/tasks", h.CreateTask)

	body := `{"start_date":"2024-01-01","end_date":"2024-12-31"}`
	req := httptest.NewRequest("POST", "/analysis/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestAnalysisHandler_GetTask 测试获取任务
func TestAnalysisHandler_GetTask(t *testing.T) {
	mockService := &MockAIService{
		GetTaskResult: &response.AnalysisTaskDetailResponse{
			ID:            1,
			TaskType:      "stock_analysis",
			Status:        "success",
			ProgressStage: "completed",
		},
	}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/tasks/:id", h.GetTask)

	req := httptest.NewRequest("GET", "/analysis/tasks/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// TestAnalysisHandler_GetTask_InvalidID 测试无效任务 ID
func TestAnalysisHandler_GetTask_InvalidID(t *testing.T) {
	mockService := &MockAIService{}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/tasks/:id", h.GetTask)

	req := httptest.NewRequest("GET", "/analysis/tasks/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestAnalysisHandler_GetTask_NotFound 测试任务不存在
func TestAnalysisHandler_GetTask_NotFound(t *testing.T) {
	mockService := &MockAIService{
		GetTaskErr: service.ErrTransactionNotFound,
	}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/tasks/:id", h.GetTask)

	req := httptest.NewRequest("GET", "/analysis/tasks/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// TestAnalysisHandler_GetTasks 测试获取任务列表
func TestAnalysisHandler_GetTasks(t *testing.T) {
	mockService := &MockAIService{
		GetTasksResult: &response.AnalysisTaskListResponse{
			Items: []response.AnalysisTaskDetailResponse{
				{ID: 1, TaskType: "stock_analysis", Status: "pending"},
				{ID: 2, TaskType: "stock_analysis", Status: "success"},
			},
			Total:    2,
			Page:     1,
			PageSize: 10,
		},
	}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/tasks", h.GetTasks)

	req := httptest.NewRequest("GET", "/analysis/tasks?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp["data"].(map[string]interface{})
	if data["total"].(float64) != 2 {
		t.Errorf("Expected total 2, got %v", data["total"])
	}
}

// TestAnalysisHandler_GetTasks_WithStatus 测试带状态筛选
func TestAnalysisHandler_GetTasks_WithStatus(t *testing.T) {
	mockService := &MockAIService{
		GetTasksResult: &response.AnalysisTaskListResponse{
			Items: []response.AnalysisTaskDetailResponse{
				{ID: 1, TaskType: "stock_analysis", Status: "pending"},
			},
			Total:    1,
			Page:     1,
			PageSize: 10,
		},
	}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/tasks", h.GetTasks)

	req := httptest.NewRequest("GET", "/analysis/tasks?status=pending", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestAnalysisHandler_GetReportDetail 测试获取报告详情
func TestAnalysisHandler_GetReportDetail(t *testing.T) {
	mockService := &MockAIService{
		GetReportDetailResult: &response.AnalysisReportDetailResponse{
			ID:          1,
			ReportType:  "summary",
			ReportTitle: "股票分析报告",
		},
	}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/reports/:id", h.GetReportDetail)

	req := httptest.NewRequest("GET", "/analysis/reports/1", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// TestAnalysisHandler_GetReportDetail_InvalidID 测试无效报告 ID
func TestAnalysisHandler_GetReportDetail_InvalidID(t *testing.T) {
	mockService := &MockAIService{}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/reports/:id", h.GetReportDetail)

	req := httptest.NewRequest("GET", "/analysis/reports/invalid", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestAnalysisHandler_GetReportDetail_NotFound 测试报告不存在
func TestAnalysisHandler_GetReportDetail_NotFound(t *testing.T) {
	mockService := &MockAIService{
		GetReportDetailErr: service.ErrTransactionNotFound,
	}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/reports/:id", h.GetReportDetail)

	req := httptest.NewRequest("GET", "/analysis/reports/999", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// TestAnalysisHandler_GenerateSummary 测试生成投资总结
func TestAnalysisHandler_GenerateSummary(t *testing.T) {
	mockService := &MockAIService{
		GenerateSummaryResult: &response.AnalysisReportResponse{
			ID:          1,
			ReportType:  "summary",
			ReportTitle: "投资总结 (2024-01-01 至 2024-12-31)",
			SummaryText: "这是投资总结内容",
		},
	}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/analysis/summary", h.GenerateSummary)

	req := httptest.NewRequest("POST", "/analysis/summary?start_date=2024-01-01&end_date=2024-12-31", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// TestAnalysisHandler_GenerateSummary_MissingDates 测试缺少日期参数
func TestAnalysisHandler_GenerateSummary_MissingDates(t *testing.T) {
	mockService := &MockAIService{}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/analysis/summary", h.GenerateSummary)

	req := httptest.NewRequest("POST", "/analysis/summary", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestAnalysisHandler_GenerateSummary_MissingStartDate 测试缺少开始日期
func TestAnalysisHandler_GenerateSummary_MissingStartDate(t *testing.T) {
	mockService := &MockAIService{}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/analysis/summary", h.GenerateSummary)

	req := httptest.NewRequest("POST", "/analysis/summary?end_date=2024-12-31", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestAnalysisHandler_GenerateSummary_ServiceError 测试服务层错误
func TestAnalysisHandler_GenerateSummary_ServiceError(t *testing.T) {
	mockService := &MockAIService{
		GenerateSummaryErr: service.ErrTransactionNotFound,
	}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/analysis/summary", h.GenerateSummary)

	req := httptest.NewRequest("POST", "/analysis/summary?start_date=2024-01-01&end_date=2024-12-31", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// TestAnalysisHandler_GetReports 测试获取历史报告
func TestAnalysisHandler_GetReports(t *testing.T) {
	mockService := &MockAIService{
		GetReportsResult: []response.AnalysisReportResponse{
			{ID: 1, ReportType: "summary", ReportTitle: "报告1"},
			{ID: 2, ReportType: "summary", ReportTitle: "报告2"},
		},
	}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/reports", h.GetReports)

	req := httptest.NewRequest("GET", "/analysis/reports?limit=10", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	data := resp["data"].([]interface{})
	if len(data) != 2 {
		t.Errorf("Expected 2 reports, got %d", len(data))
	}
}

// TestAnalysisHandler_GetReports_WithType 测试带类型筛选
func TestAnalysisHandler_GetReports_WithType(t *testing.T) {
	mockService := &MockAIService{
		GetReportsResult: []response.AnalysisReportResponse{
			{ID: 1, ReportType: "risk", ReportTitle: "风险报告"},
		},
	}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/reports", h.GetReports)

	req := httptest.NewRequest("GET", "/analysis/reports?report_type=risk", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// TestAnalysisHandler_GetReports_Empty 测试空报告列表
func TestAnalysisHandler_GetReports_Empty(t *testing.T) {
	mockService := &MockAIService{
		GetReportsResult: []response.AnalysisReportResponse{},
	}

	h := handler.NewAnalysisHandler(mockService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/reports", h.GetReports)

	req := httptest.NewRequest("GET", "/analysis/reports", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

// 禁用未使用变量警告
var _ = time.Now()
