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
	Content      string
	SystemPrompt string
	UserPrompt   string
	modelName    string
	Err          error
}

func (m *MockLLMProvider) GetContent(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	m.SystemPrompt = systemPrompt
	m.UserPrompt = userPrompt
	if m.Err != nil {
		return "", m.Err
	}
	return m.Content, nil
}

func (m *MockLLMProvider) GetContentStream(ctx context.Context, systemPrompt, userPrompt string, onToken func(string) error) (string, error) {
	if m.Err != nil {
		return "", m.Err
	}
	if onToken != nil && m.Content != "" {
		if err := onToken(m.Content); err != nil {
			return "", err
		}
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
	ItemsByReportID map[uint64][]model.AnalysisReportItem
}

func NewMockAnalysisReportItemRepository() *MockAnalysisReportItemRepository {
	return &MockAnalysisReportItemRepository{
		ItemsByReportID: make(map[uint64][]model.AnalysisReportItem),
	}
}

func (r *MockAnalysisReportItemRepository) FindByReportID(reportID uint64) ([]model.AnalysisReportItem, error) {
	items := r.ItemsByReportID[reportID]
	return append([]model.AnalysisReportItem(nil), items...), nil
}

func (r *MockAnalysisReportItemRepository) BatchCreate(items []model.AnalysisReportItem) error {
	for _, item := range items {
		r.ItemsByReportID[item.ReportID] = append(r.ItemsByReportID[item.ReportID], item)
	}
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

func TestAIService_GetAnalysisReportDetail_BuildsStructuredRiskInsights(t *testing.T) {
	reportRepo := NewMockAnalysisReportRepository()
	reportRepo.Create(&model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "风险预警报告",
		AnalysisPeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		AnalysisPeriodEnd:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		SymbolsCount:        2,
		WinningTrades:       1,
		LosingTrades:        1,
		TotalInvestment:     decimal.NewFromInt(100000),
		TotalProfit:         decimal.NewFromInt(-5000),
		ProfitRate:          decimal.NewFromFloat(-5),
		RiskLevel:           "medium",
		MarketDataStatus:    "partial",
		SummaryText:         "测试总结",
		AIModel:             "test-model",
	})

	itemRepo := NewMockAnalysisReportItemRepository()
	itemRepo.ItemsByReportID[1] = []model.AnalysisReportItem{
		{
			ID:                   1,
			ReportID:             1,
			UserID:               1,
			Symbol:               "000001.SZ",
			AssetName:            "平安银行",
			TradeCount:           9,
			BuyCount:             6,
			SellCount:            2,
			BuyAmount:            decimal.NewFromInt(60000),
			SellAmount:           decimal.NewFromInt(20000),
			NetQuantity:          decimal.NewFromInt(1000),
			RealizedProfit:       decimal.NewFromInt(-1000),
			RealizedProfitRate:   decimal.NewFromFloat(-2),
			EndingPositionQty:    decimal.NewFromInt(1000),
			EndingAvgCost:        decimal.NewFromFloat(12.5),
			LatestPrice:          decimal.NewFromFloat(10.2),
			LatestMarketValue:    decimal.NewFromInt(10200),
			UnrealizedProfit:     decimal.NewFromInt(-3000),
			TotalProfit:          decimal.NewFromInt(-4000),
			ChangePercent7D:      decimal.NewFromInt(10),
			PeriodPriceChangePct: decimal.NewFromInt(10),
			MarketDataStatus:     "partial",
			RiskLevel:            "high",
			AnalysisText:         "测试个股分析",
			Recommendation:       "reduce",
		},
		{
			ID:                   2,
			ReportID:             1,
			UserID:               1,
			Symbol:               "600519.SH",
			AssetName:            "贵州茅台",
			TradeCount:           2,
			BuyCount:             1,
			SellCount:            1,
			BuyAmount:            decimal.NewFromInt(40000),
			SellAmount:           decimal.NewFromInt(42000),
			NetQuantity:          decimal.Zero,
			RealizedProfit:       decimal.NewFromInt(2000),
			RealizedProfitRate:   decimal.NewFromInt(5),
			EndingPositionQty:    decimal.Zero,
			EndingAvgCost:        decimal.NewFromFloat(0),
			LatestPrice:          decimal.NewFromFloat(0),
			LatestMarketValue:    decimal.Zero,
			UnrealizedProfit:     decimal.Zero,
			TotalProfit:          decimal.NewFromInt(2000),
			ChangePercent7D:      decimal.NewFromInt(2),
			PeriodPriceChangePct: decimal.NewFromInt(2),
			MarketDataStatus:     "complete",
			RiskLevel:            "low",
			AnalysisText:         "测试个股分析2",
			Recommendation:       "hold",
		},
	}

	recommendationsJSON := `["控制仓位","减少高频交易"]`
	reportRepo.Reports[1].Recommendations = &recommendationsJSON

	aiService := service.NewAIService(
		NewMockAnalysisTaskRepository(),
		reportRepo,
		itemRepo,
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	detail, err := aiService.GetAnalysisReportDetail(1, 1)
	if err != nil {
		t.Fatalf("GetAnalysisReportDetail() error = %v", err)
	}

	if detail.RiskOverview.RiskLevel != "high" {
		t.Fatalf("expected overall risk level high, got %s", detail.RiskOverview.RiskLevel)
	}
	if detail.RiskOverview.RiskScore != 100 {
		t.Fatalf("expected overall risk score 100, got %d", detail.RiskOverview.RiskScore)
	}
	if len(detail.RiskAlerts) != 5 {
		t.Fatalf("expected 5 risk alerts, got %d", len(detail.RiskAlerts))
	}
	if len(detail.TopRiskSymbols) != 1 {
		t.Fatalf("expected 1 top risk symbol, got %d", len(detail.TopRiskSymbols))
	}
	if detail.TopRiskSymbols[0].Symbol != "000001.SZ" {
		t.Fatalf("expected top risk symbol 000001.SZ, got %s", detail.TopRiskSymbols[0].Symbol)
	}
	if detail.TopRiskSymbols[0].RiskScore != 100 {
		t.Fatalf("expected top risk score 100, got %d", detail.TopRiskSymbols[0].RiskScore)
	}
	if len(detail.TopRiskSymbols[0].TriggerReasons) != 5 {
		t.Fatalf("expected 5 trigger reasons, got %d", len(detail.TopRiskSymbols[0].TriggerReasons))
	}
	if len(detail.RiskOverview.Recommendations) != 2 {
		t.Fatalf("expected 2 recommendations, got %d", len(detail.RiskOverview.Recommendations))
	}
	if !strings.Contains(strings.Join(detail.RiskOverview.RiskFactors, ","), "持仓亏损预警") {
		t.Fatalf("expected drawdown alert in risk factors, got %v", detail.RiskOverview.RiskFactors)
	}
}

func TestAIService_GetAnalysisReportDetail_FetchedLiveDoesNotTriggerMarketDataAlert(t *testing.T) {
	reportRepo := NewMockAnalysisReportRepository()
	reportRepo.Create(&model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "实时补全报告",
		AnalysisPeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		AnalysisPeriodEnd:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		SymbolsCount:        1,
		WinningTrades:       1,
		LosingTrades:        0,
		TotalInvestment:     decimal.NewFromInt(20000),
		TotalProfit:         decimal.NewFromInt(1500),
		ProfitRate:          decimal.NewFromFloat(7.5),
		RiskLevel:           "low",
		MarketDataStatus:    "fetched_live",
		SummaryText:         "测试总结",
		AIModel:             "test-model",
	})

	itemRepo := NewMockAnalysisReportItemRepository()
	itemRepo.ItemsByReportID[1] = []model.AnalysisReportItem{{
		ID:                   1,
		ReportID:             1,
		UserID:               1,
		Symbol:               "600519.SH",
		AssetName:            "贵州茅台",
		TradeCount:           2,
		BuyCount:             1,
		SellCount:            1,
		BuyAmount:            decimal.NewFromInt(20000),
		SellAmount:           decimal.NewFromInt(21500),
		NetQuantity:          decimal.Zero,
		RealizedProfit:       decimal.NewFromInt(1500),
		RealizedProfitRate:   decimal.NewFromFloat(7.5),
		EndingPositionQty:    decimal.Zero,
		EndingAvgCost:        decimal.Zero,
		LatestPrice:          decimal.Zero,
		LatestMarketValue:    decimal.Zero,
		UnrealizedProfit:     decimal.Zero,
		TotalProfit:          decimal.NewFromInt(1500),
		ChangePercent7D:      decimal.NewFromInt(1),
		PeriodPriceChangePct: decimal.NewFromInt(1),
		MarketDataStatus:     "fetched_live",
		RiskLevel:            "low",
		AnalysisText:         "测试个股分析",
		Recommendation:       "hold",
	}}

	aiService := service.NewAIService(
		NewMockAnalysisTaskRepository(),
		reportRepo,
		itemRepo,
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	detail, err := aiService.GetAnalysisReportDetail(1, 1)
	if err != nil {
		t.Fatalf("GetAnalysisReportDetail() error = %v", err)
	}

	if len(detail.RiskAlerts) != 0 {
		t.Fatalf("expected 0 risk alerts for fetched_live, got %d", len(detail.RiskAlerts))
	}
	if len(detail.TopRiskSymbols) != 0 {
		t.Fatalf("expected 0 top risk symbols for fetched_live, got %d", len(detail.TopRiskSymbols))
	}
}

func TestAIService_GetAnalysisReportDetail_EmptyRiskInsights(t *testing.T) {
	reportRepo := NewMockAnalysisReportRepository()
	reportRepo.Create(&model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "低风险报告",
		AnalysisPeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		AnalysisPeriodEnd:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		SymbolsCount:        1,
		WinningTrades:       1,
		LosingTrades:        0,
		TotalInvestment:     decimal.NewFromInt(20000),
		TotalProfit:         decimal.NewFromInt(1500),
		ProfitRate:          decimal.NewFromFloat(7.5),
		RiskLevel:           "low",
		MarketDataStatus:    "complete",
		SummaryText:         "测试总结",
		AIModel:             "test-model",
	})

	itemRepo := NewMockAnalysisReportItemRepository()
	itemRepo.ItemsByReportID[1] = []model.AnalysisReportItem{
		{
			ID:                   1,
			ReportID:             1,
			UserID:               1,
			Symbol:               "600036.SH",
			AssetName:            "招商银行",
			TradeCount:           2,
			BuyCount:             1,
			SellCount:            1,
			BuyAmount:            decimal.NewFromInt(20000),
			SellAmount:           decimal.NewFromInt(21500),
			NetQuantity:          decimal.Zero,
			RealizedProfit:       decimal.NewFromInt(1500),
			RealizedProfitRate:   decimal.NewFromFloat(7.5),
			EndingPositionQty:    decimal.Zero,
			EndingAvgCost:        decimal.Zero,
			LatestPrice:          decimal.Zero,
			LatestMarketValue:    decimal.Zero,
			UnrealizedProfit:     decimal.Zero,
			TotalProfit:          decimal.NewFromInt(1500),
			ChangePercent7D:      decimal.NewFromInt(1),
			PeriodPriceChangePct: decimal.NewFromInt(1),
			MarketDataStatus:     "complete",
			RiskLevel:            "low",
			AnalysisText:         "测试个股分析",
			Recommendation:       "hold",
		},
	}

	aiService := service.NewAIService(
		NewMockAnalysisTaskRepository(),
		reportRepo,
		itemRepo,
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	detail, err := aiService.GetAnalysisReportDetail(1, 1)
	if err != nil {
		t.Fatalf("GetAnalysisReportDetail() error = %v", err)
	}

	if detail.RiskOverview.RiskLevel != "low" {
		t.Fatalf("expected overall risk level low, got %s", detail.RiskOverview.RiskLevel)
	}
	if detail.RiskOverview.RiskScore != 20 {
		t.Fatalf("expected overall risk score 20, got %d", detail.RiskOverview.RiskScore)
	}
	if len(detail.RiskAlerts) != 0 {
		t.Fatalf("expected 0 risk alerts, got %d", len(detail.RiskAlerts))
	}
	if len(detail.TopRiskSymbols) != 0 {
		t.Fatalf("expected 0 top risk symbols, got %d", len(detail.TopRiskSymbols))
	}
	if len(detail.RiskOverview.RiskFactors) != 0 {
		t.Fatalf("expected 0 risk factors, got %d", len(detail.RiskOverview.RiskFactors))
	}
}

func TestAIService_GetAnalysisReportDetail_UsesStructuredPredictionFromRawAIOutput(t *testing.T) {
	reportRepo := NewMockAnalysisReportRepository()
	predictionText := "若组合继续维持当前持仓结构，短期波动可能放大，但盈利弹性仍然存在。"
	rawOutput := `{"summary":{"report_title":"预测报告","summary_text":"组合总盈亏为正，主要收益来自强势标的。当前持仓集中度较高，组合弹性与波动并存。","risk_level":"medium","investment_style":"balanced","risk_analysis":"单一标的占比较高，若价格回撤会放大净值波动。","pattern_insights":"交易集中在少数标的，近期更偏向持有而非高频切换。","prediction_text":"若组合继续维持当前持仓结构，短期波动可能放大，但盈利弹性仍然存在。","prediction":{"bias":"看多","confidence":"较高","horizon":"未来7天","drivers":[" 600519.SH 周期涨跌幅 8.00% ","600519.SH 周期涨跌幅 8.00%","组合总盈亏为正","市场数据完整"],"scenarios":[{"condition":" 若龙头标的继续强势 ","outcome":" 组合净值有望继续修复 "},{"condition":"若龙头标的继续强势","outcome":"组合净值有望继续修复"},{"condition":"若集中仓位回撤","outcome":"组合波动将被放大"}]},"recommendations":["控制仓位","跟踪波动"]},"stocks":[{"symbol":"600519.SH","asset_name":"贵州茅台","risk_level":"low","investment_style":"value","analysis_text":"贵州茅台当前总盈亏为正，且阶段涨幅较高，对组合贡献明显。由于仓位相对集中，后续价格波动会直接影响组合表现。","recommendation":"hold","key_points":["强势标的贡献收益"]}]}`
	reportRepo.Create(&model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "预测报告",
		AnalysisPeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		AnalysisPeriodEnd:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		RiskLevel:           "medium",
		SummaryText:         "测试总结",
		PredictionText:      &predictionText,
		RawAIOutput:         &rawOutput,
		AIModel:             "test-model",
	})

	itemRepo := NewMockAnalysisReportItemRepository()
	itemRepo.ItemsByReportID[1] = []model.AnalysisReportItem{{
		ID:                   1,
		ReportID:             1,
		UserID:               1,
		Symbol:               "600519.SH",
		AssetName:            "贵州茅台",
		TradeCount:           2,
		BuyCount:             1,
		SellCount:            1,
		BuyAmount:            decimal.NewFromInt(100000),
		SellAmount:           decimal.NewFromInt(108000),
		NetQuantity:          decimal.NewFromInt(100),
		RealizedProfit:       decimal.NewFromInt(8000),
		RealizedProfitRate:   decimal.NewFromInt(8),
		EndingPositionQty:    decimal.NewFromInt(100),
		EndingAvgCost:        decimal.NewFromFloat(1500),
		LatestPrice:          decimal.NewFromFloat(1580),
		LatestMarketValue:    decimal.NewFromInt(158000),
		UnrealizedProfit:     decimal.NewFromInt(5000),
		TotalProfit:          decimal.NewFromInt(13000),
		ChangePercent7D:      decimal.NewFromInt(8),
		PeriodPriceChangePct: decimal.NewFromInt(8),
		MarketDataStatus:     "complete",
		RiskLevel:            "low",
		AnalysisText:         "测试个股分析",
		Recommendation:       "hold",
	}}

	aiService := service.NewAIService(
		NewMockAnalysisTaskRepository(),
		reportRepo,
		itemRepo,
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	detail, err := aiService.GetAnalysisReportDetail(1, 1)
	if err != nil {
		t.Fatalf("GetAnalysisReportDetail() error = %v", err)
	}
	if detail.Prediction == nil {
		t.Fatal("expected structured prediction")
	}
	if detail.Prediction.Bias != "bullish" {
		t.Fatalf("expected bullish bias, got %s", detail.Prediction.Bias)
	}
	if detail.Prediction.Confidence != "high" {
		t.Fatalf("expected high confidence, got %s", detail.Prediction.Confidence)
	}
	if detail.Prediction.Horizon != "next_7d" {
		t.Fatalf("expected next_7d horizon, got %s", detail.Prediction.Horizon)
	}
	if len(detail.Prediction.Drivers) != 3 {
		t.Fatalf("expected 3 drivers after normalization, got %d", len(detail.Prediction.Drivers))
	}
	if detail.Prediction.Drivers[0] != "600519.SH 周期涨跌幅 8.00%" {
		t.Fatalf("unexpected first driver: %s", detail.Prediction.Drivers[0])
	}
	if len(detail.Prediction.Scenarios) != 2 {
		t.Fatalf("expected 2 scenarios after normalization, got %d", len(detail.Prediction.Scenarios))
	}
	if detail.Prediction.Scenarios[0].Condition != "若龙头标的继续强势" {
		t.Fatalf("unexpected first scenario condition: %s", detail.Prediction.Scenarios[0].Condition)
	}
	if detail.Prediction.Scenarios[1].Condition != "若集中仓位回撤" {
		t.Fatalf("unexpected second scenario condition: %s", detail.Prediction.Scenarios[1].Condition)
	}
	if detail.Prediction.Narrative != predictionText {
		t.Fatalf("expected narrative fallback to prediction_text, got %s", detail.Prediction.Narrative)
	}
}

func TestAIService_GetAnalysisReportDetail_IgnoresWeakStructuredPredictionScenarios(t *testing.T) {
	reportRepo := NewMockAnalysisReportRepository()
	predictionText := "若白酒仓位止跌且新能源强势延续，组合收益有机会继续修复。"
	rawOutput := `{"summary":{"report_title":"预测报告","summary_text":"组合当前由盈利股和亏损股共同驱动。","risk_level":"medium","investment_style":"balanced","risk_analysis":"亏损仓位仍会拖累净值。","pattern_insights":"仓位表现分化明显。","prediction_text":"若白酒仓位止跌且新能源强势延续，组合收益有机会继续修复。","prediction":{"bias":"neutral","confidence":"medium","horizon":"next_30d","drivers":["宁德时代近7日涨幅4.60%","五粮液近7日跌幅3.20%"],"scenarios":[{"condition":"市场行情向好，盈利股票继续上涨","outcome":"投资组合实现盈利"},{"condition":"市场行情走弱，亏损股票进一步下跌","outcome":"投资组合亏损扩大"}]},"recommendations":["控制回撤"]},"stocks":[{"symbol":"300750.SZ","asset_name":"宁德时代","risk_level":"medium","investment_style":"growth","analysis_text":"测试分析","recommendation":"hold","key_points":["测试"]}]}`
	reportRepo.Create(&model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "预测报告",
		AnalysisPeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		AnalysisPeriodEnd:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		RiskLevel:           "medium",
		SummaryText:         "测试总结",
		PredictionText:      &predictionText,
		RawAIOutput:         &rawOutput,
		AIModel:             "test-model",
	})

	itemRepo := NewMockAnalysisReportItemRepository()
	itemRepo.ItemsByReportID[1] = []model.AnalysisReportItem{
		{
			ID:                   1,
			ReportID:             1,
			UserID:               1,
			Symbol:               "300750.SZ",
			AssetName:            "宁德时代",
			TradeCount:           1,
			BuyCount:             1,
			SellCount:            0,
			BuyAmount:            decimal.NewFromInt(50000),
			SellAmount:           decimal.Zero,
			NetQuantity:          decimal.NewFromInt(100),
			RealizedProfit:       decimal.Zero,
			RealizedProfitRate:   decimal.Zero,
			EndingPositionQty:    decimal.NewFromInt(100),
			EndingAvgCost:        decimal.NewFromFloat(200),
			LatestPrice:          decimal.NewFromFloat(220),
			LatestMarketValue:    decimal.NewFromInt(22000),
			UnrealizedProfit:     decimal.NewFromInt(2000),
			TotalProfit:          decimal.NewFromInt(2000),
			ChangePercent7D:      decimal.NewFromFloat(4.6),
			PeriodPriceChangePct: decimal.NewFromFloat(4.6),
			MarketDataStatus:     "complete",
			RiskLevel:            "medium",
			AnalysisText:         "测试个股分析",
			Recommendation:       "hold",
		},
		{
			ID:                   2,
			ReportID:             1,
			UserID:               1,
			Symbol:               "000858.SZ",
			AssetName:            "五粮液",
			TradeCount:           1,
			BuyCount:             1,
			SellCount:            0,
			BuyAmount:            decimal.NewFromInt(30000),
			SellAmount:           decimal.Zero,
			NetQuantity:          decimal.NewFromInt(100),
			RealizedProfit:       decimal.Zero,
			RealizedProfitRate:   decimal.Zero,
			EndingPositionQty:    decimal.NewFromInt(100),
			EndingAvgCost:        decimal.NewFromFloat(160),
			LatestPrice:          decimal.NewFromFloat(145),
			LatestMarketValue:    decimal.NewFromInt(14500),
			UnrealizedProfit:     decimal.NewFromInt(-1500),
			TotalProfit:          decimal.NewFromInt(-1500),
			ChangePercent7D:      decimal.NewFromFloat(-3.2),
			PeriodPriceChangePct: decimal.NewFromFloat(-3.2),
			MarketDataStatus:     "complete",
			RiskLevel:            "medium",
			AnalysisText:         "测试个股分析",
			Recommendation:       "observe",
		},
	}

	aiService := service.NewAIService(
		NewMockAnalysisTaskRepository(),
		reportRepo,
		itemRepo,
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	detail, err := aiService.GetAnalysisReportDetail(1, 1)
	if err != nil {
		t.Fatalf("GetAnalysisReportDetail() error = %v", err)
	}
	if detail.Prediction == nil {
		t.Fatal("expected fallback prediction")
	}
	if len(detail.Prediction.Scenarios) < 2 {
		t.Fatalf("expected item-based fallback scenarios, got %d", len(detail.Prediction.Scenarios))
	}
	if strings.Contains(detail.Prediction.Scenarios[0].Condition, "市场行情向好") {
		t.Fatalf("expected weak structured scenario to be ignored, got %s", detail.Prediction.Scenarios[0].Condition)
	}
	if !strings.Contains(detail.Prediction.Scenarios[0].Condition, "宁德时代") {
		t.Fatalf("expected fallback scenario to mention concrete holding, got %s", detail.Prediction.Scenarios[0].Condition)
	}
}

func TestAIService_GetAnalysisReportDetail_FallsBackToPredictionText(t *testing.T) {
	reportRepo := NewMockAnalysisReportRepository()
	predictionText := "若继续维持当前交易结构，强势标的可能继续贡献收益，但弱势仓位会放大组合回撤风险。"
	reportRepo.Create(&model.AnalysisReport{
		UserID:              1,
		ReportType:          "summary",
		ReportTitle:         "历史预测报告",
		AnalysisPeriodStart: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		AnalysisPeriodEnd:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		RiskLevel:           "medium",
		SummaryText:         "测试总结",
		PredictionText:      &predictionText,
		AIModel:             "test-model",
	})

	itemRepo := NewMockAnalysisReportItemRepository()
	itemRepo.ItemsByReportID[1] = []model.AnalysisReportItem{
		{
			ID:                   1,
			ReportID:             1,
			UserID:               1,
			Symbol:               "600519.SH",
			AssetName:            "贵州茅台",
			TradeCount:           1,
			BuyCount:             1,
			SellCount:            0,
			BuyAmount:            decimal.NewFromInt(100000),
			SellAmount:           decimal.Zero,
			NetQuantity:          decimal.NewFromInt(100),
			RealizedProfit:       decimal.Zero,
			RealizedProfitRate:   decimal.Zero,
			EndingPositionQty:    decimal.NewFromInt(100),
			EndingAvgCost:        decimal.NewFromFloat(1500),
			LatestPrice:          decimal.NewFromFloat(1580),
			LatestMarketValue:    decimal.NewFromInt(158000),
			UnrealizedProfit:     decimal.NewFromInt(5000),
			TotalProfit:          decimal.NewFromInt(5000),
			ChangePercent7D:      decimal.NewFromInt(6),
			PeriodPriceChangePct: decimal.NewFromInt(6),
			MarketDataStatus:     "complete",
			RiskLevel:            "low",
			AnalysisText:         "测试个股分析",
			Recommendation:       "hold",
		},
		{
			ID:                   2,
			ReportID:             1,
			UserID:               1,
			Symbol:               "000001.SZ",
			AssetName:            "平安银行",
			TradeCount:           2,
			BuyCount:             1,
			SellCount:            1,
			BuyAmount:            decimal.NewFromInt(20000),
			SellAmount:           decimal.NewFromInt(19000),
			NetQuantity:          decimal.Zero,
			RealizedProfit:       decimal.NewFromInt(-1000),
			RealizedProfitRate:   decimal.NewFromInt(-5),
			EndingPositionQty:    decimal.Zero,
			EndingAvgCost:        decimal.Zero,
			LatestPrice:          decimal.NewFromFloat(10),
			LatestMarketValue:    decimal.Zero,
			UnrealizedProfit:     decimal.Zero,
			TotalProfit:          decimal.NewFromInt(-1000),
			ChangePercent7D:      decimal.NewFromInt(-3),
			PeriodPriceChangePct: decimal.NewFromInt(-3),
			MarketDataStatus:     "complete",
			RiskLevel:            "medium",
			AnalysisText:         "测试个股分析2",
			Recommendation:       "observe",
		},
	}

	aiService := service.NewAIService(
		NewMockAnalysisTaskRepository(),
		reportRepo,
		itemRepo,
		&MockTransactionRepositoryForAI{},
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	detail, err := aiService.GetAnalysisReportDetail(1, 1)
	if err != nil {
		t.Fatalf("GetAnalysisReportDetail() error = %v", err)
	}
	if detail.Prediction == nil {
		t.Fatal("expected fallback prediction")
	}
	if detail.Prediction.Bias != "neutral" {
		t.Fatalf("expected neutral bias, got %s", detail.Prediction.Bias)
	}
	if detail.Prediction.Confidence != "medium" {
		t.Fatalf("expected medium confidence, got %s", detail.Prediction.Confidence)
	}
	if detail.Prediction.Horizon != "next_7d" {
		t.Fatalf("expected next_7d horizon, got %s", detail.Prediction.Horizon)
	}
	if len(detail.Prediction.Drivers) != 2 {
		t.Fatalf("expected 2 fallback drivers, got %d", len(detail.Prediction.Drivers))
	}
	if !strings.Contains(detail.Prediction.Drivers[0], "贵州茅台") {
		t.Fatalf("unexpected first driver: %s", detail.Prediction.Drivers[0])
	}
	if len(detail.Prediction.Scenarios) != 2 {
		t.Fatalf("expected 2 fallback scenarios, got %d", len(detail.Prediction.Scenarios))
	}
	if !strings.Contains(detail.Prediction.Scenarios[0].Condition, "贵州茅台") {
		t.Fatalf("unexpected fallback scenario condition: %s", detail.Prediction.Scenarios[0].Condition)
	}
	if !strings.Contains(detail.Prediction.Scenarios[1].Condition, "平安银行") {
		t.Fatalf("unexpected second fallback scenario condition: %s", detail.Prediction.Scenarios[1].Condition)
	}
	if detail.Prediction.Narrative != predictionText {
		t.Fatalf("unexpected fallback narrative: %s", detail.Prediction.Narrative)
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
				AssetType:       "stock",
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
				RealizedProfit:   decimal.NewFromFloat(5000),
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
	if !strings.Contains(report.ChartData, `"metric":"realized_profit"`) {
		t.Fatalf("expected realized_profit chart_data, got %s", report.ChartData)
	}
	if report.RiskAnalysis != "持仓集中在单一标的，若价格回撤会放大波动。" {
		t.Errorf("Expected structured risk analysis, got %s", report.RiskAnalysis)
	}
}

func TestAIService_GenerateInvestmentSummary_IgnoresNonStockAssets(t *testing.T) {
	txRepo := &MockTransactionRepositoryForAI{
		Transactions: []model.Transaction{
			{
				UserID:          1,
				AssetType:       "stock",
				AssetCode:       "600519",
				AssetName:       "贵州茅台",
				TransactionType: "buy",
				Quantity:        decimal.NewFromInt(100),
				TotalAmount:     decimal.NewFromFloat(185000),
			},
			{
				UserID:          1,
				AssetType:       "fund",
				AssetCode:       "110011",
				AssetName:       "易方达中小盘",
				TransactionType: "buy",
				Quantity:        decimal.NewFromInt(50),
				TotalAmount:     decimal.NewFromFloat(5000),
			},
		},
	}

	llmProvider := &MockLLMProvider{
		Content:   `{"summary":{"report_title":"测试总结报告","summary_text":"组合总盈亏为正，贵州茅台贡献了主要收益。","risk_level":"medium","investment_style":"balanced","risk_analysis":"持仓集中在单一标的。","pattern_insights":"交易集中在少数标的。","prediction_text":"若继续维持当前集中仓位，波动可能放大。","recommendations":["控制仓位","关注回撤"]},"stocks":[{"symbol":"600519.SH","asset_name":"贵州茅台","risk_level":"low","investment_style":"value","analysis_text":"贵州茅台当前总盈亏为正。","recommendation":"hold","key_points":["龙头白酒"]}]}`,
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
				RealizedProfit:   decimal.NewFromFloat(5000),
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

	_, err := aiService.GenerateInvestmentSummary(1, "2024-01-01", "2024-12-31")
	if err != nil {
		t.Fatalf("GenerateInvestmentSummary() error = %v", err)
	}
	if strings.Contains(llmProvider.UserPrompt, "易方达中小盘") {
		t.Fatalf("expected non-stock asset to be excluded from prompt, got %s", llmProvider.UserPrompt)
	}
	if !strings.Contains(llmProvider.UserPrompt, "贵州茅台") {
		t.Fatalf("expected stock asset to remain in prompt, got %s", llmProvider.UserPrompt)
	}
}

// TestAIService_GenerateInvestmentSummary_NoMetrics 测试无指标数据
func TestAIService_GenerateInvestmentSummary_NoMetrics(t *testing.T) {
	txRepo := &MockTransactionRepositoryForAI{
		Transactions: []model.Transaction{
			{
				UserID:          1,
				AssetType:       "stock",
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
		AssetType:       "stock",
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
		RealizedProfit:   decimal.NewFromInt(5000),
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

func TestAIService_CreateStockAnalysisTask_OnlyNonStockAssetsFails(t *testing.T) {
	txRepo := &MockTransactionRepositoryForAI{Transactions: []model.Transaction{{
		UserID:          1,
		AssetType:       "fund",
		AssetCode:       "110011",
		AssetName:       "易方达中小盘",
		TransactionType: "buy",
		Quantity:        decimal.NewFromInt(50),
		TotalAmount:     decimal.NewFromFloat(5000),
	}}}
	taskRepo := NewMockAnalysisTaskRepository()
	reportRepo := NewMockAnalysisReportRepository()
	aiService := service.NewAIService(
		taskRepo,
		reportRepo,
		NewMockAnalysisReportItemRepository(),
		txRepo,
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
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
		t.Fatal("expected no report to be persisted")
	}
}

func TestAIService_CreateStockAnalysisTask_SymbolFilterDoesNotMatchNonStockAsset(t *testing.T) {
	txRepo := &MockTransactionRepositoryForAI{Transactions: []model.Transaction{{
		UserID:          1,
		AssetType:       "fund",
		AssetCode:       "600519",
		AssetName:       "基金占位",
		TransactionType: "buy",
		Quantity:        decimal.NewFromInt(10),
		TotalAmount:     decimal.NewFromFloat(1000),
	}}}
	taskRepo := NewMockAnalysisTaskRepository()
	reportRepo := NewMockAnalysisReportRepository()
	aiService := service.NewAIService(
		taskRepo,
		reportRepo,
		NewMockAnalysisReportItemRepository(),
		txRepo,
		&MockStockMetricService{},
		&MockLLMProvider{modelName: "test-model"},
		zap.NewNop(),
	)

	task, err := aiService.CreateStockAnalysisTask(1, &request.CreateAnalysisTaskRequest{StartDate: "2024-01-01", EndDate: "2024-12-31", Symbols: []string{"600519.SH"}})
	if err != nil {
		t.Fatalf("CreateStockAnalysisTask() error = %v", err)
	}
	finished := waitForTaskStatus(t, taskRepo, task.ID)
	if finished.Status != "failed" {
		t.Fatalf("expected failed task, got %s", finished.Status)
	}
	if reportRepo.LastReport != nil {
		t.Fatal("expected no report to be persisted")
	}
}

func TestAIService_CreateStockAnalysisTask_CleansAndPersistsOutput(t *testing.T) {
	txRepo := &MockTransactionRepositoryForAI{Transactions: []model.Transaction{{
		UserID:          1,
		AssetType:       "stock",
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
		RealizedProfit:   decimal.NewFromInt(5000),
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
	if !reportRepo.LastReport.TotalProfit.Equal(decimal.NewFromInt(5000)) {
		t.Fatalf("expected report total_profit to use realized profit, got %s", reportRepo.LastReport.TotalProfit.String())
	}
	if !reportRepo.LastItems[0].TotalProfit.Equal(decimal.NewFromInt(5000)) {
		t.Fatalf("expected item total_profit to use realized profit, got %s", reportRepo.LastItems[0].TotalProfit.String())
	}
	if reportRepo.LastReport.ChartData == nil || !strings.Contains(*reportRepo.LastReport.ChartData, `"metric":"realized_profit"`) {
		t.Fatalf("expected realized_profit chart_data, got %+v", reportRepo.LastReport.ChartData)
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
		AssetType:       "stock",
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
		RealizedProfit:   decimal.NewFromInt(5000),
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
	if !strings.Contains(reportRepo.LastItems[0].AnalysisText, "已实现盈亏") {
		t.Fatalf("expected transparent fallback analysis text, got %s", reportRepo.LastItems[0].AnalysisText)
	}
}
