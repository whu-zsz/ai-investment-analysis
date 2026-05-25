package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	requestdto "stock-analysis-backend/internal/dto/request"
	responsedto "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/pkg/llm"

	"github.com/shopspring/decimal"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	analysisTaskTypeStock = "stock_analysis"

	analysisStatusPending    = "pending"
	analysisStatusProcessing = "processing"
	analysisStatusSuccess    = "success"
	analysisStatusFailed     = "failed"

	analysisStagePending             = "pending"
	analysisStageCollectTransactions = "collecting_transactions"
	analysisStagePreparingMetrics    = "preparing_metrics"
	analysisStageGeneratingStocks    = "generating_stock_reports"
	analysisStageGeneratingSummary   = "generating_summary"
	analysisStagePersisting          = "persisting_report"
	analysisStageCompleted           = "completed"

	marketDataStatusComplete    = "complete"
	marketDataStatusFetchedLive = "fetched_live"
	marketDataStatusPartial     = "partial"
	marketDataStatusUnavailable = "unavailable"
)

type AIService interface {
	GenerateInvestmentSummary(userID uint64, startDate, endDate string) (*responsedto.AnalysisReportResponse, error)
	GetReports(userID uint64, reportType string, limit int) ([]responsedto.AnalysisReportResponse, error)
	CreateStockAnalysisTask(userID uint64, req *requestdto.CreateAnalysisTaskRequest) (*responsedto.AnalysisTaskResponse, error)
	GetAnalysisTasks(userID uint64, status string, page, pageSize int) (*responsedto.AnalysisTaskListResponse, error)
	GetAnalysisTask(userID, taskID uint64) (*responsedto.AnalysisTaskDetailResponse, error)
	GetAnalysisReportDetail(userID, reportID uint64) (*responsedto.AnalysisReportDetailResponse, error)
	ExportAnalysisReportPDF(userID, reportID uint64) ([]byte, string, error)
}

type aiService struct {
	analysisTaskRepo       repository.AnalysisTaskRepository
	analysisReportRepo     repository.AnalysisReportRepository
	analysisReportItemRepo repository.AnalysisReportItemRepository
	transactionRepo        repository.TransactionRepository
	stockMetricService     StockAnalysisMetricService
	llmProvider            llm.Provider
	pdfRenderer            PDFRenderer
	logger                 *zap.Logger
}

type stockAggregate struct {
	Symbol           string
	AssetName        string
	TradeCount       int
	BuyCount         int
	SellCount        int
	BuyAmount        decimal.Decimal
	SellAmount       decimal.Decimal
	NetQuantity      decimal.Decimal
	RealizedProfit   decimal.Decimal
	LatestPrice      decimal.Decimal
	ChangePercent7D  decimal.Decimal
	Market           string
	MarketDataStatus string
	MarketSnapshots  []model.MarketSnapshot
}

type riskSymbolSummary struct {
	Symbol         string
	AssetName      string
	RiskLevel      string
	RiskScore      int
	TriggerReasons []string
}

type aiPredictionScenario struct {
	Condition string `json:"condition"`
	Outcome   string `json:"outcome"`
}

type aiPredictionOutput struct {
	Bias       string                 `json:"bias"`
	Confidence string                 `json:"confidence"`
	Horizon    string                 `json:"horizon"`
	Drivers    []string               `json:"drivers"`
	Scenarios  []aiPredictionScenario `json:"scenarios"`
}

type aiSummaryOutput struct {
	ReportTitle     string             `json:"report_title"`
	SummaryText     string             `json:"summary_text"`
	RiskLevel       string             `json:"risk_level"`
	InvestmentStyle string             `json:"investment_style"`
	RiskAnalysis    string             `json:"risk_analysis"`
	PatternInsights string             `json:"pattern_insights"`
	PredictionText  string             `json:"prediction_text"`
	Prediction      aiPredictionOutput `json:"prediction"`
	Recommendations []string           `json:"recommendations"`
}

type aiStockOutput struct {
	Symbol          string   `json:"symbol"`
	AssetName       string   `json:"asset_name"`
	RiskLevel       string   `json:"risk_level"`
	InvestmentStyle string   `json:"investment_style"`
	AnalysisText    string   `json:"analysis_text"`
	Recommendation  string   `json:"recommendation"`
	KeyPoints       []string `json:"key_points"`
}

type chartPoint struct {
	Symbol string `json:"symbol"`
	Value  string `json:"value"`
}

type chartDataEnvelope struct {
	Version int          `json:"version"`
	Kind    string       `json:"kind"`
	Metric  string       `json:"metric"`
	Points  []chartPoint `json:"points"`
}

type aiAnalysisOutput struct {
	Summary aiSummaryOutput `json:"summary"`
	Stocks  []aiStockOutput `json:"stocks"`
}

func normalizeChartData(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	var envelope chartDataEnvelope
	if err := json.Unmarshal([]byte(trimmed), &envelope); err == nil && envelope.Kind == "profit_by_symbol" {
		if envelope.Version == 0 {
			envelope.Version = 2
		}
		if strings.TrimSpace(envelope.Metric) == "" {
			envelope.Metric = "total_profit"
		}
		data, err := json.Marshal(envelope)
		if err == nil {
			return string(data)
		}
	}

	var legacyPoints []chartPoint
	if err := json.Unmarshal([]byte(trimmed), &legacyPoints); err == nil {
		wrapped := chartDataEnvelope{
			Version: 2,
			Kind:    "profit_by_symbol",
			Metric:  "total_profit",
			Points:  legacyPoints,
		}
		data, err := json.Marshal(wrapped)
		if err == nil {
			return string(data)
		}
	}

	return ""
}


func NewAIService(
	analysisTaskRepo repository.AnalysisTaskRepository,
	analysisReportRepo repository.AnalysisReportRepository,
	analysisReportItemRepo repository.AnalysisReportItemRepository,
	transactionRepo repository.TransactionRepository,
	stockMetricService StockAnalysisMetricService,
	llmProvider llm.Provider,
	logger *zap.Logger,
) AIService {
	return NewAIServiceWithPDFRenderer(
		analysisTaskRepo,
		analysisReportRepo,
		analysisReportItemRepo,
		transactionRepo,
		stockMetricService,
		llmProvider,
		NewChromedpPDFRenderer(),
		logger,
	)
}

func NewAIServiceWithPDFRenderer(
	analysisTaskRepo repository.AnalysisTaskRepository,
	analysisReportRepo repository.AnalysisReportRepository,
	analysisReportItemRepo repository.AnalysisReportItemRepository,
	transactionRepo repository.TransactionRepository,
	stockMetricService StockAnalysisMetricService,
	llmProvider llm.Provider,
	pdfRenderer PDFRenderer,
	logger *zap.Logger,
) AIService {
	if pdfRenderer == nil {
		pdfRenderer = NewChromedpPDFRenderer()
	}
	return &aiService{
		analysisTaskRepo:       analysisTaskRepo,
		analysisReportRepo:     analysisReportRepo,
		analysisReportItemRepo: analysisReportItemRepo,
		transactionRepo:        transactionRepo,
		stockMetricService:     stockMetricService,
		llmProvider:            llmProvider,
		pdfRenderer:            pdfRenderer,
		logger:                 logger,
	}
}

