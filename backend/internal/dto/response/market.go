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
	SnapshotTime string                    `json:"snapshot_time"`
	RefreshedAt  string                    `json:"refreshed_at"`
	IsStale      bool                      `json:"is_stale"`
	Source       string                    `json:"source"`
	Indices      []MarketIndexItemResponse `json:"indices"`
	MainChart    MarketChartResponse       `json:"main_chart"`
	Stats        []DashboardStatResponse   `json:"stats"`
}

type MarketIndexItemResponse struct {
	Symbol        string `json:"symbol"`
	Name          string `json:"name"`
	Market        string `json:"market"`
	LastPrice     string `json:"last_price"`
	ChangeAmount  string `json:"change_amount"`
	ChangePercent string `json:"change_percent"`
	OpenPrice     string `json:"open_price"`
	HighPrice     string `json:"high_price"`
	LowPrice      string `json:"low_price"`
	PrevClose     string `json:"prev_close"`
	Volume        string `json:"volume"`
	Turnover      string `json:"turnover"`
}

type MarketChartResponse struct {
	IndexName string             `json:"index_name"`
	Series    []MarketChartPoint `json:"series"`
}

type MarketChartPoint struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type DashboardStatResponse struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type MarketBreadthItemResponse struct {
	Symbol         string `json:"symbol"`
	Name           string `json:"name"`
	Market         string `json:"market"`
	LastPrice      string `json:"last_price"`
	ChangeAmount   string `json:"change_amount"`
	ChangePercent  string `json:"change_percent"`
	OpenPrice      string `json:"open_price"`
	HighPrice      string `json:"high_price"`
	LowPrice       string `json:"low_price"`
	PrevClose      string `json:"prev_close"`
	Volume         string `json:"volume"`
	Turnover       string `json:"turnover"`
	TurnoverRate   string `json:"turnover_rate"`
	TotalMarketCap string `json:"total_market_cap"`
	FloatMarketCap string `json:"float_market_cap"`
}

type MarketBoardItemResponse struct {
	BoardType       string `json:"board_type"`
	Code            string `json:"code"`
	Name            string `json:"name"`
	LastPrice       string `json:"last_price"`
	ChangeAmount    string `json:"change_amount"`
	ChangePercent   string `json:"change_percent"`
	Volume          string `json:"volume"`
	Turnover        string `json:"turnover"`
	TotalMarketCap  string `json:"total_market_cap"`
	FloatMarketCap  string `json:"float_market_cap"`
	StockCount      int    `json:"stock_count"`
	RiseCount       int    `json:"rise_count"`
	FallCount       int    `json:"fall_count"`
	FlatCount       int    `json:"flat_count"`
	ConstituentNode string `json:"constituent_node"`
}

type BoardConstituentResponse struct {
	Symbol         string `json:"symbol"`
	Name           string `json:"name"`
	Market         string `json:"market"`
	LastPrice      string `json:"last_price"`
	ChangeAmount   string `json:"change_amount"`
	ChangePercent  string `json:"change_percent"`
	Volume         string `json:"volume"`
	Turnover       string `json:"turnover"`
	TotalMarketCap string `json:"total_market_cap"`
	FloatMarketCap string `json:"float_market_cap"`
	Source         string `json:"source"`
	HasSnapshot    bool   `json:"has_snapshot"`
}

type MarketBoardDetailResponse struct {
	SnapshotTime string                     `json:"snapshot_time"`
	RefreshedAt  string                     `json:"refreshed_at"`
	Source       string                     `json:"source"`
	IsPartial    bool                       `json:"is_partial"`
	Message      string                     `json:"message"`
	Board        MarketBoardItemResponse    `json:"board"`
	Coverage     []DashboardStatResponse    `json:"coverage"`
	Constituents []BoardConstituentResponse `json:"constituents"`
	TopGainers   []BoardConstituentResponse `json:"top_gainers"`
	TopLosers    []BoardConstituentResponse `json:"top_losers"`
	TopTurnover  []BoardConstituentResponse `json:"top_turnover"`
}

type DistributionBucketResponse struct {
	Label string `json:"label"`
	Min   string `json:"min"`
	Max   string `json:"max"`
	Count int    `json:"count"`
}

