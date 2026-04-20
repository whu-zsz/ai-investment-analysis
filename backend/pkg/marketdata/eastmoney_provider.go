package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type eastmoneyProvider struct {
	baseURL   string
	userAgent string
	referer   string
	client    *http.Client
}

type eastmoneyResponseEnvelope struct {
	Data struct {
		Diff []eastmoneyQuoteItem `json:"diff"`
	} `json:"data"`
}

type eastmoneyQuoteItem struct {
	Code          string `json:"f12"`
	Name          string `json:"f14"`
	LastPrice     any    `json:"f2"`
	ChangePercent any    `json:"f3"`
	ChangeAmount  any    `json:"f4"`
	Volume        any    `json:"f5"`
	Turnover      any    `json:"f6"`
	OpenPrice     any    `json:"f17"`
	HighPrice     any    `json:"f15"`
	LowPrice      any    `json:"f16"`
	PrevClose     any    `json:"f18"`
}

func NewEastmoneyProvider(baseURL, userAgent, referer string, client *http.Client) Provider {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &eastmoneyProvider{
		baseURL:   baseURL,
		userAgent: userAgent,
		referer:   referer,
		client:    client,
	}
}

func (p *eastmoneyProvider) GetQuotes(ctx context.Context, symbols []string) ([]Quote, error) {
	if len(symbols) == 0 {
		return nil, fmt.Errorf("symbols are required")
	}

	secids, err := buildEastmoneySecIDs(symbols)
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("pn", "1")
	query.Set("pz", strconv.Itoa(len(symbols)))
	query.Set("po", "1")
	query.Set("np", "1")
	query.Set("fltt", "2")
	query.Set("invt", "2")
	query.Set("fid", "f3")
	query.Set("fields", "f12,f14,f2,f3,f4,f5,f6,f17,f15,f16,f18")
	query.Set("secids", strings.Join(secids, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	if p.userAgent != "" {
		req.Header.Set("User-Agent", p.userAgent)
	}
	if p.referer != "" {
		req.Header.Set("Referer", p.referer)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch eastmoney quotes: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read eastmoney response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("eastmoney request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var payload eastmoneyResponseEnvelope
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse eastmoney response: %w", err)
	}
	if len(payload.Data.Diff) == 0 {
		return nil, fmt.Errorf("eastmoney returned empty quotes")
	}

	now := time.Now().Truncate(time.Minute)
	quotes := make([]Quote, 0, len(payload.Data.Diff))
	for _, item := range payload.Data.Diff {
		market, symbol := normalizeEastmoneySymbol(item.Code)
		quote := Quote{
			Symbol:        symbol,
			Name:          item.Name,
			Market:        market,
			SnapshotTime:  now,
			LastPrice:     eastmoneyToFloat(item.LastPrice),
			ChangeAmount:  eastmoneyToFloat(item.ChangeAmount),
			ChangePercent: eastmoneyToFloat(item.ChangePercent),
			OpenPrice:     eastmoneyToFloat(item.OpenPrice),
			HighPrice:     eastmoneyToFloat(item.HighPrice),
			LowPrice:      eastmoneyToFloat(item.LowPrice),
			PrevClose:     eastmoneyToFloat(item.PrevClose),
			Volume:        eastmoneyToFloat(item.Volume),
			Turnover:      eastmoneyToFloat(item.Turnover),
			Source:        "eastmoney",
		}
		if quote.Symbol == "" {
			continue
		}
		quotes = append(quotes, quote)
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("eastmoney returned no valid quotes")
	}
	return quotes, nil
}

func buildEastmoneySecIDs(symbols []string) ([]string, error) {
	secids := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		switch {
		case strings.HasSuffix(normalized, ".SH"):
			secids = append(secids, "1."+strings.TrimSuffix(normalized, ".SH"))
		case strings.HasSuffix(normalized, ".SZ"):
			secids = append(secids, "0."+strings.TrimSuffix(normalized, ".SZ"))
		default:
			return nil, fmt.Errorf("unsupported symbol for eastmoney: %s", symbol)
		}
	}
	return secids, nil
}

func normalizeEastmoneySymbol(code string) (string, string) {
	trimmed := strings.TrimSpace(code)
	if trimmed == "" {
		return "", ""
	}
	market := "cn_stock"
	if strings.HasPrefix(trimmed, "399") || trimmed == "000300" || trimmed == "000001" {
		market = "cn_index"
	}
	if strings.HasPrefix(trimmed, "6") {
		return market, trimmed + ".SH"
	}
	return market, trimmed + ".SZ"
}

func eastmoneyToFloat(value any) float64 {
	switch v := value.(type) {
	case float64:
		return v
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err == nil {
			return parsed
		}
	}
	return 0
}
