package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"stock-analysis-backend/internal/config"
	marketResponse "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/pkg/marketdata"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type MarketStockService interface {
	GetStockDetail(symbol string, forceRefresh bool) (*marketResponse.MarketStockDetailResponse, error)
	GetStockProfile(symbol string) (*marketResponse.StockProfileResponse, error)
	GetStockKlines(symbol, period, adjust string, limit int, forceRefresh bool) (*marketResponse.MarketStockKlineResponse, error)
}

type stockProfileFetcher func(ctx context.Context, pythonPath, scriptPath string, symbol string) (*marketResponse.StockCompanyProfileResponse, error)

type marketStockService struct {
	provider              marketdata.Provider
	supplementProvider    marketdata.Provider
	rankingProvider       marketdata.MarketRankingProvider
	httpClient            *http.Client
	detailRepo            repository.StockQuoteDetailRepository
	klineRepo             repository.StockKlineRepository
	boardConstituentRepo  repository.MarketBoardConstituentRepository
	marketConfig          config.MarketConfig
	akshareProfileFetcher stockProfileFetcher
}

const minExternalRefreshInterval = 10 * time.Second

func NewMarketStockService(
	marketConfig config.MarketConfig,
	provider marketdata.Provider,
	rankingProvider marketdata.MarketRankingProvider,
	detailRepo repository.StockQuoteDetailRepository,
	klineRepo repository.StockKlineRepository,
	boardConstituentRepo repository.MarketBoardConstituentRepository,
) MarketStockService {
	return &marketStockService{
		provider: provider,
		supplementProvider: marketdata.NewEastmoneyProvider(
			"",
			"Mozilla/5.0",
			"https://quote.eastmoney.com/center/gridlist.html",
			&http.Client{Timeout: 5 * time.Second},
		),
		rankingProvider:       rankingProvider,
		httpClient:            &http.Client{Timeout: 6 * time.Second},
		detailRepo:            detailRepo,
		klineRepo:             klineRepo,
		boardConstituentRepo:  boardConstituentRepo,
		marketConfig:          marketConfig,
		akshareProfileFetcher: fetchAKShareCompanyProfile,
	}
}

func (s *marketStockService) GetStockProfile(symbol string) (*marketResponse.StockProfileResponse, error) {
	detail, err := s.GetStockDetail(symbol, false)
	if err != nil {
		return nil, err
	}

	boards := s.lookupConstituentBoards(detail.Symbol)
	industry := detail.Industry
	concepts := append([]string(nil), detail.Concepts...)
	if industry == "" || len(concepts) == 0 {
		industry, concepts = mergeBoardDerivedTaxonomy(industry, concepts, boards)
	}
	companyProfile := s.loadCompanyProfile(detail.Symbol)
	description := buildProfileDescription(detail.Name, industry, detail.Region, concepts, boards)
	if companyProfile != nil {
		description = strings.TrimSpace(companyProfile.Introduction)
		if description == "" {
			description = strings.TrimSpace(companyProfile.Business)
		}
		if description == "" {
			description = buildProfileDescription(detail.Name, industry, detail.Region, concepts, boards)
		}
	}

	return &marketResponse.StockProfileResponse{
		Symbol:           detail.Symbol,
		Name:             detail.Name,
		Market:           detail.Market,
		Description:      description,
		CompanyProfile:   companyProfile,
		Industry:         industry,
		Region:           detail.Region,
		Concepts:         concepts,
		Boards:           boards,
		LastPrice:        detail.LastPrice,
		ChangeAmount:     detail.ChangeAmount,
		ChangePercent:    detail.ChangePercent,
		Volume:           detail.Volume,
		Turnover:         detail.Turnover,
		VolumeRatio:      detail.VolumeRatio,
		TurnoverRate:     detail.TurnoverRate,
		Amplitude:        detail.Amplitude,
		LimitUp:          detail.LimitUp,
		LimitDown:        detail.LimitDown,
		TotalMarketCap:   detail.TotalMarketCap,
		FloatMarketCap:   detail.FloatMarketCap,
		Source:           detail.Source,
		FetchedAt:        detail.FetchedAt,
		IsStale:          detail.IsStale,
		RefreshTriggered: detail.RefreshTriggered,
	}, nil
}

