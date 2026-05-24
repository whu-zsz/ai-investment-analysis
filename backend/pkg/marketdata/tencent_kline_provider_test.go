package marketdata

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type errorKlineProvider struct {
	err error
}

type quoteOnlyProvider struct {
	detail *StockDetail
}

func (p *quoteOnlyProvider) GetQuotes(ctx context.Context, symbols []string) ([]Quote, error) {
	return nil, nil
}

func (p *quoteOnlyProvider) GetStockDetail(ctx context.Context, symbol string) (*StockDetail, error) {
	if p.detail != nil {
		return p.detail, nil
	}
	return &StockDetail{Symbol: symbol, Name: "腾讯详情", Source: "tencent"}, nil
}

func (p *quoteOnlyProvider) GetKlines(ctx context.Context, symbol, period, adjust string, limit int) ([]KlineBar, error) {
	return nil, errors.New("kline not supported")
}

func (p *errorKlineProvider) GetQuotes(ctx context.Context, symbols []string) ([]Quote, error) {
	return nil, nil
}

func (p *errorKlineProvider) GetStockDetail(ctx context.Context, symbol string) (*StockDetail, error) {
	return nil, nil
}

func (p *errorKlineProvider) GetKlines(ctx context.Context, symbol, period, adjust string, limit int) ([]KlineBar, error) {
	return nil, p.err
}

func TestHybridProviderRoutesMethods(t *testing.T) {
	realtime := NewMockProvider()
	historyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/appstock/app/kline/kline" {
			t.Fatalf("Path = %s, want /appstock/app/kline/kline", r.URL.Path)
		}
		if got := r.URL.Query().Get("param"); got != "sh000001,day,,,3" {
			t.Fatalf("param = %s, want sh000001,day,,,3", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"sh000001":{"day":[["2026-05-20","4096.170","4112.900","4120.090","4067.750","589459644.000"],["2026-05-21","4112.900","4105.000","4130.000","4090.000","500000000.000"],["2026-05-22","4105.000","4120.000","4135.000","4098.000","510000000.000"]]}}}`))
	}))
	defer historyServer.Close()

	history := NewTencentKlineProvider(historyServer.URL, "Mozilla/5.0", historyServer.Client())
	provider := NewHybridProvider(realtime, history)

	quotes, err := provider.GetQuotes(context.Background(), []string{"000001.SH"})
	if err != nil {
		t.Fatalf("GetQuotes() error = %v", err)
	}
	if len(quotes) != 1 {
		t.Fatalf("GetQuotes() len = %d, want 1", len(quotes))
	}

	bars, err := provider.GetKlines(context.Background(), "000001.SH", "day", "none", 3)
	if err != nil {
		t.Fatalf("GetKlines() error = %v", err)
	}
	if len(bars) != 3 {
		t.Fatalf("GetKlines() len = %d, want 3", len(bars))
	}
	if bars[0].Source != "tencent" {
		t.Fatalf("Source = %s, want tencent", bars[0].Source)
	}
}

func TestHybridProviderRoutesStockDetailToDetailProvider(t *testing.T) {
	provider := NewHybridProvider(NewMockProvider(), &quoteOnlyProvider{detail: &StockDetail{Symbol: "161725.SZ", Name: "白酒基金LOF", Source: "tencent"}})
	detail, err := provider.GetStockDetail(context.Background(), "161725.SZ")
	if err != nil {
		t.Fatalf("GetStockDetail() error = %v", err)
	}
	if detail.Source != "tencent" {
		t.Fatalf("Source = %s, want tencent", detail.Source)
	}
	if detail.Name != "白酒基金LOF" {
		t.Fatalf("Name = %s, want 白酒基金LOF", detail.Name)
	}
}

func TestHybridProviderFallsBackToRealtimeForRealtimeOnlyPeriods(t *testing.T) {
	provider := NewHybridProvider(NewMockProvider(), &errorKlineProvider{err: errors.New("unsupported kline period: 1m")})
	bars, err := provider.GetKlines(context.Background(), "000001.SH", "1m", "none", 3)
	if err != nil {
		t.Fatalf("GetKlines() error = %v", err)
	}
	if len(bars) != 3 {
		t.Fatalf("GetKlines() len = %d, want 3", len(bars))
	}
	if bars[0].Source != "mock" {
		t.Fatalf("Source = %s, want mock", bars[0].Source)
	}
}

func TestHybridProviderFallsBackToRealtimeForMinutePeriods(t *testing.T) {
	provider := NewHybridProvider(NewMockProvider(), &errorKlineProvider{err: errors.New("failed to fetch tencent minute kline")})
	bars, err := provider.GetKlines(context.Background(), "000858.SZ", "5m", "none", 3)
	if err != nil {
		t.Fatalf("GetKlines() error = %v", err)
	}
	if len(bars) != 3 {
		t.Fatalf("GetKlines() len = %d, want 3", len(bars))
	}
	if bars[0].Source != "mock" {
		t.Fatalf("Source = %s, want mock", bars[0].Source)
	}
}

func TestHybridProviderDoesNotFallbackOnContextDeadline(t *testing.T) {
	provider := NewHybridProvider(NewMockProvider(), &errorKlineProvider{err: context.DeadlineExceeded})
	_, err := provider.GetKlines(context.Background(), "000001.SH", "day", "none", 3)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("GetKlines() error = %v, want context deadline exceeded", err)
	}
}

func TestTencentKlineProviderFetchesDayBars(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/appstock/app/kline/kline" {
			t.Fatalf("Path = %s, want /appstock/app/kline/kline", r.URL.Path)
		}
		if got := r.URL.Query().Get("param"); got != "sh000001,day,,,3" {
			t.Fatalf("param = %s, want sh000001,day,,,3", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"sh000001":{"day":[["2026-05-20","4096.170","4112.900","4120.090","4067.750","589459644.000"],["2026-05-21","4112.900","4105.000","4130.000","4090.000","500000000.000"],["2026-05-22","4105.000","4120.000","4135.000","4098.000","510000000.000"]]}}}`))
	}))
	defer server.Close()

	provider := NewTencentKlineProvider(server.URL, "Mozilla/5.0", server.Client())
	bars, err := provider.GetKlines(context.Background(), "000001.SH", "day", "none", 3)
	if err != nil {
		t.Fatalf("GetKlines() error = %v", err)
	}
	if len(bars) != 3 {
		t.Fatalf("GetKlines() len = %d, want 3", len(bars))
	}
	if bars[0].Symbol != "000001.SH" {
		t.Fatalf("Symbol = %s, want 000001.SH", bars[0].Symbol)
	}
	if bars[0].Period != "day" {
		t.Fatalf("Period = %s, want day", bars[0].Period)
	}
}

