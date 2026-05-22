package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"stock-analysis-backend/internal/dto/request"
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/handler"
	"stock-analysis-backend/internal/service"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
	ExportReportPDFResult     []byte
	ExportReportPDFFilename   string
	ExportReportPDFErr        error
	GenerateSummaryResult     *response.AnalysisReportResponse
	GenerateSummaryErr        error
	GetReportsResult          []response.AnalysisReportResponse
	GetReportsErr             error
}

type MockRecommendationService struct {
	CandidatesResult      *response.AnalysisCandidatesResponse
	CandidatesErr         error
	RecommendationsResult *response.AnalysisRecommendationsResponse
	RecommendationsErr    error
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

func (m *MockAIService) ExportAnalysisReportPDF(userID, reportID uint64) ([]byte, string, error) {
	if m.ExportReportPDFErr != nil {
		return nil, "", m.ExportReportPDFErr
	}
	return m.ExportReportPDFResult, m.ExportReportPDFFilename, nil
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

func (m *MockRecommendationService) GetCandidates(userID uint64) (*response.AnalysisCandidatesResponse, error) {
	if m.CandidatesErr != nil {
		return nil, m.CandidatesErr
	}
	return m.CandidatesResult, nil
}

func (m *MockRecommendationService) GetRecommendations(userID uint64) (*response.AnalysisRecommendationsResponse, error) {
	if m.RecommendationsErr != nil {
		return nil, m.RecommendationsErr
	}
	return m.RecommendationsResult, nil
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

// TestAnalysisHandler_ExportReportPDF 测试导出 PDF
func TestAnalysisHandler_ExportReportPDF(t *testing.T) {
	mockService := &MockAIService{
		ExportReportPDFResult:   []byte("%PDF-1.4"),
		ExportReportPDFFilename: "analysis-report-1.pdf",
	}

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/reports/:id/pdf", h.ExportReportPDF)

	req := httptest.NewRequest("GET", "/analysis/reports/1/pdf", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != "application/pdf" {
		t.Errorf("Expected Content-Type application/pdf, got %s", got)
	}
	if got := w.Header().Get("Content-Disposition"); got != "attachment; filename=\"analysis-report-1.pdf\"" {
		t.Errorf("Unexpected Content-Disposition: %s", got)
	}
	if string(w.Body.Bytes()) != "%PDF-1.4" {
		t.Errorf("Unexpected response body: %s", w.Body.String())
	}
}

// TestAnalysisHandler_ExportReportPDF_InvalidID 测试导出 PDF 时无效报告 ID
func TestAnalysisHandler_ExportReportPDF_InvalidID(t *testing.T) {
	mockService := &MockAIService{}

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/reports/:id/pdf", h.ExportReportPDF)

	req := httptest.NewRequest("GET", "/analysis/reports/invalid/pdf", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

// TestAnalysisHandler_ExportReportPDF_NotFound 测试导出 PDF 时报告不存在
func TestAnalysisHandler_ExportReportPDF_NotFound(t *testing.T) {
	mockService := &MockAIService{
		ExportReportPDFErr: gorm.ErrRecordNotFound,
	}

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/reports/:id/pdf", h.ExportReportPDF)

	req := httptest.NewRequest("GET", "/analysis/reports/999/pdf", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

// TestAnalysisHandler_ExportReportPDF_ServiceError 测试导出 PDF 时服务错误
func TestAnalysisHandler_ExportReportPDF_ServiceError(t *testing.T) {
	mockService := &MockAIService{
		ExportReportPDFErr: service.ErrTransactionNotFound,
	}

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/reports/:id/pdf", h.ExportReportPDF)

	req := httptest.NewRequest("GET", "/analysis/reports/1/pdf", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

// TestAnalysisHandler_GenerateSummary 测试生成投资总结
func TestAnalysisHandler_GenerateSummary(t *testing.T) {
	mockService := &MockAIService{
		GenerateSummaryResult: &response.AnalysisReportResponse{
			ID:              1,
			ReportType:      "summary",
			ReportTitle:     "投资总结 (2024-01-01 至 2024-12-31)",
			SummaryText:     "这是投资总结内容",
			Recommendations: []string{"控制仓位"},
		},
	}

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})
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

func TestAnalysisHandler_GetRecommendations(t *testing.T) {
	mockService := &MockAIService{}
	recommendationService := &MockRecommendationService{
		RecommendationsResult: &response.AnalysisRecommendationsResponse{
			GeneratedAt: "2026-05-11 12:00:00",
			SummaryText: "更适合优先关注熟悉标的。",
			Candidates: []response.RecommendationItemResponse{{
				Symbol:        "600519.SH",
				AssetName:     "贵州茅台",
				Action:        "hold",
				Score:         "81.00",
				MatchReason:   "当前已持仓，且近期有交易记录。",
				RiskNote:      "注意波动。",
				DataStatus:    "complete",
				LatestPrice:   "1680.0000",
				ChangePercent: "1.25",
			}},
		},
	}

	h := handler.NewAnalysisHandler(mockService, recommendationService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/recommendations", h.GetRecommendations)

	req := httptest.NewRequest("GET", "/analysis/recommendations", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestAnalysisHandler_GetCandidates(t *testing.T) {
	mockService := &MockAIService{}
	recommendationService := &MockRecommendationService{
		CandidatesResult: &response.AnalysisCandidatesResponse{
			DefaultSymbol: "600519.SH",
			Candidates: []response.AnalysisCandidateResponse{{
				Symbol:        "600519.SH",
				AssetName:     "贵州茅台",
				AssetType:     "stock",
				IsHeld:        true,
				TradeCount:    4,
				LastPrice:     "1680.0000",
				ChangePercent: "1.25",
			}},
		},
	}

	h := handler.NewAnalysisHandler(mockService, recommendationService)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.GET("/analysis/candidates", h.GetCandidates)

	req := httptest.NewRequest("GET", "/analysis/candidates", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

// ========== 安全测试 ==========

func TestAnalysis_Security_SQLInjection_Query(t *testing.T) {
	// SQL 注入 payload 作为路径参数，应被 handler 拒绝（无效 ID → 400）
	mockService := &MockAIService{}
	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})

	sqlInjectionPayloads := []string{
		"'; DROP TABLE analysis_tasks; --",
		"' UNION SELECT * FROM users --",
		"1' OR '1'='1",
	}

	for _, payload := range sqlInjectionPayloads {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uint64(1))
			c.Next()
		})
		router.GET("/analysis/tasks/:id", h.GetTask)

		req := httptest.NewRequest("GET", "/analysis/tasks/"+url.PathEscape(payload), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// handler 将非数字 ID 解析为无效参数，返回 400
		if w.Code != http.StatusBadRequest {
			t.Errorf("Query with SQL injection '%s' returned %d, want %d", payload, w.Code, http.StatusBadRequest)
		}
	}
}

func TestAnalysis_Security_XSS_Summary(t *testing.T) {
	// XSS payload 作为日期参数传入 summary 接口
	// handler 不校验日期格式，mock 服务返回错误时 handler 返回 500
	mockService := &MockAIService{
		GenerateSummaryErr: service.ErrTransactionNotFound,
	}
	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})

	xssPayloads := []string{
		"<script>alert('xss')</script>",
		"<img src=x onerror=alert('xss')>",
		"javascript:alert('xss')",
	}

	for _, payload := range xssPayloads {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uint64(1))
			c.Next()
		})
		router.POST("/analysis/summary", h.GenerateSummary)

		req := httptest.NewRequest("POST", "/analysis/summary?start_date="+url.QueryEscape(payload)+"&end_date=2024-12-31", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 服务层返回错误，handler 转为 500
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Summary with XSS '%s' returned %d, want %d", payload, w.Code, http.StatusInternalServerError)
		}
	}
}

func TestAnalysis_Security_InvalidDateRange(t *testing.T) {
	// 无效日期格式传入 summary 接口
	// handler 不校验日期格式，mock 服务返回错误时 handler 返回 500
	mockService := &MockAIService{
		GenerateSummaryErr: service.ErrTransactionNotFound,
	}
	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})

	invalidRanges := []struct {
		start string
		end   string
	}{
		{"invalid", "2024-12-31"},
		{"2024-01-01", "invalid"},
		{"2024-13-01", "2024-12-31"},
		{"2024-01-01", "2024-13-31"},
	}

	for _, r := range invalidRanges {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uint64(1))
			c.Next()
		})
		router.POST("/analysis/summary", h.GenerateSummary)

		req := httptest.NewRequest("POST", "/analysis/summary?start_date="+url.QueryEscape(r.start)+"&end_date="+url.QueryEscape(r.end), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 服务层返回错误，handler 转为 500
		if w.Code != http.StatusInternalServerError {
			t.Errorf("Summary with invalid date range (%s, %s) returned %d, want %d", r.start, r.end, w.Code, http.StatusInternalServerError)
		}
	}
}

func TestAnalysis_Security_EmptyDateRange(t *testing.T) {
	// 空日期参数，handler 直接校验并返回 400
	mockService := &MockAIService{}
	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})

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
		t.Errorf("Summary with empty date range returned %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAnalysis_Security_FutureDateRange(t *testing.T) {
	// 未来日期范围，handler 不校验日期合理性，mock 服务返回错误时 handler 返回 500
	mockService := &MockAIService{
		GenerateSummaryErr: service.ErrTransactionNotFound,
	}
	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})

	futureStart := time.Now().AddDate(1, 0, 0).Format("2006-01-02")
	futureEnd := time.Now().AddDate(2, 0, 0).Format("2006-01-02")

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/analysis/summary", h.GenerateSummary)

	req := httptest.NewRequest("POST", "/analysis/summary?start_date="+futureStart+"&end_date="+futureEnd, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 服务层返回错误，handler 转为 500
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Summary with future date range returned %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestAnalysis_Security_VeryLongSymbols(t *testing.T) {
	// 超长 symbols 数组（1000 个），mock 服务返回错误时 handler 返回 400
	mockService := &MockAIService{
		CreateTaskErr: service.ErrTransactionNotFound,
	}
	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})

	longSymbols := make([]string, 1000)
	for i := range longSymbols {
		longSymbols[i] = fmt.Sprintf("symbol_%d", i)
	}
	symbolsJSON, _ := json.Marshal(longSymbols)

	reqBody := fmt.Sprintf(`{
		"symbols": %s,
		"start_date": "2024-01-01",
		"end_date": "2024-12-31"
	}`, symbolsJSON)

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", uint64(1))
		c.Next()
	})
	router.POST("/analysis/tasks", h.CreateTask)

	req := httptest.NewRequest("POST", "/analysis/tasks", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 服务层返回错误，handler 转为 400
	if w.Code != http.StatusBadRequest {
		t.Errorf("Create task with very long symbols returned %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestAnalysis_Security_InvalidSymbols(t *testing.T) {
	// 无效 symbols（XSS/SQL注入/模板注入），mock 服务返回错误时 handler 返回 400
	mockService := &MockAIService{
		CreateTaskErr: service.ErrTransactionNotFound,
	}
	h := handler.NewAnalysisHandler(mockService, &MockRecommendationService{})

	invalidSymbols := []string{
		"<script>alert('xss')</script>",
		"'; DROP TABLE stocks; --",
		"${7*7}",
		"{{7*7}}",
	}

	for _, symbol := range invalidSymbols {
		router := gin.New()
		router.Use(func(c *gin.Context) {
			c.Set("user_id", uint64(1))
			c.Next()
		})
		router.POST("/analysis/tasks", h.CreateTask)

		reqBody := fmt.Sprintf(`{
			"symbols": ["%s"],
			"start_date": "2024-01-01",
			"end_date": "2024-12-31"
		}`, symbol)

		req := httptest.NewRequest("POST", "/analysis/tasks", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 服务层返回错误，handler 转为 400
		if w.Code != http.StatusBadRequest {
			t.Errorf("Create task with invalid symbol '%s' returned %d, want %d", symbol, w.Code, http.StatusBadRequest)
		}
	}
}
