package handler

import (
	"stock-analysis-backend/internal/dto/request"
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

// CreateTask godoc
// @Summary 创建股票分析任务
// @Description 异步创建指定时间段的 AI 股票分析任务
// @Tags AI分析
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateAnalysisTaskRequest true "分析任务请求"
// @Success 200 {object} response.Response{data=response.AnalysisTaskResponse}
// @Failure 400 {object} response.Response
// @Router /api/v1/analysis/tasks [post]
func (h *AnalysisHandler) CreateTask(c *gin.Context) {
	userID := c.GetUint64("user_id")

	var req request.CreateAnalysisTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	result, err := h.aiService.CreateStockAnalysisTask(userID, &req)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, result)
}

// GetTask godoc
// @Summary 获取分析任务状态
// @Description 获取指定 AI 分析任务的状态和进度
// @Tags AI分析
// @Produce json
// @Security BearerAuth
// @Param id path int true "任务ID"
// @Success 200 {object} response.Response{data=response.AnalysisTaskDetailResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/analysis/tasks/{id} [get]
func (h *AnalysisHandler) GetTask(c *gin.Context) {
	userID := c.GetUint64("user_id")
	taskID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid task id")
		return
	}

	result, err := h.aiService.GetAnalysisTask(userID, taskID)
	if err != nil {
		response.NotFound(c, "analysis task not found")
		return
	}

	response.Success(c, result)
}

// GetTasks godoc
// @Summary 获取分析任务列表
// @Description 获取当前用户的 AI 分析任务列表
// @Tags AI分析
// @Produce json
// @Security BearerAuth
// @Param status query string false "任务状态"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=response.AnalysisTaskListResponse}
// @Router /api/v1/analysis/tasks [get]
func (h *AnalysisHandler) GetTasks(c *gin.Context) {
	userID := c.GetUint64("user_id")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	result, err := h.aiService.GetAnalysisTasks(userID, status, page, pageSize)
	if err != nil {
		response.InternalServerError(c, "failed to get analysis tasks")
		return
	}

	response.Success(c, result)
}

// GetReportDetail godoc
// @Summary 获取分析报告详情
// @Description 获取 AI 分析报告详情和个股子分析
// @Tags AI分析
// @Produce json
// @Security BearerAuth
// @Param id path int true "报告ID"
// @Success 200 {object} response.Response{data=response.AnalysisReportDetailResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/analysis/reports/{id} [get]
func (h *AnalysisHandler) GetReportDetail(c *gin.Context) {
	userID := c.GetUint64("user_id")
	reportID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid report id")
		return
	}

	result, err := h.aiService.GetAnalysisReportDetail(userID, reportID)
	if err != nil {
		response.NotFound(c, "analysis report not found")
		return
	}

	response.Success(c, result)
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
