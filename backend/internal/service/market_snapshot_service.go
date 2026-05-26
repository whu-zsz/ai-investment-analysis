package service

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	marketResponse "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type MarketSnapshotService interface {
	GetLatestSnapshots() ([]marketResponse.MarketSnapshotResponse, error)
	GetHistory(symbol string, limit int, startTime, endTime *time.Time) ([]marketResponse.MarketSnapshotResponse, error)
	SearchStocks(query string, limit int) ([]marketResponse.MarketSnapshotResponse, error)
	GetDashboardSnapshot() (*marketResponse.DashboardMarketSnapshotResponse, error)
	GetDashboardMarketBreadth(ctx context.Context, limit int) (*marketResponse.DashboardMarketBreadthResponse, error)
	GetBoardDetail(ctx context.Context, boardType, code string, limit int) (*marketResponse.MarketBoardDetailResponse, error)
	SearchRelevantBoards(ctx context.Context, query string, limit int) ([]marketResponse.MarketBoardItemResponse, error)
}

const (
	minBreadthSnapshotCoverage = 5000
	recentBreadthBatchLimit    = 6
)

var defaultDashboardTrackedSymbols = []string{
	"000001.SH",
	"399001.SZ",
	"399006.SZ",
	"000300.SH",
}

type marketSnapshotService struct {
	snapshotRepo         repository.MarketSnapshotRepository
	boardSnapshotRepo    repository.MarketBoardSnapshotRepository
	boardConstituentRepo repository.MarketBoardConstituentRepository
	detailRepo           repository.StockQuoteDetailRepository
	marketDataService    MarketDataService
	breadthRefreshActive chan struct{}
}

func NewMarketSnapshotService(snapshotRepo repository.MarketSnapshotRepository, boardSnapshotRepo repository.MarketBoardSnapshotRepository, boardConstituentRepo repository.MarketBoardConstituentRepository, detailRepo repository.StockQuoteDetailRepository, marketDataService MarketDataService) MarketSnapshotService {
	return &marketSnapshotService{
		snapshotRepo:         snapshotRepo,
		boardSnapshotRepo:    boardSnapshotRepo,
		boardConstituentRepo: boardConstituentRepo,
		detailRepo:           detailRepo,
		marketDataService:    marketDataService,
		breadthRefreshActive: make(chan struct{}, 1),
	}
}

func (s *marketSnapshotService) triggerBreadthRefresh() {
	if s == nil || s.marketDataService == nil || s.breadthRefreshActive == nil {
		return
	}
	select {
	case s.breadthRefreshActive <- struct{}{}:
		go func() {
			defer func() { <-s.breadthRefreshActive }()
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
			defer cancel()
			_, _, _ = s.marketDataService.FetchAndStoreFullMarketSnapshots(ctx)
			_, _, _ = s.marketDataService.FetchAndStoreMarketBoardSnapshots(ctx)
		}()
	default:
	}
}

func (s *marketSnapshotService) GetLatestSnapshots() ([]marketResponse.MarketSnapshotResponse, error) {
	batchNo, err := s.snapshotRepo.FindLatestBatchNo()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []marketResponse.MarketSnapshotResponse{}, nil
		}
		return nil, err
	}

	snapshots, err := s.snapshotRepo.FindByBatchNo(batchNo)
	if err != nil {
		return nil, err
	}

	return convertSnapshots(snapshots), nil
}

func (s *marketSnapshotService) GetHistory(symbol string, limit int, startTime, endTime *time.Time) ([]marketResponse.MarketSnapshotResponse, error) {
	var (
		snapshots []model.MarketSnapshot
		err       error
	)

	if symbol == "" {
		snapshots, err = s.snapshotRepo.FindHistory(limit, startTime, endTime)
	} else {
		snapshots, err = s.snapshotRepo.FindHistoryBySymbol(symbol, limit, startTime, endTime)
	}
	if err != nil {
		return nil, err
	}
	return convertSnapshots(snapshots), nil
}

func (s *marketSnapshotService) SearchStocks(query string, limit int) ([]marketResponse.MarketSnapshotResponse, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []marketResponse.MarketSnapshotResponse{}, nil
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	snapshots, err := s.snapshotRepo.SearchStocks(query, limit)
	if err != nil {
		return nil, err
	}
	return convertSnapshots(snapshots), nil
}

func (s *marketSnapshotService) GetDashboardSnapshot() (*marketResponse.DashboardMarketSnapshotResponse, error) {
	snapshots, err := s.loadDashboardTrackedSnapshots()
	if err != nil {
		return nil, err
	}
	if len(snapshots) == 0 {
		return &marketResponse.DashboardMarketSnapshotResponse{
			Indices: []marketResponse.MarketIndexItemResponse{},
			Stats:   []marketResponse.DashboardStatResponse{},
		}, nil
	}

	sort.Slice(snapshots, func(i, j int) bool {
		if snapshots[i].CreatedAt.Equal(snapshots[j].CreatedAt) {
			return snapshots[i].SnapshotTime.Before(snapshots[j].SnapshotTime)
		}
		return snapshots[i].CreatedAt.Before(snapshots[j].CreatedAt)
	})

	latest := snapshots[len(snapshots)-1]
	indices := make([]marketResponse.MarketIndexItemResponse, 0, len(snapshots))
	for _, snapshot := range snapshots {
		indices = append(indices, marketResponse.MarketIndexItemResponse{
			Symbol:        snapshot.Symbol,
			Name:          snapshot.Name,
			Market:        snapshot.Market,
			LastPrice:     snapshot.LastPrice.String(),
			ChangeAmount:  snapshot.ChangeAmount.String(),
			ChangePercent: snapshot.ChangePercent.String(),
			OpenPrice:     snapshot.OpenPrice.String(),
			HighPrice:     snapshot.HighPrice.String(),
			LowPrice:      snapshot.LowPrice.String(),
			PrevClose:     snapshot.PrevClose.String(),
			Volume:        snapshot.Volume.String(),
			Turnover:      snapshot.Turnover.String(),
		})
	}

	chartSeries := make([]marketResponse.MarketChartPoint, 0, len(snapshots))
	for _, snapshot := range snapshots {
		chartSeries = append(chartSeries, marketResponse.MarketChartPoint{
			Label: snapshot.Name,
			Value: snapshot.LastPrice.String(),
		})
	}

	stats := []marketResponse.DashboardStatResponse{
		{Label: "指数数量", Value: formatInt(len(snapshots))},
		{Label: "上涨数", Value: formatInt(countPositive(snapshots))},
		{Label: "下跌数", Value: formatInt(countNegative(snapshots))},
		{Label: "平均涨跌幅", Value: averageChangePercent(snapshots)},
		{Label: "总成交额", Value: totalTurnover(snapshots)},
	}

	return &marketResponse.DashboardMarketSnapshotResponse{
		SnapshotTime: latest.SnapshotTime.Format("2006-01-02 15:04:05"),
		RefreshedAt:  latest.CreatedAt.Format("2006-01-02 15:04:05"),
		IsStale:      time.Since(latest.CreatedAt) > 35*time.Minute,
		Source:       latest.Source,
		Indices:      indices,
		MainChart: marketResponse.MarketChartResponse{
			IndexName: snapshots[0].Name,
			Series:    chartSeries,
		},
		Stats: stats,
	}, nil
}

