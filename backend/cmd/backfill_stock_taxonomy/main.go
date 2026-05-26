package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/model"

	"gorm.io/gorm"
)

type candidateStock struct {
	Symbol string
	Name   string
	Market string
}

type stockTaxonomy struct {
	Industry string
	Region   string
	Concepts []string
	Source   string
}

type eastmoneyBoardProfilePayload struct {
	Boards []struct {
		BoardName string `json:"BOARD_NAME"`
		IsPrecise string `json:"IS_PRECISE"`
		BoardRank int    `json:"BOARD_RANK"`
	} `json:"ssbk"`
}

func main() {
	var (
		limit        = flag.Int("limit", 0, "max stocks to process, 0 means all")
		delayMS      = flag.Int("delay-ms", 700, "delay between remote metadata requests in milliseconds")
		onlyMissing  = flag.Bool("only-missing", true, "only request remote metadata for stocks missing industry or concepts after local fallback")
		skipRemote   = flag.Bool("skip-remote", false, "skip remote metadata requests and only use local board constituents")
		requestTO    = flag.Int("request-timeout-ms", 6000, "remote metadata request timeout in milliseconds")
		progressStep = flag.Int("progress-step", 100, "progress log interval")
	)
	flag.Parse()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("load config failed: %v", err)
	}

	db, err := config.InitDB(&cfg.Database)
	if err != nil {
		log.Fatalf("init db failed: %v", err)
	}
	defer func() {
		_ = config.CloseDB(db)
	}()

	candidates, err := loadCandidateStocks(db)
	if err != nil {
		log.Fatalf("load candidates failed: %v", err)
	}
	if len(candidates) == 0 {
		log.Println("no candidate stocks found")
		return
	}
	if *limit > 0 && *limit < len(candidates) {
		candidates = candidates[:*limit]
	}

	details, err := loadExistingDetails(db)
	if err != nil {
		log.Fatalf("load existing details failed: %v", err)
	}

	boardTaxonomy, err := buildBoardFallbackTaxonomy(db)
	if err != nil {
		log.Fatalf("build board fallback taxonomy failed: %v", err)
	}

	client := &http.Client{Timeout: time.Duration(*requestTO) * time.Millisecond}
	now := time.Now()
	updated := 0
	remoteUpdated := 0
	skipped := 0
	failed := 0

	for index, candidate := range candidates {
		existing := details[candidate.Symbol]
		merged := cloneOrInitDetail(existing, candidate, now)

		localChanged := applyTaxonomy(merged, boardTaxonomy[candidate.Symbol], false)
		needRemote := !*skipRemote && (!*onlyMissing || needsRemoteTaxonomy(merged))

		remoteChanged := false
		if needRemote {
			remoteTaxonomy, remoteErr := fetchEastmoneyTaxonomy(client, candidate.Symbol)
			if remoteErr != nil {
				failed++
				log.Printf("remote taxonomy failed for %s: %v", candidate.Symbol, remoteErr)
			} else {
				remoteChanged = applyTaxonomy(merged, remoteTaxonomy, true)
				if remoteChanged {
					remoteUpdated++
				}
			}
			if *delayMS > 0 {
				time.Sleep(time.Duration(*delayMS) * time.Millisecond)
			}
		}

		if !localChanged && !remoteChanged && existing != nil {
			skipped++
		} else {
			if err := db.Save(merged).Error; err != nil {
				failed++
				log.Printf("save detail failed for %s: %v", candidate.Symbol, err)
			} else {
				updated++
				details[candidate.Symbol] = merged
			}
		}

		if *progressStep > 0 && (index+1)%*progressStep == 0 {
			log.Printf("progress %d/%d updated=%d remote=%d skipped=%d failed=%d", index+1, len(candidates), updated, remoteUpdated, skipped, failed)
		}
	}

	log.Printf("done total=%d updated=%d remote=%d skipped=%d failed=%d", len(candidates), updated, remoteUpdated, skipped, failed)
}