func (s *marketStockService) loadCompanyProfile(symbol string) *marketResponse.StockCompanyProfileResponse {
	if s.akshareProfileFetcher == nil {
		return nil
	}
	profile, err := s.akshareProfileFetcher(context.Background(), strings.TrimSpace(s.marketConfig.AKSharePythonPath), strings.TrimSpace(s.marketConfig.AKShareProfileScriptPath), normalizeSymbol(symbol))
	if err != nil {
		return nil
	}
	return profile
}

func (s *marketStockService) GetStockDetail(symbol string, forceRefresh bool) (*marketResponse.MarketStockDetailResponse, error) {
	normalized := normalizeSymbol(symbol)
	if normalized == "" {
		return nil, errors.New("symbol is required")
	}

	cached, err := s.detailRepo.FindBySymbol(normalized)
	if forceRefresh && cached != nil && !cached.FetchedAt.IsZero() && time.Since(cached.FetchedAt) < minExternalRefreshInterval {
		if !needsSupplementalStockDetail(cached) || s.supplementProvider == nil {
			return marketResponse.NewMarketStockDetailResponse(cached, false, false), nil
		}
		if supplemented := s.trySupplementStockDetail(normalized, cached); supplemented != nil {
			return marketResponse.NewMarketStockDetailResponse(supplemented, false, false), nil
		}
		return marketResponse.NewMarketStockDetailResponse(cached, false, false), nil
	}
	if err == nil && cached != nil && !forceRefresh {
		if !needsSupplementalStockDetail(cached) || s.supplementProvider == nil {
			return marketResponse.NewMarketStockDetailResponse(cached, false, false), nil
		}
		if supplemented := s.trySupplementStockDetail(normalized, cached); supplemented != nil {
			return marketResponse.NewMarketStockDetailResponse(supplemented, false, true), nil
		}
		return marketResponse.NewMarketStockDetailResponse(cached, false, false), nil
	}

	if s.provider == nil {
		if err == nil && cached != nil {
			return marketResponse.NewMarketStockDetailResponse(cached, true, false), nil
		}
		return nil, errors.New("market provider is unavailable")
	}

	fetched, fetchErr := s.provider.GetStockDetail(context.Background(), normalized)
	if fetchErr != nil {
		if supplemented := s.trySupplementStockDetail(normalized, cached); supplemented != nil {
			return marketResponse.NewMarketStockDetailResponse(supplemented, true, true), nil
		}
		if cached == nil && s.supplementProvider != nil {
			detail, supplementErr := s.supplementProvider.GetStockDetail(context.Background(), normalized)
			if supplementErr == nil && detail != nil {
				supplement := convertStockDetailToModel(detail)
				if hasMeaningfulQuoteFields(supplement) {
					if saveErr := s.detailRepo.Upsert(supplement); saveErr != nil {
						return nil, saveErr
					}
					return marketResponse.NewMarketStockDetailResponse(supplement, true, true), nil
				}
			}
		}
		if err == nil && cached != nil {
			return marketResponse.NewMarketStockDetailResponse(cached, true, false), nil
		}
		return nil, fetchErr
	}

	entity := convertStockDetailToModel(fetched)
	if needsSupplementalStockDetail(entity) {
		if supplemented := s.trySupplementStockDetail(normalized, entity); supplemented != nil {
			entity = supplemented
		}
	}
	if saveErr := s.detailRepo.Upsert(entity); saveErr != nil {
		return nil, saveErr
	}
	return marketResponse.NewMarketStockDetailResponse(entity, false, true), nil
}