func (s *marketSnapshotService) loadDashboardTrackedSnapshots() ([]model.MarketSnapshot, error) {
	symbols := dashboardTrackedSymbols()
	snapshots := make([]model.MarketSnapshot, 0, len(symbols))
	for _, symbol := range symbols {
		history, err := s.snapshotRepo.FindHistoryBySymbol(symbol, 12, nil, nil)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				continue
			}
			return nil, err
		}
		expectedName := dashboardTrackedIndexName(symbol)
		var latestIndex *model.MarketSnapshot
		var fallbackIndex *model.MarketSnapshot
		for i := range history {
			if history[i].Market != "cn_index" {
				continue
			}
			if fallbackIndex == nil {
				fallbackIndex = &history[i]
			}
			if expectedName != "" && strings.TrimSpace(history[i].Name) == expectedName {
				latestIndex = &history[i]
				break
			}
		}
		if latestIndex == nil {
			latestIndex = fallbackIndex
		}
		if latestIndex == nil {
			continue
		}
		snapshots = append(snapshots, *latestIndex)
	}
	return snapshots, nil
}

func dashboardTrackedSymbols() []string {
	raw := strings.TrimSpace(os.Getenv("MARKET_SYMBOLS"))
	if raw == "" {
		return append([]string(nil), defaultDashboardTrackedSymbols...)
	}
	parts := strings.Split(raw, ",")
	symbols := make([]string, 0, len(parts))
	for _, part := range parts {
		symbol := normalizeSymbol(part)
		if symbol == "" {
			continue
		}
		symbols = append(symbols, symbol)
	}
	if len(symbols) == 0 {
		return append([]string(nil), defaultDashboardTrackedSymbols...)
	}
	return symbols
}

func dashboardTrackedIndexName(symbol string) string {
	switch strings.ToUpper(strings.TrimSpace(symbol)) {
	case "000001.SH":
		return "上证指数"
	case "399001.SZ":
		return "深证成指"
	case "399006.SZ":
		return "创业板指"
	case "000300.SH":
		return "沪深300"
	default:
		return ""
	}
}

func (s *marketSnapshotService) GetDashboardMarketBreadth(ctx context.Context, limit int) (*marketResponse.DashboardMarketBreadthResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 50 {
		limit = 50
	}

	isPartial := false
	message := ""
	_, snapshots, usedFallbackBatch, err := s.loadPreferredBreadthSnapshots()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if s.marketDataService != nil && shouldRefreshBreadthSnapshots(snapshots) {
		isPartial = true
		message = fallbackMessage(message, "全市场雷达正在后台刷新，当前先展示最近一次可用快照")
		s.triggerBreadthRefresh()
	}
	if usedFallbackBatch {
		isPartial = true
		message = fallbackMessage(message, "最新全市场批次覆盖不足，已回退到最近可用批次")
	}

	if len(snapshots) == 0 {
		return &marketResponse.DashboardMarketBreadthResponse{
			Source:               "sina",
			IsPartial:            true,
			Message:              fallbackMessage(message, "暂无全市场行情数据"),
			Coverage:             []marketResponse.DashboardStatResponse{},
			TopGainers:           []marketResponse.MarketBreadthItemResponse{},
			TopLosers:            []marketResponse.MarketBreadthItemResponse{},
			TopTurnover:          []marketResponse.MarketBreadthItemResponse{},
			Sectors:              []marketResponse.MarketBoardItemResponse{},
			Concepts:             []marketResponse.MarketBoardItemResponse{},
			ChangeDistribution:   []marketResponse.DistributionBucketResponse{},
			TurnoverDistribution: []marketResponse.DistributionBucketResponse{},
		}, nil
	}

	sectors, _ := s.latestBoards("industry", maxInt(limit*4, 40))
	concepts, _ := s.latestBoards("concept", maxInt(limit*2, 20))
	if s.marketDataService != nil && shouldRefreshBoardSnapshots(sectors, concepts) {
		isPartial = true
		message = fallbackMessage(message, "板块雷达正在后台刷新，当前先展示最近一次可用结果")
		s.triggerBreadthRefresh()
	}

	fallbackUsed := false
	if len(sectors) == 0 || len(concepts) == 0 {
		fallbackSectors, fallbackConcepts, fallbackErr := s.buildFallbackBoards(snapshots, limit)
		if fallbackErr != nil {
			isPartial = true
			message = fallbackMessage(message, "板块聚合回退失败："+fallbackErr.Error())
		} else {
			if len(sectors) == 0 && len(fallbackSectors) > 0 {
				sectors = fallbackSectors
				fallbackUsed = true
			}
			if len(concepts) == 0 && len(fallbackConcepts) > 0 {
				concepts = fallbackConcepts
				fallbackUsed = true
			}
		}
	}
	if fallbackUsed {
		isPartial = true
		message = fallbackMessage(message, "部分板块数据由本地快照聚合生成")
	}

	latest := latestSnapshotByCreatedAt(snapshots)
	sectors = selectDashboardBoards(sectors, "industry", limit)
	concepts = selectDashboardBoards(concepts, "concept", limit)

	return &marketResponse.DashboardMarketBreadthResponse{
		SnapshotTime:         latest.SnapshotTime.Format("2006-01-02 15:04:05"),
		RefreshedAt:          latest.CreatedAt.Format("2006-01-02 15:04:05"),
		Source:               latest.Source,
		IsPartial:            isPartial,
		Message:              message,
		Coverage:             buildBreadthCoverage(snapshots),
		TopGainers:           convertBreadthItems(sortSnapshots(snapshots, func(a, b model.MarketSnapshot) bool { return a.ChangePercent.GreaterThan(b.ChangePercent) }), limit),
		TopLosers:            convertBreadthItems(sortSnapshots(snapshots, func(a, b model.MarketSnapshot) bool { return a.ChangePercent.LessThan(b.ChangePercent) }), limit),
		TopTurnover:          convertBreadthItems(sortSnapshots(snapshots, func(a, b model.MarketSnapshot) bool { return a.Turnover.GreaterThan(b.Turnover) }), limit),
		Sectors:              convertBoardItems(sectors),
		Concepts:             convertBoardItems(concepts),
		ChangeDistribution:   buildChangeDistribution(snapshots),
		TurnoverDistribution: buildTurnoverDistribution(snapshots),
	}, nil
}