func (s *aiService) GenerateInvestmentSummary(userID uint64, startDate, endDate string) (*responsedto.AnalysisReportResponse, error) {
	startTime, endTime, err := validateAnalysisRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	transactions, err := s.transactionRepo.FindByDateRange(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	transactions = filterStockTransactions(transactions)
	if len(transactions) == 0 {
		return nil, fmt.Errorf("no stock transactions found in the specified period")
	}

	metrics, err := s.stockMetricService.PrepareMetrics(context.Background(), userID, nil, startTime, endTime, nil, false, false)
	if err != nil {
		return nil, err
	}
	if len(metrics) == 0 {
		return nil, fmt.Errorf("no stock metrics found in the specified period")
	}

	output, rawOutput, err := s.generateStructuredAnalysis(startTime, endTime, metrics, transactions)
	if err != nil {
		return nil, err
	}

	report, err := buildSummaryReport(userID, startTime, endTime, rawOutput, output, metrics, s.llmProvider.ModelName())
	if err != nil {
		return nil, err
	}

	if err := s.analysisReportRepo.Create(report); err != nil {
		return nil, err
	}
	return s.convertToResponse(report), nil
}

func (s *aiService) GetReports(userID uint64, reportType string, limit int) ([]responsedto.AnalysisReportResponse, error) {
	if limit <= 0 {
		limit = 10
	}
	reports, err := s.analysisReportRepo.FindByUserID(userID, reportType, limit)
	if err != nil {
		return nil, err
	}
	results := make([]responsedto.AnalysisReportResponse, 0, len(reports))
	for _, report := range reports {
		results = append(results, *s.convertToResponse(&report))
	}
	return results, nil
}

func (s *aiService) CreateStockAnalysisTask(userID uint64, req *requestdto.CreateAnalysisTaskRequest) (*responsedto.AnalysisTaskResponse, error) {
	startTime, endTime, err := validateAnalysisRange(req.StartDate, req.EndDate)
	if err != nil {
		return nil, err
	}

	running, err := s.analysisTaskRepo.HasRunningTask(userID, analysisTaskTypeStock)
	if err != nil {
		return nil, err
	}
	if running {
		return nil, fmt.Errorf("analysis task is already running")
	}

	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	payload := string(payloadBytes)
	task := &model.AnalysisTask{
		UserID:              userID,
		TaskType:            analysisTaskTypeStock,
		Status:              analysisStatusPending,
		ProgressStage:       analysisStagePending,
		AnalysisPeriodStart: startTime,
		AnalysisPeriodEnd:   endTime,
		RequestPayload:      &payload,
	}
	if err := s.analysisTaskRepo.Create(task); err != nil {
		return nil, err
	}

	go s.runAnalysisTask(task.ID, userID, req, startTime, endTime)

	return &responsedto.AnalysisTaskResponse{
		ID:            task.ID,
		Status:        task.Status,
		ProgressStage: task.ProgressStage,
		CreatedAt:     task.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *aiService) GetAnalysisTasks(userID uint64, status string, page, pageSize int) (*responsedto.AnalysisTaskListResponse, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	tasks, total, err := s.analysisTaskRepo.FindByUserID(userID, status, pageSize, offset)
	if err != nil {
		return nil, err
	}
	items := make([]responsedto.AnalysisTaskDetailResponse, 0, len(tasks))
	for _, task := range tasks {
		items = append(items, s.convertTaskToDetail(&task))
	}
	return &responsedto.AnalysisTaskListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

func (s *aiService) GetAnalysisTask(userID, taskID uint64) (*responsedto.AnalysisTaskDetailResponse, error) {
	task, err := s.analysisTaskRepo.FindByIDAndUserID(taskID, userID)
	if err != nil {
		return nil, err
	}
	result := s.convertTaskToDetail(task)
	if task.ResultReportID != nil {
		result.ResultReportID = *task.ResultReportID
	}
	return &result, nil
}

func (s *aiService) GetAnalysisReportDetail(userID, reportID uint64) (*responsedto.AnalysisReportDetailResponse, error) {
	report, err := s.analysisReportRepo.FindByIDAndUserID(reportID, userID)
	if err != nil {
		return nil, err
	}
	items, err := s.analysisReportItemRepo.FindByReportID(report.ID)
	if err != nil {
		return nil, err
	}
	return s.convertToDetailResponse(report, items), nil
}

func (s *aiService) ExportAnalysisReportPDF(userID, reportID uint64) ([]byte, string, error) {
	detail, err := s.GetAnalysisReportDetail(userID, reportID)
	if err != nil {
		return nil, "", err
	}

	html, err := renderAnalysisReportPDFHTML(detail)
	if err != nil {
		return nil, "", err
	}

	pdfBytes, err := s.pdfRenderer.RenderHTMLToPDF(context.Background(), html)
	if err != nil {
		return nil, "", err
	}

	return pdfBytes, buildAnalysisReportPDFFilename(detail), nil
}

func (s *aiService) runAnalysisTask(taskID, userID uint64, req *requestdto.CreateAnalysisTaskRequest, startTime, endTime time.Time) {
	startedAt := time.Now()
	_ = s.analysisTaskRepo.UpdateProgress(taskID, analysisStatusProcessing, analysisStageCollectTransactions, nil, nil, &startedAt, nil)

	if err := s.executeAnalysisTask(taskID, userID, req, startTime, endTime); err != nil {
		finishedAt := time.Now()
		msg := err.Error()
		_ = s.analysisTaskRepo.UpdateProgress(taskID, analysisStatusFailed, analysisStageCompleted, &msg, nil, nil, &finishedAt)
		if s.logger != nil {
			s.logger.Warn("analysis task failed", zap.Uint64("task_id", taskID), zap.Error(err))
		}
	}
}

func (s *aiService) executeAnalysisTask(taskID, userID uint64, req *requestdto.CreateAnalysisTaskRequest, startTime, endTime time.Time) error {
	transactions, err := s.transactionRepo.FindByDateRange(userID, req.StartDate, req.EndDate)
	if err != nil {
		return err
	}
	transactions = filterStockTransactions(transactions)
	if len(transactions) == 0 {
		return fmt.Errorf("no stock transactions found in the specified period")
	}

	normalizedSymbols := normalizeSymbols(req.Symbols)
	if len(normalizedSymbols) > 0 {
		filter := make(map[string]struct{}, len(normalizedSymbols))
		for _, symbol := range normalizedSymbols {
			filter[symbol] = struct{}{}
		}
		hasMatched := false
		for _, tx := range transactions {
			if _, ok := filter[normalizeSymbol(tx.AssetCode)]; ok {
				hasMatched = true
				break
			}
		}
		if !hasMatched {
			return fmt.Errorf("no stock transactions found for the specified symbols in the specified period")
		}
	}

	if err := s.analysisTaskRepo.UpdateProgress(taskID, analysisStatusProcessing, analysisStagePreparingMetrics, nil, nil, nil, nil); err != nil {
		return err
	}
	metrics, err := s.stockMetricService.PrepareMetrics(context.Background(), userID, &taskID, startTime, endTime, req.Symbols, req.ForceRefreshMarket, req.ForceRefreshMetrics)
	if err != nil {
		return err
	}
	if len(metrics) == 0 {
		return fmt.Errorf("no stock metrics found in the specified period")
	}

	if err := s.analysisTaskRepo.UpdateProgress(taskID, analysisStatusProcessing, analysisStageGeneratingStocks, nil, nil, nil, nil); err != nil {
		return err
	}
	output, rawOutput, err := s.generateStructuredAnalysis(startTime, endTime, metrics, transactions)
	if err != nil {
		return err
	}

	if err := s.analysisTaskRepo.UpdateProgress(taskID, analysisStatusProcessing, analysisStageGeneratingSummary, nil, nil, nil, nil); err != nil {
		return err
	}
	stockOutputMap := make(map[string]aiStockOutput, len(output.Stocks))
	for _, item := range output.Stocks {
		stockOutputMap[strings.ToUpper(strings.TrimSpace(item.Symbol))] = item
	}

	report, items := buildReportModels(taskID, userID, startTime, endTime, rawOutput, output, metrics, stockOutputMap, s.llmProvider.ModelName())

	if err := s.analysisTaskRepo.UpdateProgress(taskID, analysisStatusProcessing, analysisStagePersisting, nil, nil, nil, nil); err != nil {
		return err
	}
	if err := s.analysisReportRepo.CreateWithItems(report, items); err != nil {
		return err
	}

	finishedAt := time.Now()
	if err := s.analysisTaskRepo.UpdateProgress(taskID, analysisStatusSuccess, analysisStageCompleted, nil, &report.ID, nil, &finishedAt); err != nil {
		return err
	}
	return nil
}

func (s *aiService) generateStructuredAnalysis(startTime, endTime time.Time, metrics []model.StockAnalysisMetric, transactions []model.Transaction) (*aiAnalysisOutput, string, error) {
	systemPrompt := `你是一位专业的股票交易分析助手。你只能输出一个合法 JSON 对象，禁止输出 markdown、代码块、解释性文字。JSON 顶层必须包含 summary 和 stocks 两个字段。summary 中 risk_level 只能是 low、medium、high。stocks 中 recommendation 只能是 buy、hold、reduce、sell、observe。`
	userPrompt := s.buildStructuredPrompt(startTime, endTime, metrics, transactions)
	content, err := s.llmProvider.GetContent(context.Background(), systemPrompt, userPrompt)
	if err != nil {
		return nil, "", err
	}
	parsed, err := parseAIAnalysisOutput(content)
	if err != nil {
		return nil, content, err
	}
	sanitized := sanitizeAIAnalysisOutput(parsed, metrics)
	if err := validateAIAnalysisOutput(sanitized, metrics); err != nil {
		return nil, content, err
	}
	return sanitized, content, nil
}

func (s *aiService) buildStructuredPrompt(startTime, endTime time.Time, metrics []model.StockAnalysisMetric, transactions []model.Transaction) string {
	totalInvestment := modelDecimalZero()
	totalRealizedProfit := modelDecimalZero()
	lines := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		totalInvestment = totalInvestment.Add(metric.BuyAmount)
		totalRealizedProfit = totalRealizedProfit.Add(metric.RealizedProfit)
		lines = append(lines, fmt.Sprintf("- %s %s: 交易%d次, 买入%d次/卖出%d次, 买入金额%s, 卖出金额%s, 买入股数%s, 卖出股数%s, 净持仓%s, 已实现盈亏%s(%s%%), 期末持仓%s, 持仓均价%s, 最新价%s, 最新市值%s, 当前持仓浮盈/浮亏%s, 参考总盈亏%s, 周期涨跌幅%s%%, 周期最高%s, 周期最低%s, 市场数据状态=%s", metric.Symbol, metric.AssetName, metric.TradeCount, metric.BuyCount, metric.SellCount, metric.BuyAmount.StringFixed(2), metric.SellAmount.StringFixed(2), metric.BuyQuantity.StringFixed(2), metric.SellQuantity.StringFixed(2), metric.NetQuantity.StringFixed(2), metric.RealizedProfit.StringFixed(2), metric.RealizedProfitRate.StringFixed(2), metric.EndingPositionQty.StringFixed(2), metric.EndingAvgCost.StringFixed(4), metric.LatestPrice.StringFixed(4), metric.LatestMarketValue.StringFixed(2), metric.UnrealizedProfit.StringFixed(2), metric.TotalProfit.StringFixed(2), metric.PeriodPriceChangePct.StringFixed(2), metric.PeriodHighPrice.StringFixed(4), metric.PeriodLowPrice.StringFixed(4), metric.MarketDataStatus))
	}
	return fmt.Sprintf(`请基于以下股票投资分析指标生成一份结构化分析报告。
分析周期：%s 至 %s
总交易数：%d
股票数：%d
总买入金额：%s
累计已实现盈亏：%s

硬性要求：
- 只能输出一个合法 JSON 对象，禁止 markdown、代码块、解释文字。
- 不要使用空泛套话，例如“市场有风险”“建议谨慎观察”“结合自身情况”。
- 所有叙述都必须直接引用输入里的真实指标，不要编造外部新闻、宏观判断或未提供的数据。
- summary_text、risk_analysis、pattern_insights、prediction_text、analysis_text 不能只写一句空话，必须包含具体指标。
- recommendations 和 key_points 必须是可执行、可核对的短句；recommendations 每条尽量写成完整动作句，key_points 每条必须带出明确事实或指标依据，去重后仍要保留有效信息。

字段写作规则：
1. summary.report_title：具体、简短，要体现分析周期或结果倾向。
2. summary.summary_text：按“现状 + 指标依据 + 总体判断”组织，2-3 句或分句，至少提到 2 个真实指标，并给出总体判断。
3. summary.risk_level：只能是 low、medium、high。
4. summary.investment_style：只能输出 conservative、balanced、aggressive 三者之一。
5. summary.risk_analysis：按“风险点 + 指标依据/触发条件 + 可能影响”组织，2-3 句或分句，必须能对应到持仓、盈亏、交易频率或数据状态。
6. summary.pattern_insights：按“行为模式 + 证据”组织，2-3 句或分句，总结交易行为模式，不要重复 summary_text。
7. summary.prediction_text：按“条件/场景 + 可能结果”组织，用条件式或场景式表达，描述后续可能走势。
8. summary.prediction：输出结构化预测对象，必须包含 bias、confidence、horizon、drivers、scenarios；bias 只能是 bullish、neutral、bearish，confidence 只能是 low、medium、high；drivers 需要 2-3 条直接引用输入指标的驱动因素；scenarios 需要 2 条，每条包含 condition 和 outcome。
9. summary.recommendations：3-5 条，每条一个动作，优先使用“减仓、持有、买入、卖出、观察、复盘、分散、止损”这类动词开头。
10. stocks[].analysis_text：按“现状 + 指标依据 + 结论”组织，每只股票 2-3 句，至少提到 2 个指标。
11. stocks[].recommendation：只能是 buy、hold、reduce、sell、observe。
12. stocks[].key_points：2-4 条短句，只写最关键的事实或结论。

个股指标：
%s

输出 JSON 结构：
{
  "summary": {
    "report_title": "string",
    "summary_text": "string",
    "risk_level": "low|medium|high",
    "investment_style": "string",
    "risk_analysis": "string",
    "pattern_insights": "string",
    "prediction_text": "string",
    "prediction": {
      "bias": "bullish|neutral|bearish",
      "confidence": "low|medium|high",
      "horizon": "next_7d|next_30d",
      "drivers": ["string"],
      "scenarios": [
        {
          "condition": "string",
          "outcome": "string"
        }
      ]
    },
    "recommendations": ["string"]
  },
  "stocks": [
    {
      "symbol": "string",
      "asset_name": "string",
      "risk_level": "low|medium|high",
      "investment_style": "string",
      "analysis_text": "string",
      "recommendation": "buy|hold|reduce|sell|observe",
      "key_points": ["string"]
    }
  ]
}
`, startTime.Format("2006-01-02"), endTime.Format("2006-01-02"), len(transactions), len(metrics), totalInvestment.StringFixed(2), totalRealizedProfit.StringFixed(2), strings.Join(lines, "\n"))
}

func (s *aiService) buildSummaryPrompt(transactions []model.Transaction, startDate, endDate string) string {
	return fmt.Sprintf(`请根据以下投资数据生成投资总结：

投资周期：%s 至 %s
交易次数：%d

主要投资记录：
%s

请提供以下内容：
1. 投资总结
2. 投资风格分析
3. 风险评估
4. 改进建议`, startDate, endDate, len(transactions), s.formatTransactions(transactions))
}

func (s *aiService) formatTransactions(transactions []model.Transaction) string {
	var result string
	for i, t := range transactions {
		if i >= 10 {
			result += fmt.Sprintf("...还有 %d 条记录\n", len(transactions)-10)
			break
		}
		result += fmt.Sprintf("- %s %s %s %s 数量:%s 金额:%s\n",
			t.TransactionDate.Format("2006-01-02"),
			t.TransactionType,
			t.AssetCode,
			t.AssetName,
			t.Quantity.String(),
			t.TotalAmount.String(),
		)
	}
	return result
}

func (s *aiService) convertToResponse(report *model.AnalysisReport) *responsedto.AnalysisReportResponse {
	return &responsedto.AnalysisReportResponse{
		ID:                  report.ID,
		ReportType:          report.ReportType,
		ReportTitle:         report.ReportTitle,
		AnalysisPeriodStart: report.AnalysisPeriodStart.Format("2006-01-02"),
		AnalysisPeriodEnd:   report.AnalysisPeriodEnd.Format("2006-01-02"),
		TotalInvestment:     report.TotalInvestment.String(),
		TotalProfit:         report.TotalProfit.String(),
		ProfitRate:          report.ProfitRate.String(),
		RiskLevel:           report.RiskLevel,
		MarketDataStatus:    report.MarketDataStatus,
		InvestmentStyle:     derefString(report.InvestmentStyle),
		SummaryText:         report.SummaryText,
		RiskAnalysis:        derefString(report.RiskAnalysis),
		PatternInsights:     derefString(report.PatternInsights),
		PredictionText:      derefString(report.PredictionText),
		ChartData:           normalizeChartData(derefString(report.ChartData)),
		Recommendations:     splitJSONOrLines(derefString(report.Recommendations)),
		AIModel:             report.AIModel,
		CreatedAt:           report.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func (s *aiService) convertTaskToDetail(task *model.AnalysisTask) responsedto.AnalysisTaskDetailResponse {
	result := responsedto.AnalysisTaskDetailResponse{
		ID:                  task.ID,
		TaskType:            task.TaskType,
		Status:              task.Status,
		ProgressStage:       task.ProgressStage,
		AnalysisPeriodStart: task.AnalysisPeriodStart.Format("2006-01-02"),
		AnalysisPeriodEnd:   task.AnalysisPeriodEnd.Format("2006-01-02"),
		ErrorMessage:        derefString(task.ErrorMessage),
		CreatedAt:           task.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:           task.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
	if task.ResultReportID != nil {
		result.ResultReportID = *task.ResultReportID
	}
	if task.StartedAt != nil {
		result.StartedAt = task.StartedAt.Format("2006-01-02 15:04:05")
	}
	if task.FinishedAt != nil {
		result.FinishedAt = task.FinishedAt.Format("2006-01-02 15:04:05")
	}
	return result
}

func (s *aiService) convertToDetailResponse(report *model.AnalysisReport, items []model.AnalysisReportItem) *responsedto.AnalysisReportDetailResponse {
	result := &responsedto.AnalysisReportDetailResponse{
		ID:                  report.ID,
		TaskID:              derefUint64(report.TaskID),
		ReportType:          report.ReportType,
		ReportTitle:         report.ReportTitle,
		AnalysisPeriodStart: report.AnalysisPeriodStart.Format("2006-01-02"),
		AnalysisPeriodEnd:   report.AnalysisPeriodEnd.Format("2006-01-02"),
		SymbolsCount:        report.SymbolsCount,
		WinningTrades:       report.WinningTrades,
		LosingTrades:        report.LosingTrades,
		TotalInvestment:     report.TotalInvestment.String(),
		TotalProfit:         report.TotalProfit.String(),
		ProfitRate:          report.ProfitRate.String(),
		RiskLevel:           report.RiskLevel,
		MarketDataStatus:    report.MarketDataStatus,
		InvestmentStyle:     derefString(report.InvestmentStyle),
		SummaryText:         report.SummaryText,
		RiskAnalysis:        derefString(report.RiskAnalysis),
		PatternInsights:     derefString(report.PatternInsights),
		PredictionText:      derefString(report.PredictionText),
		ChartData:           normalizeChartData(derefString(report.ChartData)),
		Recommendations:     splitJSONOrLines(derefString(report.Recommendations)),
		RiskOverview:        responsedto.RiskAnalysisResponse{RiskLevel: normalizeRiskLevel(report.RiskLevel), RiskScore: 0, RiskFactors: []string{}, Recommendations: []string{}},
		RiskAlerts:          []responsedto.RiskAlertItemResponse{},
		TopRiskSymbols:      []responsedto.RiskSymbolResponse{},
		AIModel:             report.AIModel,
		CreatedAt:           report.CreatedAt.Format("2006-01-02 15:04:05"),
		Items:               make([]responsedto.AnalysisReportItemResponse, 0, len(items)),
	}
	for _, item := range items {
		result.Items = append(result.Items, toAnalysisReportItemResponse(item))
	}
	recalculateDetailReportProfit(result)
	result.Prediction = buildPredictionResponse(report, result.Items)
	riskOverview, riskAlerts, topRiskSymbols := buildRiskInsights(result.Items, result.RiskLevel, result.Recommendations)
	result.RiskOverview = riskOverview
	result.RiskAlerts = riskAlerts
	result.TopRiskSymbols = topRiskSymbols
	return result
}

func recalculateDetailReportProfit(detail *responsedto.AnalysisReportDetailResponse) {
	if detail == nil || len(detail.Items) == 0 {
		return
	}

	totalInvestment := decimal.Zero
	totalProfit := decimal.Zero
	winningTrades := 0
	losingTrades := 0
	points := make([]chartPoint, 0, len(detail.Items))

	for _, item := range detail.Items {
		buyAmount := parseDecimalOrZero(item.BuyAmount)
		realizedProfit := parseDecimalOrZero(item.RealizedProfit)
		totalInvestment = totalInvestment.Add(buyAmount)
		totalProfit = totalProfit.Add(realizedProfit)
		if realizedProfit.GreaterThan(decimal.Zero) {
			winningTrades++
		}
		if realizedProfit.LessThan(decimal.Zero) {
			losingTrades++
		}
		if strings.TrimSpace(item.Symbol) != "" {
			points = append(points, chartPoint{Symbol: item.Symbol, Value: realizedProfit.StringFixed(2)})
		}
	}

	detail.TotalInvestment = totalInvestment.StringFixed(2)
	detail.TotalProfit = totalProfit.StringFixed(2)
	detail.WinningTrades = winningTrades
	detail.LosingTrades = losingTrades
	if totalInvestment.IsZero() {
		detail.ProfitRate = decimal.Zero.StringFixed(2)
	} else {
		detail.ProfitRate = totalProfit.Div(totalInvestment).Mul(decimal.NewFromInt(100)).StringFixed(2)
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Symbol < points[j].Symbol })
	if data, err := json.Marshal(chartDataEnvelope{Version: 2, Kind: "profit_by_symbol", Metric: "realized_profit", Points: points}); err == nil {
		detail.ChartData = string(data)
	}
}

func buildPredictionResponse(report *model.AnalysisReport, items []responsedto.AnalysisReportItemResponse) *responsedto.PredictionResponse {
	predictionText := derefString(report.PredictionText)
	rawOutput := strings.TrimSpace(derefString(report.RawAIOutput))
	if rawOutput != "" {
		if parsed, err := parseAIAnalysisOutput(rawOutput); err == nil && hasStructuredPrediction(parsed.Summary.Prediction) {
			return toPredictionResponse(parsed.Summary.Prediction, fallbackString(parsed.Summary.PredictionText, predictionText))
		}
	}
	if strings.TrimSpace(predictionText) == "" {
		return nil
	}
	return buildPredictionFallback(predictionText, items)
}

func hasStructuredPrediction(prediction aiPredictionOutput) bool {
	return normalizePredictionBias(prediction.Bias) != "" || normalizePredictionConfidence(prediction.Confidence) != "" || normalizePredictionHorizon(prediction.Horizon) != "" || len(normalizePredictionDrivers(prediction.Drivers)) > 0 || len(normalizePredictionScenarios(prediction.Scenarios)) > 0
}

func toPredictionResponse(prediction aiPredictionOutput, narrative string) *responsedto.PredictionResponse {
	drivers := normalizePredictionDrivers(prediction.Drivers)
	scenarios := normalizePredictionScenarios(prediction.Scenarios)
	response := &responsedto.PredictionResponse{
		Bias:       normalizePredictionBias(prediction.Bias),
		Confidence: normalizePredictionConfidence(prediction.Confidence),
		Horizon:    normalizePredictionHorizon(prediction.Horizon),
		Drivers:    drivers,
		Scenarios:  make([]responsedto.PredictionScenarioResponse, 0, len(scenarios)),
		Narrative:  normalizeNarrativeText(narrative),
	}
	for _, item := range scenarios {
		response.Scenarios = append(response.Scenarios, responsedto.PredictionScenarioResponse{Condition: item.Condition, Outcome: item.Outcome})
	}
	if response.Bias == "" && response.Confidence == "" && response.Horizon == "" && len(response.Drivers) == 0 && len(response.Scenarios) == 0 && response.Narrative == "" {
		return nil
	}
	return response
}

func buildPredictionFallback(predictionText string, items []responsedto.AnalysisReportItemResponse) *responsedto.PredictionResponse {
	bias := "neutral"
	positiveMomentum := 0
	negativeMomentum := 0
	for _, item := range items {
		changePct := parseDecimalOrZero(item.PeriodPriceChangePct)
		if changePct.GreaterThan(decimal.Zero) {
			positiveMomentum++
		}
		if changePct.LessThan(decimal.Zero) {
			negativeMomentum++
		}
	}
	if positiveMomentum > negativeMomentum {
		bias = "bullish"
		} else if negativeMomentum > positiveMomentum {
		bias = "bearish"
	}
	drivers := []string{}
	for _, item := range items {
		if len(drivers) >= 3 {
			break
		}
		changePct := parseDecimalOrZero(item.PeriodPriceChangePct)
		if !changePct.IsZero() {
			drivers = append(drivers, fmt.Sprintf("%s 周期涨跌幅 %s%%", item.Symbol, changePct.StringFixed(2)))
		}
	}
	return &responsedto.PredictionResponse{
		Bias:       bias,
		Confidence: "medium",
		Horizon:    "next_7d",
		Drivers:    drivers,
		Scenarios: []responsedto.PredictionScenarioResponse{{
			Condition: "若维持当前交易结构与持仓分布",
			Outcome:   normalizeNarrativeText(predictionText),
		}},
		Narrative: normalizeNarrativeText(predictionText),
	}
}

func summarizeMarketDataStatus(statuses []string) string {
	hasComplete := false
	hasLive := false
	hasMissing := false
	for _, status := range statuses {
		switch status {
		case marketDataStatusComplete:
			hasComplete = true
		case marketDataStatusFetchedLive:
			hasLive = true
		case marketDataStatusPartial, marketDataStatusUnavailable:
			hasMissing = true
		}
	}
	switch {
	case hasMissing && (hasComplete || hasLive):
		return marketDataStatusPartial
	case hasMissing:
		return marketDataStatusUnavailable
	case hasLive:
		return marketDataStatusFetchedLive
	default:
		return marketDataStatusComplete
	}
}

func toAnalysisReportItemResponse(item model.AnalysisReportItem) responsedto.AnalysisReportItemResponse {
	return responsedto.AnalysisReportItemResponse{
		ID:                   item.ID,
		Symbol:               item.Symbol,
		AssetName:            item.AssetName,
		Market:               item.Market,
		TradeCount:           item.TradeCount,
		BuyCount:             item.BuyCount,
		SellCount:            item.SellCount,
		BuyAmount:            item.BuyAmount.String(),
		SellAmount:           item.SellAmount.String(),
		NetQuantity:          item.NetQuantity.String(),
		RealizedProfit:       item.RealizedProfit.String(),
		RealizedProfitRate:   item.RealizedProfitRate.String(),
		EndingPositionQty:    item.EndingPositionQty.String(),
		EndingAvgCost:        item.EndingAvgCost.String(),
		LatestPrice:          item.LatestPrice.String(),
		LatestMarketValue:    item.LatestMarketValue.String(),
		UnrealizedProfit:     item.UnrealizedProfit.String(),
		TotalProfit:          item.TotalProfit.String(),
		ChangePercent7D:      item.ChangePercent7D.String(),
		PeriodPriceChangePct: item.PeriodPriceChangePct.String(),
		MarketDataStatus:     item.MarketDataStatus,
		RiskLevel:            item.RiskLevel,
		InvestmentStyle:      derefString(item.InvestmentStyle),
		AnalysisText:         item.AnalysisText,
		Recommendation:       item.Recommendation,
		KeyPoints:            splitJSONOrLines(derefString(item.KeyPoints)),
		CreatedAt:            item.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func buildRiskInsights(items []responsedto.AnalysisReportItemResponse, reportRiskLevel string, recommendations []string) (responsedto.RiskAnalysisResponse, []responsedto.RiskAlertItemResponse, []responsedto.RiskSymbolResponse) {
	alertsByType := map[string]*responsedto.RiskAlertItemResponse{}
	topRiskSymbols := make([]riskSymbolSummary, 0, len(items))

	appendAlert := func(level, alertType, title, description, symbol string) {
		alert, ok := alertsByType[alertType]
		if !ok {
			alert = &responsedto.RiskAlertItemResponse{
				Level:       level,
				Type:        alertType,
				Title:       title,
				Description: description,
				Symbols:     []string{},
			}
			alertsByType[alertType] = alert
		}
		if symbol != "" {
			for _, existing := range alert.Symbols {
				if existing == symbol {
					return
				}
			}
			alert.Symbols = append(alert.Symbols, symbol)
		}
	}

	for _, item := range items {
		totalProfit := parseDecimalOrZero(item.TotalProfit)
		endingPositionQty := parseDecimalOrZero(item.EndingPositionQty)
		priceChangePct := parseDecimalOrZero(item.PeriodPriceChangePct)

		riskScore := 0
		triggerReasons := make([]string, 0, 4)

		if item.MarketDataStatus == marketDataStatusPartial || item.MarketDataStatus == marketDataStatusUnavailable {
			riskScore += 20
			triggerReasons = append(triggerReasons, "市场数据存在缺口")
			appendAlert("medium", "market_data", "市场数据完整性预警", "部分标的缺少完整历史市场数据，分析结论需要结合数据可用性理解。", item.Symbol)
		}
		if totalProfit.LessThan(decimal.Zero) && endingPositionQty.GreaterThan(decimal.Zero) {
			riskScore += 30
			triggerReasons = append(triggerReasons, "当前仍持仓且总盈亏为负")
			appendAlert("high", "drawdown", "持仓亏损预警", "部分标的在仍有持仓的情况下处于亏损状态，后续回撤会继续拖累组合。", item.Symbol)
		}
		if absDecimal(priceChangePct).GreaterThanOrEqual(decimal.NewFromInt(8)) && endingPositionQty.GreaterThan(decimal.Zero) {
			riskScore += 20
			triggerReasons = append(triggerReasons, "持仓标的阶段波动较大")
			appendAlert("medium", "volatility", "价格波动预警", "部分持仓标的在分析期内波动较大，短期收益和回撤的不确定性更高。", item.Symbol)
		}
		if item.TradeCount >= 8 {
			riskScore += 15
			triggerReasons = append(triggerReasons, "交易频率偏高")
			appendAlert("medium", "high_frequency", "高频交易预警", "部分标的交易频率偏高，容易放大追涨杀跌和交易摩擦成本。", item.Symbol)
		}
		if item.BuyCount >= item.SellCount*2 && item.BuyCount >= 3 && endingPositionQty.GreaterThan(decimal.Zero) {
			riskScore += 20
			triggerReasons = append(triggerReasons, "持续加仓且仓位偏重")
			appendAlert("medium", "concentration", "仓位集中预警", "部分标的呈现持续加仓且仍保留较多仓位的特征，集中度风险偏高。", item.Symbol)
		}

		if riskScore > 100 {
			riskScore = 100
		}
		if riskScore == 0 {
			continue
		}

		topRiskSymbols = append(topRiskSymbols, riskSymbolSummary{
			Symbol:         item.Symbol,
			AssetName:      item.AssetName,
			RiskLevel:      scoreToRiskLevel(riskScore),
			RiskScore:      riskScore,
			TriggerReasons: triggerReasons,
		})
	}

	alertList := make([]responsedto.RiskAlertItemResponse, 0, len(alertsByType))
	for _, alert := range alertsByType {
		alertList = append(alertList, *alert)
	}
	sort.Slice(alertList, func(i, j int) bool {
		return riskLevelRank(alertList[i].Level) > riskLevelRank(alertList[j].Level)
	})

	sort.Slice(topRiskSymbols, func(i, j int) bool {
		if topRiskSymbols[i].RiskScore == topRiskSymbols[j].RiskScore {
			return topRiskSymbols[i].Symbol < topRiskSymbols[j].Symbol
		}
		return topRiskSymbols[i].RiskScore > topRiskSymbols[j].RiskScore
	})
	if len(topRiskSymbols) > 5 {
		topRiskSymbols = topRiskSymbols[:5]
	}

	riskFactors := make([]string, 0, len(alertList))
	for _, alert := range alertList {
		riskFactors = append(riskFactors, alert.Title)
	}

	overallScore := riskLevelBaseScore(reportRiskLevel)
	if len(topRiskSymbols) > 0 && topRiskSymbols[0].RiskScore > overallScore {
		overallScore = topRiskSymbols[0].RiskScore
	}
	if overallScore > 100 {
		overallScore = 100
	}

	topRiskResponses := make([]responsedto.RiskSymbolResponse, 0, len(topRiskSymbols))
	for _, item := range topRiskSymbols {
		topRiskResponses = append(topRiskResponses, responsedto.RiskSymbolResponse{
			Symbol:         item.Symbol,
			AssetName:      item.AssetName,
			RiskLevel:      item.RiskLevel,
			RiskScore:      item.RiskScore,
			TriggerReasons: item.TriggerReasons,
		})
	}

	return responsedto.RiskAnalysisResponse{
		RiskLevel:       scoreToRiskLevel(overallScore),
		RiskScore:       overallScore,
		RiskFactors:     riskFactors,
		Recommendations: recommendations,
	}, alertList, topRiskResponses
}

func parseDecimalOrZero(value string) decimal.Decimal {
	parsed, err := decimal.NewFromString(strings.TrimSpace(value))
	if err != nil {
		return decimal.Zero
	}
	return parsed
}

func absDecimal(value decimal.Decimal) decimal.Decimal {
	if value.IsNegative() {
		return value.Neg()
	}
	return value
}

func scoreToRiskLevel(score int) string {
	switch {
	case score >= 60:
		return "high"
	case score >= 30:
		return "medium"
	default:
		return "low"
	}
}

func riskLevelBaseScore(level string) int {
	switch normalizeRiskLevel(level) {
	case "high":
		return 80
	case "medium":
		return 50
	default:
		return 20
	}
}

func riskLevelRank(level string) int {
	switch normalizeRiskLevel(level) {
	case "high":
		return 3
	case "medium":
		return 2
	default:
		return 1
	}
}

func summarizeMetricRows(metrics []model.StockAnalysisMetric) (decimal.Decimal, decimal.Decimal, []string) {
	totalInvestment := modelDecimalZero()
	totalRealizedProfit := modelDecimalZero()
	marketStatuses := make([]string, 0, len(metrics))
	for _, metric := range metrics {
		totalInvestment = totalInvestment.Add(metric.BuyAmount)
		totalRealizedProfit = totalRealizedProfit.Add(metric.RealizedProfit)
		marketStatuses = append(marketStatuses, metric.MarketDataStatus)
	}
	return totalInvestment, totalRealizedProfit, marketStatuses
}

func buildSummaryReport(userID uint64, startTime, endTime time.Time, rawOutput string, output *aiAnalysisOutput, metrics []model.StockAnalysisMetric, modelName string) (*model.AnalysisReport, error) {
	totalInvestment, totalProfit, marketStatuses := summarizeMetricRows(metrics)
	recommendationsJSON := marshalJSONArray(output.Summary.Recommendations)
	chartData := buildChartData(metrics)
	raw := rawOutput
	report := &model.AnalysisReport{
		UserID:              userID,
		ReportType:          "summary",
		ReportTitle:         fallbackString(output.Summary.ReportTitle, fmt.Sprintf("投资总结 (%s 至 %s)", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))),
		AnalysisPeriodStart: startTime,
		AnalysisPeriodEnd:   endTime,
		SymbolsCount:        len(metrics),
		RiskLevel:           normalizeRiskLevel(output.Summary.RiskLevel),
		MarketDataStatus:    summarizeMarketDataStatus(marketStatuses),
		InvestmentStyle:     stringPointerIfNotEmpty(output.Summary.InvestmentStyle),
		SummaryText:         output.Summary.SummaryText,
		RiskAnalysis:        stringPointerIfNotEmpty(output.Summary.RiskAnalysis),
		PatternInsights:     stringPointerIfNotEmpty(output.Summary.PatternInsights),
		PredictionText:      stringPointerIfNotEmpty(output.Summary.PredictionText),
		ChartData:           stringPointerIfNotEmpty(chartData),
		Recommendations:     stringPointerIfNotEmpty(recommendationsJSON),
		RawAIOutput:         stringPointerIfNotEmpty(raw),
		AIModel:             fallbackString(modelName, "unknown"),
		TotalInvestment:     totalInvestment,
		TotalProfit:         totalProfit,
		ProfitRate:          modelDecimalZero(),
	}
	winningTrades := 0
	losingTrades := 0
	for _, metric := range metrics {
		if metric.RealizedProfit.GreaterThan(modelDecimalZero()) {
			winningTrades++
		}
		if metric.RealizedProfit.LessThan(modelDecimalZero()) {
			losingTrades++
		}
	}
	report.WinningTrades = winningTrades
	report.LosingTrades = losingTrades
	if !totalInvestment.IsZero() {
		report.ProfitRate = totalProfit.Div(totalInvestment).Mul(decimal.NewFromInt(100))
	}
	return report, nil
}

func buildReportModels(taskID, userID uint64, startTime, endTime time.Time, rawOutput string, output *aiAnalysisOutput, metrics []model.StockAnalysisMetric, stockOutputMap map[string]aiStockOutput, modelName string) (*model.AnalysisReport, []model.AnalysisReportItem) {
	totalInvestment, totalProfit, marketStatuses := summarizeMetricRows(metrics)
	winningTrades := 0
	losingTrades := 0
	chartData := buildChartData(metrics)
	recommendationsJSON := marshalJSONArray(output.Summary.Recommendations)
	raw := rawOutput
	report := &model.AnalysisReport{
		TaskID:              &taskID,
		UserID:              userID,
		ReportType:          "summary",
		ReportTitle:         fallbackString(output.Summary.ReportTitle, fmt.Sprintf("股票分析报告 (%s 至 %s)", startTime.Format("2006-01-02"), endTime.Format("2006-01-02"))),
		AnalysisPeriodStart: startTime,
		AnalysisPeriodEnd:   endTime,
		SymbolsCount:        len(metrics),
		RiskLevel:           normalizeRiskLevel(output.Summary.RiskLevel),
		InvestmentStyle:     stringPointerIfNotEmpty(output.Summary.InvestmentStyle),
		SummaryText:         output.Summary.SummaryText,
		RiskAnalysis:        stringPointerIfNotEmpty(output.Summary.RiskAnalysis),
		PatternInsights:     stringPointerIfNotEmpty(output.Summary.PatternInsights),
		PredictionText:      stringPointerIfNotEmpty(output.Summary.PredictionText),
		ChartData:           stringPointerIfNotEmpty(chartData),
		Recommendations:     stringPointerIfNotEmpty(recommendationsJSON),
		RawAIOutput:         stringPointerIfNotEmpty(raw),
		AIModel:             fallbackString(modelName, "unknown"),
		TotalInvestment:     modelDecimalZero(),
		TotalProfit:         modelDecimalZero(),
		ProfitRate:          modelDecimalZero(),
	}
	items := make([]model.AnalysisReportItem, 0, len(metrics))
	for _, metric := range metrics {
		if metric.RealizedProfit.GreaterThan(modelDecimalZero()) {
			winningTrades++
		}
		if metric.RealizedProfit.LessThan(modelDecimalZero()) {
			losingTrades++
		}
		aiStock := stockOutputMap[metric.Symbol]
		keyPoints := marshalJSONArray(aiStock.KeyPoints)
		analysisText := normalizeNarrativeText(aiStock.AnalysisText)
		if isWeakNarrative(analysisText, 18) {
			analysisText = buildTransparentStockAnalysis(metric)
		}
		item := model.AnalysisReportItem{
			UserID:               userID,
			Symbol:               metric.Symbol,
			AssetName:            fallbackString(aiStock.AssetName, metric.AssetName),
			Market:               metric.Market,
			TradeCount:           metric.TradeCount,
			BuyCount:             metric.BuyCount,
			SellCount:            metric.SellCount,
			BuyAmount:            metric.BuyAmount,
			SellAmount:           metric.SellAmount,
			NetQuantity:          metric.NetQuantity,
			RealizedProfit:       metric.RealizedProfit,
			RealizedProfitRate:   metric.RealizedProfitRate,
			EndingPositionQty:    metric.EndingPositionQty,
			EndingAvgCost:        metric.EndingAvgCost,
			LatestPrice:          metric.LatestPrice,
			LatestMarketValue:    metric.LatestMarketValue,
			UnrealizedProfit:     metric.UnrealizedProfit,
			TotalProfit:          metric.RealizedProfit,
			ChangePercent7D:      metric.PeriodPriceChangePct,
			PeriodPriceChangePct: metric.PeriodPriceChangePct,
			MarketDataStatus:     metric.MarketDataStatus,
			RiskLevel:            normalizeRiskLevel(aiStock.RiskLevel),
			InvestmentStyle:      stringPointerIfNotEmpty(aiStock.InvestmentStyle),
			AnalysisText:         analysisText,
			Recommendation:       normalizeRecommendation(aiStock.Recommendation),
			KeyPoints:            stringPointerIfNotEmpty(keyPoints),
			RawAIOutput:          stringPointerIfNotEmpty(raw),
		}
		items = append(items, item)
	}
	report.TotalInvestment = totalInvestment
	report.TotalProfit = totalProfit
	report.WinningTrades = winningTrades
	report.LosingTrades = losingTrades
	report.MarketDataStatus = summarizeMarketDataStatus(marketStatuses)
	if !totalInvestment.IsZero() {
		report.ProfitRate = totalProfit.Div(totalInvestment).Mul(decimal.NewFromInt(100))
	}
	return report, items
}

func parseAIAnalysisOutput(content string) (*aiAnalysisOutput, error) {
	cleaned := strings.TrimSpace(content)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)
	start := strings.Index(cleaned, "{")
	end := strings.LastIndex(cleaned, "}")
	if start >= 0 && end >= start {
		cleaned = cleaned[start : end+1]
	}
	var output aiAnalysisOutput
	if err := json.Unmarshal([]byte(cleaned), &output); err != nil {
		return nil, fmt.Errorf("failed to parse AI output: %w", err)
	}
	output.Summary.ReportTitle = normalizeNarrativeText(output.Summary.ReportTitle)
	output.Summary.SummaryText = normalizeNarrativeText(output.Summary.SummaryText)
	output.Summary.RiskLevel = normalizeRiskLevel(output.Summary.RiskLevel)
	output.Summary.InvestmentStyle = normalizeInvestmentStyle(output.Summary.InvestmentStyle)
	output.Summary.RiskAnalysis = normalizeNarrativeText(output.Summary.RiskAnalysis)
	output.Summary.PatternInsights = normalizeNarrativeText(output.Summary.PatternInsights)
	output.Summary.PredictionText = normalizeNarrativeText(output.Summary.PredictionText)
	output.Summary.Prediction = normalizePredictionOutput(output.Summary.Prediction)
	output.Summary.Recommendations = normalizeRecommendationList(output.Summary.Recommendations)
	for i := range output.Stocks {
		output.Stocks[i].Symbol = normalizeSymbol(output.Stocks[i].Symbol)
		output.Stocks[i].AssetName = normalizeNarrativeText(output.Stocks[i].AssetName)
		output.Stocks[i].RiskLevel = normalizeRiskLevel(output.Stocks[i].RiskLevel)
		output.Stocks[i].InvestmentStyle = normalizeInvestmentStyle(output.Stocks[i].InvestmentStyle)
		output.Stocks[i].AnalysisText = normalizeNarrativeText(output.Stocks[i].AnalysisText)
		output.Stocks[i].Recommendation = normalizeRecommendation(output.Stocks[i].Recommendation)
		output.Stocks[i].KeyPoints = normalizeKeyPointsList(output.Stocks[i].KeyPoints)
	}
	return &output, nil
}

func sanitizeAIAnalysisOutput(output *aiAnalysisOutput, metrics []model.StockAnalysisMetric) *aiAnalysisOutput {
	if output == nil {
		return nil
	}
	metricMap := make(map[string]model.StockAnalysisMetric, len(metrics))
	for _, metric := range metrics {
		metricMap[normalizeSymbol(metric.Symbol)] = metric
	}
	seen := make(map[string]struct{}, len(metrics))
	filteredStocks := make([]aiStockOutput, 0, len(output.Stocks))
	for _, stock := range output.Stocks {
		normalizedSymbol := normalizeSymbol(stock.Symbol)
		metric, ok := metricMap[normalizedSymbol]
		if !ok || normalizedSymbol == "" {
			continue
		}
		if _, exists := seen[normalizedSymbol]; exists {
			continue
		}
		seen[normalizedSymbol] = struct{}{}
		if strings.TrimSpace(stock.AssetName) == "" {
			stock.AssetName = metric.AssetName
		}
		stock.Symbol = normalizedSymbol
		stock.AssetName = normalizeNarrativeText(stock.AssetName)
		stock.RiskLevel = normalizeRiskLevel(stock.RiskLevel)
		stock.InvestmentStyle = normalizeInvestmentStyle(stock.InvestmentStyle)
		stock.AnalysisText = normalizeNarrativeText(stock.AnalysisText)
		stock.Recommendation = normalizeRecommendation(stock.Recommendation)
		stock.KeyPoints = normalizeKeyPointsList(stock.KeyPoints)
		filteredStocks = append(filteredStocks, stock)
	}
	output.Stocks = filteredStocks
	output.Summary.ReportTitle = normalizeNarrativeText(output.Summary.ReportTitle)
	output.Summary.SummaryText = normalizeNarrativeText(output.Summary.SummaryText)
	output.Summary.InvestmentStyle = normalizeInvestmentStyle(output.Summary.InvestmentStyle)
	output.Summary.RiskAnalysis = normalizeNarrativeText(output.Summary.RiskAnalysis)
	output.Summary.PatternInsights = normalizeNarrativeText(output.Summary.PatternInsights)
	output.Summary.PredictionText = normalizeNarrativeText(output.Summary.PredictionText)
	output.Summary.Prediction = normalizePredictionOutput(output.Summary.Prediction)
	output.Summary.Recommendations = normalizeRecommendationList(output.Summary.Recommendations)
	return output
}

func validateAIAnalysisOutput(output *aiAnalysisOutput, metrics []model.StockAnalysisMetric) error {
	if output == nil {
		return fmt.Errorf("AI output is empty")
	}
	if isWeakNarrative(output.Summary.SummaryText, 10) {
		return fmt.Errorf("AI summary_text is too weak")
	}
	if !hasSummaryNarrativeShape(output.Summary.SummaryText) {
		return fmt.Errorf("AI summary_text lacks structured narrative")
	}
	if len(output.Summary.Recommendations) == 0 {
		return fmt.Errorf("AI recommendations are empty")
	}
	strongSections := 0
	if !isWeakNarrative(output.Summary.RiskAnalysis, 8) {
		if !hasRiskNarrativeShape(output.Summary.RiskAnalysis) {
			return fmt.Errorf("AI risk_analysis lacks structured narrative")
		}
		strongSections++
	}
	if !isWeakNarrative(output.Summary.PatternInsights, 8) {
		if !hasPatternNarrativeShape(output.Summary.PatternInsights) {
			return fmt.Errorf("AI pattern_insights lacks structured narrative")
		}
		strongSections++
	}
	if !isWeakNarrative(output.Summary.PredictionText, 8) {
		if !hasPredictionNarrativeShape(output.Summary.PredictionText) {
			return fmt.Errorf("AI prediction_text lacks conditional narrative")
		}
		strongSections++
	}
	if strongSections == 0 {
		return fmt.Errorf("AI summary sections are too weak")
	}
	if len(output.Stocks) == 0 {
		return fmt.Errorf("AI stock analysis is empty")
	}
	return nil
}

func normalizeNarrativeText(value string) string {
	text := strings.TrimSpace(value)
	if text == "" {
		return ""
	}
	text = strings.Join(strings.Fields(text), " ")
	return strings.TrimSpace(text)
}

func countNarrativeSegments(value string) int {
	text := normalizeNarrativeText(value)
	if text == "" {
		return 0
	}
	segments := 0
	tokenStart := 0
	for i, r := range text {
		switch r {
		case '。', '；', ';', '!', '！', '?', '？':
			if strings.TrimSpace(text[tokenStart:i]) != "" {
				segments++
			}
			tokenStart = i + utf8.RuneLen(r)
		}
	}
	if strings.TrimSpace(text[tokenStart:]) != "" {
		segments++
	}
	return segments
}

func containsAnyFold(text string, terms []string) bool {
	lowered := strings.ToLower(normalizeNarrativeText(text))
	for _, term := range terms {
		if strings.Contains(lowered, strings.ToLower(term)) {
			return true
		}
	}
	return false
}

func hasSummaryNarrativeShape(value string) bool {
	if countNarrativeSegments(value) >= 2 {
		return true
	}
	return containsAnyFold(value, []string{"偏高", "偏低", "放大", "改善", "承压", "贡献", "集中", "分散", "盈利", "亏损"})
}

func hasRiskNarrativeShape(value string) bool {
	return containsAnyFold(value, []string{"风险", "回撤", "波动", "集中", "亏损", "过期", "触发", "影响"})
}

func hasPatternNarrativeShape(value string) bool {
	return containsAnyFold(value, []string{"交易", "持有", "加仓", "减仓", "频率", "集中", "低频", "高频"})
}

func hasPredictionNarrativeShape(value string) bool {
	return containsAnyFold(value, []string{"若", "如果", "一旦", "继续", "当"}) && containsAnyFold(value, []string{"可能", "将", "则", "容易", "风险", "回撤", "改善", "放大"})
}

func normalizePredictionOutput(prediction aiPredictionOutput) aiPredictionOutput {
	prediction.Bias = normalizePredictionBias(prediction.Bias)
	prediction.Confidence = normalizePredictionConfidence(prediction.Confidence)
	prediction.Horizon = normalizePredictionHorizon(prediction.Horizon)
	prediction.Drivers = normalizePredictionDrivers(prediction.Drivers)
	prediction.Scenarios = normalizePredictionScenarios(prediction.Scenarios)
	return prediction
}

func normalizePredictionBias(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "bullish", "up", "positive", "看多", "偏多", "乐观":
		return "bullish"
	case "bearish", "down", "negative", "看空", "偏空", "悲观":
		return "bearish"
	case "neutral", "flat", "mixed", "中性", "震荡", "观望":
		return "neutral"
	default:
		return ""
	}
}

func normalizePredictionConfidence(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "high", "strong", "高", "高置信", "较高":
		return "high"
	case "medium", "mid", "moderate", "中", "中等", "一般":
		return "medium"
	case "low", "weak", "低", "较低":
		return "low"
	default:
		return ""
	}
}

func normalizePredictionHorizon(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "next_7d", "7d", "7_days", "short_term", "短期", "未来7天", "近7天":
		return "next_7d"
	case "next_30d", "30d", "30_days", "mid_term", "中期", "未来30天", "近30天":
		return "next_30d"
	default:
		return ""
	}
}

func normalizePredictionDrivers(values []string) []string {
	items := normalizeAndDeduplicateList(values)
	if len(items) > 3 {
		items = items[:3]
	}
	return items
}

func normalizePredictionScenarios(values []aiPredictionScenario) []aiPredictionScenario {
	result := make([]aiPredictionScenario, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		condition := normalizeNarrativeText(value.Condition)
		outcome := normalizeNarrativeText(value.Outcome)
		if condition == "" || outcome == "" {
			continue
		}
		key := condition + "|" + outcome
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, aiPredictionScenario{Condition: condition, Outcome: outcome})
		if len(result) >= 2 {
			break
		}
	}
	return result
}

