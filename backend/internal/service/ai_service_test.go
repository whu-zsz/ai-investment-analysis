package service_test

import (
	"context"
	"stock-analysis-backend/internal/dto/request"
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/service"
	"strings"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MockLLMProvider 模拟 LLM 提供者
type MockLLMProvider struct {
	Content     string
	modelName   string
	Err         error
}

func (m *MockLLMProvider) GetContent(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	return m.Content, nil
}

func (m *MockLLMProvider) ModelName() string {
	if m.modelName == "" {
		return "test-model"
	}
	return m.modelName
}

// MockAnalysisTaskRepository 模拟分析任务仓储
type MockAnalysisTaskRepository struct {
	Tasks        map[uint64]*model.AnalysisTask
	NextID       uint64
	HasRunning   bool
	RunningError error
}

func NewMockAnalysisTaskRepository() *MockAnalysisTaskRepository {
	return &MockAnalysisTaskRepository{
		Tasks:  make(map[uint64]*model.AnalysisTask),
		NextID: 1,
	}
}

func (r *MockAnalysisTaskRepository) Create(task *model.AnalysisTask) error {
	task.ID = r.NextID
	r.Tasks[r.NextID] = task
	r.NextID++
	return nil
}

func (r *MockAnalysisTaskRepository) FindByIDAndUserID(id, userID uint64) (*model.AnalysisTask, error) {
	task, ok := r.Tasks[id]
	if !ok || task.UserID != userID {
		return nil, gorm.ErrRecordNotFound
	}
	return task, nil
}

func (r *MockAnalysisTaskRepository) FindByUserID(userID uint64, status string, limit, offset int) ([]model.AnalysisTask, int64, error) {
	var result []model.AnalysisTask
	for _, t := range r.Tasks {
		if t.UserID == userID {
			result = append(result, *t)
		}
	}
	return result, int64(len(result)), nil
}

func (r *MockAnalysisTaskRepository) UpdateProgress(id uint64, status, stage string, errorMsg *string, reportID *uint64, startedAt, finishedAt *time.Time) error {
	task, ok := r.Tasks[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	task.Status = status
	task.ProgressStage = stage
	if reportID != nil {
		task.ResultReportID = reportID
	}
	return nil
}

func (r *MockAnalysisTaskRepository) HasRunningTask(userID uint64, taskType string) (bool, error) {
	if r.RunningError != nil {
		return false, r.RunningError
	}
	return r.HasRunning, nil
}

// MockAnalysisReportRepository 模拟分析报告仓储
type MockAnalysisReportRepository struct {
	Reports    map[uint64]*model.AnalysisReport
	NextID     uint64
	LastReport *model.AnalysisReport
	LastItems  []model.AnalysisReportItem
}

func NewMockAnalysisReportRepository() *MockAnalysisReportRepository {
	return &MockAnalysisReportRepository{
		Reports: make(map[uint64]*model.AnalysisReport),
		NextID:  1,
	}
}

func (r *MockAnalysisReportRepository) Create(report *model.AnalysisReport) error {
	report.ID = r.NextID
	r.Reports[r.NextID] = report
	r.LastReport = report
	r.NextID++
	return nil
}

func (r *MockAnalysisReportRepository) FindByUserID(userID uint64, reportType string, limit int) ([]model.AnalysisReport, error) {
	var result []model.AnalysisReport
	for _, rp := range r.Reports {
		if rp.UserID == userID {
			result = append(result, *rp)
		}
	}
	return result, nil
}

func (r *MockAnalysisReportRepository) FindByIDAndUserID(id, userID uint64) (*model.AnalysisReport, error) {
	report, ok := r.Reports[id]
	if !ok || report.UserID != userID {
		return nil, gorm.ErrRecordNotFound
	}
	return report, nil
}

func (r *MockAnalysisReportRepository) FindByID(id uint64) (*model.AnalysisReport, error) {
	report, ok := r.Reports[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return report, nil
}

func (r *MockAnalysisReportRepository) FindByTaskID(taskID uint64) (*model.AnalysisReport, error) {
	for _, rp := range r.Reports {
		if rp.TaskID != nil && *rp.TaskID == taskID {
			return rp, nil
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func (r *MockAnalysisReportRepository) FindLatestByUser(userID uint64, reportType string) (*model.AnalysisReport, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *MockAnalysisReportRepository) CreateWithItems(report *model.AnalysisReport, items []model.AnalysisReportItem) error {
	if err := r.Create(report); err != nil {
		return err
	}
	r.LastItems = append([]model.AnalysisReportItem(nil), items...)
	return nil
}

func (r *MockAnalysisReportRepository) Delete(id uint64) error {
	_, ok := r.Reports[id]
	if !ok {
		return gorm.ErrRecordNotFound
	}
	delete(r.Reports, id)
	return nil
}

// MockAnalysisReportItemRepository 模拟分析报告项仓储
type MockAnalysisReportItemRepository struct {
	Items map[uint64]*model.AnalysisReportItem
}

func NewMockAnalysisReportItemRepository() *MockAnalysisReportItemRepository {
	return &MockAnalysisReportItemRepository{
		Items: make(map[uint64]*model.AnalysisReportItem),
	}
}

func (r *MockAnalysisReportItemRepository) FindByReportID(reportID uint64) ([]model.AnalysisReportItem, error) {
	return []model.AnalysisReportItem{}, nil
}

func (r *MockAnalysisReportItemRepository) BatchCreate(items []model.AnalysisReportItem) error {
	return nil
}

// MockTransactionRepositoryForAI 模拟交易仓储
type MockTransactionRepositoryForAI struct {
	Transactions []model.Transaction
	Err          error
}

func (r *MockTransactionRepositoryForAI) Create(transaction *model.Transaction) error {
	return nil
}

func (r *MockTransactionRepositoryForAI) BatchCreate(transactions []model.Transaction) error {
	return nil
}

func (r *MockTransactionRepositoryForAI) FindByID(id uint64) (*model.Transaction, error) {
	return nil, gorm.ErrRecordNotFound
}

func (r *MockTransactionRepositoryForAI) FindByUserID(userID uint64, limit, offset int) ([]model.Transaction, int64, error) {
	return []model.Transaction{}, 0, nil
}

func (r *MockTransactionRepositoryForAI) FindByAssetCode(userID uint64, assetCode string) ([]model.Transaction, error) {
	return []model.Transaction{}, nil
}

func (r *MockTransactionRepositoryForAI) FindByDateRange(userID uint64, startDate, endDate string) ([]model.Transaction, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	return r.Transactions, nil
}

func (r *MockTransactionRepositoryForAI) Update(transaction *model.Transaction) error {
	return nil
}

func (r *MockTransactionRepositoryForAI) Delete(id uint64) error {
	return nil
}

func (r *MockTransactionRepositoryForAI) GetTransactionStats(userID uint64) (*response.TransactionStats, error) {
	return &response.TransactionStats{}, nil
}

// MockStockMetricService 模拟股票指标服务
type MockStockMetricService struct {
	Metrics []model.StockAnalysisMetric
	Err     error
}

func (m *MockStockMetricService) PrepareMetrics(ctx context.Context, userID uint64, taskID *uint64, startTime, endTime time.Time, symbols []string, forceRefreshMarket, forceRefreshMetrics bool) ([]model.StockAnalysisMetric, error) {
	if m.Err != nil {
		return nil, m.Err
	}
	return m.Metrics, nil
}

// 辅助函数：创建测试用 AIService
func createTestAIService() (service.AIService, *MockAnalysisTaskRepository, *MockAnalysisReportRepository, *MockTransactionRepositoryForAI) {
	taskRepo := NewMockAnalysisTaskRepository()
	reportRepo := NewMockAnalysisReportRepository()
	txRepo := &MockTransactionRepositoryForAI{}

	aiService := service.NewAIService(
		taskRepo,
		reportRepo,
		NewMockAnalysisReportItemRepository(),
		txRepo,
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	return aiService, taskRepo, reportRepo, txRepo
}

// TestAIService_GetReports 测试获取报告列表
func TestAIService_GetReports(t *testing.T) {
	reportRepo := NewMockAnalysisReportRepository()
	reportRepo.Create(&model.AnalysisReport{
		UserID:      1,
		ReportType:  "summary",
		ReportTitle: "测试报告",
	})

	aiService := service.NewAIService(
		NewMockAnalysisTaskRepository(),
		reportRepo,
		NewMockAnalysisReportItemRepository(),
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	reports, err := aiService.GetReports(1, "", 10)
	if err != nil {
		t.Fatalf("GetReports() error = %v", err)
	}

	if len(reports) != 1 {
		t.Errorf("Expected 1 report, got %d", len(reports))
	}
}

// TestAIService_GetReports_Empty 测试空报告列表

func TestAIService_GetReports_NormalizesLegacyChartData(t *testing.T) {
	reportRepo := NewMockAnalysisReportRepository()
	legacyChartData := `[{"symbol":"600519.SH","value":"5000.00"}]`
	reportRepo.Create(&model.AnalysisReport{
		UserID:      1,
		ReportType:  "summary",
		ReportTitle: "历史图表报告",
		ChartData:   &legacyChartData,
	})

	aiService := service.NewAIService(
		NewMockAnalysisTaskRepository(),
		reportRepo,
		NewMockAnalysisReportItemRepository(),
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	reports, err := aiService.GetReports(1, "", 10)
	if err != nil {
		t.Fatalf("GetReports() error = %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("Expected 1 report, got %d", len(reports))
	}

	expected := `{"version":2,"kind":"profit_by_symbol","metric":"total_profit","points":[{"symbol":"600519.SH","value":"5000.00"}]}`
	if reports[0].ChartData != expected {
		t.Errorf("Expected normalized chart_data %s, got %s", expected, reports[0].ChartData)
	}
}

func TestAIService_GetAnalysisReportDetail_NormalizesLegacyChartData(t *testing.T) {
	reportRepo := NewMockAnalysisReportRepository()
	legacyChartData := `[{"symbol":"000001.SH","value":"-120.50"}]`
	reportRepo.Create(&model.AnalysisReport{
		UserID:      1,
		ReportType:  "summary",
		ReportTitle: "历史详情报告",
		ChartData:   &legacyChartData,
	})

	aiService := service.NewAIService(
		NewMockAnalysisTaskRepository(),
		reportRepo,
		NewMockAnalysisReportItemRepository(),
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	detail, err := aiService.GetAnalysisReportDetail(1, 1)
	if err != nil {
		t.Fatalf("GetAnalysisReportDetail() error = %v", err)
	}

	expected := `{"version":2,"kind":"profit_by_symbol","metric":"total_profit","points":[{"symbol":"000001.SH","value":"-120.50"}]}`
	if detail.ChartData != expected {
		t.Errorf("Expected normalized chart_data %s, got %s", expected, detail.ChartData)
	}
}

func TestAIService_GetReports_DropsInvalidChartData(t *testing.T) {
	reportRepo := NewMockAnalysisReportRepository()
	invalidChartData := `{"radar":[1,2,3]}`
	reportRepo.Create(&model.AnalysisReport{
		UserID:      1,
		ReportType:  "summary",
		ReportTitle: "非法图表报告",
		ChartData:   &invalidChartData,
	})

	aiService := service.NewAIService(
		NewMockAnalysisTaskRepository(),
		reportRepo,
		NewMockAnalysisReportItemRepository(),
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	reports, err := aiService.GetReports(1, "", 10)
	if err != nil {
		t.Fatalf("GetReports() error = %v", err)
	}
	if len(reports) != 1 {
		t.Fatalf("Expected 1 report, got %d", len(reports))
	}
	if reports[0].ChartData != "" {
		t.Errorf("Expected empty chart_data for invalid input, got %s", reports[0].ChartData)
	}
}

func TestAIService_GetReports_DefaultLimit(t *testing.T) {
	aiService, _, _, _ := createTestAIService()

	// limit <= 0 应该使用默认值 10
	reports, err := aiService.GetReports(1, "", 0)
	if err != nil {
		t.Fatalf("GetReports() error = %v", err)
	}

	// 空结果，不应报错
	if reports == nil {
		t.Error("Expected empty slice, got nil")
	}
}

// TestAIService_GetAnalysisTasks 测试获取分析任务列表
func TestAIService_GetAnalysisTasks(t *testing.T) {
	taskRepo := NewMockAnalysisTaskRepository()
	taskRepo.Create(&model.AnalysisTask{
		UserID:   1,
		TaskType: "stock_analysis",
		Status:   "pending",
	})

	aiService := service.NewAIService(
		taskRepo,
		NewMockAnalysisReportRepository(),
		NewMockAnalysisReportItemRepository(),
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	result, err := aiService.GetAnalysisTasks(1, "", 1, 10)
	if err != nil {
		t.Fatalf("GetAnalysisTasks() error = %v", err)
	}

	if result.Total != 1 {
		t.Errorf("Expected total 1, got %d", result.Total)
	}

	if result.Page != 1 {
		t.Errorf("Expected page 1, got %d", result.Page)
	}

	if result.PageSize != 10 {
		t.Errorf("Expected pageSize 10, got %d", result.PageSize)
	}
}

// TestAIService_GetAnalysisTasks_DefaultPagination 测试默认分页
func TestAIService_GetAnalysisTasks_DefaultPagination(t *testing.T) {
	aiService, _, _, _ := createTestAIService()

	// page <= 0 和 pageSize <= 0 应该使用默认值
	result, err := aiService.GetAnalysisTasks(1, "", 0, 0)
	if err != nil {
		t.Fatalf("GetAnalysisTasks() error = %v", err)
	}

	if result.Page != 1 {
		t.Errorf("Expected default page 1, got %d", result.Page)
	}

	if result.PageSize != 10 {
		t.Errorf("Expected default pageSize 10, got %d", result.PageSize)
	}
}

// TestAIService_GetAnalysisTask 测试获取单个分析任务
func TestAIService_GetAnalysisTask(t *testing.T) {
	taskRepo := NewMockAnalysisTaskRepository()
	taskRepo.Create(&model.AnalysisTask{
		UserID:   1,
		TaskType: "stock_analysis",
		Status:   "pending",
	})

	aiService := service.NewAIService(
		taskRepo,
		NewMockAnalysisReportRepository(),
		NewMockAnalysisReportItemRepository(),
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	task, err := aiService.GetAnalysisTask(1, 1)
	if err != nil {
		t.Fatalf("GetAnalysisTask() error = %v", err)
	}

	if task.ID != 1 {
		t.Errorf("Expected task ID 1, got %d", task.ID)
	}

	if task.TaskType != "stock_analysis" {
		t.Errorf("Expected taskType stock_analysis, got %s", task.TaskType)
	}
}

// TestAIService_GetAnalysisTask_NotFound 测试获取不存在的任务
func TestAIService_GetAnalysisTask_NotFound(t *testing.T) {
	aiService, _, _, _ := createTestAIService()

	_, err := aiService.GetAnalysisTask(1, 999)
	if err == nil {
		t.Error("Expected error for non-existent task")
	}
}

// TestAIService_GetAnalysisTask_WrongUser 测试获取其他用户的任务
func TestAIService_GetAnalysisTask_WrongUser(t *testing.T) {
	taskRepo := NewMockAnalysisTaskRepository()
	taskRepo.Create(&model.AnalysisTask{
		UserID:   1,
		TaskType: "stock_analysis",
		Status:   "pending",
	})

	aiService := service.NewAIService(
		taskRepo,
		NewMockAnalysisReportRepository(),
		NewMockAnalysisReportItemRepository(),
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	// 用户 2 尝试访问用户 1 的任务
	_, err := aiService.GetAnalysisTask(2, 1)
	if err == nil {
		t.Error("Expected error for accessing other user's task")
	}
}

// TestAIService_CreateStockAnalysisTask_InvalidDate 测试无效日期
func TestAIService_CreateStockAnalysisTask_InvalidDate(t *testing.T) {
	aiService, _, _, _ := createTestAIService()

	req := &request.CreateAnalysisTaskRequest{
		StartDate: "invalid-date",
		EndDate:   "2024-12-31",
	}

	_, err := aiService.CreateStockAnalysisTask(1, req)
	if err == nil {
		t.Error("Expected error for invalid date")
	}
}

// TestAIService_CreateStockAnalysisTask_InvalidEndDate 测试无效结束日期
func TestAIService_CreateStockAnalysisTask_InvalidEndDate(t *testing.T) {
	aiService, _, _, _ := createTestAIService()

	req := &request.CreateAnalysisTaskRequest{
		StartDate: "2024-01-01",
		EndDate:   "invalid-date",
	}

	_, err := aiService.CreateStockAnalysisTask(1, req)
	if err == nil {
		t.Error("Expected error for invalid end date")
	}
}

// TestAIService_CreateStockAnalysisTask_EndBeforeStart 测试结束日期早于开始日期
func TestAIService_CreateStockAnalysisTask_EndBeforeStart(t *testing.T) {
	aiService, _, _, _ := createTestAIService()

	req := &request.CreateAnalysisTaskRequest{
		StartDate: "2024-12-31",
		EndDate:   "2024-01-01",
	}

	_, err := aiService.CreateStockAnalysisTask(1, req)
	if err == nil {
		t.Error("Expected error for end date before start date")
	}
}

// TestAIService_CreateStockAnalysisTask_AlreadyRunning 测试任务已在运行
func TestAIService_CreateStockAnalysisTask_AlreadyRunning(t *testing.T) {
	taskRepo := NewMockAnalysisTaskRepository()
	taskRepo.HasRunning = true

	aiService := service.NewAIService(
		taskRepo,
		NewMockAnalysisReportRepository(),
		NewMockAnalysisReportItemRepository(),
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	req := &request.CreateAnalysisTaskRequest{
		StartDate: "2024-01-01",
		EndDate:   "2024-12-31",
	}

	_, err := aiService.CreateStockAnalysisTask(1, req)
	if err == nil {
		t.Error("Expected error for already running task")
	}
}

// TestAIService_GetAnalysisReportDetail 测试获取报告详情
func TestAIService_GetAnalysisReportDetail(t *testing.T) {
	reportRepo := NewMockAnalysisReportRepository()
	reportRepo.Create(&model.AnalysisReport{
		UserID:      1,
		ReportType:  "summary",
		ReportTitle: "测试报告",
	})

	aiService := service.NewAIService(
		NewMockAnalysisTaskRepository(),
		reportRepo,
		NewMockAnalysisReportItemRepository(),
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	detail, err := aiService.GetAnalysisReportDetail(1, 1)
	if err != nil {
		t.Fatalf("GetAnalysisReportDetail() error = %v", err)
	}

	if detail.ID != 1 {
		t.Errorf("Expected report ID 1, got %d", detail.ID)
	}

	if detail.ReportTitle != "测试报告" {
		t.Errorf("Expected title '测试报告', got %s", detail.ReportTitle)
	}
}

// TestAIService_GetAnalysisReportDetail_NotFound 测试获取不存在的报告
func TestAIService_GetAnalysisReportDetail_NotFound(t *testing.T) {
	aiService, _, _, _ := createTestAIService()

	_, err := aiService.GetAnalysisReportDetail(1, 999)
	if err == nil {
		t.Error("Expected error for non-existent report")
	}
}

// TestAIService_GetAnalysisReportDetail_WrongUser 测试获取其他用户的报告
func TestAIService_GetAnalysisReportDetail_WrongUser(t *testing.T) {
	reportRepo := NewMockAnalysisReportRepository()
	reportRepo.Create(&model.AnalysisReport{
		UserID:      1,
		ReportType:  "summary",
		ReportTitle: "测试报告",
	})

	aiService := service.NewAIService(
		NewMockAnalysisTaskRepository(),
		reportRepo,
		NewMockAnalysisReportItemRepository(),
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	// 用户 2 尝试访问用户 1 的报告
	_, err := aiService.GetAnalysisReportDetail(2, 1)
	if err == nil {
		t.Error("Expected error for accessing other user's report")
	}
}

// TestAIService_GenerateInvestmentSummary_NoTransactions 测试无交易记录
func TestAIService_GenerateInvestmentSummary_NoTransactions(t *testing.T) {
	aiService, _, _, txRepo := createTestAIService()
	txRepo.Transactions = []model.Transaction{} // 空交易

	_, err := aiService.GenerateInvestmentSummary(1, "2024-01-01", "2024-12-31")
	if err == nil {
		t.Error("Expected error for no transactions")
	}
}

// TestAIService_GenerateInvestmentSummary_Success 测试生成投资总结成功
func TestAIService_GenerateInvestmentSummary_Success(t *testing.T) {
	txRepo := &MockTransactionRepositoryForAI{
		Transactions: []model.Transaction{
			{
				UserID:          1,
				AssetCode:       "600519",
				AssetName:       "贵州茅台",
				TransactionType: "buy",
				Quantity:        decimal.NewFromInt(100),
				TotalAmount:     decimal.NewFromFloat(185000),
			},
		},
	}

	llmProvider := &MockLLMProvider{
		Content:   `{"summary":{"report_title":"测试总结报告","summary_text":"组合总盈亏为正，贵州茅台贡献了主要收益。当前交易次数较少，但单一标的集中度偏高。","risk_level":"medium","investment_style":"balanced","risk_analysis":"持仓集中在单一标的，若价格回撤会放大波动。","pattern_insights":"交易集中在少数标的，偏向低频持有。","prediction_text":"若继续维持当前集中仓位，后续收益弹性与回撤风险都会同时放大。","recommendations":["控制仓位","关注回撤"]},"stocks":[{"symbol":"600519.SH","asset_name":"贵州茅台","risk_level":"low","investment_style":"value","analysis_text":"贵州茅台当前总盈亏为正，且交易次数较少，更接近持有型行为。由于仓位集中，后续波动会更直接影响组合表现。","recommendation":"hold","key_points":["龙头白酒"]}]}`,
		modelName: "test-model",
	}

	metricService := &MockStockMetricService{
		Metrics: []model.StockAnalysisMetric{
			{
				Symbol:           "600519.SH",
				AssetName:        "贵州茅台",
				TradeCount:       1,
				BuyCount:         1,
				BuyAmount:        decimal.NewFromFloat(185000),
				TotalProfit:      decimal.NewFromFloat(5000),
				MarketDataStatus: "complete",
			},
		},
	}

	aiService := service.NewAIService(
		NewMockAnalysisTaskRepository(),
		NewMockAnalysisReportRepository(),
		NewMockAnalysisReportItemRepository(),
		txRepo,
		metricService,
		llmProvider,
		zap.NewNop(),
	)

	report, err := aiService.GenerateInvestmentSummary(1, "2024-01-01", "2024-12-31")
	if err != nil {
		t.Fatalf("GenerateInvestmentSummary() error = %v", err)
	}

	if report.SummaryText != "组合总盈亏为正，贵州茅台贡献了主要收益。当前交易次数较少，但单一标的集中度偏高。" {
		t.Errorf("Expected summary text, got %s", report.SummaryText)
	}
	if len(report.Recommendations) != 2 {
		t.Fatalf("Expected 2 recommendations, got %d", len(report.Recommendations))
	}
	if report.Recommendations[0] != "控制仓位" {
		t.Errorf("Expected first recommendation to be 控制仓位, got %s", report.Recommendations[0])
	}
	if report.ChartData == "" {
		t.Error("Expected chart_data to be populated")
	}
	if report.RiskAnalysis != "持仓集中在单一标的，若价格回撤会放大波动。" {
		t.Errorf("Expected structured risk analysis, got %s", report.RiskAnalysis)
	}
}


// TestAIService_GenerateInvestmentSummary_NoMetrics 测试无指标数据
func TestAIService_GenerateInvestmentSummary_NoMetrics(t *testing.T) {
	txRepo := &MockTransactionRepositoryForAI{
		Transactions: []model.Transaction{
			{
				UserID:          1,
				AssetCode:       "600519",
				AssetName:       "贵州茅台",
				TransactionType: "buy",
				Quantity:        decimal.NewFromInt(100),
				TotalAmount:     decimal.NewFromFloat(185000),
			},
		},
	}

	aiService := service.NewAIService(
		NewMockAnalysisTaskRepository(),
		NewMockAnalysisReportRepository(),
		NewMockAnalysisReportItemRepository(),
		txRepo,
		&MockStockMetricService{Metrics: []model.StockAnalysisMetric{}},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	_, err := aiService.GenerateInvestmentSummary(1, "2024-01-01", "2024-12-31")
	if err == nil {
		t.Error("Expected error for no stock metrics")
	}
}

// TestAIService_CreateStockAnalysisTask_Success 测试创建分析任务成功
func waitForTaskStatus(t *testing.T, taskRepo *MockAnalysisTaskRepository, taskID uint64) *model.AnalysisTask {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		task := taskRepo.Tasks[taskID]
		if task != nil && (task.Status == "success" || task.Status == "failed") {
			return task
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("task %d did not finish", taskID)
	return nil
}

func TestAIService_CreateStockAnalysisTask_WeakOutputFails(t *testing.T) {
	txRepo := &MockTransactionRepositoryForAI{Transactions: []model.Transaction{{
		UserID:          1,
		AssetCode:       "600519",
		AssetName:       "贵州茅台",
		TransactionType: "buy",
		Quantity:        decimal.NewFromInt(100),
		TotalAmount:     decimal.NewFromFloat(185000),
	}}}
	metricService := &MockStockMetricService{Metrics: []model.StockAnalysisMetric{{
		Symbol:           "600519.SH",
		AssetName:        "贵州茅台",
		TradeCount:       1,
		BuyCount:         1,
		TotalProfit:      decimal.NewFromInt(5000),
		MarketDataStatus: "complete",
	}}}
	taskRepo := NewMockAnalysisTaskRepository()
	reportRepo := NewMockAnalysisReportRepository()
	aiService := service.NewAIService(
		taskRepo,
		reportRepo,
		NewMockAnalysisReportItemRepository(),
		txRepo,
		metricService,
		&MockLLMProvider{Content: `{"summary":{"report_title":"测试报告","summary_text":"测试总结","risk_level":"medium","investment_style":"balanced","risk_analysis":"","pattern_insights":"","prediction_text":"","recommendations":[]},"stocks":[{"symbol":"600519.SH","asset_name":"贵州茅台","risk_level":"low","investment_style":"value","analysis_text":"分析","recommendation":"hold","key_points":[]}]}`,
			modelName: "test-model",
		},
		zap.NewNop(),
	)
	task, err := aiService.CreateStockAnalysisTask(1, &request.CreateAnalysisTaskRequest{StartDate: "2024-01-01", EndDate: "2024-12-31"})
	if err != nil {
		t.Fatalf("CreateStockAnalysisTask() error = %v", err)
	}
	finished := waitForTaskStatus(t, taskRepo, task.ID)
	if finished.Status != "failed" {
		t.Fatalf("expected failed task, got %s", finished.Status)
	}
	if reportRepo.LastReport != nil {
		t.Fatalf("expected no report, got %+v", reportRepo.LastReport)
	}
}

func TestAIService_CreateStockAnalysisTask_CleansAndPersistsOutput(t *testing.T) {
	txRepo := &MockTransactionRepositoryForAI{Transactions: []model.Transaction{{
		UserID:          1,
		AssetCode:       "600519",
		AssetName:       "贵州茅台",
		TransactionType: "buy",
		Quantity:        decimal.NewFromInt(100),
		TotalAmount:     decimal.NewFromFloat(185000),
	}}}
	metricService := &MockStockMetricService{Metrics: []model.StockAnalysisMetric{{
		Symbol:           "600519.SH",
		AssetName:        "贵州茅台",
		TradeCount:       1,
		BuyCount:         1,
		TotalProfit:      decimal.NewFromInt(5000),
		MarketDataStatus: "complete",
	}}}
	taskRepo := NewMockAnalysisTaskRepository()
	reportRepo := NewMockAnalysisReportRepository()
	aiService := service.NewAIService(
		taskRepo,
		reportRepo,
		NewMockAnalysisReportItemRepository(),
		txRepo,
		metricService,
		&MockLLMProvider{Content: `{"summary":{"report_title":"  测试报告  ","summary_text":"  组合总盈亏为正，贵州茅台贡献明显。  ","risk_level":"medium","investment_style":"稳健型","risk_analysis":"  风险在于集中度较高  ","pattern_insights":"  重复交易集中在单一标的  ","prediction_text":"  若继续集中持仓，波动可能放大  ","recommendations":["- 减仓","1. 减仓","  复盘  ","复盘"]},"stocks":[{"symbol":"600519.SH","asset_name":"贵州茅台","risk_level":"low","investment_style":"价值型","analysis_text":"  该标的交易次数较少但总盈亏为正，持仓较集中。  ","recommendation":"hold","key_points":["- 贡献利润","1. 贡献利润","关注集中度"]}]}`,
			modelName: "test-model",
		},
		zap.NewNop(),
	)
	task, err := aiService.CreateStockAnalysisTask(1, &request.CreateAnalysisTaskRequest{StartDate: "2024-01-01", EndDate: "2024-12-31"})
	if err != nil {
		t.Fatalf("CreateStockAnalysisTask() error = %v", err)
	}
	finished := waitForTaskStatus(t, taskRepo, task.ID)
	if finished.Status != "success" {
		t.Fatalf("expected success task, got %s", finished.Status)
	}
	if reportRepo.LastReport == nil {
		t.Fatal("expected report to be persisted")
	}
	if reportRepo.LastReport.InvestmentStyle == nil || *reportRepo.LastReport.InvestmentStyle != "balanced" {
		t.Fatalf("expected normalized summary investment style, got %+v", reportRepo.LastReport.InvestmentStyle)
	}
	if reportRepo.LastReport.SummaryText != "组合总盈亏为正，贵州茅台贡献明显。" {
		t.Fatalf("unexpected summary text: %s", reportRepo.LastReport.SummaryText)
	}
	if reportRepo.LastReport.Recommendations == nil || *reportRepo.LastReport.Recommendations == "" {
		t.Fatal("expected recommendations to be persisted")
	}
	if len(reportRepo.LastItems) != 1 {
		t.Fatalf("expected 1 item, got %d", len(reportRepo.LastItems))
	}
	if reportRepo.LastItems[0].InvestmentStyle == nil || *reportRepo.LastItems[0].InvestmentStyle != "balanced" {
		t.Fatalf("expected normalized item investment style, got %+v", reportRepo.LastItems[0].InvestmentStyle)
	}
	if reportRepo.LastItems[0].KeyPoints == nil || *reportRepo.LastItems[0].KeyPoints == "" {
		t.Fatal("expected key points to be persisted")
	}
	if strings.Contains(*reportRepo.LastItems[0].KeyPoints, "贡献利润") {
		t.Fatalf("expected weak key point to be filtered, got %s", *reportRepo.LastItems[0].KeyPoints)
	}
	if !strings.Contains(*reportRepo.LastItems[0].KeyPoints, "关注集中度") {
		t.Fatalf("unexpected key points: %s", *reportRepo.LastItems[0].KeyPoints)
	}
}



func TestAIService_CreateStockAnalysisTask_TransparentFallbackForWeakStockAnalysis(t *testing.T) {
	txRepo := &MockTransactionRepositoryForAI{Transactions: []model.Transaction{{
		UserID:          1,
		AssetCode:       "600519",
		AssetName:       "贵州茅台",
		TransactionType: "buy",
		Quantity:        decimal.NewFromInt(100),
		TotalAmount:     decimal.NewFromFloat(185000),
	}}}
	metricService := &MockStockMetricService{Metrics: []model.StockAnalysisMetric{{
		Symbol:           "600519.SH",
		AssetName:        "贵州茅台",
		TradeCount:       1,
		BuyCount:         1,
		TotalProfit:      decimal.NewFromInt(5000),
		MarketDataStatus: "complete",
	}}}
	taskRepo := NewMockAnalysisTaskRepository()
	reportRepo := NewMockAnalysisReportRepository()
	aiService := service.NewAIService(
		taskRepo,
		reportRepo,
		NewMockAnalysisReportItemRepository(),
		txRepo,
		metricService,
		&MockLLMProvider{Content: `{"summary":{"report_title":"测试报告","summary_text":"组合总盈亏为正，贵州茅台贡献主要收益。当前持仓较集中，整体收益弹性较高。","risk_level":"medium","investment_style":"稳健型","risk_analysis":"单一标的集中度较高，若价格回撤则组合波动会被放大。","pattern_insights":"交易主要集中在贵州茅台，买入后以持有为主，呈现低频集中持有特征。","prediction_text":"若继续集中持仓，收益弹性可能提升，但回撤风险也会放大。","recommendations":["减仓","复盘"]},"stocks":[{"symbol":"600519.SH","asset_name":"贵州茅台","risk_level":"low","investment_style":"价值型","analysis_text":"分析","recommendation":"hold","key_points":["关注集中度"]}]}`,
			modelName: "test-model",
		},
		zap.NewNop(),
	)
	task, err := aiService.CreateStockAnalysisTask(1, &request.CreateAnalysisTaskRequest{StartDate: "2024-01-01", EndDate: "2024-12-31"})
	if err != nil {
		t.Fatalf("CreateStockAnalysisTask() error = %v", err)
	}
	finished := waitForTaskStatus(t, taskRepo, task.ID)
	if finished.Status != "success" {
		t.Fatalf("expected success task, got %s", finished.Status)
	}
	if reportRepo.LastReport == nil || len(reportRepo.LastItems) != 1 {
		t.Fatal("expected persisted report and items")
	}
	if !strings.Contains(reportRepo.LastItems[0].AnalysisText, "总盈亏") {
		t.Fatalf("expected transparent fallback analysis text, got %s", reportRepo.LastItems[0].AnalysisText)
	}
}
