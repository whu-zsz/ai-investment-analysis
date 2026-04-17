package request

type CreateAnalysisTaskRequest struct {
	StartDate           string   `json:"start_date" binding:"required"`
	EndDate             string   `json:"end_date" binding:"required"`
	Symbols             []string `json:"symbols"`
	ForceRefreshMarket  bool     `json:"force_refresh_market"`
	ForceRefreshMetrics bool     `json:"force_refresh_metrics"`
}
