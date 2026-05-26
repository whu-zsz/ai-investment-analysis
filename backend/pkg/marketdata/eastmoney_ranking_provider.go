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
	"sync"
	"time"
)

const defaultEastmoneyBoardURL = "https://push2.eastmoney.com/api/qt/clist/get"
const sinaFallbackCooldown = time.Hour
const eastmoneyRetryAttempts = 3
const eastmoneyRetryDelay = 400 * time.Millisecond

type eastmoneyRankingProvider struct {
	baseURL   string
	userAgent string
	referer   string
	client    *http.Client
}

type fallbackRankingProvider struct {
	primary         MarketRankingProvider
	fallback        MarketRankingProvider
	mu              sync.RWMutex
	fallbackBlocked time.Time
}

func NewEastmoneyRankingProvider(baseURL, userAgent, referer string, client *http.Client) MarketRankingProvider {
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	if strings.TrimSpace(baseURL) == "" {
		baseURL = defaultEastmoneyBoardURL
	}
	if strings.TrimSpace(userAgent) == "" {
		userAgent = "Mozilla/5.0"
	}
	if strings.TrimSpace(referer) == "" {
		referer = "https://quote.eastmoney.com/center/gridlist.html"
	}
	return &eastmoneyRankingProvider{
		baseURL:   strings.TrimRight(baseURL, "/"),
		userAgent: userAgent,
		referer:   referer,
		client:    client,
	}
}

func NewFallbackRankingProvider(primary, fallback MarketRankingProvider) MarketRankingProvider {
	if primary == nil {
		return fallback
	}
	if fallback == nil {
		return primary
	}
	return &fallbackRankingProvider{primary: primary, fallback: fallback}
}

func (p *fallbackRankingProvider) GetMarketRankings(ctx context.Context, sort string, asc bool, page, size int) ([]MarketRankItem, error) {
	items, err := p.primary.GetMarketRankings(ctx, sort, asc, page, size)
	if err == nil && len(items) > 0 {
		return items, nil
	}
	if blockedErr := p.fallbackBlockedError("primary ranking provider failed", err); blockedErr != nil {
		return nil, blockedErr
	}
	fallbackItems, fallbackErr := p.fallback.GetMarketRankings(ctx, sort, asc, page, size)
	if fallbackErr != nil {
		p.trackFallbackFailure(fallbackErr)
		if err != nil {
			return nil, fmt.Errorf("primary ranking provider failed: %w; fallback failed: %v", err, fallbackErr)
		}
		return nil, fallbackErr
	}
	return fallbackItems, nil
}

func (p *fallbackRankingProvider) GetMarketNodes(ctx context.Context) ([]MarketNode, error) {
	nodes, err := p.primary.GetMarketNodes(ctx)
	if err == nil && len(nodes) > 0 {
		return nodes, nil
	}
	if blockedErr := p.fallbackBlockedError("primary board provider failed", err); blockedErr != nil {
		return nil, blockedErr
	}
	fallbackNodes, fallbackErr := p.fallback.GetMarketNodes(ctx)
	if fallbackErr != nil {
		p.trackFallbackFailure(fallbackErr)
		if err != nil {
			return nil, fmt.Errorf("primary board provider failed: %w; fallback failed: %v", err, fallbackErr)
		}
		return nil, fallbackErr
	}
	return fallbackNodes, nil
}

func (p *fallbackRankingProvider) GetNodeStocks(ctx context.Context, node, sort string, asc bool, page, size int) ([]MarketRankItem, error) {
	items, err := p.primary.GetNodeStocks(ctx, node, sort, asc, page, size)
	if err == nil && len(items) > 0 {
		return items, nil
	}
	if blockedErr := p.fallbackBlockedError("primary node provider failed", err); blockedErr != nil {
		return nil, blockedErr
	}
	fallbackItems, fallbackErr := p.fallback.GetNodeStocks(ctx, node, sort, asc, page, size)
	if fallbackErr != nil {
		p.trackFallbackFailure(fallbackErr)
		if err != nil {
			return nil, fmt.Errorf("primary node provider failed: %w; fallback failed: %v", err, fallbackErr)
		}
		return nil, fallbackErr
	}
	return fallbackItems, nil
}