func (s *marketStockService) trySupplementStockDetail(symbol string, current *model.StockQuoteDetail) *model.StockQuoteDetail {
	if current == nil {
		return nil
	}

	merged := current
	if s.supplementProvider != nil {
		detail, err := s.supplementProvider.GetStockDetail(context.Background(), symbol)
		if err == nil && detail != nil {
			supplement := convertStockDetailToModel(detail)
			merged = mergeStockQuoteDetail(merged, supplement)
		}
	}
	if needsSupplementalStockDetail(merged) {
		if boardMeta := s.lookupBoardMetadata(symbol); boardMeta != nil {
			merged = mergeStockQuoteDetail(merged, boardMeta)
		}
	}
	if merged == nil || merged == current {
		return merged
	}
	if saveErr := s.detailRepo.Upsert(merged); saveErr != nil {
		return nil
	}
	return merged
}

type eastmoneyBoardProfilePayload struct {
	Boards []struct {
		BoardName string `json:"BOARD_NAME"`
		IsPrecise string `json:"IS_PRECISE"`
		BoardRank int    `json:"BOARD_RANK"`
	} `json:"ssbk"`
}

func (s *marketStockService) lookupBoardMetadata(symbol string) *model.StockQuoteDetail {
	if s.httpClient == nil {
		return nil
	}
	code := toEastmoneySecurityCode(symbol)
	if code == "" {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	query := url.Values{}
	query.Set("code", code)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://emweb.securities.eastmoney.com/PC_HSF10/OperationsRequired/PageAjax?"+query.Encode(), nil)
	if err != nil {
		return nil
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Referer", "https://quote.eastmoney.com/")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != http.StatusOK {
		return nil
	}

	var payload eastmoneyBoardProfilePayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil
	}
	if len(payload.Boards) == 0 {
		return nil
	}

	industry := ""
	region := ""
	concepts := make([]string, 0, 8)
	seenConcepts := make(map[string]struct{}, 8)

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
			if _, ok := seenConcepts[concept]; ok {
				continue
			}
			seenConcepts[concept] = struct{}{}
			concepts = append(concepts, concept)
		}
	}

	if industry == "" && region == "" && len(concepts) == 0 {
		return nil
	}

	return &model.StockQuoteDetail{
		Symbol:   normalizeSymbol(symbol),
		Industry: industry,
		Region:   region,
		Concepts: strings.Join(model.NormalizeConceptList(concepts), ","),
		Source:   "em-board",
	}
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
	if precise == "1" {
		return false
	}
	if rank > 3 {
		return false
	}
	if strings.Contains(name, "板块") || strings.Contains(name, "概念") {
		return false
	}
	if strings.Contains(name, "指数") || strings.Contains(name, "HS") || strings.Contains(name, "MSCI") || strings.Contains(name, "富时") || strings.Contains(name, "标普") || strings.Contains(name, "央视") {
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

func buildProfileDescription(name, industry, region string, concepts []string, boards []marketResponse.StockBoardMembershipResponse) string {
	label := strings.TrimSpace(name)
	if label == "" {
		label = "该证券"
	}
	parts := make([]string, 0, 4)
	if normalizedIndustry := model.NormalizeIndustryLabel(industry); normalizedIndustry != "" {
		parts = append(parts, "所属行业为 "+normalizedIndustry)
	}
	if normalizedRegion := model.NormalizeRegionLabel(region); normalizedRegion != "" {
		parts = append(parts, "注册地区为 "+normalizedRegion)
	}
	if len(concepts) > 0 {
		limit := len(concepts)
		if limit > 4 {
			limit = 4
		}
		parts = append(parts, "概念标签包含 "+strings.Join(concepts[:limit], "、"))
	}
	if len(boards) > 0 {
		boardNames := make([]string, 0, 3)
		for _, board := range boards {
			if strings.TrimSpace(board.Name) == "" {
				continue
			}
			boardNames = append(boardNames, board.Name)
			if len(boardNames) == 3 {
				break
			}
		}
		if len(boardNames) > 0 {
			parts = append(parts, "可映射到板块 "+strings.Join(boardNames, "、"))
		}
	}
	if len(parts) == 0 {
		return label + " 当前暂无完整画像，页面将优先展示行情、趋势、新闻和已识别板块。"
	}
	return label + "，" + strings.Join(parts, "；") + "。"
}

type akshareCompanyProfilePayload struct {
	CompanyName         string `json:"company_name"`
	EnglishName         string `json:"english_name"`
	MarketLabel         string `json:"market_label"`
	IndustryLabel       string `json:"industry_label"`
	LegalRepresentative string `json:"legal_representative"`
	RegisteredCapital   string `json:"registered_capital"`
	FoundedAt           string `json:"founded_at"`
	ListedAt            string `json:"listed_at"`
	Website             string `json:"website"`
	Email               string `json:"email"`
	Phone               string `json:"phone"`
	Address             string `json:"address"`
	OfficeAddress       string `json:"office_address"`
	Business            string `json:"business"`
	BusinessScope       string `json:"business_scope"`
	Introduction        string `json:"introduction"`
	Source              string `json:"source"`
}

func fetchAKShareCompanyProfile(ctx context.Context, pythonPath, scriptPath string, symbol string) (*marketResponse.StockCompanyProfileResponse, error) {
	if strings.TrimSpace(symbol) == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	output, err := runAKShareScript(ctx, pythonPath, scriptPath, []string{"AKSHARE_PROFILE_SYMBOL=" + normalizeSymbol(symbol)})
	if err != nil {
		return nil, err
	}
	var payload akshareCompanyProfilePayload
	if err := json.Unmarshal(output, &payload); err != nil {
		return nil, fmt.Errorf("parse akshare company profile payload failed: %w", err)
	}
	if strings.TrimSpace(payload.CompanyName) == "" && strings.TrimSpace(payload.Introduction) == "" && strings.TrimSpace(payload.Business) == "" {
		return nil, nil
	}
	return &marketResponse.StockCompanyProfileResponse{
		CompanyName:         strings.TrimSpace(payload.CompanyName),
		EnglishName:         strings.TrimSpace(payload.EnglishName),
		MarketLabel:         strings.TrimSpace(payload.MarketLabel),
		IndustryLabel:       strings.TrimSpace(payload.IndustryLabel),
		LegalRepresentative: strings.TrimSpace(payload.LegalRepresentative),
		RegisteredCapital:   strings.TrimSpace(payload.RegisteredCapital),
		FoundedAt:           strings.TrimSpace(payload.FoundedAt),
		ListedAt:            strings.TrimSpace(payload.ListedAt),
		Website:             strings.TrimSpace(payload.Website),
		Email:               strings.TrimSpace(payload.Email),
		Phone:               strings.TrimSpace(payload.Phone),
		Address:             strings.TrimSpace(payload.Address),
		OfficeAddress:       strings.TrimSpace(payload.OfficeAddress),
		Business:            strings.TrimSpace(payload.Business),
		BusinessScope:       strings.TrimSpace(payload.BusinessScope),
		Introduction:        strings.TrimSpace(payload.Introduction),
		Source:              strings.TrimSpace(payload.Source),
	}, nil
}

func (s *marketStockService) lookupConstituentBoards(symbol string) []marketResponse.StockBoardMembershipResponse {
	if s.boardConstituentRepo == nil {
		return nil
	}
	normalized := normalizeSymbol(symbol)
	if normalized == "" {
		return nil
	}
	items, err := s.boardConstituentRepo.FindBySymbol(normalized)
	if err != nil || len(items) == 0 {
		return nil
	}
	boards := make([]marketResponse.StockBoardMembershipResponse, 0, 8)
	seen := make(map[string]struct{}, 8)
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.BoardType)) + "|" + strings.TrimSpace(item.BoardCode)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		boards = append(boards, marketResponse.StockBoardMembershipResponse{
			BoardType: strings.TrimSpace(item.BoardType),
			Code:      strings.TrimSpace(item.BoardCode),
			Name:      decodeBoardDisplayName(strings.TrimSpace(item.BoardName)),
			Source:    strings.TrimSpace(item.Source),
		})
	}
	sort.Slice(boards, func(i, j int) bool {
		if boards[i].BoardType != boards[j].BoardType {
			return boards[i].BoardType < boards[j].BoardType
		}
		if boards[i].Name != boards[j].Name {
			return boards[i].Name < boards[j].Name
		}
		return boards[i].Code < boards[j].Code
	})
	return boards
}

