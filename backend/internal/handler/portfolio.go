package handler

import (
	dtoResponse "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/service"
	"stock-analysis-backend/pkg/response"

	"github.com/gin-gonic/gin"
)

type PortfolioHandler struct {
	portfolioService service.PortfolioService
}

func NewPortfolioHandler(portfolioService service.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{
		portfolioService: portfolioService,
	}
}

// GetPortfolios godoc
// @Summary 获取持仓列表
// @Description 获取用户当前所有持仓信息
// @Tags 持仓管理
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=[]dtoResponse.PortfolioResponse}
// @Router /api/v1/portfolios [get]
func (h *PortfolioHandler) GetPortfolios(c *gin.Context) {
	userID := c.GetUint64("user_id")

	portfolios, err := h.portfolioService.GetPortfolios(userID)
	if err != nil {
		response.InternalServerError(c, "failed to get portfolios")
		return
	}

	// 转换为响应格式
	var portfolioResponses []dtoResponse.PortfolioResponse
	for _, p := range portfolios {
		currentPrice := "0"
		if p.CurrentPrice != nil {
			currentPrice = p.CurrentPrice.String()
		}

		portfolioResponses = append(portfolioResponses, dtoResponse.PortfolioResponse{
			ID:                p.ID,
			AssetCode:         p.AssetCode,
			AssetName:         p.AssetName,
			AssetType:         p.AssetType,
			TotalQuantity:     p.TotalQuantity.String(),
			AvailableQuantity: p.AvailableQuantity.String(),
			AverageCost:       p.AverageCost.String(),
			CurrentPrice:      currentPrice,
			MarketValue:       p.MarketValue.String(),
			ProfitLoss:        p.ProfitLoss.String(),
			ProfitLossPercent: p.ProfitLossPercent.String(),
			LastUpdated:       p.LastUpdated.Format("2006-01-02 15:04:05"),
		})
	}

	response.Success(c, portfolioResponses)
}