func (s *marketSnapshotService) GetBoardDetail(ctx context.Context, boardType, code string, limit int) (*marketResponse.MarketBoardDetailResponse, error) {
	_ = ctx
	boardType = strings.ToLower(strings.TrimSpace(boardType))
	code = strings.TrimSpace(code)
	if boardType == "" || code == "" {
		return nil, fmt.Errorf("board type and code are required")
	}
	if limit <= 0 {
		limit = 60
	}
	if limit > 200 {
		limit = 200
	}
	if s.boardSnapshotRepo == nil {
		return nil, gorm.ErrRecordNotFound
	}

	boards, err := s.latestBoards(boardType, 500)
	if err != nil {
		return nil, err
	}
	var target *model.MarketBoardSnapshot
	for i := range boards {
		if strings.EqualFold(strings.TrimSpace(boards[i].Code), code) {
			target = &boards[i]
			break
		}
	}
	if target == nil {
		return nil, gorm.ErrRecordNotFound
	}

	_, snapshots, usedFallbackBatch, err := s.loadPreferredBreadthSnapshots()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	snapshotBySymbol := make(map[string]model.MarketSnapshot, len(snapshots))
	for _, snapshot := range snapshots {
		if existing, ok := snapshotBySymbol[snapshot.Symbol]; ok {
			if snapshot.CreatedAt.After(existing.CreatedAt) || snapshot.SnapshotTime.After(existing.SnapshotTime) {
				snapshotBySymbol[snapshot.Symbol] = snapshot
			}
			continue
		}
		snapshotBySymbol[snapshot.Symbol] = snapshot
	}

	constituents := make([]model.MarketBoardConstituent, 0)
	if s.boardConstituentRepo != nil {
		allItems, findErr := s.boardConstituentRepo.FindAll()
		if findErr != nil && !errors.Is(findErr, gorm.ErrRecordNotFound) {
			return nil, findErr
		}
		for _, item := range allItems {
			if strings.EqualFold(item.BoardType, boardType) && strings.EqualFold(item.BoardCode, code) {
				constituents = append(constituents, item)
			}
		}
	}

	items := make([]marketResponse.BoardConstituentResponse, 0, len(constituents))
	for _, item := range constituents {
		snapshot, ok := snapshotBySymbol[item.Symbol]
		source := item.Source
		responseItem := marketResponse.BoardConstituentResponse{
			Symbol:         item.Symbol,
			Name:           fallbackString(item.StockName, item.Symbol),
			Market:         "cn_stock",
			LastPrice:      "",
			ChangeAmount:   "",
			ChangePercent:  "",
			Volume:         "",
			Turnover:       "",
			TotalMarketCap: item.TotalMarketCap.String(),
			FloatMarketCap: item.FloatMarketCap.String(),
			Source:         source,
			HasSnapshot:    ok,
		}
		if ok {
			responseItem.Name = fallbackString(snapshot.Name, responseItem.Name)
			responseItem.Market = fallbackString(snapshot.Market, responseItem.Market)
			responseItem.LastPrice = snapshot.LastPrice.String()
			responseItem.ChangeAmount = snapshot.ChangeAmount.String()
			responseItem.ChangePercent = snapshot.ChangePercent.String()
			responseItem.Volume = snapshot.Volume.String()
			responseItem.Turnover = snapshot.Turnover.String()
			source = snapshot.Source
			responseItem.Source = source
		}
		items = append(items, responseItem)
	}

	sort.Slice(items, func(i, j int) bool {
		leftTurnover := toBoardDecimal(items[i].Turnover)
		rightTurnover := toBoardDecimal(items[j].Turnover)
		if !leftTurnover.Equal(rightTurnover) {
			return leftTurnover.GreaterThan(rightTurnover)
		}
		leftChange := toBoardDecimal(items[i].ChangePercent)
		rightChange := toBoardDecimal(items[j].ChangePercent)
		if !leftChange.Equal(rightChange) {
			return leftChange.GreaterThan(rightChange)
		}
		return items[i].Symbol < items[j].Symbol
	})

	displayItems := items
	if len(displayItems) > limit {
		displayItems = displayItems[:limit]
	}
	topGainers := append([]marketResponse.BoardConstituentResponse(nil), items...)
	sort.Slice(topGainers, func(i, j int) bool {
		left := toBoardDecimal(topGainers[i].ChangePercent)
		right := toBoardDecimal(topGainers[j].ChangePercent)
		if !left.Equal(right) {
			return left.GreaterThan(right)
		}
		return topGainers[i].Symbol < topGainers[j].Symbol
	})
	if len(topGainers) > 8 {
		topGainers = topGainers[:8]
	}
	topLosers := append([]marketResponse.BoardConstituentResponse(nil), items...)
	sort.Slice(topLosers, func(i, j int) bool {
		left := toBoardDecimal(topLosers[i].ChangePercent)
		right := toBoardDecimal(topLosers[j].ChangePercent)
		if !left.Equal(right) {
			return left.LessThan(right)
		}
		return topLosers[i].Symbol < topLosers[j].Symbol
	})
	if len(topLosers) > 8 {
		topLosers = topLosers[:8]
	}
	topTurnover := append([]marketResponse.BoardConstituentResponse(nil), items...)
	if len(topTurnover) > 8 {
		topTurnover = topTurnover[:8]
	}

	pricedCount := 0
	for _, item := range items {
		if item.HasSnapshot {
			pricedCount++
		}
	}
	coverage := []marketResponse.DashboardStatResponse{
		{Label: "成分股数", Value: formatInt(target.StockCount)},
		{Label: "可定价数", Value: formatInt(pricedCount)},
		{Label: "上涨家数", Value: formatInt(target.RiseCount)},
		{Label: "下跌家数", Value: formatInt(target.FallCount)},
		{Label: "平盘家数", Value: formatInt(target.FlatCount)},
		{Label: "板块成交额", Value: formatInt(0)},
	}
	if !target.Turnover.IsZero() {
		coverage[5].Value = totalTurnover([]model.MarketSnapshot{{Turnover: target.Turnover}})
	}

	message := ""
	if len(constituents) == 0 {
		message = fallbackMessage(message, "当前板块暂无成分股清单")
	}

	snapshotTime := target.SnapshotTime.Format("2006-01-02 15:04:05")
	refreshedAt := target.CreatedAt.Format("2006-01-02 15:04:05")
	return &marketResponse.MarketBoardDetailResponse{
		SnapshotTime: snapshotTime,
		RefreshedAt:  refreshedAt,
		Source:       target.Source,
		IsPartial:    usedFallbackBatch || pricedCount < len(constituents),
		Message:      message,
		Board: marketResponse.MarketBoardItemResponse{
			BoardType:       target.BoardType,
			Code:            target.Code,
			Name:            decodeBoardDisplayName(target.Name),
			LastPrice:       target.LastPrice.String(),
			ChangeAmount:    target.ChangeAmount.String(),
			ChangePercent:   target.ChangePercent.String(),
			Volume:          target.Volume.String(),
			Turnover:        target.Turnover.String(),
			TotalMarketCap:  target.TotalMarketCap.String(),
			FloatMarketCap:  target.FloatMarketCap.String(),
			StockCount:      target.StockCount,
			RiseCount:       target.RiseCount,
			FallCount:       target.FallCount,
			FlatCount:       target.FlatCount,
			ConstituentNode: target.ConstituentNode,
		},
		Coverage:     coverage,
		Constituents: displayItems,
		TopGainers:   topGainers,
		TopLosers:    topLosers,
		TopTurnover:  topTurnover,
	}, nil
}