func mergeBoardDerivedTaxonomy(industry string, concepts []string, boards []marketResponse.StockBoardMembershipResponse) (string, []string) {
	normalizedIndustry := model.NormalizeIndustryLabel(industry)
	mergedConcepts := model.NormalizeConceptList(concepts)
	seen := make(map[string]struct{}, len(mergedConcepts))
	for _, concept := range mergedConcepts {
		seen[concept] = struct{}{}
	}
	for _, board := range boards {
		name := strings.TrimSpace(board.Name)
		if name == "" {
			continue
		}
		switch strings.ToLower(strings.TrimSpace(board.BoardType)) {
		case "industry":
			if normalizedIndustry == "" {
				normalizedIndustry = name
			}
		case "concept":
			if _, ok := seen[name]; ok {
				continue
			}
			seen[name] = struct{}{}
			mergedConcepts = append(mergedConcepts, name)
		}
	}
	return normalizedIndustry, mergedConcepts
}

func needsSupplementalStockDetail(detail *model.StockQuoteDetail) bool {
	if detail == nil {
		return false
	}
	if detail.Market == "cn_index" {
		return false
	}
	concepts := model.NormalizeConceptList(strings.Split(detail.Concepts, ","))
	return model.NormalizeIndustryLabel(detail.Industry) == "" ||
		model.NormalizeRegionLabel(detail.Region) == "" ||
		len(concepts) == 0 ||
		detail.VolumeRatio.IsZero()
}

