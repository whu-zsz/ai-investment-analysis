package news

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html/charset"
)

type GoogleNewsProvider struct {
	client *http.Client
}

type googleNewsRSS struct {
	Channel struct {
		Items []struct {
			Title       string `xml:"title"`
			Link        string `xml:"link"`
			Description string `xml:"description"`
			PubDate     string `xml:"pubDate"`
			Source      struct {
				Text string `xml:",chardata"`
			} `xml:"source"`
		} `xml:"item"`
	} `xml:"channel"`
}

func NewGoogleNewsProvider(client *http.Client) *GoogleNewsProvider {
	if client == nil {
		client = &http.Client{Timeout: 12 * time.Second}
	}
	return &GoogleNewsProvider{client: client}
}

func (p *GoogleNewsProvider) Name() string {
	return "google_news_rss"
}

func (p *GoogleNewsProvider) FetchByStock(ctx context.Context, symbol, assetName string) ([]Item, error) {
	query := strings.TrimSpace(assetName)
	if query == "" {
		query = strings.TrimSuffix(strings.TrimSuffix(strings.ToUpper(symbol), ".SH"), ".SZ")
	}
	feedURL := "https://news.google.com/rss/search?q=" + url.QueryEscape(query+" when:3d") + "&hl=zh-CN&gl=CN&ceid=CN:zh-Hans"
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
	var feed googleNewsRSS
	if err := decoder.Decode(&feed); err != nil {
		return nil, err
	}
	items := make([]Item, 0, len(feed.Channel.Items))
	for _, entry := range feed.Channel.Items {
		title := NormalizeWhitespace(entry.Title)
		if idx := strings.LastIndex(title, " - "); idx > 0 {
			title = NormalizeWhitespace(title[:idx])
		}
		items = append(items, Item{
			Title:       title,
			Summary:     entry.Description,
			Source:      NormalizeWhitespace(entry.Source.Text),
			URL:         entry.Link,
			PublishedAt: ParseNewsTime(entry.PubDate),
			Provider:    p.Name(),
		})
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("google news returned empty feed")
	}
	return items, nil
}
