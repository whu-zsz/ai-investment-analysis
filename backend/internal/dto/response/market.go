package response

type MarketSnapshotResponse struct {
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	Market        string `json:"market"`
	SnapshotTime  string `json:"snapshot_time"`
	LastPrice     string `json:"last_price"`
	ChangeAmount  string `json:"change_amount"`
	ChangePercent string `json:"change_percent"`
	OpenPrice     string `json:"open_price"`
	HighPrice     string `json:"high_price"`
	LowPrice      string `json:"low_price"`
	PrevClose     string `json:"prev_close"`
	Volume        string `json:"volume"`
	Turnover      string `json:"turnover"`
	Source        string `json:"source"`
	BatchNo       string `json:"batch_no"`
}

type DashboardMarketSnapshotResponse struct {
	SnapshotTime string                     `json:"snapshot_time"`
	IsStale      bool                       `json:"is_stale"`
	Source       string                     `json:"source"`
	Indices      []MarketIndexItemResponse  `json:"indices"`
	MainChart    MarketChartResponse        `json:"main_chart"`
	Stats        []DashboardStatResponse    `json:"stats"`
}

type MarketIndexItemResponse struct {
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	LastPrice     string `json:"last_price"`
	ChangeAmount  string `json:"change_amount"`
	ChangePercent string `json:"change_percent"`
}

type MarketChartResponse struct {
	IndexName string              `json:"index_name"`
	Series    []MarketChartPoint  `json:"series"`
}

type MarketChartPoint struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type DashboardStatResponse struct {
	Label string `json:"label"`
	Value string `json:"value"`
}
