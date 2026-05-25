package marketdata

import (
	"stock-analysis-backend/internal/config"
	"testing"
)

func TestNewProviderCreatesMock(t *testing.T) {
	provider, err := NewProvider(config.MarketConfig{Provider: "mock", TimeoutSeconds: 5})
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	if provider == nil {
		t.Fatal("NewProvider() returned nil provider")
	}
}

func TestNewProviderCreatesEastmoney(t *testing.T) {
	provider, err := NewProvider(config.MarketConfig{
		Provider:           "eastmoney",
		TimeoutSeconds:     5,
		EastmoneyBaseURL:   "https://push2.eastmoney.com/api/qt/ulist.np/get",
		EastmoneyUserAgent: "Mozilla/5.0",
		EastmoneyReferer:   "https://quote.eastmoney.com/center/gridlist.html",
	})
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	if provider == nil {
		t.Fatal("NewProvider() returned nil provider")
	}
}

func TestNewProviderCreatesTencent(t *testing.T) {
	provider, err := NewProvider(config.MarketConfig{
		Provider:         "tencent",
		TimeoutSeconds:   5,
		TencentBaseURL:   "https://web.ifzq.gtimg.cn",
		TencentUserAgent: "Mozilla/5.0",
	})
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	if provider == nil {
		t.Fatal("NewProvider() returned nil provider")
	}
}

func TestNewProviderCreatesHybrid(t *testing.T) {
	provider, err := NewProvider(config.MarketConfig{
		Provider:           "hybrid",
		TimeoutSeconds:     5,
		EastmoneyBaseURL:   "https://push2.eastmoney.com/api/qt/ulist.np/get",
		EastmoneyUserAgent: "Mozilla/5.0",
		EastmoneyReferer:   "https://quote.eastmoney.com/center/gridlist.html",
		TencentBaseURL:     "https://web.ifzq.gtimg.cn",
		TencentUserAgent:   "Mozilla/5.0",
	})
	if err != nil {
		t.Fatalf("NewProvider() error = %v", err)
	}
	if provider == nil {
		t.Fatal("NewProvider() returned nil provider")
	}
}

func TestNewProviderRejectsUnknown(t *testing.T) {
	_, err := NewProvider(config.MarketConfig{Provider: "unknown", TimeoutSeconds: 5})
	if err == nil {
		t.Fatal("NewProvider() expected error for unknown provider")
	}
}

func TestNewRankingProviderCreatesFallback(t *testing.T) {
	provider := NewRankingProvider(config.MarketConfig{
		TimeoutSeconds:     5,
		SinaRequestDelayMS: 350,
		EastmoneyUserAgent: "Mozilla/5.0",
		EastmoneyReferer:   "https://quote.eastmoney.com/center/gridlist.html",
	})
	if provider == nil {
		t.Fatal("NewRankingProvider() returned nil provider")
	}
}

func TestToTencentSymbol(t *testing.T) {
	tests := []struct {
		symbol string
		want   string
	}{
		{symbol: "000001.SH", want: "sh000001"},
		{symbol: "399001.SZ", want: "sz399001"},
		{symbol: "000858", want: "sz000858"},
	}

	for _, tt := range tests {
		got, err := toTencentSymbol(tt.symbol)
		if err != nil {
			t.Fatalf("toTencentSymbol(%q) error = %v", tt.symbol, err)
		}
		if got != tt.want {
			t.Fatalf("toTencentSymbol(%q) = %q, want %q", tt.symbol, got, tt.want)
		}
	}
}

func TestBuildTencentMinuteEndpoint(t *testing.T) {
	if got := buildTencentMinuteEndpoint("https://web.ifzq.gtimg.cn", "/appstock/app/kline/mkline"); got != "https://ifzq.gtimg.cn/appstock/app/kline/mkline" {
		t.Fatalf("buildTencentMinuteEndpoint() = %q, want %q", got, "https://ifzq.gtimg.cn/appstock/app/kline/mkline")
	}
	if got := buildTencentMinuteEndpoint("", "/appstock/app/kline/mkline"); got != "https://ifzq.gtimg.cn/appstock/app/kline/mkline" {
		t.Fatalf("buildTencentMinuteEndpoint() empty = %q, want %q", got, "https://ifzq.gtimg.cn/appstock/app/kline/mkline")
	}
}

func TestTencentQuoteEndpoint(t *testing.T) {
	provider := NewTencentKlineProvider("https://web.ifzq.gtimg.cn", "Mozilla/5.0", nil)
	if got := provider.(*tencentKlineProvider).quoteEndpoint(); got != "https://qt.gtimg.cn" {
		t.Fatalf("quoteEndpoint() = %q, want %q", got, "https://qt.gtimg.cn")
	}
}

func TestParseTencentBarTime(t *testing.T) {
	if _, err := parseTencentBarTime("day", "2026-05-22"); err != nil {
		t.Fatalf("parseTencentBarTime day error = %v", err)
	}
	if _, err := parseTencentBarTime("5m", "202605221500"); err != nil {
		t.Fatalf("parseTencentBarTime 5m error = %v", err)
	}
}

func TestParseTencentRow(t *testing.T) {
	row := []any{"2026-05-22", "4096.170", "4112.900", "4120.090", "4067.750", "589459644.000"}
	bar, err := parseTencentRow("000001.SH", "day", "none", row, 4077.28)
	if err != nil {
		t.Fatalf("parseTencentRow() error = %v", err)
	}
	if bar.Symbol != "000001.SH" {
		t.Fatalf("Symbol = %s, want 000001.SH", bar.Symbol)
	}
	if bar.Period != "day" {
		t.Fatalf("Period = %s, want day", bar.Period)
	}
	if bar.AdjustType != "none" {
		t.Fatalf("AdjustType = %s, want none", bar.AdjustType)
	}
	if bar.Source != "tencent" {
		t.Fatalf("Source = %s, want tencent", bar.Source)
	}
	if bar.Volume <= 0 {
		t.Fatalf("Volume = %v, want > 0", bar.Volume)
	}
}