func TestTencentKlineProviderFetchesQuotes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "" && r.URL.Path != "/" {
			t.Fatalf("Path = %s, want empty path", r.URL.Path)
		}
		if got := r.URL.Query().Get("q"); got != "sh000001,sz399001" {
			t.Fatalf("q = %s, want sh000001,sz399001", got)
		}
		_, _ = w.Write([]byte("v_sh000001=\"1~上证指数~000001~4112.90~4077.28~4096.17~589459644~0~0~0.00~0~0.00~0~0.00~0~0.00~0~0.00~0~0.00~0~0.00~0~0.00~0~0.00~0~0.00~0~~20260522161415~35.62~0.87~4120.09~4067.75~4112.90/589459644/1285847521737~589459644~128584752~1.23~18.11~~4120.09~4067.75~1.28~635766.64~686015.27~0.00~-1~-1~0.89~0~4094.27~~~~~~128584752.1737~0.0000~0~ ~ZS~3.63~-0.54~~~~4258.86~3332.49~-1.60~0.68~0.76~4810586917924~~3.68~10.35~4810586917924~~~21.68~-0.01~~CNY~0~~0.00~0~\";v_sz399001=\"51~深证成指~399001~15597.30~15247.27~15382.29~763067278~0~0~0.00~0~0.00~0~0.00~0~0.00~0~0.00~0~0.00~0~0.00~0~0.00~0~0.00~0~0.00~0~~20260522161400~350.03~2.30~15624.97~15295.57~15597.30/763067278/1617472538532~763067278~161747254~3.14~52.73~~15624.97~15295.57~2.16~424003.70~496780.06~0.00~-1~-1~0.97~0~15452.35~~~~~~161747253.8532~0.0000~0~ ~ZS~15.32~0.23~~~~16207.75~9950.14~0.22~4.11~10.62~2429024238356~~3.20~31.94~2429024238356~~~52.62~-0.00~~CNY~0~~0.00~0~\";"))
	}))
	defer server.Close()

	provider := NewTencentKlineProvider(server.URL, "Mozilla/5.0", server.Client())
	quotes, err := provider.GetQuotes(context.Background(), []string{"000001.SH", "399001.SZ"})
	if err != nil {
		t.Fatalf("GetQuotes() error = %v", err)
	}
	if len(quotes) != 2 {
		t.Fatalf("GetQuotes() len = %d, want 2", len(quotes))
	}
	if quotes[0].Source != "tencent" {
		t.Fatalf("Source = %s, want tencent", quotes[0].Source)
	}
	if quotes[0].Symbol != "000001.SH" {
		t.Fatalf("Symbol = %s, want 000001.SH", quotes[0].Symbol)
	}
	if quotes[0].LastPrice != 4112.9 {
		t.Fatalf("LastPrice = %v, want 4112.9", quotes[0].LastPrice)
	}
	if quotes[1].Symbol != "399001.SZ" {
		t.Fatalf("Symbol = %s, want 399001.SZ", quotes[1].Symbol)
	}
}