func normalizeListItem(value string) string {
	text := normalizeNarrativeText(value)
	for {
		trimmed := strings.TrimSpace(text)
		if trimmed == "" {
			return ""
		}
		if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "•") || strings.HasPrefix(trimmed, "·") {
			text = strings.TrimSpace(trimmed[1:])
			continue
		}
		if len(trimmed) >= 2 && trimmed[0] >= '0' && trimmed[0] <= '9' {
			rest := strings.TrimSpace(trimmed[1:])
			if rest != "" {
				r, size := utf8.DecodeRuneInString(rest)
				switch r {
				case '.', '、', ')', '）':
					text = strings.TrimSpace(rest[size:])
					continue
				}
			}
		}
		return trimmed
	}
}

func normalizeAndDeduplicateList(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		item := normalizeListItem(value)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		result = append(result, item)
	}
	return result
}

func normalizeRecommendationList(values []string) []string {
	items := normalizeAndDeduplicateList(values)
	if len(items) > 5 {
		items = items[:5]
	}
	return items
}

func normalizeKeyPointsList(values []string) []string {
	items := normalizeAndDeduplicateList(values)
	result := make([]string, 0, len(items))
	for _, item := range items {
		if isWeakKeyPoint(item) {
			continue
		}
		result = append(result, item)
		if len(result) >= 4 {
			break
		}
	}
	return result
}

