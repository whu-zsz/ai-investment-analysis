package handler

import (
	"errors"
	dtoRequest "stock-analysis-backend/internal/dto/request"
	dtoResponse "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/service"
	pkgResponse "stock-analysis-backend/pkg/response"
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
// @Param request body dtoRequest.CreateTransactionRequest true "交易记录"
// @Success 200 {object} response.Response
// @Failure 400 {object} response.Response
// @Router /api/v1/transactions [post]
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	userID := c.GetUint64("user_id")

	var req dtoRequest.CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgResponse.ValidationError(c, err)
		return
	}

	if err := h.transactionService.CreateTransaction(userID, &req); err != nil {
		pkgResponse.BadRequest(c, err.Error())
		return
	}

	pkgResponse.Success(c, nil)
}

// GetTransactions godoc
// @Summary 获取交易记录列表
// @Description 获取用户的交易记录列表（分页）
// @Tags 交易记录
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=dtoResponse.TransactionListResponse}
// @Router /api/v1/transactions [get]
func (h *TransactionHandler) GetTransactions(c *gin.Context) {
	userID := c.GetUint64("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.transactionService.GetTransactions(userID, page, pageSize)
	if err != nil {
		pkgResponse.InternalServerError(c, "failed to get transactions")
		return
	}

	pkgResponse.Success(c, result)
}

// GetTransaction godoc
// @Summary 获取交易记录详情
// @Description 获取当前用户指定交易记录详情
// @Tags 交易记录
// @Produce json
// @Security BearerAuth
// @Param id path int true "交易记录ID"
// @Success 200 {object} response.Response{data=dtoResponse.TransactionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/transactions/{id} [get]
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	userID := c.GetUint64("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		pkgResponse.BadRequest(c, "invalid transaction id")
		return
	}

	transaction, err := h.transactionService.GetTransactionByID(userID, id)
	if err != nil {
		if errors.Is(err, service.ErrTransactionNotFound) {
			pkgResponse.NotFound(c, "transaction not found")
			return
		}
		pkgResponse.InternalServerError(c, "failed to get transaction")
		return
	}

	pkgResponse.Success(c, dtoResponse.NewTransactionResponse(transaction))
}

// UpdateTransaction godoc
// @Summary 更新交易记录
// @Description 以完整对象语义更新指定交易记录，并在更新后触发持仓重算
// @Tags 交易记录
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "交易记录ID"
// @Param request body dtoRequest.UpdateTransactionRequest true "更新交易记录"
// @Success 200 {object} response.Response{data=dtoResponse.TransactionResponse}
// @Failure 400 {object} response.Response
// @Failure 404 {object} response.Response
// @Router /api/v1/transactions/{id} [put]
func (h *TransactionHandler) UpdateTransaction(c *gin.Context) {
	userID := c.GetUint64("user_id")

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		pkgResponse.BadRequest(c, "invalid transaction id")
		return
	}

	var req dtoRequest.UpdateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		pkgResponse.ValidationError(c, err)
		return
	}

	transaction, err := h.transactionService.UpdateTransaction(userID, id, &req)
	if err != nil {
		if errors.Is(err, service.ErrTransactionNotFound) {
			pkgResponse.NotFound(c, "transaction not found")
			return
		}
		pkgResponse.BadRequest(c, err.Error())
		return
	}

	pkgResponse.Success(c, dtoResponse.NewTransactionResponse(transaction))
}

// GetTransactionStats godoc
// @Summary 获取交易统计
// @Description 获取用户的交易统计数据
// @Tags 交易记录
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=dtoResponse.TransactionStats}
// @Router /api/v1/transactions/stats [get]
func (h *TransactionHandler) GetTransactionStats(c *gin.Context) {
	userID := c.GetUint64("user_id")

	stats, err := h.transactionService.GetTransactionStats(userID)
	if err != nil {
		pkgResponse.InternalServerError(c, "failed to get transaction stats")
		return
	}

	pkgResponse.Success(c, stats)
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
		pkgResponse.BadRequest(c, "invalid transaction id")
		return
	}

	if err := h.transactionService.DeleteTransaction(userID, id); err != nil {
		if errors.Is(err, service.ErrTransactionNotFound) {
			pkgResponse.NotFound(c, "transaction not found")
			return
		}
		pkgResponse.InternalServerError(c, "failed to delete transaction")
		return
	}

	pkgResponse.Success(c, nil)
}
