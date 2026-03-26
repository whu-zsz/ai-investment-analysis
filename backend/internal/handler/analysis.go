package handler

import (
	"stock-analysis-backend/internal/service"
	"stock-analysis-backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AnalysisHandler struct {
	aiService service.AIService
}

func NewAnalysisHandler(aiService service.AIService) *AnalysisHandler {
	return &AnalysisHandler{
		aiService: aiService,
	}
}

// GenerateSummary godoc
// @Summary 生成投资总结
// @Description 使用AI生成指定时间段的投资总结报告
// @Tags AI分析
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "开始日期 (YYYY-MM-DD)"
// @Param end_date query string true "结束日期 (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=response.AnalysisReportResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/analysis/summary [post]
func (h *AnalysisHandler) GenerateSummary(c *gin.Context) {
	userID := c.GetUint64("user_id")

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		response.BadRequest(c, "start_date and end_date are required")
		return
	}

	report, err := h.aiService.GenerateInvestmentSummary(userID, startDate, endDate)
	if err != nil {
		response.InternalServerError(c, "failed to generate summary: "+err.Error())
		return
	}

	response.Success(c, report)
}

// GetReports godoc
// @Summary 获取历史报告
// @Description 获取用户的AI分析历史报告
// @Tags AI分析
// @Produce json
// @Security BearerAuth
// @Param report_type query string false "报告类型" Enums(summary, risk, prediction, pattern)
// @Param limit query int false "返回数量限制" default(10)
// @Success 200 {object} response.Response{data=[]response.AnalysisReportResponse}
// @Router /api/v1/analysis/reports [get]
func (h *AnalysisHandler) GetReports(c *gin.Context) {
	userID := c.GetUint64("user_id")

	reportType := c.Query("report_type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	reports, err := h.aiService.GetReports(userID, reportType, limit)
	if err != nil {
		response.InternalServerError(c, "failed to get reports")
		return
	}

	response.Success(c, reports)
}
