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
	"unicode/utf8"

	"golang.org/x/text/encoding/simplifiedchinese"
)

const defaultTencentBaseURL = "https://web.ifzq.gtimg.cn"
const defaultTencentQuoteBaseURL = "https://qt.gtimg.cn"

type tencentKlineProvider struct {
	baseURL   string
	userAgent string
	client    *http.Client
}

type tencentKlineResponse struct {
	Code int                                   `json:"code"`
	Msg  string                                `json:"msg"`
	Data map[string]map[string]json.RawMessage `json:"data"`
}

type tencentQuoteRecord struct {
	key   string
	parts []string
}

func NewTencentKlineProvider(baseURL, userAgent string, client *http.Client) Provider {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &tencentKlineProvider{
		baseURL:   strings.TrimSpace(baseURL),
		userAgent: strings.TrimSpace(userAgent),
		client:    client,
	}
}

func (p *tencentKlineProvider) GetQuotes(ctx context.Context, symbols []string) ([]Quote, error) {
	normalizedSymbols := make([]string, 0, len(symbols))
	tencentSymbols := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		normalized := normalizeProviderSymbol(symbol)
		if normalized == "" {
			continue
		}
		tencentSymbol, err := toTencentSymbol(normalized)
		if err != nil {
			return nil, err
		}
		normalizedSymbols = append(normalizedSymbols, normalized)
		tencentSymbols = append(tencentSymbols, tencentSymbol)
	}
	if len(normalizedSymbols) == 0 {
		return nil, fmt.Errorf("symbols are required")
	}
	body, err := p.doQuoteRequest(ctx, strings.Join(tencentSymbols, ","))
	if err != nil {
		return nil, err
	}
	records, err := parseTencentQuotePayload(body)
	if err != nil {
		return nil, err
	}
	recordByKey := make(map[string]tencentQuoteRecord, len(records))
	for _, record := range records {
		recordByKey[record.key] = record
	}
	now := time.Now().Truncate(time.Minute)
	quotes := make([]Quote, 0, len(normalizedSymbols))
	for idx, normalized := range normalizedSymbols {
		record, ok := recordByKey[tencentSymbols[idx]]
		if !ok {
			continue
		}
		quote, err := buildTencentQuote(normalized, now, record.parts)
		if err != nil {
			continue
		}
		quotes = append(quotes, quote)
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("tencent returned no valid quotes")
	}
	return quotes, nil
}

func (p *tencentKlineProvider) GetStockDetail(ctx context.Context, symbol string) (*StockDetail, error) {
	normalizedSymbol := normalizeProviderSymbol(symbol)
	if normalizedSymbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	quoteSymbol, err := toTencentSymbol(normalizedSymbol)
	if err != nil {
		return nil, err
	}
	body, err := p.doQuoteRequest(ctx, quoteSymbol)
	if err != nil {
		return nil, err
	}
	return parseTencentStockDetail(normalizedSymbol, body)
}

func (p *tencentKlineProvider) GetKlines(ctx context.Context, symbol, period, adjust string, limit int) ([]KlineBar, error) {
	normalizedSymbol := normalizeProviderSymbol(symbol)
	if normalizedSymbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	if limit <= 0 {
		limit = 60
	}
	normalizedPeriod := normalizePeriod(period)
	normalizedAdjust := normalizeAdjust(adjust)
	tencentSymbol, err := toTencentSymbol(normalizedSymbol)
	if err != nil {
		return nil, err
	}

	endpoint, param, keys, err := p.buildRequest(tencentSymbol, normalizedPeriod, normalizedAdjust, limit)
	if err != nil {
		return nil, err
	}

	body, err := p.doRequest(ctx, endpoint, param)
	if err != nil {
		return nil, err
	}

	var payload tencentKlineResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse tencent kline response: %w", err)
	}
	if payload.Code != 0 {
		if strings.TrimSpace(payload.Msg) == "" {
			return nil, fmt.Errorf("tencent kline request failed with code %d", payload.Code)
		}
		return nil, fmt.Errorf("tencent kline request failed: %s", payload.Msg)
	}
	if len(payload.Data) == 0 {
		return nil, fmt.Errorf("tencent returned empty kline data")
	}
	data, ok := payload.Data[tencentSymbol]
	if !ok {
		return nil, fmt.Errorf("tencent returned no data for symbol %s", normalizedSymbol)
	}

	rows, err := extractTencentRows(data, keys)
	if err != nil {
		return nil, err
	}
	bars, err := parseTencentRows(normalizedSymbol, normalizedPeriod, normalizedAdjust, rows)
	if err != nil {
		return nil, err
	}
	if len(bars) == 0 {
		return nil, fmt.Errorf("tencent returned no valid klines")
	}
	return bars, nil
}