func (p *fallbackRankingProvider) fallbackBlockedError(prefix string, primaryErr error) error {
	p.mu.RLock()
	blockedUntil := p.fallbackBlocked
	p.mu.RUnlock()
	if blockedUntil.IsZero() || time.Now().After(blockedUntil) {
		return nil
	}
	if primaryErr != nil {
		return fmt.Errorf("%s: %w; sina fallback paused until %s after 456 block", prefix, primaryErr, blockedUntil.Format(time.RFC3339))
	}
	return fmt.Errorf("sina fallback paused until %s after 456 block", blockedUntil.Format(time.RFC3339))
}

func (p *fallbackRankingProvider) trackFallbackFailure(err error) {
	if err == nil || !strings.Contains(err.Error(), "status 456") {
		return
	}
	p.mu.Lock()
	p.fallbackBlocked = time.Now().Add(sinaFallbackCooldown)
	p.mu.Unlock()
}

func (p *eastmoneyRankingProvider) GetMarketRankings(ctx context.Context, sort string, asc bool, page, size int) ([]MarketRankItem, error) {
	query := p.baseQuery(page, size, sort, asc)
	query.Set("fs", "m:0+t:6,m:0+t:80,m:1+t:2,m:1+t:23,m:0+t:81+s:2048")
	query.Set("fields", strings.Join([]string{
		"f12", "f14", "f2", "f3", "f4", "f5", "f6", "f8", "f15", "f16", "f17", "f18", "f20", "f21",
	}, ","))
	return p.fetchRankItems(ctx, query, false)
}

func (p *eastmoneyRankingProvider) GetMarketNodes(ctx context.Context) ([]MarketNode, error) {
	nodes := make([]MarketNode, 0, 1024)
	appendNodes := func(boardType, fs string) error {
		for page := 1; page <= 10; page++ {
			query := p.baseQuery(page, 200, "changepercent", false)
			query.Set("fs", fs)
			query.Set("fields", "f12,f14")
			body, err := p.doRequest(ctx, query)
			if err != nil {
				return err
			}
			var payload eastmoneyResponseEnvelope
			if err := json.Unmarshal(body, &payload); err != nil {
				return fmt.Errorf("failed to parse eastmoney nodes response: %w", err)
			}
			if len(payload.Data.Diff) == 0 {
				break
			}
			for _, item := range payload.Data.Diff {
				code := strings.TrimSpace(item.Code)
				name := strings.TrimSpace(item.Name)
				if code == "" || name == "" {
					continue
				}
				nodes = append(nodes, MarketNode{Type: boardType, Name: name, Node: code})
			}
			if len(payload.Data.Diff) < 200 {
				break
			}
		}
		return nil
	}
	if err := appendNodes("industry", "m:90+t:2"); err != nil {
		return nil, err
	}
	if err := appendNodes("concept", "m:90+t:3"); err != nil {
		return nil, err
	}
	return nodes, nil
}

func (p *eastmoneyRankingProvider) GetNodeStocks(ctx context.Context, node, sort string, asc bool, page, size int) ([]MarketRankItem, error) {
	trimmedNode := strings.TrimSpace(node)
	if trimmedNode == "" {
		return nil, fmt.Errorf("node is required")
	}
	query := p.baseQuery(page, size, sort, asc)
	query.Set("fields", strings.Join([]string{
		"f12", "f14", "f2", "f3", "f4", "f5", "f6", "f8", "f15", "f16", "f17", "f18", "f20", "f21",
	}, ","))
	switch {
	case strings.HasPrefix(trimmedNode, "BK"):
		query.Set("fs", "b:"+trimmedNode)
		return p.fetchRankItems(ctx, query, true)
	case strings.HasPrefix(trimmedNode, "m:"):
		query.Set("fs", trimmedNode)
		return p.fetchRankItems(ctx, query, false)
	default:
		return nil, fmt.Errorf("unsupported eastmoney node: %s", node)
	}
}

func (p *eastmoneyRankingProvider) baseQuery(page, size int, sort string, asc bool) url.Values {
	if page <= 0 {
		page = 1
	}
	if size <= 0 {
		size = 100
	}
	if size > 200 {
		size = 200
	}
	query := url.Values{}
	query.Set("pn", strconv.Itoa(page))
	query.Set("pz", strconv.Itoa(size))
	if asc {
		query.Set("po", "0")
	} else {
		query.Set("po", "1")
	}
	query.Set("np", "1")
	query.Set("ut", "bd1d9ddb04089700cf9c27f6f7426281")
	query.Set("fltt", "2")
	query.Set("invt", "2")
	query.Set("fid", eastmoneyRankingSortField(sort))
	return query
}