func (s *marketSnapshotService) SearchRelevantBoards(ctx context.Context, query string, limit int) ([]marketResponse.MarketBoardItemResponse, error) {
	_ = ctx
	query = strings.TrimSpace(query)
	if query == "" {
		return []marketResponse.MarketBoardItemResponse{}, nil
	}
	if limit <= 0 {
		limit = 8
	}
	if limit > 20 {
		limit = 20
	}

	industryBoards, industryErr := s.latestBoards("industry", 500)
	conceptBoards, conceptErr := s.latestBoards("concept", 500)
	if industryErr != nil && conceptErr != nil {
		return nil, industryErr
	}

	constituents := make([]model.MarketBoardConstituent, 0)
	if s.boardConstituentRepo != nil {
		items, err := s.boardConstituentRepo.FindAll()
		if err == nil {
			constituents = items
		}
	}
	constituentByBoard := make(map[string][]model.MarketBoardConstituent, 512)
	for _, item := range constituents {
		key := strings.ToLower(strings.TrimSpace(item.BoardType)) + ":" + strings.TrimSpace(item.BoardCode)
		constituentByBoard[key] = append(constituentByBoard[key], item)
	}

	type boardMatch struct {
		board marketResponse.MarketBoardItemResponse
		score int
	}
	terms := boardSearchTerms(query)
	matches := make(map[string]boardMatch, 128)
	boardLookup := make(map[string]marketResponse.MarketBoardItemResponse, 512)
	appendMatches := func(items []model.MarketBoardSnapshot) {
		for _, item := range items {
			response := marketResponse.MarketBoardItemResponse{
				BoardType:       item.BoardType,
				Code:            item.Code,
				Name:            item.Name,
				LastPrice:       item.LastPrice.StringFixed(4),
				ChangeAmount:    item.ChangeAmount.StringFixed(4),
				ChangePercent:   item.ChangePercent.StringFixed(4),
				Volume:          item.Volume.StringFixed(0),
				Turnover:        item.Turnover.StringFixed(0),
				TotalMarketCap:  item.TotalMarketCap.StringFixed(0),
				FloatMarketCap:  item.FloatMarketCap.StringFixed(0),
				StockCount:      item.StockCount,
				RiseCount:       item.RiseCount,
				FallCount:       item.FallCount,
				FlatCount:       item.FlatCount,
				ConstituentNode: item.ConstituentNode,
			}
			key := strings.ToLower(strings.TrimSpace(item.BoardType)) + ":" + strings.ToLower(strings.TrimSpace(item.Name))
			boardLookup[key] = response
			directScore := scoreBoardQueryMatch(response, constituentByBoard[strings.ToLower(strings.TrimSpace(item.BoardType))+":"+strings.TrimSpace(item.Code)], terms)
			if directScore <= 0 {
				continue
			}
			matches[key] = boardMatch{board: response, score: directScore}
		}
	}
	appendMatches(industryBoards)
	appendMatches(conceptBoards)

	if s.detailRepo != nil && s.snapshotRepo != nil {
		_, snapshots, _, err := s.loadPreferredBreadthSnapshots()
		if err == nil && len(snapshots) > 0 {
			symbols := make([]string, 0, len(snapshots))
			for _, snapshot := range snapshots {
				symbols = append(symbols, snapshot.Symbol)
			}
			details, detailErr := s.detailRepo.FindBySymbols(symbols)
			if detailErr == nil && len(details) > 0 {
				detailBySymbol := make(map[string]model.StockQuoteDetail, len(details))
				for _, detail := range details {
					detailBySymbol[detail.Symbol] = detail
				}
				ragScores := make(map[string]int, 128)
				for _, snapshot := range snapshots {
					detail, ok := detailBySymbol[snapshot.Symbol]
					if !ok {
						continue
					}
					docScore := scoreTaxonomyDocument(snapshot.Name, detail, terms)
					if docScore <= 0 {
						continue
					}
					if industry := model.NormalizeIndustryLabel(detail.Industry); industry != "" {
						key := "industry:" + strings.ToLower(strings.TrimSpace(industry))
						if _, ok := boardLookup[key]; ok {
							ragScores[key] += docScore * 3
						}
					}
					for _, concept := range model.NormalizeConceptList(strings.Split(detail.Concepts, ",")) {
						key := "concept:" + strings.ToLower(strings.TrimSpace(concept))
						if _, ok := boardLookup[key]; ok {
							ragScores[key] += docScore * 2
						}
					}
				}
				for key, score := range ragScores {
					board, ok := boardLookup[key]
					if !ok || score <= 0 {
						continue
					}
					if existing, exists := matches[key]; exists {
						existing.score += score
						matches[key] = existing
						continue
					}
					matches[key] = boardMatch{board: board, score: score}
				}
			}
		}
	}

	ordered := make([]boardMatch, 0, len(matches))
	for _, item := range matches {
		ordered = append(ordered, item)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].score != ordered[j].score {
			return ordered[i].score > ordered[j].score
		}
		left := parseDecimalOrZero(ordered[i].board.ChangePercent)
		right := parseDecimalOrZero(ordered[j].board.ChangePercent)
		if !left.Equal(right) {
			return left.GreaterThan(right)
		}
		return ordered[i].board.Name < ordered[j].board.Name
	})

	result := make([]marketResponse.MarketBoardItemResponse, 0, limit)
	for _, item := range ordered {
		result = append(result, item.board)
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func minBoardLimit(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func selectDashboardBoards(items []model.MarketBoardSnapshot, boardType string, limit int) []model.MarketBoardSnapshot {
	if limit <= 0 || len(items) == 0 {
		return []model.MarketBoardSnapshot{}
	}
	if boardType != "industry" {
		if len(items) <= limit {
			return items
		}
		return items[:limit]
	}

	minTurnover := decimal.NewFromInt(5_000_000_000)
	filtered := make([]model.MarketBoardSnapshot, 0, len(items))
	for _, item := range items {
		if item.StockCount < 15 {
			continue
		}
		if item.Turnover.LessThan(minTurnover) {
			continue
		}
		filtered = append(filtered, item)
	}
	if len(filtered) >= minBoardLimit(limit, 6) {
		if len(filtered) <= limit {
			return filtered
		}
		return filtered[:limit]
	}
	if len(items) <= limit {
		return items
	}
	return items[:limit]
}

func (s *marketSnapshotService) latestBoards(boardType string, limit int) ([]model.MarketBoardSnapshot, error) {
	if s.boardSnapshotRepo == nil {
		return []model.MarketBoardSnapshot{}, nil
	}
	batchNo, err := s.boardSnapshotRepo.FindLatestBatchNo(boardType)
	if err != nil {
		return []model.MarketBoardSnapshot{}, err
	}
	return s.boardSnapshotRepo.FindByBatchNo(boardType, batchNo, limit)
}

func (s *marketSnapshotService) loadPreferredBreadthSnapshots() (string, []model.MarketSnapshot, bool, error) {
	batchNo, err := s.snapshotRepo.FindLatestBatchNo()
	if err != nil {
		return "", nil, false, err
	}
	snapshots, err := s.snapshotRepo.FindByBatchNo(batchNo)
	if err != nil {
		return "", nil, false, err
	}
	if isBreadthBatchUsable(snapshots) {
		return batchNo, snapshots, false, nil
	}

	batchNos, err := s.snapshotRepo.FindRecentBatchNos(recentBreadthBatchLimit)
	if err != nil {
		return batchNo, snapshots, false, err
	}
	for _, candidate := range batchNos {
		if candidate == "" || candidate == batchNo {
			continue
		}
		candidateSnapshots, candidateErr := s.snapshotRepo.FindByBatchNo(candidate)
		if candidateErr != nil {
			continue
		}
		if isBreadthBatchUsable(candidateSnapshots) {
			return candidate, candidateSnapshots, true, nil
		}
	}
	return batchNo, snapshots, false, nil
}

func (s *marketSnapshotService) buildFallbackBoards(snapshots []model.MarketSnapshot, limit int) ([]model.MarketBoardSnapshot, []model.MarketBoardSnapshot, error) {
	if s.detailRepo == nil || len(snapshots) == 0 {
		return []model.MarketBoardSnapshot{}, []model.MarketBoardSnapshot{}, nil
	}

	symbols := make([]string, 0, len(snapshots))
	for _, snapshot := range snapshots {
		symbols = append(symbols, snapshot.Symbol)
	}

	details, err := s.detailRepo.FindBySymbols(symbols)
	if err != nil {
		return nil, nil, err
	}
	if len(details) == 0 {
		return []model.MarketBoardSnapshot{}, []model.MarketBoardSnapshot{}, nil
	}

	detailMap := make(map[string]model.StockQuoteDetail, len(details))
	for _, detail := range details {
		detailMap[detail.Symbol] = detail
	}

	industryAgg := make(map[string]*boardAggregate)
	conceptAgg := make(map[string]*boardAggregate)
	latest := latestSnapshotByCreatedAt(snapshots)

	for _, snapshot := range snapshots {
		detail, ok := detailMap[snapshot.Symbol]
		if !ok {
			continue
		}
		industry := model.NormalizeIndustryLabel(detail.Industry)
		if industry != "" {
			accumulateBoardAggregate(industryAgg, industry, snapshot)
		}
		conceptValues := model.NormalizeConceptList(strings.Split(detail.Concepts, ","))
		for _, concept := range conceptValues {
			if concept == "" {
				continue
			}
			accumulateBoardAggregate(conceptAgg, concept, snapshot)
		}
	}

	industries := convertAggregatesToBoards(industryAgg, "industry", latest)
	concepts := convertAggregatesToBoards(conceptAgg, "concept", latest)
	industries = selectDashboardBoards(industries, "industry", maxInt(limit*4, 40))
	concepts = selectDashboardBoards(concepts, "concept", maxInt(limit*2, 20))
	return industries, concepts, nil
}

type boardAggregate struct {
	name            string
	stockCount      int
	riseCount       int
	fallCount       int
	flatCount       int
	changeSum       decimal.Decimal
	turnover        decimal.Decimal
	volume          decimal.Decimal
	totalMarketCap  decimal.Decimal
	floatMarketCap  decimal.Decimal
	lastPriceSum    decimal.Decimal
	changeAmountSum decimal.Decimal
}

func accumulateBoardAggregate(target map[string]*boardAggregate, name string, snapshot model.MarketSnapshot) {
	item, ok := target[name]
	if !ok {
		item = &boardAggregate{name: name}
		target[name] = item
	}
	item.stockCount++
	item.changeSum = item.changeSum.Add(snapshot.ChangePercent)
	item.turnover = item.turnover.Add(snapshot.Turnover)
	item.volume = item.volume.Add(snapshot.Volume)
	item.lastPriceSum = item.lastPriceSum.Add(snapshot.LastPrice)
	item.changeAmountSum = item.changeAmountSum.Add(snapshot.ChangeAmount)
	switch {
	case snapshot.ChangePercent.GreaterThan(decimal.Zero):
		item.riseCount++
	case snapshot.ChangePercent.LessThan(decimal.Zero):
		item.fallCount++
	default:
		item.flatCount++
	}
}

func convertAggregatesToBoards(aggregates map[string]*boardAggregate, boardType string, latest model.MarketSnapshot) []model.MarketBoardSnapshot {
	boards := make([]model.MarketBoardSnapshot, 0, len(aggregates))
	for _, item := range aggregates {
		if item == nil || item.stockCount == 0 {
			continue
		}
		count := decimal.NewFromInt(int64(item.stockCount))
		boards = append(boards, model.MarketBoardSnapshot{
			BoardType:      boardType,
			Code:           item.name,
			Name:           item.name,
			LastPrice:      item.lastPriceSum.Div(count),
			ChangeAmount:   item.changeAmountSum.Div(count),
			ChangePercent:  item.changeSum.Div(count),
			Volume:         item.volume,
			Turnover:       item.turnover,
			TotalMarketCap: item.totalMarketCap,
			FloatMarketCap: item.floatMarketCap,
			StockCount:     item.stockCount,
			RiseCount:      item.riseCount,
			FallCount:      item.fallCount,
			FlatCount:      item.flatCount,
			Source:         latest.Source + "+snapshot-fallback",
			BatchNo:        latest.BatchNo,
			SnapshotTime:   latest.SnapshotTime,
			CreatedAt:      latest.CreatedAt,
			UpdatedAt:      latest.UpdatedAt,
		})
	}
	sort.Slice(boards, func(i, j int) bool {
		left := boards[i]
		right := boards[j]
		if !left.ChangePercent.Equal(right.ChangePercent) {
			return left.ChangePercent.GreaterThan(right.ChangePercent)
		}
		if !left.Turnover.Equal(right.Turnover) {
			return left.Turnover.GreaterThan(right.Turnover)
		}
		return left.Name < right.Name
	})
	return boards
}

func boardSearchTerms(query string) []string {
	normalized := strings.ToLower(strings.TrimSpace(query))
	if normalized == "" {
		return nil
	}
	fields := strings.FieldsFunc(normalized, func(r rune) bool {
		switch r {
		case ' ', '\n', '\t', '，', ',', '。', '；', ';', '、', '：', ':', '（', '）', '(', ')':
			return true
		default:
			return false
		}
	})
	terms := make([]string, 0, len(fields)+12)
	terms = append(terms, normalized)
	for _, field := range fields {
		field = strings.TrimSpace(field)
		if len([]rune(field)) >= 2 {
			terms = append(terms, field)
		}
	}
	terms = append(terms, extractCJKQueryTerms(normalized)...)
	return uniqueBoardTerms(terms)
}

func extractCJKQueryTerms(query string) []string {
	stopwords := map[string]struct{}{
		"股票": {}, "个股": {}, "相关": {}, "板块": {}, "概念": {}, "行业": {},
		"我想": {}, "想买": {}, "买入": {}, "推荐": {}, "看看": {}, "一下": {}, "哪些": {},
	}
	sequences := make([]string, 0, 4)
	current := make([]rune, 0, len(query))
	flush := func() {
		if len(current) >= 2 {
			sequences = append(sequences, string(current))
		}
		current = current[:0]
	}
	for _, r := range query {
		if r >= 0x4e00 && r <= 0x9fff {
			current = append(current, r)
			continue
		}
		flush()
	}
	flush()
	terms := make([]string, 0, len(sequences)*4)
	for _, seq := range sequences {
		runes := []rune(seq)
		for size := 2; size <= 4; size++ {
			if len(runes) < size {
				continue
			}
			for i := 0; i+size <= len(runes); i++ {
				term := strings.TrimSpace(string(runes[i : i+size]))
				if _, blocked := stopwords[term]; blocked {
					continue
				}
				terms = append(terms, term)
			}
		}
	}
	return uniqueBoardTerms(terms)
}

func scoreBoardQueryMatch(board marketResponse.MarketBoardItemResponse, constituents []model.MarketBoardConstituent, terms []string) int {
	if len(terms) == 0 {
		return 0
	}
	score := 0
	name := strings.ToLower(strings.TrimSpace(board.Name))
	code := strings.ToLower(strings.TrimSpace(board.Code))
	for _, term := range terms {
		switch {
		case term == "":
			continue
		case strings.Contains(name, term):
			score += 12
		case strings.Contains(code, term):
			score += 8
		}
	}
	if score == 0 && len(constituents) > 0 {
		for _, constituent := range constituents {
			stockName := strings.ToLower(strings.TrimSpace(constituent.StockName))
			symbol := strings.ToLower(strings.TrimSpace(constituent.Symbol))
			for _, term := range terms {
				if term == "" {
					continue
				}
				if strings.Contains(stockName, term) {
					score += 5
					break
				}
				if strings.Contains(symbol, term) {
					score += 3
					break
				}
			}
			if score >= 15 {
				break
			}
		}
	}
	return score
}

func scoreTaxonomyDocument(snapshotName string, detail model.StockQuoteDetail, terms []string) int {
	if len(terms) == 0 {
		return 0
	}
	name := strings.ToLower(strings.TrimSpace(fallbackString(snapshotName, detail.Name)))
	industry := strings.ToLower(strings.TrimSpace(model.NormalizeIndustryLabel(detail.Industry)))
	concepts := model.NormalizeConceptList(strings.Split(detail.Concepts, ","))
	score := 0
	for _, rawTerm := range terms {
		term := strings.ToLower(strings.TrimSpace(rawTerm))
		if len([]rune(term)) < 2 {
			continue
		}
		if name != "" && strings.Contains(name, term) {
			score += 6
		}
		if industry != "" && strings.Contains(industry, term) {
			score += 10
		}
		for _, concept := range concepts {
			lowerConcept := strings.ToLower(strings.TrimSpace(concept))
			if lowerConcept != "" && strings.Contains(lowerConcept, term) {
				score += 8
				break
			}
		}
	}
	return score
}

func uniqueBoardTerms(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.ToLower(strings.TrimSpace(item))
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

func convertSnapshots(snapshots []model.MarketSnapshot) []marketResponse.MarketSnapshotResponse {
	responses := make([]marketResponse.MarketSnapshotResponse, 0, len(snapshots))
	for _, snapshot := range snapshots {
		responses = append(responses, marketResponse.MarketSnapshotResponse{
			Symbol:        snapshot.Symbol,
			Name:          snapshot.Name,
			Market:        snapshot.Market,
			SnapshotTime:  snapshot.SnapshotTime.Format("2006-01-02 15:04:05"),
			LastPrice:     snapshot.LastPrice.String(),
			ChangeAmount:  snapshot.ChangeAmount.String(),
			ChangePercent: snapshot.ChangePercent.String(),
			OpenPrice:     snapshot.OpenPrice.String(),
			HighPrice:     snapshot.HighPrice.String(),
			LowPrice:      snapshot.LowPrice.String(),
			PrevClose:     snapshot.PrevClose.String(),
			Volume:        snapshot.Volume.String(),
			Turnover:      snapshot.Turnover.String(),
			Source:        snapshot.Source,
			BatchNo:       snapshot.BatchNo,
		})
	}
	return responses
}

func convertBreadthItems(snapshots []model.MarketSnapshot, limit int) []marketResponse.MarketBreadthItemResponse {
	if limit > len(snapshots) {
		limit = len(snapshots)
	}
	responses := make([]marketResponse.MarketBreadthItemResponse, 0, limit)
	for _, snapshot := range snapshots[:limit] {
		responses = append(responses, marketResponse.MarketBreadthItemResponse{
			Symbol:         snapshot.Symbol,
			Name:           snapshot.Name,
			Market:         snapshot.Market,
			LastPrice:      snapshot.LastPrice.String(),
			ChangeAmount:   snapshot.ChangeAmount.String(),
			ChangePercent:  snapshot.ChangePercent.String(),
			OpenPrice:      snapshot.OpenPrice.String(),
			HighPrice:      snapshot.HighPrice.String(),
			LowPrice:       snapshot.LowPrice.String(),
			PrevClose:      snapshot.PrevClose.String(),
			Volume:         snapshot.Volume.String(),
			Turnover:       snapshot.Turnover.String(),
			TurnoverRate:   "",
			TotalMarketCap: "",
			FloatMarketCap: "",
		})
	}
	return responses
}

func convertBoardItems(items []model.MarketBoardSnapshot) []marketResponse.MarketBoardItemResponse {
	responses := make([]marketResponse.MarketBoardItemResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, marketResponse.MarketBoardItemResponse{
			BoardType:       item.BoardType,
			Code:            item.Code,
			Name:            decodeBoardDisplayName(item.Name),
			LastPrice:       item.LastPrice.String(),
			ChangeAmount:    item.ChangeAmount.String(),
			ChangePercent:   item.ChangePercent.String(),
			Volume:          item.Volume.String(),
			Turnover:        item.Turnover.String(),
			TotalMarketCap:  item.TotalMarketCap.String(),
			FloatMarketCap:  item.FloatMarketCap.String(),
			StockCount:      item.StockCount,
			RiseCount:       item.RiseCount,
			FallCount:       item.FallCount,
			FlatCount:       item.FlatCount,
			ConstituentNode: item.ConstituentNode,
		})
	}
	return responses
}

func decodeBoardDisplayName(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || !strings.Contains(trimmed, `\u`) {
		return trimmed
	}
	decoded, err := strconv.Unquote(`"` + strings.ReplaceAll(trimmed, `"`, `\"`) + `"`)
	if err != nil {
		return trimmed
	}
	return strings.TrimSpace(decoded)
}

func sortSnapshots(snapshots []model.MarketSnapshot, less func(a, b model.MarketSnapshot) bool) []model.MarketSnapshot {
	copied := append([]model.MarketSnapshot(nil), snapshots...)
	sort.Slice(copied, func(i, j int) bool { return less(copied[i], copied[j]) })
	return copied
}

func latestSnapshotByCreatedAt(snapshots []model.MarketSnapshot) model.MarketSnapshot {
	latest := snapshots[0]
	for _, snapshot := range snapshots[1:] {
		if snapshot.CreatedAt.After(latest.CreatedAt) || snapshot.ID > latest.ID {
			latest = snapshot
		}
	}
	return latest
}

func shouldRefreshBreadthSnapshots(snapshots []model.MarketSnapshot) bool {
	if len(snapshots) < minBreadthSnapshotCoverage {
		return true
	}
	latest := latestSnapshotByCreatedAt(snapshots)
	return time.Since(latest.CreatedAt) > 30*time.Minute
}

func isBreadthBatchUsable(snapshots []model.MarketSnapshot) bool {
	return len(snapshots) >= minBreadthSnapshotCoverage
}

func shouldRefreshBoardSnapshots(sectors, concepts []model.MarketBoardSnapshot) bool {
	if len(sectors) == 0 && len(concepts) == 0 {
		return true
	}

	var latest time.Time
	for _, item := range sectors {
		if item.CreatedAt.After(latest) {
			latest = item.CreatedAt
		}
	}
	for _, item := range concepts {
		if item.CreatedAt.After(latest) {
			latest = item.CreatedAt
		}
	}
	if latest.IsZero() {
		return true
	}
	return time.Since(latest) > 30*time.Minute
}

func buildBreadthCoverage(snapshots []model.MarketSnapshot) []marketResponse.DashboardStatResponse {
	return []marketResponse.DashboardStatResponse{
		{Label: "覆盖标的", Value: formatInt(len(snapshots))},
		{Label: "上涨数", Value: formatInt(countPositiveStrict(snapshots))},
		{Label: "下跌数", Value: formatInt(countNegative(snapshots))},
		{Label: "平盘数", Value: formatInt(countFlat(snapshots))},
		{Label: "平均涨跌幅", Value: averageChangePercent(snapshots)},
		{Label: "总成交额", Value: totalTurnover(snapshots)},
	}
}

func buildChangeDistribution(snapshots []model.MarketSnapshot) []marketResponse.DistributionBucketResponse {
	buckets := []marketResponse.DistributionBucketResponse{
		{Label: "<-5%", Min: "", Max: "-5", Count: 0},
		{Label: "-5~-2%", Min: "-5", Max: "-2", Count: 0},
		{Label: "-2~0%", Min: "-2", Max: "0", Count: 0},
		{Label: "0~2%", Min: "0", Max: "2", Count: 0},
		{Label: "2~5%", Min: "2", Max: "5", Count: 0},
		{Label: ">5%", Min: "5", Max: "", Count: 0},
	}
	for _, snapshot := range snapshots {
		value, _ := snapshot.ChangePercent.Float64()
		switch {
		case value < -5:
			buckets[0].Count++
		case value < -2:
			buckets[1].Count++
		case value < 0:
			buckets[2].Count++
		case value < 2:
			buckets[3].Count++
		case value < 5:
			buckets[4].Count++
		default:
			buckets[5].Count++
		}
	}
	return buckets
}

func buildTurnoverDistribution(snapshots []model.MarketSnapshot) []marketResponse.DistributionBucketResponse {
	buckets := []marketResponse.DistributionBucketResponse{
		{Label: "<1亿", Min: "", Max: "100000000", Count: 0},
		{Label: "1~5亿", Min: "100000000", Max: "500000000", Count: 0},
		{Label: "5~10亿", Min: "500000000", Max: "1000000000", Count: 0},
		{Label: "10~50亿", Min: "1000000000", Max: "5000000000", Count: 0},
		{Label: ">50亿", Min: "5000000000", Max: "", Count: 0},
	}
	for _, snapshot := range snapshots {
		value, _ := snapshot.Turnover.Float64()
		switch {
		case value < 100000000:
			buckets[0].Count++
		case value < 500000000:
			buckets[1].Count++
		case value < 1000000000:
			buckets[2].Count++
		case value < 5000000000:
			buckets[3].Count++
		default:
			buckets[4].Count++
		}
	}
	return buckets
}

func countPositiveStrict(snapshots []model.MarketSnapshot) int {
	count := 0
	for _, snapshot := range snapshots {
		if snapshot.ChangeAmount.GreaterThan(modelDecimalZero()) {
			count++
		}
	}
	return count
}

func countFlat(snapshots []model.MarketSnapshot) int {
	count := 0
	for _, snapshot := range snapshots {
		if snapshot.ChangeAmount.Equal(modelDecimalZero()) {
			count++
		}
	}
	return count
}

func fallbackMessage(message, fallback string) string {
	if strings.TrimSpace(message) != "" {
		return message
	}
	return fallback
}

func countPositive(snapshots []model.MarketSnapshot) int {
	count := 0
	for _, snapshot := range snapshots {
		if snapshot.ChangeAmount.GreaterThanOrEqual(modelDecimalZero()) {
			count++
		}
	}
	return count
}

func countNegative(snapshots []model.MarketSnapshot) int {
	count := 0
	for _, snapshot := range snapshots {
		if snapshot.ChangeAmount.LessThan(modelDecimalZero()) {
			count++
		}
	}
	return count
}

func averageChangePercent(snapshots []model.MarketSnapshot) string {
	if len(snapshots) == 0 {
		return "0%"
	}
	total := modelDecimalZero()
	for _, snapshot := range snapshots {
		total = total.Add(snapshot.ChangePercent)
	}
	return total.Div(modelDecimalFromInt(len(snapshots))).StringFixed(2) + "%"
}

func totalTurnover(snapshots []model.MarketSnapshot) string {
	total := modelDecimalZero()
	for _, snapshot := range snapshots {
		total = total.Add(snapshot.Turnover)
	}
	return total.Div(modelDecimalFromInt(100000000)).StringFixed(2) + "亿"
}

func formatInt(value int) string {
	return modelDecimalFromInt(value).String()
}

func toBoardDecimal(value string) decimal.Decimal {
	parsed, err := decimal.NewFromString(strings.TrimSpace(value))
	if err != nil {
		return decimal.Zero
	}
	return parsed
}