func (p *tencentKlineProvider) buildRequest(symbol, period, adjust string, limit int) (string, string, []string, error) {
	switch period {
	case "day", "week", "month":
		if adjust == "none" {
			return p.rawKlineEndpoint(), fmt.Sprintf("%s,%s,,,%d", symbol, period, limit), []string{period}, nil
		}
		return p.fqKlineEndpoint(), fmt.Sprintf("%s,%s,,,%d,%s", symbol, period, limit, adjust), []string{adjust + period, period}, nil
	case "5m", "15m", "60m":
		minutePeriod := "m" + strings.TrimSuffix(period, "m")
		return p.minuteKlineEndpoint(), fmt.Sprintf("%s,%s,,%d", symbol, minutePeriod, limit), []string{minutePeriod}, nil
	default:
		return "", "", nil, fmt.Errorf("unsupported kline period: %s", period)
	}
}

func (p *tencentKlineProvider) doRequest(ctx context.Context, endpoint, param string) ([]byte, error) {
	query := url.Values{}
	query.Set("param", param)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create tencent request: %w", err)
	}
	if p.userAgent != "" {
		req.Header.Set("User-Agent", p.userAgent)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tencent kline data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tencent kline response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tencent kline request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (p *tencentKlineProvider) doQuoteRequest(ctx context.Context, symbol string) ([]byte, error) {
	query := url.Values{}
	query.Set("q", symbol)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.quoteEndpoint()+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create tencent quote request: %w", err)
	}
	if p.userAgent != "" {
		req.Header.Set("User-Agent", p.userAgent)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch tencent quote data: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read tencent quote response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tencent quote request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (p *tencentKlineProvider) fqKlineEndpoint() string {
	return buildTencentEndpoint(p.baseURL, "/appstock/app/fqkline/get")
}

func (p *tencentKlineProvider) rawKlineEndpoint() string {
	return buildTencentEndpoint(p.baseURL, "/appstock/app/kline/kline")
}

func (p *tencentKlineProvider) minuteKlineEndpoint() string {
	return buildTencentMinuteEndpoint(p.baseURL, "/appstock/app/kline/mkline")
}

func (p *tencentKlineProvider) quoteEndpoint() string {
	base := strings.TrimSpace(p.baseURL)
	if base == "" {
		return defaultTencentQuoteBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return defaultTencentQuoteBaseURL
	}
	host := parsed.Hostname()
	switch {
	case strings.Contains(host, "web.ifzq.gtimg.cn"):
		parsed.Host = strings.Replace(parsed.Host, "web.ifzq.gtimg.cn", "qt.gtimg.cn", 1)
	case strings.Contains(host, "web3.ifzq.gtimg.cn"):
		parsed.Host = strings.Replace(parsed.Host, "web3.ifzq.gtimg.cn", "qt.gtimg.cn", 1)
	case strings.Contains(host, "ifzq.gtimg.cn"):
		parsed.Host = strings.Replace(parsed.Host, "ifzq.gtimg.cn", "qt.gtimg.cn", 1)
	}
	parsed.Path = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func buildTencentMinuteEndpoint(baseURL, path string) string {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		return "https://ifzq.gtimg.cn" + path
	}
	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "https://ifzq.gtimg.cn" + path
	}
	host := parsed.Hostname()
	switch {
	case strings.Contains(host, "web.ifzq.gtimg.cn"):
		parsed.Host = strings.Replace(parsed.Host, "web.ifzq.gtimg.cn", "ifzq.gtimg.cn", 1)
	case strings.Contains(host, "web3.ifzq.gtimg.cn"):
		parsed.Host = strings.Replace(parsed.Host, "web3.ifzq.gtimg.cn", "ifzq.gtimg.cn", 1)
	}
	parsed.Path = path
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func buildTencentEndpoint(baseURL, path string) string {
	base := strings.TrimSpace(baseURL)
	if base == "" {
		base = defaultTencentBaseURL
	}
	parsed, err := url.Parse(base)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return defaultTencentBaseURL + path
	}
	parsed.Path = path
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}

func toTencentSymbol(symbol string) (string, error) {
	normalized := normalizeProviderSymbol(symbol)
	switch {
	case strings.HasSuffix(normalized, ".SH"):
		return "sh" + strings.TrimSuffix(normalized, ".SH"), nil
	case strings.HasSuffix(normalized, ".SZ"):
		return "sz" + strings.TrimSuffix(normalized, ".SZ"), nil
	default:
		return "", fmt.Errorf("unsupported symbol for tencent kline: %s", symbol)
	}
}

func extractTencentRows(data map[string]json.RawMessage, keys []string) ([][]any, error) {
	for _, key := range keys {
		raw, ok := data[key]
		if !ok || len(raw) == 0 {
			continue
		}
		var rows [][]any
		if err := json.Unmarshal(raw, &rows); err == nil && len(rows) > 0 {
			return rows, nil
		}
	}
	return nil, fmt.Errorf("tencent returned no supported kline series")
}

func parseTencentRows(symbol, period, adjust string, rows [][]any) ([]KlineBar, error) {
	bars := make([]KlineBar, 0, len(rows))
	var prevClose float64
	for _, row := range rows {
		bar, err := parseTencentRow(symbol, period, adjust, row, prevClose)
		if err != nil {
			continue
		}
		prevClose = bar.Close
		bars = append(bars, bar)
	}
	if len(bars) == 0 {
		return nil, fmt.Errorf("tencent returned no valid kline rows")
	}
	return bars, nil
}

func parseTencentRow(symbol, period, adjust string, row []any, prevClose float64) (KlineBar, error) {
	if len(row) < 6 {
		return KlineBar{}, fmt.Errorf("invalid tencent kline row")
	}
	barTime, err := parseTencentBarTime(period, stringFromAny(row[0]))
	if err != nil {
		return KlineBar{}, err
	}
	openPrice := parseFloat(stringFromAny(row[1]))
	closePrice := parseFloat(stringFromAny(row[2]))
	highPrice := parseFloat(stringFromAny(row[3]))
	lowPrice := parseFloat(stringFromAny(row[4]))
	volume := parseFloat(stringFromAny(row[5]))
	basePrice := prevClose
	if basePrice == 0 {
		basePrice = openPrice
	}
	changeAmount := 0.0
	changePercent := 0.0
	amplitude := 0.0
	if basePrice != 0 {
		changeAmount = closePrice - basePrice
		changePercent = changeAmount / basePrice * 100
		amplitude = (highPrice - lowPrice) / basePrice * 100
	}
	return KlineBar{
		Symbol:        symbol,
		Period:        period,
		AdjustType:    adjust,
		BarTime:       barTime,
		Open:          Round(openPrice),
		Close:         Round(closePrice),
		High:          Round(highPrice),
		Low:           Round(lowPrice),
		Volume:        Round(volume),
		Amount:        0,
		Amplitude:     Round(amplitude),
		ChangePercent: Round(changePercent),
		ChangeAmount:  Round(changeAmount),
		TurnoverRate:  0,
		Source:        "tencent",
	}, nil
}

func parseTencentBarTime(period, value string) (time.Time, error) {
	trimmed := strings.TrimSpace(value)
	switch period {
	case "5m", "15m", "60m":
		return time.ParseInLocation("200601021504", trimmed, time.Local)
	case "month", "week", "day":
		return time.ParseInLocation("2006-01-02", trimmed, time.Local)
	default:
		return time.Time{}, fmt.Errorf("unsupported tencent period: %s", period)
	}
}

func parseTencentQuotePayload(body []byte) ([]tencentQuoteRecord, error) {
	decoded := decodeTencentPayload(body)
	lines := strings.Split(decoded, ";")
	records := make([]tencentQuoteRecord, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "v_") {
			continue
		}
		eqIdx := strings.Index(trimmed, "=")
		if eqIdx <= 2 {
			continue
		}
		key := strings.TrimSpace(trimmed[2:eqIdx])
		raw := strings.Trim(strings.TrimSpace(trimmed[eqIdx+1:]), "\"")
		parts := strings.Split(raw, "~")
		if len(parts) < 38 {
			continue
		}
		records = append(records, tencentQuoteRecord{key: key, parts: parts})
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("tencent returned invalid quote payload")
	}
	return records, nil
}

