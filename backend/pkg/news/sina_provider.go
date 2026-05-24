package news

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/html/charset"
)

type SinaProvider struct {
	client   *http.Client
	feedURLs []string
}

type sinaRSS struct {
	Channel struct {
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
		} `xml:"item"`
	} `xml:"channel"`
}

func NewSinaProvider(client *http.Client) *SinaProvider {
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	return &SinaProvider{
		client: client,
		feedURLs: []string{
			"https://rss.sina.com.cn/roll/stock/hot_roll.xml",
			"https://rss.sina.com.cn/news/allnews/finance.xml",
			"https://rss.sina.com.cn/finance/hkstock.xml",
		},
	}
}

func (p *SinaProvider) Name() string {
	return "sina_finance_rss"
}

func (p *SinaProvider) FetchByStock(ctx context.Context, symbol, assetName string) ([]Item, error) {
	items := make([]Item, 0, 24)
	for _, feedURL := range p.feedURLs {
		feedItems, err := p.fetchFeed(ctx, feedURL)
		if err != nil {
			continue
		}
		items = append(items, feedItems...)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("failed to fetch sina finance feeds")
	}
	return items, nil
}

func (p *SinaProvider) fetchFeed(ctx context.Context, feedURL string) ([]Item, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = charset.NewReaderLabel
	var feed sinaRSS
	if err := decoder.Decode(&feed); err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(feed.Channel.Items))
	for _, entry := range feed.Channel.Items {
		items = append(items, Item{
			Title:       entry.Title,
			Summary:     entry.Description,
			Source:      "新浪财经",
			URL:         entry.Link,
			PublishedAt: ParseNewsTime(entry.PubDate),
			Provider:    p.Name(),
		})
	}
	return items, nil
}