func isWeakKeyPoint(value string) bool {
	text := normalizeNarrativeText(value)
	if text == "" || utf8.RuneCountInString(text) < 4 {
		return true
	}
	if containsAnyFold(text, []string{"总盈亏", "已实现盈亏", "未实现盈亏", "净持仓", "期末持仓", "交易次数", "买入金额", "卖出金额", "买入股数", "卖出股数", "最新价", "最新市值", "持仓均价", "周期涨跌幅", "回撤", "波动", "集中度", "仓位", "成本", "收益率", "风险"}) {
		return false
	}
	for _, r := range text {
		if r >= '0' && r <= '9' {
			return false
		}
	}
	return true
}


func isWeakNarrative(value string, minRunes int) bool {
	text := normalizeNarrativeText(value)
	if text == "" || utf8.RuneCountInString(text) < minRunes {
		return true
	}
	lowered := strings.ToLower(text)
	weakPhrases := []string{
		"市场有风险",
		"建议谨慎",
		"结合自身情况",
		"请结合市场情况",
		"总体表现一般",
		"暂无总结",
		"无法判断",
		"观望为主",
		"保持关注",
	}
	for _, phrase := range weakPhrases {
		if strings.Contains(lowered, strings.ToLower(phrase)) {
			return true
		}
	}
	return false
}