func loadCandidateStocks(db *gorm.DB) ([]candidateStock, error) {
	candidates := make(map[string]candidateStock, 8192)

	var details []struct {
		Symbol string
		Name   string
		Market string
	}
	if err := db.Model(&model.StockQuoteDetail{}).Select("symbol", "name", "market").Find(&details).Error; err != nil {
		return nil, err
	}
	for _, item := range details {
		symbol := normalizeSymbol(item.Symbol)
		if symbol == "" || inferMarket(symbol) != "cn_stock" {
			continue
		}
		candidates[symbol] = candidateStock{Symbol: symbol, Name: strings.TrimSpace(item.Name), Market: fallbackText(strings.TrimSpace(item.Market), "cn_stock")}
	}

	var latestBatch struct {
		BatchNo string
	}
	if err := db.Model(&model.MarketSnapshot{}).
		Select("batch_no").
		Where("market = ?", "cn_stock").
		Order("created_at DESC, id DESC").
		Limit(1).
		Scan(&latestBatch).Error; err != nil {
		return nil, err
	}
	if strings.TrimSpace(latestBatch.BatchNo) != "" {
		var snapshots []struct {
			Symbol string
			Name   string
			Market string
		}
		if err := db.Model(&model.MarketSnapshot{}).
			Select("symbol", "name", "market").
			Where("market = ? AND batch_no = ?", "cn_stock", latestBatch.BatchNo).
			Find(&snapshots).Error; err != nil {
			return nil, err
		}
		for _, item := range snapshots {
			symbol := normalizeSymbol(item.Symbol)
			if symbol == "" {
				continue
			}
			current := candidates[symbol]
			if current.Symbol == "" {
				candidates[symbol] = candidateStock{Symbol: symbol, Name: strings.TrimSpace(item.Name), Market: fallbackText(strings.TrimSpace(item.Market), "cn_stock")}
				continue
			}
			if current.Name == "" && strings.TrimSpace(item.Name) != "" {
				current.Name = strings.TrimSpace(item.Name)
				candidates[symbol] = current
			}
		}
	}

	var boards []struct {
		Symbol    string
		StockName string
	}
	if err := db.Model(&model.MarketBoardConstituent{}).Select("symbol", "stock_name").Find(&boards).Error; err != nil {
		return nil, err
	}
	for _, item := range boards {
		symbol := normalizeSymbol(item.Symbol)
		if symbol == "" {
			continue
		}
		current := candidates[symbol]
		if current.Symbol == "" {
			candidates[symbol] = candidateStock{Symbol: symbol, Name: strings.TrimSpace(item.StockName), Market: "cn_stock"}
			continue
		}
		if current.Name == "" && strings.TrimSpace(item.StockName) != "" {
			current.Name = strings.TrimSpace(item.StockName)
			candidates[symbol] = current
		}
	}

	list := make([]candidateStock, 0, len(candidates))
	for _, item := range candidates {
		list = append(list, item)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].Symbol < list[j].Symbol })
	return list, nil
}

func loadExistingDetails(db *gorm.DB) (map[string]*model.StockQuoteDetail, error) {
	var rows []model.StockQuoteDetail
	if err := db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]*model.StockQuoteDetail, len(rows))
	for i := range rows {
		row := rows[i]
		result[normalizeSymbol(row.Symbol)] = &row
	}
	return result, nil
}

func buildBoardFallbackTaxonomy(db *gorm.DB) (map[string]stockTaxonomy, error) {
	var rows []model.MarketBoardConstituent
	if err := db.Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[string]stockTaxonomy, 8192)
	for _, row := range rows {
		symbol := normalizeSymbol(row.Symbol)
		if symbol == "" {
			continue
		}
		item := result[symbol]
		name := decodeBoardName(strings.TrimSpace(row.BoardName))
		switch strings.ToLower(strings.TrimSpace(row.BoardType)) {
		case "industry":
			if item.Industry == "" && name != "" {
				item.Industry = name
			}
		case "concept":
			if name != "" {
				item.Concepts = appendUnique(item.Concepts, name)
			}
		}
		item.Source = mergeSource(item.Source, "boardmap")
		result[symbol] = item
	}
	return result, nil
}