func buildTencentQuote(symbol string, defaultTime time.Time, parts []string) (Quote, error) {
	if len(parts) < 38 {
		return Quote{}, fmt.Errorf("tencent returned invalid quote fields")
	}
	market := marketFromSymbol(symbol)
	snapshotTime := defaultTime
	if parsedTime, err := time.ParseInLocation("20060102150405", strings.TrimSpace(parts[30]), time.Local); err == nil {
		snapshotTime = parsedTime
	}
	prevClose := parseFloat(parts[4])
	lastPrice := parseFloat(parts[3])
	changeAmount := parseFloat(parts[31])
	changePercent := parseFloat(parts[32])
	openPrice := parseFloat(parts[5])
	highPrice := parseFloat(parts[33])
	lowPrice := parseFloat(parts[34])
	volume := parseFloat(parts[36])
	turnover := parseFloat(parts[37]) * 10000
	if changeAmount == 0 && lastPrice != 0 && prevClose != 0 {
		changeAmount = Round(lastPrice - prevClose)
	}
	if changePercent == 0 && changeAmount != 0 && prevClose != 0 {
		changePercent = Round(changeAmount / prevClose * 100)
	}
	return Quote{
		Symbol:        symbol,
		Name:          fallbackString(parts[1], DefaultName(symbol)),
		Market:        market,
		SnapshotTime:  snapshotTime,
		LastPrice:     Round(lastPrice),
		ChangeAmount:  Round(changeAmount),
		ChangePercent: Round(changePercent),
		OpenPrice:     Round(openPrice),
		HighPrice:     Round(highPrice),
		LowPrice:      Round(lowPrice),
		PrevClose:     Round(prevClose),
		Volume:        Round(volume),
		Turnover:      Round(turnover),
		Source:        "tencent",
	}, nil
}

