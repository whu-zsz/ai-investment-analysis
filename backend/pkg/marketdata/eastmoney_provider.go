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
	Amplitude     any    `json:"f7"`
	TurnoverRate  any    `json:"f8"`
	VolumeRatio   any    `json:"f10"`
	OpenPrice     any    `json:"f17"`
	HighPrice     any    `json:"f15"`
	LowPrice      any    `json:"f16"`
	PrevClose     any    `json:"f18"`
	TotalMarketCap any   `json:"f20"`
	FloatMarketCap any   `json:"f21"`
	F50           any    `json:"f50"`
	F51           any    `json:"f51"`
	F52           any    `json:"f52"`
	F55           any    `json:"f55"`
	F71           any    `json:"f71"`
	F84           any    `json:"f84"`
	F85           any    `json:"f85"`
	F116          any    `json:"f116"`
	F117          any    `json:"f117"`
	F127          any    `json:"f127"`
	F128          any    `json:"f128"`
	F129          any    `json:"f129"`
}

type eastmoneyStockDetailEnvelope struct {
	Data map[string]any `json:"data"`
}

type eastmoneyKlineEnvelope struct {
	Data *struct {
		Code   string   `json:"code"`
		Name   string   `json:"name"`
		Market int      `json:"market"`
		Klines []string `json:"klines"`
	} `json:"data"`
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

	respBody, err := p.doRequest(ctx, p.quoteEndpoint(), query)
	if err != nil {
		return nil, err
	}

	var payload eastmoneyResponseEnvelope
	if err := json.Unmarshal(respBody, &payload); err != nil {
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

func (p *eastmoneyProvider) GetStockDetail(ctx context.Context, symbol string) (*StockDetail, error) {
	normalized := normalizeProviderSymbol(symbol)
	if normalized == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	secids, err := buildEastmoneySecIDs([]string{normalized})
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("pn", "1")
	query.Set("pz", "1")
	query.Set("po", "1")
	query.Set("np", "1")
	query.Set("fltt", "2")
	query.Set("invt", "2")
	query.Set("fid", "f3")
	query.Set("fields", strings.Join([]string{
		"f12", "f14", "f2", "f3", "f4", "f5", "f6", "f7", "f8", "f10", "f17", "f15", "f16", "f18", "f20", "f21",
		"f50", "f51", "f52", "f55", "f71", "f84", "f85", "f116", "f117", "f127", "f128", "f129",
	}, ","))
	query.Set("secids", secids[0])

	respBody, err := p.doRequest(ctx, p.quoteEndpoint(), query)
	if err != nil {
		return nil, err
	}

	var payload eastmoneyResponseEnvelope
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse eastmoney detail response: %w", err)
	}
	if len(payload.Data.Diff) == 0 {
		return nil, fmt.Errorf("eastmoney returned empty stock detail")
	}

	item := payload.Data.Diff[0]
	market, normalizedSymbol := normalizeEastmoneySymbol(item.Code)
	if normalizedSymbol == "" {
		normalizedSymbol = normalized
	}
	if market == "" {
		market = marketFromSymbol(normalizedSymbol)
	}

	return buildStockDetailFromQuoteItem(item, normalizedSymbol, market), nil
}

func (p *eastmoneyProvider) GetKlines(ctx context.Context, symbol, period, adjust string, limit int) ([]KlineBar, error) {
	normalized := normalizeProviderSymbol(symbol)
	if normalized == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if limit <= 0 {
		limit = 60
	}
	klt, err := eastmoneyKlineType(period)
	if err != nil {
		return nil, err
	}
	fqt, err := eastmoneyAdjustType(adjust)
	if err != nil {
		return nil, err
	}

	secids, err := buildEastmoneySecIDs([]string{normalized})
	if err != nil {
		return nil, err
	}

	query := url.Values{}
	query.Set("secid", secids[0])
	query.Set("klt", klt)
	query.Set("fqt", fqt)
	query.Set("lmt", strconv.Itoa(limit))
	query.Set("end", "20500101")
	query.Set("fields1", "f1,f2,f3,f4,f5,f6")
	query.Set("fields2", "f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61")

	respBody, err := p.doRequest(ctx, p.klineEndpoint(), query)
	if err != nil {
		return nil, err
	}

	var payload eastmoneyKlineEnvelope
	if err := json.Unmarshal(respBody, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse eastmoney kline response: %w", err)
	}
	if payload.Data == nil || len(payload.Data.Klines) == 0 {
		return nil, fmt.Errorf("eastmoney returned empty klines")
	}

	bars := make([]KlineBar, 0, len(payload.Data.Klines))
	for _, line := range payload.Data.Klines {
		bar, err := parseEastmoneyKlineLine(normalized, normalizePeriod(period), normalizeAdjust(adjust), line)
		if err != nil {
			continue
		}
		bar.Source = "eastmoney"
		bars = append(bars, bar)
	}
	if len(bars) == 0 {
		return nil, fmt.Errorf("eastmoney returned no valid klines")
	}
	return bars, nil
}

func (p *eastmoneyProvider) quoteEndpoint() string {
	if strings.TrimSpace(p.baseURL) != "" {
		return p.baseURL
	}
	return "https://push2.eastmoney.com/api/qt/ulist.np/get"
}

func (p *eastmoneyProvider) detailEndpoint() string {
	return buildEastmoneyEndpoint(p.quoteEndpoint(), false, "/api/qt/stock/get")
}

func (p *eastmoneyProvider) klineEndpoint() string {
	return buildEastmoneyEndpoint(p.quoteEndpoint(), true, "/api/qt/stock/kline/get")
}

func buildEastmoneyEndpoint(base string, historical bool, path string) string {
	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		if historical {
			return "https://push2his.eastmoney.com" + path
		}
		return "https://push2.eastmoney.com" + path
	}
	if historical {
		host := parsed.Host
		if strings.HasPrefix(host, "push2.") {
			host = strings.Replace(host, "push2.", "push2his.", 1)
		} else if !strings.HasPrefix(host, "push2his.") {
			host = "push2his.eastmoney.com"
		}
		parsed.Host = host
	}
	parsed.Path = path
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func (p *eastmoneyProvider) doRequest(ctx context.Context, endpoint string, query url.Values) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
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
		return nil, fmt.Errorf("failed to fetch eastmoney data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read eastmoney response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("eastmoney request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func buildEastmoneySecIDs(symbols []string) ([]string, error) {
	secids := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		normalized := normalizeProviderSymbol(symbol)
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
	if trimmed == "000001" || trimmed == "000300" {
		return market, trimmed + ".SH"
	}
	if strings.HasPrefix(trimmed, "6") {
		return market, trimmed + ".SH"
	}
	return market, trimmed + ".SZ"
}

func normalizeProviderSymbol(value string) string {
	trimmed := strings.ToUpper(strings.TrimSpace(value))
	if trimmed == "" {
		return ""
	}
	if trimmed == "000001.SH" || trimmed == "000001.SZ" {
		return "000001.SH"
	}
	if trimmed == "000300.SH" || trimmed == "000300.SZ" {
		return "000300.SH"
	}
	if strings.HasSuffix(trimmed, ".SH") || strings.HasSuffix(trimmed, ".SZ") {
		return trimmed
	}
	if len(trimmed) == 6 {
		if trimmed == "000001" || trimmed == "000300" {
			return trimmed + ".SH"
		}
		if strings.HasPrefix(trimmed, "6") {
			return trimmed + ".SH"
		}
		return trimmed + ".SZ"
	}
	return trimmed
}

func marketFromSymbol(symbol string) string {
	if strings.HasPrefix(symbol, "399") || strings.HasPrefix(symbol, "000300") || strings.HasPrefix(symbol, "000001") {
		return "cn_index"
	}
	return "cn_stock"
}

func eastmoneyKlineType(period string) (string, error) {
	switch normalizePeriod(period) {
	case "1m":
		return "1", nil
	case "5m":
		return "5", nil
	case "15m":
		return "15", nil
	case "30m":
		return "30", nil
	case "60m":
		return "60", nil
	case "week":
		return "102", nil
	case "month":
		return "103", nil
	case "day":
		return "101", nil
	default:
		return "", fmt.Errorf("unsupported kline period: %s", period)
	}
}

func eastmoneyAdjustType(adjust string) (string, error) {
	switch normalizeAdjust(adjust) {
	case "none":
		return "0", nil
	case "hfq":
		return "2", nil
	case "qfq":
		return "1", nil
	default:
		return "", fmt.Errorf("unsupported adjust type: %s", adjust)
	}
}

func normalizePeriod(period string) string {
	switch strings.ToLower(strings.TrimSpace(period)) {
	case "1m", "5m", "15m", "30m", "60m", "week", "month":
		return strings.ToLower(strings.TrimSpace(period))
	case "daily", "d", "day", "":
		return "day"
	default:
		return strings.ToLower(strings.TrimSpace(period))
	}
}

func normalizeAdjust(adjust string) string {
	switch strings.ToLower(strings.TrimSpace(adjust)) {
	case "", "qfq", "forward":
		return "qfq"
	case "hfq", "backward":
		return "hfq"
	case "none", "raw":
		return "none"
	default:
		return strings.ToLower(strings.TrimSpace(adjust))
	}
}

func parseEastmoneyKlineLine(symbol, period, adjust, line string) (KlineBar, error) {
	parts := strings.Split(line, ",")
	if len(parts) < 11 {
		return KlineBar{}, fmt.Errorf("invalid kline line")
	}
	barTime, err := parseEastmoneyBarTime(period, parts[0])
	if err != nil {
		return KlineBar{}, err
	}
	return KlineBar{
		Symbol:        symbol,
		Period:        period,
		AdjustType:    adjust,
		BarTime:       barTime,
		Open:          parseFloat(parts[1]),
		Close:         parseFloat(parts[2]),
		High:          parseFloat(parts[3]),
		Low:           parseFloat(parts[4]),
		Volume:        parseFloat(parts[5]),
		Amount:        parseFloat(parts[6]),
		Amplitude:     parseFloat(parts[7]),
		ChangePercent: parseFloat(parts[8]),
		ChangeAmount:  parseFloat(parts[9]),
		TurnoverRate:  parseFloat(parts[10]),
	}, nil
}

func buildStockDetailFromQuoteItem(item eastmoneyQuoteItem, symbol, market string) *StockDetail {
	lastPrice := eastmoneyToFloat(item.LastPrice)
	prevClose := eastmoneyToFloat(item.PrevClose)
	changeAmount := eastmoneyToFloat(item.ChangeAmount)
	changePercent := eastmoneyToFloat(item.ChangePercent)
	openPrice := eastmoneyToFloat(item.OpenPrice)
	highPrice := eastmoneyToFloat(item.HighPrice)
	lowPrice := eastmoneyToFloat(item.LowPrice)
	volume := eastmoneyToFloat(item.Volume)
	turnover := eastmoneyToFloat(item.Turnover)
	if changeAmount == 0 && lastPrice != 0 && prevClose != 0 {
		changeAmount = Round(lastPrice - prevClose)
	}
	if changePercent == 0 && changeAmount != 0 && prevClose != 0 {
		changePercent = Round(changeAmount / prevClose * 100)
	}

	amplitude := eastmoneyToFloat(item.Amplitude)
	turnoverRate := eastmoneyToFloat(item.TurnoverRate)
	volumeRatio := eastmoneyToFloat(item.VolumeRatio)
	totalMarketCap := eastmoneyToFloat(item.TotalMarketCap)
	floatMarketCap := eastmoneyToFloat(item.FloatMarketCap)
	totalShares := eastmoneyToFloat(item.F84)
	floatShares := eastmoneyToFloat(item.F85)
	averagePrice := eastmoneyToFloat(item.F71)
	if averagePrice == 0 && volume != 0 {
		averagePrice = Round(turnover / volume)
	}
	if averagePrice == 0 {
		averagePrice = lastPrice
	}

	limitUp := 0.0
	limitDown := 0.0
	if prevClose != 0 {
		limitUp = Round(prevClose * 1.1)
		limitDown = Round(prevClose * 0.9)
	}
	if market == "cn_index" {
		limitUp = 0
		limitDown = 0
		turnoverRate = 0
		volumeRatio = 0
		floatShares = 0
		totalShares = 0
		totalMarketCap = 0
		floatMarketCap = 0
	}

	return &StockDetail{
		Symbol:         symbol,
		Name:           fallbackString(strings.TrimSpace(item.Name), DefaultName(symbol)),
		Market:         fallbackString(strings.TrimSpace(market), marketFromSymbol(symbol)),
		LastPrice:      Round(lastPrice),
		OpenPrice:      Round(openPrice),
		HighPrice:      Round(highPrice),
		LowPrice:       Round(lowPrice),
		PrevClose:      Round(prevClose),
		ChangeAmount:   Round(changeAmount),
		ChangePercent:  Round(changePercent),
		Volume:         Round(volume),
		Turnover:       Round(turnover),
		VolumeRatio:    Round(volumeRatio),
		TurnoverRate:   Round(turnoverRate),
		Amplitude:      Round(amplitude),
		LimitUp:        limitUp,
		LimitDown:      limitDown,
		AveragePrice:   Round(averagePrice),
		TotalShares:    Round(totalShares),
		FloatShares:    Round(floatShares),
		TotalMarketCap: Round(totalMarketCap),
		FloatMarketCap: Round(floatMarketCap),
		Industry:       fallbackString(stringValue(item.F127), ""),
		Region:         fallbackString(stringValue(item.F128), ""),
		Concepts:       splitConcepts(stringValue(item.F129)),
		Source:         "eastmoney",
		FetchedAt:      time.Now().Truncate(time.Minute),
	}
}

func fallbackString(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed != "" && trimmed != "-" {
		return trimmed
	}
	return strings.TrimSpace(fallback)
}

func parseEastmoneyBarTime(period, value string) (time.Time, error) {
	layout := "2006-01-02"
	if strings.HasSuffix(normalizePeriod(period), "m") {
		layout = "2006-01-02 15:04"
	}
	return time.ParseInLocation(layout, strings.TrimSpace(value), time.Local)
}

func stringValue(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return strings.TrimSpace(strconv.FormatFloat(v, 'f', -1, 64))
	case json.Number:
		return strings.TrimSpace(v.String())
	default:
		return ""
	}
}

func parseFloat(value string) float64 {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0
	}
	return parsed
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
	case json.Number:
		parsed, err := v.Float64()
		if err == nil {
			return parsed
		}
	}
	return 0
}

func eastmoneyPriceToFloat(value any) float64 {
	return eastmoneyToFloat(value) / 100
}

func eastmoneyRateToFloat(value any) float64 {
	return eastmoneyToFloat(value) / 100
}

func splitConcepts(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{}
	}
	parts := strings.Split(raw, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		result = append(result, trimmed)
	}
	return result
}
