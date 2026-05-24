package news

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

var htmlTagPattern = regexp.MustCompile(`<[^>]+>`)

func NormalizeWhitespace(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	return strings.Join(strings.Fields(trimmed), " ")
}

func StripHTML(value string) string {
	cleaned := htmlTagPattern.ReplaceAllString(value, " ")
	cleaned = strings.NewReplacer("&nbsp;", " ", "&amp;", "&", "&lt;", "<", "&gt;", ">", "&#39;", "'", "&quot;", `"`).Replace(cleaned)
	return NormalizeWhitespace(cleaned)
}

func BuildKeywords(symbol, assetName string) []string {
	result := make([]string, 0, 6)
	seen := make(map[string]struct{})
	appendKeyword := func(value string) {
		value = NormalizeWhitespace(strings.TrimSuffix(strings.TrimSuffix(strings.ToUpper(value), ".SH"), ".SZ"))
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	appendKeyword(symbol)
	appendKeyword(strings.TrimSuffix(strings.TrimSuffix(symbol, ".SH"), ".SZ"))
	appendKeyword(assetName)
	appendKeyword(strings.ReplaceAll(assetName, " ", ""))
	return result
}

func ScoreItem(item Item, keywords []string) int {
	content := strings.ToUpper(NormalizeWhitespace(item.Title + " " + item.Summary))
	score := 0
	for _, keyword := range keywords {
		upperKeyword := strings.ToUpper(keyword)
		if upperKeyword == "" {
			continue
		}
		if strings.Contains(content, upperKeyword) {
			score += 2
			if strings.Contains(strings.ToUpper(item.Title), upperKeyword) {
				score += 4
			}
		}
	}
	if !item.PublishedAt.IsZero() {
		hours := time.Since(item.PublishedAt).Hours()
		switch {
		case hours <= 12:
			score += 8
		case hours <= 24:
			score += 6
		case hours <= 72:
			score += 4
		case hours <= 168:
			score += 2
		}
	}
	return score
}

func MergeAndRankItems(items []Item, keywords []string, limit int) []Item {
	if limit <= 0 {
		limit = 8
	}
	seen := make(map[string]struct{}, len(items))
	result := make([]Item, 0, len(items))
	for _, item := range items {
		item.Title = NormalizeWhitespace(item.Title)
		item.Summary = StripHTML(item.Summary)
		item.Source = NormalizeWhitespace(item.Source)
		item.URL = strings.TrimSpace(item.URL)
		item.Provider = NormalizeWhitespace(item.Provider)
		if item.Title == "" || item.URL == "" {
			continue
		}
		item.Score = ScoreItem(item, keywords)
		if item.Score <= 0 {
			continue
		}
		key := item.URL
		if key == "" {
			key = item.Title
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		iTime := result[i].PublishedAt
		jTime := result[j].PublishedAt
		if !iTime.IsZero() && !jTime.IsZero() {
			iHours := time.Since(iTime).Hours()
			jHours := time.Since(jTime).Hours()
			if absFloat(iHours-jHours) > 24 {
				return iTime.After(jTime)
			}
		}
		if result[i].Score == result[j].Score {
			return result[i].PublishedAt.After(result[j].PublishedAt)
		}
		return result[i].Score > result[j].Score
	})
	if len(result) > limit {
		result = result[:limit]
	}
	return result
}

func BuildCoverage(successProviders []string, failedProviders []string, itemCount int) string {
	parts := []string{fmt.Sprintf("共筛选出 %d 条相关新闻", itemCount)}
	if len(successProviders) > 0 {
		parts = append(parts, "成功来源："+strings.Join(successProviders, "、"))
	}
	if len(failedProviders) > 0 {
		parts = append(parts, "失败来源："+strings.Join(failedProviders, "、"))
	}
	return strings.Join(parts, "；")
}

func BuildSummary(items []Item) string {
	if len(items) == 0 {
		return ""
	}
	parts := make([]string, 0, minInt(len(items), 4))
	for _, item := range items[:minInt(len(items), 4)] {
		publishedAt := ""
		if !item.PublishedAt.IsZero() {
			publishedAt = item.PublishedAt.Format("2006-01-02 15:04")
		}
		segment := item.Title
		meta := make([]string, 0, 3)
		if item.Source != "" {
			meta = append(meta, item.Source)
		}
		if publishedAt != "" {
			meta = append(meta, publishedAt)
		}
		if len(meta) > 0 {
			segment += "（" + strings.Join(meta, "，") + "）"
		}
		parts = append(parts, segment)
	}
	return strings.Join(parts, "；")
}

func ParseNewsTime(value string) time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}
	}
	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
		"Mon, 02 Jan 2006 15:04:05 MST",
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006/01/02 15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed
		}
	}
	return time.Time{}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func absFloat(value float64) float64 {
	if value < 0 {
		return -value
	}
	return value
}
