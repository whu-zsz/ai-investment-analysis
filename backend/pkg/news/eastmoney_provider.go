package news

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var eastmoneyHighlightTagPattern = regexp.MustCompile(`</?em>`)

type EastmoneyProvider struct {
	client *http.Client
}

type eastmoneyJSONPResponse struct {
	Result struct {
		CMSArticleWebOld []struct {
			Date      string `json:"date"`
			Title     string `json:"title"`
			Content   string `json:"content"`
			MediaName string `json:"mediaName"`
			URL       string `json:"url"`
		} `json:"cmsArticleWebOld"`
	} `json:"result"`
}

func NewEastmoneyProvider(client *http.Client) *EastmoneyProvider {
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	return &EastmoneyProvider{client: client}
}

func (p *EastmoneyProvider) Name() string {
	return "eastmoney_search"
}

func (p *EastmoneyProvider) FetchByStock(ctx context.Context, symbol, assetName string) ([]Item, error) {
	keyword := strings.TrimSpace(assetName)
	if keyword == "" {
		keyword = strings.TrimSuffix(strings.TrimSuffix(strings.ToUpper(strings.TrimSpace(symbol)), ".SH"), ".SZ")
	}
	payload := map[string]any{
		"uid":           "",
		"keyword":       keyword,
		"type":          []string{"cmsArticleWebOld"},
		"client":        "web",
		"clientType":    "web",
		"clientVersion": "curr",
		"param": map[string]any{
			"cmsArticleWebOld": map[string]any{
				"searchScope": "default",
				"sort":        "default",
				"pageIndex":   1,
				"pageSize":    10,
				"preTag":      "<em>",
				"postTag":     "</em>",
			},
		},
	}
	paramJSON, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("cb", "jQuery11240987654321_"+fmt.Sprintf("%d", time.Now().UnixMilli()))
	query.Set("param", string(paramJSON))
	endpoint := "https://search-api-web.eastmoney.com/search/jsonp?" + query.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://so.eastmoney.com/news/s?keyword="+url.QueryEscape(keyword))
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	body := strings.TrimSpace(string(raw))
	start := strings.IndexByte(body, '(')
	end := strings.LastIndexByte(body, ')')
	if start < 0 || end <= start {
		return nil, fmt.Errorf("invalid eastmoney jsonp response")
	}
	jsonText := body[start+1 : end]
	var parsed eastmoneyJSONPResponse
	if err := json.Unmarshal([]byte(jsonText), &parsed); err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(parsed.Result.CMSArticleWebOld))
	for _, article := range parsed.Result.CMSArticleWebOld {
		title := StripHTML(eastmoneyHighlightTagPattern.ReplaceAllString(article.Title, ""))
		summary := StripHTML(eastmoneyHighlightTagPattern.ReplaceAllString(article.Content, ""))
		items = append(items, Item{
			Title:       title,
			Summary:     summary,
			Source:      NormalizeWhitespace(article.MediaName),
			URL:         strings.TrimSpace(article.URL),
			PublishedAt: ParseNewsTime(article.Date),
			Provider:    p.Name(),
		})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("eastmoney search returned empty feed")
	}
	return items, nil
}
