package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/pkg/marketdata"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type stubStockDetailProvider struct {
	detail *marketdata.StockDetail
	err    error
}

func (s *stubStockDetailProvider) GetQuotes(ctx context.Context, symbols []string) ([]marketdata.Quote, error) {
	return nil, nil
}

func (s *stubStockDetailProvider) GetStockDetail(ctx context.Context, symbol string) (*marketdata.StockDetail, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.detail, nil
}

func (s *stubStockDetailProvider) GetKlines(ctx context.Context, symbol, period, adjust string, limit int) ([]marketdata.KlineBar, error) {
	return nil, nil
}

type stubStockQuoteDetailRepo struct {
	detail    *model.StockQuoteDetail
	findErr   error
	upserted  *model.StockQuoteDetail
	upsertErr error
}

func (r *stubStockQuoteDetailRepo) Upsert(detail *model.StockQuoteDetail) error {
	r.upserted = detail
	r.detail = detail
	return r.upsertErr
}

func (r *stubStockQuoteDetailRepo) FindBySymbol(symbol string) (*model.StockQuoteDetail, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	return r.detail, nil
}

func (r *stubStockQuoteDetailRepo) FindBySymbols(symbols []string) ([]model.StockQuoteDetail, error) {
	if r.findErr != nil {
		return nil, r.findErr
	}
	if r.detail == nil {
		return []model.StockQuoteDetail{}, nil
	}
	return []model.StockQuoteDetail{*r.detail}, nil
}