func stringFromAny(value any) string {
	switch v := value.(type) {
	case string:
		return strings.TrimSpace(v)
	case float64:
		return strings.TrimSpace(strconv.FormatFloat(v, 'f', -1, 64))
	case json.Number:
		return strings.TrimSpace(v.String())
	default:
		return strings.TrimSpace(fmt.Sprint(v))
	}
}

func parseTencentStockDetail(symbol string, body []byte) (*StockDetail, error) {
	decoded := decodeTencentPayload(body)
	parts := strings.Split(decoded, "~")
	if len(parts) < 74 {
		return nil, fmt.Errorf("tencent returned invalid stock detail payload")
	}
	if strings.TrimSpace(parts[2]) == "" {
		return nil, fmt.Errorf("tencent returned empty stock detail payload")
	}
	name := fallbackString(parts[1], DefaultName(symbol))
	market := marketFromSymbol(symbol)
	prevClose := parseFloat(parts[4])
	lastPrice := parseFloat(parts[3])
	changeAmount := parseFloat(parts[31])
	changePercent := parseFloat(parts[32])
	openPrice := parseFloat(parts[5])
	highPrice := parseFloat(parts[33])
	lowPrice := parseFloat(parts[34])
	volume := parseFloat(parts[36])
	turnover := parseFloat(parts[37]) * 10000
	turnoverRate := parseFloat(parts[38])
	amplitude := parseFloat(parts[43])
	totalMarketCap := parseFloat(parts[44]) * 100000000
	floatMarketCap := parseFloat(parts[45]) * 100000000
	averagePrice := parseFloat(parts[51])
	floatShares := parseFloat(parts[72])
	totalShares := parseFloat(parts[73])
	limitUp := parseFloat(parts[67])
	limitDown := parseFloat(parts[68])
	if limitUp == 0 && prevClose != 0 && market != "cn_index" {
		limitUp = Round(prevClose * 1.1)
	}
	if limitDown == 0 && prevClose != 0 && market != "cn_index" {
		limitDown = Round(prevClose * 0.9)
	}
	if changeAmount == 0 && lastPrice != 0 && prevClose != 0 {
		changeAmount = Round(lastPrice - prevClose)
	}
	if changePercent == 0 && changeAmount != 0 && prevClose != 0 {
		changePercent = Round(changeAmount / prevClose * 100)
	}
	if averagePrice == 0 {
		averagePrice = lastPrice
	}
	fetchedAt := time.Now().Truncate(time.Minute)
	if parsedTime, err := time.ParseInLocation("20060102150405", strings.TrimSpace(parts[30]), time.Local); err == nil {
		fetchedAt = parsedTime
	}
	if market == "cn_index" {
		turnoverRate = 0
		totalMarketCap = 0
		floatMarketCap = 0
		totalShares = 0
		floatShares = 0
		limitUp = 0
		limitDown = 0
	}
	return &StockDetail{
		Symbol:         symbol,
		Name:           name,
		Market:         market,
		LastPrice:      Round(lastPrice),
		OpenPrice:      Round(openPrice),
		HighPrice:      Round(highPrice),
		LowPrice:       Round(lowPrice),
		PrevClose:      Round(prevClose),
		ChangeAmount:   Round(changeAmount),
		ChangePercent:  Round(changePercent),
		Volume:         Round(volume),
		Turnover:       Round(turnover),
		VolumeRatio:    0,
		TurnoverRate:   Round(turnoverRate),
		Amplitude:      Round(amplitude),
		LimitUp:        Round(limitUp),
		LimitDown:      Round(limitDown),
		AveragePrice:   Round(averagePrice),
		TotalShares:    Round(totalShares),
		FloatShares:    Round(floatShares),
		TotalMarketCap: Round(totalMarketCap),
		FloatMarketCap: Round(floatMarketCap),
		Industry:       "",
		Region:         "",
		Concepts:       nil,
		Source:         "tencent",
		FetchedAt:      fetchedAt,
	}, nil
}

func decodeTencentPayload(body []byte) string {
	if utf8.Valid(body) {
		return strings.TrimSpace(string(body))
	}
	decoded, err := simplifiedchinese.GBK.NewDecoder().Bytes(body)
	if err == nil && utf8.Valid(decoded) {
		return strings.TrimSpace(string(decoded))
	}
	return strings.TrimSpace(string(body))
}
