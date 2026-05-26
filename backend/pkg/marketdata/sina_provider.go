package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const defaultSinaMarketCenterURL = "https://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php"

type MarketRankingProvider interface {
	GetMarketRankings(ctx context.Context, sort string, asc bool, page, size int) ([]MarketRankItem, error)
	GetMarketNodes(ctx context.Context) ([]MarketNode, error)
	GetNodeStocks(ctx context.Context, node, sort string, asc bool, page, size int) ([]MarketRankItem, error)
}

type MarketRankItem struct {
	Symbol         string
	Code           string
	Name           string
	Market         string
	LastPrice      float64
	ChangeAmount   float64
	ChangePercent  float64
	OpenPrice      float64
	HighPrice      float64
	LowPrice       float64
	PrevClose      float64
	Volume         float64
	Turnover       float64
	TurnoverRate   float64
	TotalMarketCap float64
	FloatMarketCap float64
	SnapshotTime   time.Time
	Source         string
}

type MarketNode struct {
	Type string
	Name string
	Node string
}

type sinaRankingProvider struct {
	baseURL   string
	userAgent string
	client    *http.Client
	minDelay  time.Duration
	mu        sync.Mutex
	lastCall  time.Time
}

type sinaMarketItem struct {
	Symbol        string      `json:"symbol"`
	Code          string      `json:"code"`
	Name          string      `json:"name"`
	Trade         string      `json:"trade"`
	PriceChange   interface{} `json:"pricechange"`
	ChangePercent interface{} `json:"changepercent"`
	Settlement    string      `json:"settlement"`
	Open          string      `json:"open"`
	High          string      `json:"high"`
	Low           string      `json:"low"`
	Volume        interface{} `json:"volume"`
	Amount        interface{} `json:"amount"`
	TickTime      string      `json:"ticktime"`
	MarketCap     interface{} `json:"mktcap"`
	FloatCap      interface{} `json:"nmc"`
	TurnoverRatio interface{} `json:"turnoverratio"`
}

func NewSinaRankingProvider(baseURL, userAgent string, client *http.Client, minDelay time.Duration) MarketRankingProvider {
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultSinaMarketCenterURL
	}
	if strings.TrimSpace(userAgent) == "" {
		userAgent = "Mozilla/5.0"
	}
	if minDelay < 0 {
		minDelay = 0
	}
	return &sinaRankingProvider{baseURL: strings.TrimRight(baseURL, "/"), userAgent: userAgent, client: client, minDelay: minDelay}
}

func (p *sinaRankingProvider) GetMarketRankings(ctx context.Context, sort string, asc bool, page, size int) ([]MarketRankItem, error) {
	return p.GetNodeStocks(ctx, "hs_a", sort, asc, page, size)
}

func (p *sinaRankingProvider) GetNodeStocks(ctx context.Context, node, sort string, asc bool, page, size int) ([]MarketRankItem, error) {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 80
	}
	if size > 200 {
		size = 200
	}
	if strings.TrimSpace(sort) == "" {
		sort = "changepercent"
	}
	query := url.Values{}
	query.Set("page", strconv.Itoa(page))
	query.Set("num", strconv.Itoa(size))
	query.Set("sort", sort)
	if asc {
		query.Set("asc", "1")
	} else {
		query.Set("asc", "0")
	}
	query.Set("node", strings.TrimSpace(node))
	query.Set("symbol", "")
	query.Set("_s_r_a", "page")

	body, err := p.doRequest(ctx, "/Market_Center.getHQNodeData", query, "https://finance.sina.com.cn/")
	if err != nil {
		return nil, err
	}
	var payload []sinaMarketItem
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse sina ranking response: %w", err)
	}
	items := make([]MarketRankItem, 0, len(payload))
	for _, item := range payload {
		converted := convertSinaMarketItem(item)
		if converted.Symbol == "" {
			continue
		}
		items = append(items, converted)
	}
	return items, nil
}