func TestTencentKlineProviderFetchesStockDetail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "" && r.URL.Path != "/" {
			t.Fatalf("Path = %s, want empty path", r.URL.Path)
		}
		if got := r.URL.Query().Get("q"); got != "sz161725" {
			t.Fatalf("q = %s, want sz161725", got)
		}
		_, _ = w.Write([]byte("v_sz161725=\"51~白酒基金LOF~161725~0.570~0.583~0.582~1171605~545551~626054~0.570~8444~0.569~6268~0.568~9008~0.567~1644~0.566~1819~0.571~8388~0.572~6649~0.573~3345~0.574~4701~0.575~3089~~20260522161418~-0.013~-2.23~0.583~0.570~0.570/1171605/67202729~1171605~6720~3.35~~~0.583~0.570~2.23~19.93~19.93~0.00~0.641~0.525~1.57~1011~0.574~~~~~~6720.2729~0.0000~0~ ~LOF~-19.83~-3.06~~~~0.850~0.570~-7.47~-9.09~-18.80~3496777665~3496777665~1.89~-28.48~3496777665~0.00~~-26.45~-0.18~0.5700~CNY~0~~0.580~-14295~\";"))
	}))
	defer server.Close()

	provider := NewTencentKlineProvider(server.URL, "Mozilla/5.0", server.Client())
	detail, err := provider.GetStockDetail(context.Background(), "161725.SZ")
	if err != nil {
		t.Fatalf("GetStockDetail() error = %v", err)
	}
	if detail.Symbol != "161725.SZ" {
		t.Fatalf("Symbol = %s, want 161725.SZ", detail.Symbol)
	}
	if detail.Name != "白酒基金LOF" {
		t.Fatalf("Name = %s, want 白酒基金LOF", detail.Name)
	}
	if detail.Source != "tencent" {
		t.Fatalf("Source = %s, want tencent", detail.Source)
	}
	if detail.LastPrice != 0.57 {
		t.Fatalf("LastPrice = %v, want 0.57", detail.LastPrice)
	}
	if detail.AveragePrice != 0.574 {
		t.Fatalf("AveragePrice = %v, want 0.574", detail.AveragePrice)
	}
	if detail.TotalMarketCap != 1993000000 {
		t.Fatalf("TotalMarketCap = %v, want 1993000000", detail.TotalMarketCap)
	}
	if detail.FloatShares != 3496777665 {
		t.Fatalf("FloatShares = %v, want 3496777665", detail.FloatShares)
	}
	if detail.Industry != "" {
		t.Fatalf("Industry = %q, want empty", detail.Industry)
	}
	if detail.Region != "" {
		t.Fatalf("Region = %q, want empty", detail.Region)
	}
	if len(detail.Concepts) != 0 {
		t.Fatalf("Concepts = %#v, want empty", detail.Concepts)
	}
}

func TestParseTencentMinuteRow(t *testing.T) {
	row := []any{"202605221500", "84.07", "84.03", "84.07", "84.03", "9962.00", map[string]any{}, "2.57"}
	bar, err := parseTencentRow("000858.SZ", "5m", "qfq", row, 84.07)
	if err != nil {
		t.Fatalf("parseTencentRow() error = %v", err)
	}
	if bar.Period != "5m" {
		t.Fatalf("Period = %s, want 5m", bar.Period)
	}
	if bar.Symbol != "000858.SZ" {
		t.Fatalf("Symbol = %s, want 000858.SZ", bar.Symbol)
	}
	if bar.Volume <= 0 {
		t.Fatalf("Volume = %v, want > 0", bar.Volume)
	}
}