func hasMeaningfulQuoteFields(detail *model.StockQuoteDetail) bool {
	if detail == nil {
		return false
	}
	return !detail.LastPrice.IsZero() ||
		!detail.OpenPrice.IsZero() ||
		!detail.HighPrice.IsZero() ||
		!detail.LowPrice.IsZero() ||
		!detail.PrevClose.IsZero() ||
		!detail.ChangeAmount.IsZero() ||
		!detail.ChangePercent.IsZero() ||
		!detail.Volume.IsZero() ||
		!detail.Turnover.IsZero() ||
		!detail.TurnoverRate.IsZero() ||
		!detail.Amplitude.IsZero() ||
		!detail.TotalMarketCap.IsZero() ||
		!detail.FloatMarketCap.IsZero()
}

func mergeStockQuoteDetail(current, supplement *model.StockQuoteDetail) *model.StockQuoteDetail {
	if current == nil {
		return supplement
	}
	if supplement == nil {
		return current
	}

	merged := *current
	if model.NormalizeIndustryLabel(merged.Industry) == "" {
		merged.Industry = model.NormalizeIndustryLabel(supplement.Industry)
	}
	if model.NormalizeRegionLabel(merged.Region) == "" {
		merged.Region = model.NormalizeRegionLabel(supplement.Region)
	}
	if len(model.NormalizeConceptList(strings.Split(merged.Concepts, ","))) == 0 {
		merged.Concepts = strings.Join(model.NormalizeConceptList(strings.Split(supplement.Concepts, ",")), ",")
	}
	if merged.VolumeRatio.IsZero() && !supplement.VolumeRatio.IsZero() {
		merged.VolumeRatio = supplement.VolumeRatio
	}
	if merged.TotalShares.IsZero() && !supplement.TotalShares.IsZero() {
		merged.TotalShares = supplement.TotalShares
	}
	if merged.FloatShares.IsZero() && !supplement.FloatShares.IsZero() {
		merged.FloatShares = supplement.FloatShares
	}
	if supplement.Source != "" && !strings.Contains(merged.Source, supplement.Source) {
		merged.Source = strings.TrimSpace(merged.Source + "+" + supplement.Source)
	}
	if supplement.FetchedAt.After(merged.FetchedAt) {
		merged.FetchedAt = supplement.FetchedAt
	}
	return &merged
}