func eastmoneyRankingSortField(sort string) string {
	switch strings.ToLower(strings.TrimSpace(sort)) {
	case "symbol", "code":
		return "f12"
	case "amount", "turnover":
		return "f6"
	case "volume":
		return "f5"
	case "turnoverratio":
		return "f8"
	case "mktcap", "marketcap":
		return "f20"
	case "nmc", "floatcap":
		return "f21"
	case "changepercent", "":
		return "f3"
	default:
		return "f3"
	}
}

func (p *eastmoneyRankingProvider) fetchRankItems(ctx context.Context, query url.Values, boardMode bool) ([]MarketRankItem, error) {
	body, err := p.doRequest(ctx, query)
	if err != nil {
		return nil, err
	}
	var payload eastmoneyResponseEnvelope
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, fmt.Errorf("failed to parse eastmoney ranking response: %w", err)
	}
	items := make([]MarketRankItem, 0, len(payload.Data.Diff))
	for _, item := range payload.Data.Diff {
		converted := convertEastmoneyRankItem(item, boardMode)
		if converted.Symbol == "" && !boardMode {
			continue
		}
		items = append(items, converted)
	}
	return items, nil
}

func (p *eastmoneyRankingProvider) doRequest(ctx context.Context, query url.Values) ([]byte, error) {
	var lastErr error
	for attempt := 1; attempt <= eastmoneyRetryAttempts; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"?"+query.Encode(), nil)
		if err != nil {
			return nil, err
		}
		req.Close = true
		req.Header.Set("User-Agent", p.userAgent)
		req.Header.Set("Referer", p.referer)
		req.Header.Set("Connection", "close")

		resp, err := p.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to fetch eastmoney ranking data: %w", err)
		} else {
			body, readErr := io.ReadAll(resp.Body)
			resp.Body.Close()
			if readErr != nil {
				lastErr = fmt.Errorf("failed to read eastmoney ranking response: %w", readErr)
			} else if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("eastmoney ranking request failed with status %d: %s", resp.StatusCode, string(body))
			} else {
				return body, nil
			}
		}

		if !shouldRetryEastmoneyError(lastErr) || attempt == eastmoneyRetryAttempts {
			break
		}
		if err := sleepWithContext(ctx, eastmoneyRetryDelay*time.Duration(attempt)); err != nil {
			return nil, err
		}
	}
	return nil, lastErr
}

func shouldRetryEastmoneyError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "eof") ||
		strings.Contains(message, "unexpected eof") ||
		strings.Contains(message, "connection reset") ||
		strings.Contains(message, "timeout")
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func convertEastmoneyRankItem(item eastmoneyQuoteItem, boardMode bool) MarketRankItem {
	market := "cn_board"
	symbol := strings.TrimSpace(item.Code)
	if !boardMode {
		market, symbol = normalizeEastmoneySymbol(item.Code)
	}
	return MarketRankItem{
		Symbol:         symbol,
		Code:           strings.TrimSpace(item.Code),
		Name:           strings.TrimSpace(item.Name),
		Market:         market,
		LastPrice:      eastmoneyToFloat(item.LastPrice),
		ChangeAmount:   eastmoneyToFloat(item.ChangeAmount),
		ChangePercent:  eastmoneyToFloat(item.ChangePercent),
		OpenPrice:      eastmoneyToFloat(item.OpenPrice),
		HighPrice:      eastmoneyToFloat(item.HighPrice),
		LowPrice:       eastmoneyToFloat(item.LowPrice),
		PrevClose:      eastmoneyToFloat(item.PrevClose),
		Volume:         eastmoneyToFloat(item.Volume),
		Turnover:       eastmoneyToFloat(item.Turnover),
		TurnoverRate:   eastmoneyToFloat(item.TurnoverRate),
		TotalMarketCap: eastmoneyToFloat(item.TotalMarketCap),
		FloatMarketCap: eastmoneyToFloat(item.FloatMarketCap),
		SnapshotTime:   time.Now().Truncate(time.Minute),
		Source:         "eastmoney",
	}
}