func buildTransparentStockAnalysis(metric model.StockAnalysisMetric) string {
	return fmt.Sprintf("%s 在分析期内共 %d 次交易，买入%d次、卖出%d次，已实现盈亏%s，最新价%s，期末持仓%s。", metric.Symbol, metric.TradeCount, metric.BuyCount, metric.SellCount, metric.RealizedProfit.StringFixed(2), metric.LatestPrice.StringFixed(4), metric.EndingPositionQty.StringFixed(2))
}

func buildChartData(metrics []model.StockAnalysisMetric) string {
	points := make([]chartPoint, 0, len(metrics))
	for _, metric := range metrics {
		points = append(points, chartPoint{Symbol: metric.Symbol, Value: metric.RealizedProfit.StringFixed(2)})
	}
	sort.Slice(points, func(i, j int) bool { return points[i].Symbol < points[j].Symbol })
	data, _ := json.Marshal(chartDataEnvelope{
		Version: 2,
		Kind:    "profit_by_symbol",
		Metric:  "realized_profit",
		Points:  points,
	})
	return string(data)
}

func splitJSONOrLines(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err == nil {
		return values
	}
	parts := strings.Split(raw, "\n")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func marshalJSONArray(values []string) string {
	if len(values) == 0 {
		return ""
	}
	data, _ := json.Marshal(values)
	return string(data)
}

func validateAnalysisRange(startDate, endDate string) (time.Time, time.Time, error) {
	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date, use YYYY-MM-DD")
	}
	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date, use YYYY-MM-DD")
	}
	if endTime.Before(startTime) {
		return time.Time{}, time.Time{}, fmt.Errorf("end_date must be greater than or equal to start_date")
	}
	return startTime, endTime, nil
}