func (s *marketStockService) GetStockKlines(symbol, period, adjust string, limit int, forceRefresh bool) (*marketResponse.MarketStockKlineResponse, error) {
	normalized := normalizeSymbol(symbol)
	if normalized == "" {
		return nil, errors.New("symbol is required")
	}
	normalizedPeriod := normalizeKlinePeriod(period)
	normalizedAdjust := normalizeAdjustType(adjust)
	if limit <= 0 {
		limit = 120
	}
	if s.klineRepo == nil {
		return nil, errors.New("stock kline repository is unavailable")
	}

	cachedBars, cachedErr := s.klineRepo.FindBars(normalized, normalizedPeriod, normalizedAdjust, limit)
	if forceRefresh && cachedErr == nil && len(cachedBars) > 0 {
		latestCachedAt := cachedBars[0].UpdatedAt
		for _, bar := range cachedBars[1:] {
			if bar.UpdatedAt.After(latestCachedAt) {
				latestCachedAt = bar.UpdatedAt
			}
		}
		if !latestCachedAt.IsZero() && time.Since(latestCachedAt) < minExternalRefreshInterval {
			return marketResponse.NewMarketStockKlineResponse(normalized, normalizedPeriod, normalizedAdjust, cachedBars, false, false), nil
		}
	}
	if cachedErr == nil && !forceRefresh && len(cachedBars) >= limit {
		return marketResponse.NewMarketStockKlineResponse(normalized, normalizedPeriod, normalizedAdjust, cachedBars, false, false), nil
	}

	if s.provider == nil {
		if cachedErr == nil && len(cachedBars) > 0 {
			return marketResponse.NewMarketStockKlineResponse(normalized, normalizedPeriod, normalizedAdjust, cachedBars, true, false), nil
		}
		return nil, errors.New("market provider is unavailable")
	}

	fetchedBars, fetchErr := s.provider.GetKlines(context.Background(), normalized, normalizedPeriod, normalizedAdjust, limit)
	if fetchErr != nil {
		if cachedErr == nil && len(cachedBars) > 0 {
			return marketResponse.NewMarketStockKlineResponse(normalized, normalizedPeriod, normalizedAdjust, cachedBars, true, false), nil
		}
		return nil, fetchErr
	}
	if len(fetchedBars) == 0 {
		if cachedErr == nil && len(cachedBars) > 0 {
			return marketResponse.NewMarketStockKlineResponse(normalized, normalizedPeriod, normalizedAdjust, cachedBars, true, false), nil
		}
		return nil, errors.New("market provider returned empty klines")
	}

	entities := convertKlinesToModels(fetchedBars)
	if saveErr := s.klineRepo.UpsertBars(entities); saveErr != nil {
		return nil, saveErr
	}
	return marketResponse.NewMarketStockKlineResponse(normalized, normalizedPeriod, normalizedAdjust, entities, false, true), nil
}

