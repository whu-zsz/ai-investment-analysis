package marketdata

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEastmoneyProviderGetStockDetailUsesQuoteEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/qt/ulist.np/get" {
			t.Fatalf("Path = %s, want /api/qt/ulist.np/get", r.URL.Path)
		}
		if got := r.URL.Query().Get("secids"); got != "0.161725" {
			t.Fatalf("secids = %s, want 0.161725", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"diff":[{"f2":0.57,"f3":-2.23,"f4":-0.013,"f5":1171605,"f6":67202729.466,"f7":2.23,"f8":0,"f10":1.57,"f12":"161725","f14":"白酒基金LOF","f15":0.583,"f16":0.57,"f17":0.582,"f18":0.583,"f20":1993000000,"f21":1993000000,"f71":0.574,"f84":3496777665,"f85":3496777665,"f127":"基金","f128":"","f129":"白酒,LOF"}]}}`))
	}))
	defer server.Close()

	provider := NewEastmoneyProvider(server.URL+"/api/qt/ulist.np/get", "Mozilla/5.0", "https://quote.eastmoney.com", server.Client())
	res, err := provider.GetStockDetail(context.Background(), "161725.SZ")
	if err != nil {
		t.Fatalf("GetStockDetail() error = %v", err)
	}
	if res.Symbol != "161725.SZ" {
		t.Fatalf("Symbol = %s, want 161725.SZ", res.Symbol)
	}
	if res.Name != "白酒基金LOF" {
		t.Fatalf("Name = %s, want 白酒基金LOF", res.Name)
	}
	if res.LastPrice != 0.57 {
		t.Fatalf("LastPrice = %v, want 0.57", res.LastPrice)
	}
	if res.AveragePrice != 0.574 {
		t.Fatalf("AveragePrice = %v, want 0.574", res.AveragePrice)
	}
	if res.TotalMarketCap != 1993000000 {
		t.Fatalf("TotalMarketCap = %v, want 1993000000", res.TotalMarketCap)
	}
	if res.FloatShares != 3496777665 {
		t.Fatalf("FloatShares = %v, want 3496777665", res.FloatShares)
	}
	if len(res.Concepts) != 2 || res.Concepts[0] != "白酒" {
		t.Fatalf("Concepts = %#v, want [白酒 LOF]", res.Concepts)
	}
	if res.Source != "eastmoney" {
		t.Fatalf("Source = %s, want eastmoney", res.Source)
	}
}

func TestNormalizeEastmoneySymbol_Beijing(t *testing.T) {
	market, symbol := normalizeEastmoneySymbol("920125")
	if market != "cn_stock" {
		t.Fatalf("market = %q, want cn_stock", market)
	}
	if symbol != "920125.BJ" {
		t.Fatalf("symbol = %q, want 920125.BJ", symbol)
	}
}

func TestNormalizeProviderSymbol_DistinguishesPingAnBankAndShanghaiIndex(t *testing.T) {
	if got := normalizeProviderSymbol("000001"); got != "000001.SZ" {
		t.Fatalf("normalizeProviderSymbol(000001) = %q, want 000001.SZ", got)
	}
	if got := normalizeProviderSymbol("000001.SH"); got != "000001.SH" {
		t.Fatalf("normalizeProviderSymbol(000001.SH) = %q, want 000001.SH", got)
	}
	if got := marketFromSymbol("000001.SZ"); got != "cn_stock" {
		t.Fatalf("marketFromSymbol(000001.SZ) = %q, want cn_stock", got)
	}
	if got := marketFromSymbol("000001.SH"); got != "cn_index" {
		t.Fatalf("marketFromSymbol(000001.SH) = %q, want cn_index", got)
	}
	if got := DefaultName("000001.SZ"); got != "000001.SZ" {
		t.Fatalf("DefaultName(000001.SZ) = %q, want 000001.SZ", got)
	}
}
