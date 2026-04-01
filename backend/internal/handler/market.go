package handler

import (
	"errors"
	"stock-analysis-backend/internal/service"
	"stock-analysis-backend/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

type MarketHandler struct {
	marketSnapshotService service.MarketSnapshotService
}

func NewMarketHandler(marketSnapshotService service.MarketSnapshotService) *MarketHandler {
	return &MarketHandler{marketSnapshotService: marketSnapshotService}
}

func (h *MarketHandler) GetLatestSnapshots(c *gin.Context) {
	snapshots, err := h.marketSnapshotService.GetLatestSnapshots()
	if err != nil {
		response.InternalServerError(c, "failed to get latest market snapshots")
		return
	}
	response.Success(c, snapshots)
}

func (h *MarketHandler) GetSnapshotHistory(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		response.BadRequest(c, "symbol is required")
		return
	}

	limit := 60
	if value := c.Query("limit"); value != "" {
		if parsed, err := strconvAtoi(value); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	startTime, err := parseOptionalTime(c.Query("start_time"))
	if err != nil {
		response.BadRequest(c, "invalid start_time, use YYYY-MM-DD or YYYY-MM-DD HH:MM:SS")
		return
	}
	endTime, err := parseOptionalTime(c.Query("end_time"))
	if err != nil {
		response.BadRequest(c, "invalid end_time, use YYYY-MM-DD or YYYY-MM-DD HH:MM:SS")
		return
	}

	snapshots, err := h.marketSnapshotService.GetHistory(symbol, limit, startTime, endTime)
	if err != nil {
		response.InternalServerError(c, "failed to get market snapshot history")
		return
	}
	response.Success(c, snapshots)
}

func (h *MarketHandler) GetDashboardSnapshot(c *gin.Context) {
	snapshot, err := h.marketSnapshotService.GetDashboardSnapshot()
	if err != nil {
		response.InternalServerError(c, "failed to get dashboard market snapshot")
		return
	}
	response.Success(c, snapshot)
}

func parseOptionalTime(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return &parsed, nil
		}
	}
	return nil, errors.New("invalid time format")
}

func strconvAtoi(value string) (int, error) {
	var result int
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return 0, errors.New("invalid integer")
		}
		result = result*10 + int(ch-'0')
	}
	return result, nil
}