func convertStockDetailToModel(detail *marketdata.StockDetail) *model.StockQuoteDetail {
	changeAmount := detail.ChangeAmount
	changePercent := detail.ChangePercent
	if changeAmount == 0 && detail.LastPrice != 0 && detail.PrevClose != 0 {
		changeAmount = detail.LastPrice - detail.PrevClose
	}
	if changePercent == 0 && detail.PrevClose != 0 {
		changePercent = changeAmount / detail.PrevClose * 100
	}
	return &model.StockQuoteDetail{
		Symbol:         detail.Symbol,
		Name:           fallbackString(detail.Name, detail.Symbol),
		Market:         fallbackString(detail.Market, marketFromNormalizedSymbol(detail.Symbol)),
		LastPrice:      decimal.NewFromFloat(detail.LastPrice),
		OpenPrice:      decimal.NewFromFloat(detail.OpenPrice),
		HighPrice:      decimal.NewFromFloat(detail.HighPrice),
		LowPrice:       decimal.NewFromFloat(detail.LowPrice),
		PrevClose:      decimal.NewFromFloat(detail.PrevClose),
		ChangeAmount:   decimal.NewFromFloat(changeAmount),
		ChangePercent:  decimal.NewFromFloat(changePercent),
		Volume:         decimal.NewFromFloat(detail.Volume),
		Turnover:       decimal.NewFromFloat(detail.Turnover),
		VolumeRatio:    decimal.NewFromFloat(detail.VolumeRatio),
		TurnoverRate:   decimal.NewFromFloat(detail.TurnoverRate),
		Amplitude:      decimal.NewFromFloat(detail.Amplitude),
		LimitUp:        decimal.NewFromFloat(detail.LimitUp),
		LimitDown:      decimal.NewFromFloat(detail.LimitDown),
		AveragePrice:   decimal.NewFromFloat(detail.AveragePrice),
		TotalShares:    decimal.NewFromFloat(detail.TotalShares),
		FloatShares:    decimal.NewFromFloat(detail.FloatShares),
		TotalMarketCap: decimal.NewFromFloat(detail.TotalMarketCap),
		FloatMarketCap: decimal.NewFromFloat(detail.FloatMarketCap),
		Industry:       model.NormalizeIndustryLabel(detail.Industry),
		Region:         model.NormalizeRegionLabel(detail.Region),
		Concepts:       strings.Join(model.NormalizeConceptList(detail.Concepts), ","),
		Source:         detail.Source,
		FetchedAt:      detail.FetchedAt,
	}
}

func convertKlinesToModels(bars []marketdata.KlineBar) []model.StockKlineBar {
	entities := make([]model.StockKlineBar, 0, len(bars))
	for _, bar := range bars {
		entities = append(entities, model.StockKlineBar{
			Symbol:        bar.Symbol,
			Period:        normalizeKlinePeriod(bar.Period),
			AdjustType:    normalizeAdjustType(bar.AdjustType),
			BarTime:       bar.BarTime,
			OpenPrice:     decimal.NewFromFloat(bar.Open),
			ClosePrice:    decimal.NewFromFloat(bar.Close),
			HighPrice:     decimal.NewFromFloat(bar.High),
			LowPrice:      decimal.NewFromFloat(bar.Low),
			Volume:        decimal.NewFromFloat(bar.Volume),
			Turnover:      decimal.NewFromFloat(bar.Amount),
			Amplitude:     decimal.NewFromFloat(bar.Amplitude),
			ChangePercent: decimal.NewFromFloat(bar.ChangePercent),
			ChangeAmount:  decimal.NewFromFloat(bar.ChangeAmount),
			TurnoverRate:  decimal.NewFromFloat(bar.TurnoverRate),
			Source:        bar.Source,
		})
	}
	return entities
}

func normalizeKlinePeriod(period string) string {
	switch strings.ToLower(strings.TrimSpace(period)) {
	case "1m", "5m", "15m", "30m", "60m", "week", "month":
		return strings.ToLower(strings.TrimSpace(period))
	case "", "day", "daily", "d":
		return "day"
	default:
		return strings.ToLower(strings.TrimSpace(period))
	}
}

func normalizeAdjustType(adjust string) string {
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

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func marketFromNormalizedSymbol(symbol string) string {
	switch symbol {
	case "000001.SH", "000300.SH", "399001.SZ", "399006.SZ":
		return "cn_index"
	}
	return "cn_stock"
}

var _ = gorm.ErrRecordNotFound
