package service

import (
	"context"
	"fmt"
	"stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/pkg/deepseek"
	"time"
)

type AIService interface {
	GenerateInvestmentSummary(userID uint64, startDate, endDate string) (*response.AnalysisReportResponse, error)
	GetReports(userID uint64, reportType string, limit int) ([]response.AnalysisReportResponse, error)
}

type aiService struct {
	analysisReportRepo repository.AnalysisReportRepository
	transactionRepo    repository.TransactionRepository
	deepseekClient     *deepseek.Client
}

func NewAIService(
	analysisReportRepo repository.AnalysisReportRepository,
	transactionRepo repository.TransactionRepository,
	deepseekClient *deepseek.Client,
) AIService {
	return &aiService{
		analysisReportRepo: analysisReportRepo,
		transactionRepo:    transactionRepo,
		deepseekClient:     deepseekClient,
	}
}

func (s *aiService) GenerateInvestmentSummary(userID uint64, startDate, endDate string) (*response.AnalysisReportResponse, error) {
	// 获取交易记录
	transactions, err := s.transactionRepo.FindByDateRange(userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	if len(transactions) == 0 {
		return nil, fmt.Errorf("no transactions found in the specified period")
	}

	// 构建AI提示词
	systemPrompt := "你是一位专业的投资顾问，擅长分析投资数据并提供专业建议。"
	userPrompt := s.buildSummaryPrompt(transactions, startDate, endDate)

	// 调用Deepseek API
	aiContent, err := s.deepseekClient.GetContent(context.Background(), systemPrompt, userPrompt)
	if err != nil {
		return nil, err
	}

	// 创建分析报告
	report := &model.AnalysisReport{
		UserID:              userID,
		ReportType:          "summary",
		ReportTitle:         fmt.Sprintf("投资总结 (%s 至 %s)", startDate, endDate),
		AnalysisPeriodStart: parseDate(startDate),
		AnalysisPeriodEnd:   parseDate(endDate),
		SummaryText:         aiContent,
		AIModel:             "deepseek-chat",
		CreatedAt:           time.Now(),
	}

	// 保存到数据库
	if err := s.analysisReportRepo.Create(report); err != nil {
		return nil, err
	}

	return s.convertToResponse(report), nil
}

func (s *aiService) GetReports(userID uint64, reportType string, limit int) ([]response.AnalysisReportResponse, error) {
	if limit <= 0 {
		limit = 10
	}

	reports, err := s.analysisReportRepo.FindByUserID(userID, reportType, limit)
	if err != nil {
		return nil, err
	}

	var reportResponses []response.AnalysisReportResponse
	for _, report := range reports {
		reportResponses = append(reportResponses, *s.convertToResponse(&report))
	}

	return reportResponses, nil
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
		if i >= 10 { // 只显示前10条
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

func (s *aiService) convertToResponse(report *model.AnalysisReport) *response.AnalysisReportResponse {
	investmentStyle := ""
	if report.InvestmentStyle != nil {
		investmentStyle = *report.InvestmentStyle
	}
	riskAnalysis := ""
	if report.RiskAnalysis != nil {
		riskAnalysis = *report.RiskAnalysis
	}
	patternInsights := ""
	if report.PatternInsights != nil {
		patternInsights = *report.PatternInsights
	}
	predictionText := ""
	if report.PredictionText != nil {
		predictionText = *report.PredictionText
	}
	chartData := ""
	if report.ChartData != nil {
		chartData = *report.ChartData
	}
	recommendations := ""
	if report.Recommendations != nil {
		recommendations = *report.Recommendations
	}

	return &response.AnalysisReportResponse{
		ID:                  report.ID,
		ReportType:          report.ReportType,
		ReportTitle:         report.ReportTitle,
		AnalysisPeriodStart: report.AnalysisPeriodStart.Format("2006-01-02"),
		AnalysisPeriodEnd:   report.AnalysisPeriodEnd.Format("2006-01-02"),
		TotalInvestment:     report.TotalInvestment.String(),
		TotalProfit:         report.TotalProfit.String(),
		ProfitRate:          report.ProfitRate.String(),
		RiskLevel:           report.RiskLevel,
		InvestmentStyle:     investmentStyle,
		SummaryText:         report.SummaryText,
		RiskAnalysis:        riskAnalysis,
		PatternInsights:     patternInsights,
		PredictionText:      predictionText,
		ChartData:           chartData,
		Recommendations:     recommendations,
		AIModel:             report.AIModel,
		CreatedAt:           report.CreatedAt.Format("2006-01-02 15:04:05"),
	}
}

func parseDate(dateStr string) time.Time {
	t, _ := time.Parse("2006-01-02", dateStr)
	return t
}