type DashboardMarketBreadthResponse struct {
	SnapshotTime         string                       `json:"snapshot_time"`
	RefreshedAt          string                       `json:"refreshed_at"`
	Source               string                       `json:"source"`
	IsPartial            bool                         `json:"is_partial"`
	Message              string                       `json:"message"`
	Coverage             []DashboardStatResponse      `json:"coverage"`
	TopGainers           []MarketBreadthItemResponse  `json:"top_gainers"`
	TopLosers            []MarketBreadthItemResponse  `json:"top_losers"`
	TopTurnover          []MarketBreadthItemResponse  `json:"top_turnover"`
	Sectors              []MarketBoardItemResponse    `json:"sectors"`
	Concepts             []MarketBoardItemResponse    `json:"concepts"`
	ChangeDistribution   []DistributionBucketResponse `json:"change_distribution"`
	TurnoverDistribution []DistributionBucketResponse `json:"turnover_distribution"`
}

type StockBoardMembershipResponse struct {
	BoardType string `json:"board_type"`
	Code      string `json:"code"`
	Name      string `json:"name"`
	Source    string `json:"source"`
}

type StockCompanyProfileResponse struct {
	CompanyName     string `json:"company_name"`
	EnglishName     string `json:"english_name"`
	MarketLabel     string `json:"market_label"`
	IndustryLabel   string `json:"industry_label"`
	LegalRepresentative string `json:"legal_representative"`
	RegisteredCapital string `json:"registered_capital"`
	FoundedAt       string `json:"founded_at"`
	ListedAt        string `json:"listed_at"`
	Website         string `json:"website"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Address         string `json:"address"`
	OfficeAddress   string `json:"office_address"`
	Business        string `json:"business"`
	BusinessScope   string `json:"business_scope"`
	Introduction    string `json:"introduction"`
	Source          string `json:"source"`
}

type StockProfileResponse struct {
	Symbol           string                         `json:"symbol"`
	Name             string                         `json:"name"`
	Market           string                         `json:"market"`
	Description      string                         `json:"description"`
	CompanyProfile   *StockCompanyProfileResponse   `json:"company_profile,omitempty"`
	Industry         string                         `json:"industry"`
	Region           string                         `json:"region"`
	Concepts         []string                       `json:"concepts"`
	Boards           []StockBoardMembershipResponse `json:"boards"`
	LastPrice        string                         `json:"last_price"`
	ChangeAmount     string                         `json:"change_amount"`
	ChangePercent    string                         `json:"change_percent"`
	Volume           string                         `json:"volume"`
	Turnover         string                         `json:"turnover"`
	VolumeRatio      string                         `json:"volume_ratio"`
	TurnoverRate     string                         `json:"turnover_rate"`
	Amplitude        string                         `json:"amplitude"`
	LimitUp          string                         `json:"limit_up"`
	LimitDown        string                         `json:"limit_down"`
	TotalMarketCap   string                         `json:"total_market_cap"`
	FloatMarketCap   string                         `json:"float_market_cap"`
	Source           string                         `json:"source"`
	FetchedAt        string                         `json:"fetched_at"`
	IsStale          bool                           `json:"is_stale"`
	RefreshTriggered bool                           `json:"refresh_triggered"`
}

type StockNewsItemResponse struct {
	Title       string `json:"title"`
	Summary     string `json:"summary"`
	Source      string `json:"source"`
	URL         string `json:"url"`
	PublishedAt string `json:"published_at"`
	Provider    string `json:"provider"`
	IsRecent    bool   `json:"is_recent"`
}

type StockNewsResponse struct {
	Symbol      string                  `json:"symbol"`
	AssetName   string                  `json:"asset_name"`
	GeneratedAt string                  `json:"generated_at"`
	Coverage    string                  `json:"coverage"`
	Items       []StockNewsItemResponse `json:"items"`
}

type BoardNewsResponse struct {
	BoardType    string                  `json:"board_type"`
	Code         string                  `json:"code"`
	BoardName    string                  `json:"board_name"`
	GeneratedAt  string                  `json:"generated_at"`
	Coverage     string                  `json:"coverage"`
	Items        []StockNewsItemResponse `json:"items"`
}