func normalizeSymbol(value string) string {
	trimmed := strings.ToUpper(strings.TrimSpace(value))
	if trimmed == "" {
		return ""
	}
	if trimmed == "000001.SH" || trimmed == "000001.SZ" {
		return "000001.SH"
	}
	if trimmed == "000300.SH" || trimmed == "000300.SZ" {
		return "000300.SH"
	}
	if strings.HasSuffix(trimmed, ".SH") || strings.HasSuffix(trimmed, ".SZ") {
		return trimmed
	}
	if len(trimmed) == 6 {
		if trimmed == "000001" || trimmed == "000300" {
			return trimmed + ".SH"
		}
		if strings.HasPrefix(trimmed, "6") {
			return trimmed + ".SH"
		}
		return trimmed + ".SZ"
	}
	return trimmed
}

func normalizeRiskLevel(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "low":
		return "low"
	case "high":
		return "high"
	default:
		return "medium"
	}
}

func normalizeRecommendation(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "buy", "hold", "reduce", "sell":
		return normalized
	case "observe", "watch", "watchlist", "monitor", "观察", "观望":
		return "observe"
	case "加仓", "买入", "增持":
		return "buy"
	case "持有", "继续持有", "拿住":
		return "hold"
	case "减仓", "降低仓位", "部分卖出":
		return "reduce"
	case "卖出", "止盈", "止损", "清仓":
		return "sell"
	default:
		return "observe"
	}
}

func normalizeInvestmentStyle(value string) string {
	normalized := strings.ToLower(strings.TrimSpace(value))
	switch normalized {
	case "conservative", "保守", "保守型", "保守防御", "稳健偏保守":
		return "conservative"
	case "aggressive", "激进", "激进型", "进攻型", "高弹性":
		return "aggressive"
	case "balanced", "均衡", "稳健", "稳健型", "平衡型", "balanced_growth", "value":
		return "balanced"
	default:
		return "balanced"
	}
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func derefUint64(value *uint64) uint64 {
	if value == nil {
		return 0
	}
	return *value
}

func stringPointerIfNotEmpty(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func fallbackString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func parseDate(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}

func errorsIsRecordNotFound(err error) bool {
	return err == nil || err == gorm.ErrRecordNotFound
}