func fetchEastmoneyTaxonomy(client *http.Client, symbol string) (stockTaxonomy, error) {
	code := toEastmoneySecurityCode(symbol)
	if code == "" {
		return stockTaxonomy{}, fmt.Errorf("unsupported symbol %s", symbol)
	}

	ctx, cancel := context.WithTimeout(context.Background(), client.Timeout)
	defer cancel()

	query := url.Values{}
	query.Set("code", code)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://emweb.securities.eastmoney.com/PC_HSF10/OperationsRequired/PageAjax?"+query.Encode(), nil)
	if err != nil {
		return stockTaxonomy{}, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://quote.eastmoney.com/")

	resp, err := client.Do(req)
	if err != nil {
		return stockTaxonomy{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return stockTaxonomy{}, err
	}
	if resp.StatusCode != http.StatusOK {
		return stockTaxonomy{}, fmt.Errorf("status %d", resp.StatusCode)
	}

	var payload eastmoneyBoardProfilePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return stockTaxonomy{}, err
	}
	if len(payload.Boards) == 0 {
		return stockTaxonomy{}, nil
	}

	industry := ""
	region := ""
	concepts := make([]string, 0, 8)
	for _, board := range payload.Boards {
		name := strings.TrimSpace(board.BoardName)
		if name == "" {
			continue
		}
		if region == "" {
			if inferredRegion := inferRegionFromBoard(name); inferredRegion != "" {
				region = inferredRegion
				continue
			}
		}
		if industry == "" && isPrimaryIndustryBoard(name, board.BoardRank, board.IsPrecise) {
			industry = name
			continue
		}
		if concept := normalizeBoardConcept(name, industry, board.IsPrecise); concept != "" {
			concepts = appendUnique(concepts, concept)
		}
	}

	return stockTaxonomy{
		Industry: model.NormalizeIndustryLabel(industry),
		Region:   model.NormalizeRegionLabel(region),
		Concepts: model.NormalizeConceptList(concepts),
		Source:   "em-board",
	}, nil
}

func cloneOrInitDetail(existing *model.StockQuoteDetail, candidate candidateStock, now time.Time) *model.StockQuoteDetail {
	if existing != nil {
		cloned := *existing
		if cloned.Name == "" {
			cloned.Name = candidate.Name
		}
		if cloned.Market == "" {
			cloned.Market = candidate.Market
		}
		return &cloned
	}
	return &model.StockQuoteDetail{
		Symbol:    candidate.Symbol,
		Name:      fallbackText(candidate.Name, candidate.Symbol),
		Market:    fallbackText(candidate.Market, "cn_stock"),
		Source:    "taxonomy-script",
		FetchedAt: now,
	}
}

func applyTaxonomy(detail *model.StockQuoteDetail, taxonomy stockTaxonomy, overwrite bool) bool {
	changed := false
	if detail == nil {
		return false
	}
	if overwrite {
		if taxonomy.Industry != "" && taxonomy.Industry != detail.Industry {
			detail.Industry = taxonomy.Industry
			changed = true
		}
		if taxonomy.Region != "" && taxonomy.Region != detail.Region {
			detail.Region = taxonomy.Region
			changed = true
		}
		if len(taxonomy.Concepts) > 0 {
			nextConcepts := strings.Join(model.NormalizeConceptList(taxonomy.Concepts), ",")
			if nextConcepts != detail.Concepts {
				detail.Concepts = nextConcepts
				changed = true
			}
		}
	} else {
		if model.NormalizeIndustryLabel(detail.Industry) == "" && taxonomy.Industry != "" {
			detail.Industry = taxonomy.Industry
			changed = true
		}
		if model.NormalizeRegionLabel(detail.Region) == "" && taxonomy.Region != "" {
			detail.Region = taxonomy.Region
			changed = true
		}
		if len(model.NormalizeConceptList(strings.Split(detail.Concepts, ","))) == 0 && len(taxonomy.Concepts) > 0 {
			detail.Concepts = strings.Join(model.NormalizeConceptList(taxonomy.Concepts), ",")
			changed = true
		}
	}
	if changed {
		detail.Source = mergeSource(detail.Source, taxonomy.Source)
		detail.FetchedAt = time.Now()
	}
	return changed
}

func needsRemoteTaxonomy(detail *model.StockQuoteDetail) bool {
	if detail == nil {
		return true
	}
	if model.NormalizeIndustryLabel(detail.Industry) == "" {
		return true
	}
	if len(model.NormalizeConceptList(strings.Split(detail.Concepts, ","))) == 0 {
		return true
	}
	return false
}

func appendUnique(items []string, value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return items
	}
	for _, item := range items {
		if item == value {
			return items
		}
	}
	return append(items, value)
}