func (p *sinaRankingProvider) GetMarketNodes(ctx context.Context) ([]MarketNode, error) {
	body, err := p.doRequest(ctx, "/Market_Center.getHQNodes", url.Values{}, "https://finance.sina.com.cn/")
	if err != nil {
		return nil, err
	}
	text := string(body)
	nodes := make([]MarketNode, 0, 160)
	seen := make(map[string]struct{})
	appendMatchedNodes(text, regexp.MustCompile(`\["([^"]+)","","(new_[^"]+|sw\d?_[^"]+|hangye_[^"]+)"(?:,"cn")?\]`), "industry", seen, &nodes)
	appendMatchedNodes(text, regexp.MustCompile(`\["([^"]+)","","(gn_[^"]+)"(?:,"cn")?\]`), "concept", seen, &nodes)
	return nodes, nil
}

func (p *sinaRankingProvider) doRequest(ctx context.Context, path string, query url.Values, referer string) ([]byte, error) {
	if err := p.waitTurn(ctx); err != nil {
		return nil, err
	}
	endpoint := p.baseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", p.userAgent)
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sina market data: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("sina market request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

func (p *sinaRankingProvider) waitTurn(ctx context.Context) error {
	if p.minDelay <= 0 {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now()
	wait := p.minDelay - now.Sub(p.lastCall)
	if wait > 0 {
		timer := time.NewTimer(wait)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
	p.lastCall = time.Now()
	return nil
}

func convertSinaMarketItem(item sinaMarketItem) MarketRankItem {
	symbol := normalizeSinaSymbol(item.Symbol, item.Code)
	return MarketRankItem{
		Symbol:         symbol,
		Code:           item.Code,
		Name:           strings.TrimSpace(item.Name),
		Market:         "cn_stock",
		LastPrice:      parseFloat(item.Trade),
		ChangeAmount:   sinaToFloat(item.PriceChange),
		ChangePercent:  sinaToFloat(item.ChangePercent),
		OpenPrice:      parseFloat(item.Open),
		HighPrice:      parseFloat(item.High),
		LowPrice:       parseFloat(item.Low),
		PrevClose:      parseFloat(item.Settlement),
		Volume:         sinaToFloat(item.Volume),
		Turnover:       sinaToFloat(item.Amount),
		TurnoverRate:   sinaToFloat(item.TurnoverRatio),
		TotalMarketCap: sinaToFloat(item.MarketCap) * 10000,
		FloatMarketCap: sinaToFloat(item.FloatCap) * 10000,
		SnapshotTime:   time.Now().Truncate(time.Minute),
		Source:         "sina",
	}
}

func sinaToFloat(value any) float64 {
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

func normalizeSinaSymbol(rawSymbol, code string) string {
	raw := strings.ToLower(strings.TrimSpace(rawSymbol))
	trimmedCode := strings.TrimSpace(code)
	switch {
	case strings.HasPrefix(raw, "sh"):
		return strings.TrimPrefix(raw, "sh") + ".SH"
	case strings.HasPrefix(raw, "sz"):
		return strings.TrimPrefix(raw, "sz") + ".SZ"
	case strings.HasPrefix(raw, "bj"):
		return strings.TrimPrefix(raw, "bj") + ".BJ"
	case len(trimmedCode) == 6:
		if strings.HasPrefix(trimmedCode, "6") {
			return trimmedCode + ".SH"
		}
		if strings.HasPrefix(trimmedCode, "8") || strings.HasPrefix(trimmedCode, "9") {
			return trimmedCode + ".BJ"
		}
		return trimmedCode + ".SZ"
	default:
		return ""
	}
}

func appendMatchedNodes(text string, re *regexp.Regexp, nodeType string, seen map[string]struct{}, result *[]MarketNode) {
	matches := re.FindAllStringSubmatch(text, -1)
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		name := decodeSinaEscapedText(strings.TrimSpace(match[1]))
		node := strings.TrimSpace(match[2])
		if name == "" || node == "" {
			continue
		}
		if _, ok := seen[node]; ok {
			continue
		}
		seen[node] = struct{}{}
		*result = append(*result, MarketNode{
			Type: nodeType,
			Name: name,
			Node: node,
		})
	}
}

func decodeSinaEscapedText(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	unquoted, err := strconv.Unquote(`"` + strings.ReplaceAll(trimmed, `"`, `\"`) + `"`)
	if err != nil {
		return trimmed
	}
	return strings.TrimSpace(unquoted)
}
