package handler

import (
	"stock-analysis-backend/internal/dto/request"
	"stock-analysis-backend/internal/service"
	"stock-analysis-backend/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TransactionHandler struct {
	transactionService service.TransactionService
}

func NewTransactionHandler(transactionService service.TransactionService) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
	}
}

// CreateTransaction godoc
// @Summary 创建交易记录
// @Description 创建新的交易记录
// @Tags 交易记录
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.CreateTransactionRequest true "交易记录"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/transactions [post]
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	userID := c.GetUint64("user_id")

	var req request.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.transactionService.CreateTransaction(userID, &req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetTransactions godoc
// @Summary 获取交易记录列表
// @Description 获取用户的交易记录列表（分页）
// @Tags 交易记录
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=response.TransactionListResponse}
// @Router /api/v1/transactions [get]
func (h *TransactionHandler) GetTransactions(c *gin.Context) {
	userID := c.GetUint64("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.transactionService.GetTransactions(userID, page, pageSize)
	if err != nil {
		response.InternalServerError(c, "failed to get transactions")
		return
	}

	response.Success(c, result)
}

// GetTransactionStats godoc
// @Summary 获取交易统计
// @Description 获取用户的交易统计数据
// @Tags 交易记录
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=response.TransactionStats}
// @Router /api/v1/transactions/stats [get]
func (h *TransactionHandler) GetTransactionStats(c *gin.Context) {
	userID := c.GetUint64("user_id")

	stats, err := h.transactionService.GetTransactionStats(userID)
	if err != nil {
		response.InternalServerError(c, "failed to get transaction stats")
		return
	}

	response.Success(c, stats)
}

// DeleteTransaction godoc
// @Summary 删除交易记录
// @Description 删除指定的交易记录
// @Tags 交易记录
// @Produce json
// @Security BearerAuth
// @Param id path int true "交易记录ID"
// @Success 200 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/transactions/{id} [delete]
func (h *TransactionHandler) DeleteTransaction(c *gin.Context) {
	userID := c.GetUint64("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "invalid transaction id")
		return
	}

	if err := h.transactionService.DeleteTransaction(userID, id); err != nil {
		response.NotFound(c, err.Error())
		return
	}

	response.Success(c, nil)
}