func mergeSource(existing, extra string) string {
	existing = strings.TrimSpace(existing)
	extra = strings.TrimSpace(extra)
	if extra == "" {
		return existing
	}
	if existing == "" {
		return extra
	}
	for _, part := range strings.Split(existing, "+") {
		if strings.TrimSpace(part) == extra {
			return existing
		}
	}
	merged := existing + "+" + extra
	if len(merged) > 32 {
		return existing
	}
	return merged
}

func fallbackText(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func inferMarket(symbol string) string {
	normalized := normalizeSymbol(symbol)
	if normalized == "" {
		return ""
	}
	if strings.HasSuffix(normalized, ".SH") || strings.HasSuffix(normalized, ".SZ") || strings.HasSuffix(normalized, ".BJ") {
		return "cn_stock"
	}
	return ""
}

func normalizeSymbol(symbol string) string {
	text := strings.ToUpper(strings.TrimSpace(symbol))
	if strings.HasSuffix(text, ".SH") || strings.HasSuffix(text, ".SZ") || strings.HasSuffix(text, ".BJ") {
		return text
	}
	return text
}

func toEastmoneySecurityCode(symbol string) string {
	normalized := normalizeSymbol(symbol)
	switch {
	case strings.HasSuffix(normalized, ".SH"):
		return "SH" + strings.TrimSuffix(normalized, ".SH")
	case strings.HasSuffix(normalized, ".SZ"):
		return "SZ" + strings.TrimSuffix(normalized, ".SZ")
	case strings.HasSuffix(normalized, ".BJ"):
		return "BJ" + strings.TrimSuffix(normalized, ".BJ")
	default:
		return ""
	}
}

func inferRegionFromBoard(name string) string {
	if !strings.HasSuffix(name, "板块") {
		return ""
	}
	trimmed := strings.TrimSuffix(name, "板块")
	switch trimmed {
	case "北京", "上海", "广东", "江苏", "浙江", "山东", "四川", "贵州", "安徽", "湖北", "湖南", "福建", "江西", "重庆", "天津", "河北", "河南", "云南", "陕西", "辽宁", "吉林", "黑龙江", "海南", "广西", "山西", "内蒙古", "甘肃", "宁夏", "青海", "新疆", "西藏":
		return trimmed
	default:
		return ""
	}
}

func isPrimaryIndustryBoard(name string, rank int, precise string) bool {
	if precise == "1" || rank > 3 {
		return false
	}
	if strings.Contains(name, "板块") || strings.Contains(name, "概念") || strings.Contains(name, "指数") {
		return false
	}
	if strings.Contains(name, "HS") || strings.Contains(name, "MSCI") || strings.Contains(name, "富时") || strings.Contains(name, "标普") || strings.Contains(name, "央视") {
		return false
	}
	if strings.Contains(name, "股") && !strings.Contains(name, "饮料") {
		return false
	}
	return true
}

func normalizeBoardConcept(name, industry, precise string) string {
	switch {
	case name == "", name == industry:
		return ""
	case inferRegionFromBoard(name) != "":
		return ""
	case strings.Contains(name, "指数"):
		return ""
	case strings.Contains(name, "HS300"), strings.Contains(name, "上证"), strings.Contains(name, "MSCI"), strings.Contains(name, "富时"), strings.Contains(name, "标普"), strings.Contains(name, "央视"):
		return ""
	case strings.Contains(name, "融资融券"), strings.Contains(name, "证金持股"), strings.Contains(name, "机构重仓"), strings.Contains(name, "权重股"), strings.Contains(name, "大盘股"), strings.Contains(name, "行业龙头"):
		return ""
	case precise == "1":
		return name
	case strings.Contains(name, "概念"):
		return name
	default:
		return ""
	}
}

func decodeBoardName(name string) string {
	return strings.TrimSpace(name)
}