type stubStockKlineRepo struct{
	bars []model.StockKlineBar
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func (r *stubStockKlineRepo) UpsertBars(bars []model.StockKlineBar) error {
	return nil
}

func (r *stubStockKlineRepo) FindBars(symbol, period, adjust string, limit int) ([]model.StockKlineBar, error) {
	return r.bars, nil
}

func (r *stubStockKlineRepo) FindLatestBar(symbol, period, adjust string) (*model.StockKlineBar, error) {
	return nil, gorm.ErrRecordNotFound
}

func TestMarketStockService_GetStockDetail_SupplementsCachedMetadata(t *testing.T) {
	repo := &stubStockQuoteDetailRepo{
		detail: &model.StockQuoteDetail{
			Symbol:         "600519.SH",
			Name:           "贵州茅台",
			Market:         "cn_stock",
			LastPrice:      decimal.NewFromFloat(1290.2),
			VolumeRatio:    decimal.Zero,
			Industry:       "",
			Region:         "",
			Concepts:       "",
			Source:         "tencent",
			FetchedAt:      time.Date(2026, 5, 25, 14, 0, 0, 0, time.Local),
			TotalShares:    decimal.Zero,
			FloatShares:    decimal.Zero,
			TotalMarketCap: decimal.NewFromFloat(1615679000000),
			FloatMarketCap: decimal.NewFromFloat(1615679000000),
		},
	}

	svc := &marketStockService{
		provider: nil,
		supplementProvider: &stubStockDetailProvider{
			detail: &marketdata.StockDetail{
				Symbol:         "600519.SH",
				Name:           "贵州茅台",
				Market:         "cn_stock",
				Industry:       "酿酒行业",
				Region:         "贵州",
				Concepts:       []string{"白酒", "沪股通"},
				VolumeRatio:    1.28,
				TotalShares:    1256197800,
				FloatShares:    1256197800,
				Source:         "eastmoney",
				FetchedAt:      time.Date(2026, 5, 25, 14, 5, 0, 0, time.Local),
				LastPrice:      1290.2,
				TotalMarketCap: 1615679000000,
				FloatMarketCap: 1615679000000,
			},
		},
		detailRepo: repo,
		klineRepo:  &stubStockKlineRepo{},
	}

	resp, err := svc.GetStockDetail("600519.SH", false)
	if err != nil {
		t.Fatalf("GetStockDetail() error = %v", err)
	}
	if resp.Industry != "酿酒行业" {
		t.Fatalf("Industry = %q, want %q", resp.Industry, "酿酒行业")
	}
	if resp.Region != "贵州" {
		t.Fatalf("Region = %q, want %q", resp.Region, "贵州")
	}
	if resp.VolumeRatio != "1.28" {
		t.Fatalf("VolumeRatio = %q, want %q", resp.VolumeRatio, "1.28")
	}
	if resp.Source != "tencent+eastmoney" {
		t.Fatalf("Source = %q, want %q", resp.Source, "tencent+eastmoney")
	}
	if repo.upserted == nil {
		t.Fatalf("expected supplemented detail to be persisted")
	}
}

func TestMarketStockService_GetStockDetail_ForceRefreshWithinWindowUsesCache(t *testing.T) {
	repo := &stubStockQuoteDetailRepo{
		detail: &model.StockQuoteDetail{
			Symbol:    "600519.SH",
			Name:      "贵州茅台",
			Market:    "cn_stock",
			LastPrice: decimal.NewFromFloat(1540.12),
			Source:    "tencent",
			FetchedAt: time.Now().Add(-3 * time.Second),
		},
	}
	provider := &stubStockDetailProvider{
		detail: &marketdata.StockDetail{
			Symbol:    "600519.SH",
			Name:      "贵州茅台",
			LastPrice: 1600.00,
			Source:    "tencent",
			FetchedAt: time.Now(),
		},
	}
	svc := &marketStockService{
		provider:   provider,
		detailRepo: repo,
		klineRepo:  &stubStockKlineRepo{},
	}

	resp, err := svc.GetStockDetail("600519.SH", true)
	if err != nil {
		t.Fatalf("GetStockDetail() error = %v", err)
	}
	if resp.LastPrice != "1540.12" {
		t.Fatalf("LastPrice = %q, want %q", resp.LastPrice, "1540.12")
	}
	if repo.upserted != nil {
		t.Fatal("expected cache short-circuit without upsert")
	}
}

func TestMergeStockQuoteDetail_PreservesExistingSourceSuffix(t *testing.T) {
	current := &model.StockQuoteDetail{
		Symbol:      "600519.SH",
		Source:      "tencent+eastmoney",
		VolumeRatio: decimal.Zero,
	}
	supplement := &model.StockQuoteDetail{
		Symbol:      "600519.SH",
		Source:      "eastmoney",
		VolumeRatio: decimal.NewFromFloat(1.1),
	}

	merged := mergeStockQuoteDetail(current, supplement)
	if merged.Source != "tencent+eastmoney" {
		t.Fatalf("Source = %q, want %q", merged.Source, "tencent+eastmoney")
	}
	if merged.VolumeRatio.String() != "1.1" {
		t.Fatalf("VolumeRatio = %q, want %q", merged.VolumeRatio.String(), "1.1")
	}
}

func TestMarketFromNormalizedSymbol_DistinguishesStockAndIndex(t *testing.T) {
	if got := marketFromNormalizedSymbol("000001.SZ"); got != "cn_stock" {
		t.Fatalf("marketFromNormalizedSymbol(000001.SZ) = %q, want %q", got, "cn_stock")
	}
	if got := marketFromNormalizedSymbol("000001.SH"); got != "cn_index" {
		t.Fatalf("marketFromNormalizedSymbol(000001.SH) = %q, want %q", got, "cn_index")
	}
}

func TestMarketStockService_LookupBoardMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("code"); got != "SH600519" {
			t.Fatalf("code = %q, want SH600519", got)
		}
		_, _ = w.Write([]byte(`{
			"ssbk": [
				{"BOARD_NAME":"食品饮料","IS_PRECISE":"0","BOARD_RANK":1},
				{"BOARD_NAME":"贵州板块","IS_PRECISE":"0","BOARD_RANK":4},
				{"BOARD_NAME":"白酒","IS_PRECISE":"1","BOARD_RANK":24},
				{"BOARD_NAME":"超级品牌","IS_PRECISE":"1","BOARD_RANK":25},
				{"BOARD_NAME":"沪股通","IS_PRECISE":"0","BOARD_RANK":15}
			]
		}`))
	}))
	defer server.Close()

	client := server.Client()
	client.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = server.Listener.Addr().String()
		return http.DefaultTransport.RoundTrip(req)
	})

	svc := &marketStockService{httpClient: client}
	meta := svc.lookupBoardMetadata("600519.SH")
	if meta == nil {
		t.Fatal("lookupBoardMetadata() returned nil")
	}
	if meta.Industry != "食品饮料" {
		t.Fatalf("Industry = %q, want %q", meta.Industry, "食品饮料")
	}
	if meta.Region != "贵州" {
		t.Fatalf("Region = %q, want %q", meta.Region, "贵州")
	}
	if meta.Concepts != "白酒,超级品牌" {
		t.Fatalf("Concepts = %q, want %q", meta.Concepts, "白酒,超级品牌")
	}
	if meta.Source != "em-board" {
		t.Fatalf("Source = %q, want %q", meta.Source, "em-board")
	}
}

func TestMarketStockService_GetStockKlines_ForceRefreshWithinWindowUsesCache(t *testing.T) {
	now := time.Now()
	repo := &stubStockKlineRepo{
		bars: []model.StockKlineBar{
			{
				Symbol:     "600519.SH",
				Period:     "5m",
				AdjustType: "qfq",
				BarTime:    now.Add(-5 * time.Minute),
				ClosePrice: decimal.NewFromFloat(1540.12),
				Source:     "tencent",
				UpdatedAt:  now.Add(-4 * time.Second),
			},
		},
	}
	svc := &marketStockService{
		provider:   &stubStockDetailProvider{},
		detailRepo: &stubStockQuoteDetailRepo{},
		klineRepo:  repo,
	}

	resp, err := svc.GetStockKlines("600519.SH", "5m", "qfq", 1, true)
	if err != nil {
		t.Fatalf("GetStockKlines() error = %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(resp.Items))
	}
	if resp.RefreshTriggered {
		t.Fatal("expected cached kline response without external refresh")
	}
}

var _ repository.StockQuoteDetailRepository = (*stubStockQuoteDetailRepo)(nil)
var _ repository.StockKlineRepository = (*stubStockKlineRepo)(nil)
var _ marketdata.Provider = (*stubStockDetailProvider)(nil)
