package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/pkg/marketdata"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type MarketDataService interface {
	FetchAndStoreMarketSnapshots(ctx context.Context) (string, int, error)
	FetchAndStoreQuotesBySymbols(ctx context.Context, symbols []string) ([]model.MarketSnapshot, error)
	FetchAndStoreFullMarketSnapshots(ctx context.Context) (string, int, error)
	FetchAndStoreMarketBoardSnapshots(ctx context.Context) (string, int, error)
	EnsureTrackedIndexHistory(ctx context.Context) error
}

const (
	fullMarketPageSize         = 100
	fullMarketMaxPages         = 60
	fullMarketBatchSize        = 200
	minFullMarketCoverage      = 5000
	defaultFullMarketDelay     = 350 * time.Millisecond
	rankingRetryAttempts       = 3
	boardUniverseSyncInterval  = 24 * time.Hour
	recentFullMarketBatchLimit = 6
)

type akshareFullSnapshotFetcher func(ctx context.Context, pythonPath, scriptPath string) ([]marketdata.Quote, error)
type akshareBoardUniverseFetcher func(ctx context.Context, pythonPath, scriptPath string, extraEnv []string) ([]akshareBoardPayload, error)

type marketDataService struct {
	marketConfig         config.MarketConfig
	provider             marketdata.Provider
	rankingProvider      marketdata.MarketRankingProvider
	snapshotRepo         repository.MarketSnapshotRepository
	boardSnapshotRepo    repository.MarketBoardSnapshotRepository
	boardConstituentRepo repository.MarketBoardConstituentRepository
	klineRepo            repository.StockKlineRepository
	universeMu           sync.RWMutex
	fullMarketSymbols    []string
	fullMarketSource     string
	akshareFetcher       akshareFullSnapshotFetcher
	akshareBoardFetcher  akshareBoardUniverseFetcher
}

func NewMarketDataService(
	marketConfig config.MarketConfig,
	provider marketdata.Provider,
	rankingProvider marketdata.MarketRankingProvider,
	snapshotRepo repository.MarketSnapshotRepository,
	boardSnapshotRepo repository.MarketBoardSnapshotRepository,
	boardConstituentRepo repository.MarketBoardConstituentRepository,
	klineRepo repository.StockKlineRepository,
) MarketDataService {
	return &marketDataService{
		marketConfig:         marketConfig,
		provider:             provider,
		rankingProvider:      rankingProvider,
		snapshotRepo:         snapshotRepo,
		boardSnapshotRepo:    boardSnapshotRepo,
		boardConstituentRepo: boardConstituentRepo,
		klineRepo:            klineRepo,
		akshareFetcher:       fetchAKShareFullMarketQuotes,
		akshareBoardFetcher:  fetchAKShareBoardUniverse,
	}
}

func (s *marketDataService) FetchAndStoreMarketSnapshots(ctx context.Context) (string, int, error) {
	symbols := s.symbols()
	if len(symbols) == 0 {
		return "", 0, fmt.Errorf("market symbols are empty")
	}
	snapshots, err := s.fetchAndStoreBySymbols(ctx, symbols)
	if err != nil {
		return "", 0, err
	}
	if len(snapshots) == 0 {
		return "", 0, fmt.Errorf("no quotes returned")
	}
	return snapshots[0].BatchNo, len(snapshots), nil
}

func (s *marketDataService) FetchAndStoreQuotesBySymbols(ctx context.Context, symbols []string) ([]model.MarketSnapshot, error) {
	normalized := normalizeSymbols(symbols)
	if len(normalized) == 0 {
		return []model.MarketSnapshot{}, nil
	}
	return s.fetchAndStoreBySymbols(ctx, normalized)
}

func (s *marketDataService) FetchAndStoreFullMarketSnapshots(ctx context.Context) (string, int, error) {
	if s.usesAKShareFullSnapshot() {
		return s.fetchAndStoreAKShareFullMarketSnapshots(ctx)
	}
	if s.provider == nil {
		return "", 0, fmt.Errorf("market quote provider is unavailable")
	}
	if s.rankingProvider == nil {
		return "", 0, fmt.Errorf("market ranking provider is unavailable")
	}
	symbols, _, err := s.ensureFullMarketUniverse(ctx)
	if err != nil {
		return "", 0, err
	}

	delay := s.fullMarketRequestDelay()
	quotes := make([]marketdata.Quote, 0, len(symbols))
	quoteBatchSource := ""
	for start := 0; start < len(symbols); start += fullMarketBatchSize {
		end := start + fullMarketBatchSize
		if end > len(symbols) {
			end = len(symbols)
		}

		batchQuotes, err := s.provider.GetQuotes(ctx, symbols[start:end])
		if err != nil {
			return "", 0, fmt.Errorf("fetch full market quote batch %d-%d failed: %w", start, end, err)
		}
		if len(batchQuotes) == 0 {
			return "", 0, fmt.Errorf("fetch full market quote batch %d-%d returned empty quotes", start, end)
		}
		batchSource := quoteSource(batchQuotes)
		if quoteBatchSource == "" {
			quoteBatchSource = batchSource
		}
		if quoteBatchSource != "" && batchSource != "" && !strings.EqualFold(quoteBatchSource, batchSource) {
			return "", 0, fmt.Errorf("full market quote source changed from %s to %s", quoteBatchSource, batchSource)
		}
		quotes = append(quotes, batchQuotes...)

		if end < len(symbols) {
			if err := sleepWithContext(ctx, delay); err != nil {
				return "", 0, err
			}
		}
	}
	if len(quotes) < minFullMarketCoverage {
		return "", 0, fmt.Errorf("insufficient full market quotes: got %d, want at least %d", len(quotes), minFullMarketCoverage)
	}

	snapshots := convertQuotesToSnapshots(quotes)
	if err := s.snapshotRepo.BatchCreate(snapshots); err != nil {
		return "", 0, err
	}
	return snapshots[0].BatchNo, len(snapshots), nil
}

func (s *marketDataService) fetchAndStoreAKShareFullMarketSnapshots(ctx context.Context) (string, int, error) {
	if s.akshareFetcher == nil {
		return "", 0, fmt.Errorf("akshare full snapshot fetcher is unavailable")
	}
	quotes, err := s.akshareFetcher(ctx, strings.TrimSpace(s.marketConfig.AKSharePythonPath), strings.TrimSpace(s.marketConfig.AKShareScriptPath))
	if err != nil {
		return "", 0, err
	}
	if len(quotes) < minFullMarketCoverage {
		return "", 0, fmt.Errorf("insufficient akshare full market quotes: got %d, want at least %d", len(quotes), minFullMarketCoverage)
	}
	snapshots := convertQuotesToSnapshots(quotes)
	if err := s.snapshotRepo.BatchCreate(snapshots); err != nil {
		return "", 0, err
	}
	return snapshots[0].BatchNo, len(snapshots), nil
}

func (s *marketDataService) FetchAndStoreMarketBoardSnapshots(ctx context.Context) (string, int, error) {
	if s.boardSnapshotRepo == nil {
		return "", 0, fmt.Errorf("market board repository is unavailable")
	}
	if s.boardConstituentRepo == nil {
		return "", 0, fmt.Errorf("market board constituent repository is unavailable")
	}
	constituents, err := s.loadBoardConstituents(ctx)
	if err != nil {
		return "", 0, err
	}
	if len(constituents) == 0 {
		return "", 0, fmt.Errorf("no board constituents available")
	}

	latestBatchNo, snapshots, err := s.loadLatestUsableFullMarketSnapshots()
	if err != nil {
		return "", 0, err
	}
	_ = latestBatchNo
	if len(snapshots) == 0 {
		return "", 0, fmt.Errorf("no full market snapshots available")
	}

	batchNo := time.Now().Format("20060102150405") + "-" + uuid.NewString()
	boards := buildBoardSnapshotsFromConstituents(constituents, snapshots, batchNo)
	if len(boards) == 0 {
		return "", 0, fmt.Errorf("no market board snapshots built from board constituents")
	}
	if err := s.boardSnapshotRepo.BatchCreate(boards); err != nil {
		return "", 0, err
	}
	return batchNo, len(boards), nil
}

func (s *marketDataService) fetchAndStoreBySymbols(ctx context.Context, symbols []string) ([]model.MarketSnapshot, error) {
	quotes, err := s.provider.GetQuotes(ctx, symbols)
	if err != nil {
		return nil, err
	}
	if len(quotes) == 0 {
		return nil, fmt.Errorf("no quotes returned")
	}

	snapshots := convertQuotesToSnapshots(quotes)
	if err := s.snapshotRepo.BatchCreate(snapshots); err != nil {
		return nil, err
	}
	return snapshots, nil
}

func (s *marketDataService) EnsureTrackedIndexHistory(ctx context.Context) error {
	if s.provider == nil || s.klineRepo == nil {
		return nil
	}

	var firstErr error
	for _, symbol := range s.symbols() {
		normalized := normalizeSymbol(symbol)
		if _, err := s.klineRepo.FindLatestBar(normalized, "day", "none"); err == nil {
			continue
		}

		bars, err := s.provider.GetKlines(ctx, normalized, "day", "none", 30)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if len(bars) == 0 {
			continue
		}
		if err := s.klineRepo.UpsertBars(convertKlinesToModels(bars)); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func convertQuotesToSnapshots(quotes []marketdata.Quote) []model.MarketSnapshot {
	batchNo := time.Now().Format("20060102150405") + "-" + uuid.NewString()
	snapshots := make([]model.MarketSnapshot, 0, len(quotes))
	for _, quote := range quotes {
		snapshots = append(snapshots, model.MarketSnapshot{
			Symbol:        quote.Symbol,
			Name:          quote.Name,
			Market:        quote.Market,
			SnapshotTime:  quote.SnapshotTime,
			LastPrice:     decimal.NewFromFloat(quote.LastPrice),
			ChangeAmount:  decimal.NewFromFloat(quote.ChangeAmount),
			ChangePercent: decimal.NewFromFloat(quote.ChangePercent),
			OpenPrice:     decimal.NewFromFloat(quote.OpenPrice),
			HighPrice:     decimal.NewFromFloat(quote.HighPrice),
			LowPrice:      decimal.NewFromFloat(quote.LowPrice),
			PrevClose:     decimal.NewFromFloat(quote.PrevClose),
			Volume:        decimal.NewFromFloat(quote.Volume),
			Turnover:      decimal.NewFromFloat(quote.Turnover),
			Source:        quote.Source,
			BatchNo:       batchNo,
		})
	}
	return snapshots
}

func convertRankItemsToSnapshots(items []marketdata.MarketRankItem) []model.MarketSnapshot {
	batchNo := time.Now().Format("20060102150405") + "-" + uuid.NewString()
	snapshots := make([]model.MarketSnapshot, 0, len(items))
	for _, item := range items {
		snapshots = append(snapshots, model.MarketSnapshot{
			Symbol:        item.Symbol,
			Name:          item.Name,
			Market:        item.Market,
			SnapshotTime:  item.SnapshotTime,
			LastPrice:     decimal.NewFromFloat(item.LastPrice),
			ChangeAmount:  decimal.NewFromFloat(item.ChangeAmount),
			ChangePercent: decimal.NewFromFloat(item.ChangePercent),
			OpenPrice:     decimal.NewFromFloat(item.OpenPrice),
			HighPrice:     decimal.NewFromFloat(item.HighPrice),
			LowPrice:      decimal.NewFromFloat(item.LowPrice),
			PrevClose:     decimal.NewFromFloat(item.PrevClose),
			Volume:        decimal.NewFromFloat(item.Volume),
			Turnover:      decimal.NewFromFloat(item.Turnover),
			Source:        item.Source,
			BatchNo:       batchNo,
		})
	}
	return snapshots
}

func (s *marketDataService) ensureFullMarketUniverse(ctx context.Context) ([]string, string, error) {
	s.universeMu.RLock()
	if len(s.fullMarketSymbols) > 0 {
		cached := append([]string(nil), s.fullMarketSymbols...)
		source := s.fullMarketSource
		s.universeMu.RUnlock()
		return cached, source, nil
	}
	s.universeMu.RUnlock()

	collected := make([]string, 0, 6000)
	seen := make(map[string]struct{}, 6000)
	universeSource := ""
	for page := 1; page <= fullMarketMaxPages; page++ {
		items, err := retryMarketDataCall(ctx, rankingRetryAttempts, s.fullMarketRequestDelay(), func() ([]marketdata.MarketRankItem, error) {
			return s.rankingProvider.GetMarketRankings(ctx, "symbol", false, page, fullMarketPageSize)
		})
		if err != nil {
			return nil, "", fmt.Errorf("fetch full market universe page %d failed: %w", page, err)
		}
		if len(items) == 0 {
			break
		}
		pageSource := rankItemSource(items)
		if universeSource == "" {
			universeSource = pageSource
		}
		if universeSource != "" && pageSource != "" && !strings.EqualFold(universeSource, pageSource) {
			return nil, "", fmt.Errorf("full market universe source changed from %s to %s on page %d", universeSource, pageSource, page)
		}
		for _, item := range items {
			if item.Symbol == "" {
				continue
			}
			if _, ok := seen[item.Symbol]; ok {
				continue
			}
			seen[item.Symbol] = struct{}{}
			collected = append(collected, item.Symbol)
		}
		if len(items) < fullMarketPageSize {
			break
		}
		if err := sleepWithContext(ctx, s.fullMarketRequestDelay()); err != nil {
			return nil, "", err
		}
	}
	if len(collected) < minFullMarketCoverage {
		return nil, "", fmt.Errorf("insufficient full market universe size: got %d, want at least %d", len(collected), minFullMarketCoverage)
	}

	s.universeMu.Lock()
	s.fullMarketSymbols = append([]string(nil), collected...)
	s.fullMarketSource = universeSource
	s.universeMu.Unlock()
	return append([]string(nil), collected...), universeSource, nil
}

func (s *marketDataService) fullMarketRequestDelay() time.Duration {
	if s.marketConfig.SinaRequestDelayMS > 0 {
		return time.Duration(s.marketConfig.SinaRequestDelayMS) * time.Millisecond
	}
	return defaultFullMarketDelay
}

func (s *marketDataService) usesAKShareFullSnapshot() bool {
	return strings.EqualFold(strings.TrimSpace(s.marketConfig.FullSnapshotSource), "akshare")
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func retryMarketDataCall[T any](ctx context.Context, attempts int, delay time.Duration, fn func() (T, error)) (T, error) {
	var zero T
	if attempts <= 0 {
		attempts = 1
	}
	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		value, err := fn()
		if err == nil {
			return value, nil
		}
		lastErr = err
		if attempt == attempts {
			break
		}
		if waitErr := sleepWithContext(ctx, delay); waitErr != nil {
			return zero, waitErr
		}
	}
	return zero, lastErr
}

func rankItemSource(items []marketdata.MarketRankItem) string {
	for _, item := range items {
		if strings.TrimSpace(item.Source) != "" {
			return strings.TrimSpace(item.Source)
		}
	}
	return ""
}

func quoteSource(quotes []marketdata.Quote) string {
	for _, quote := range quotes {
		if strings.TrimSpace(quote.Source) != "" {
			return strings.TrimSpace(quote.Source)
		}
	}
	return ""
}

type akshareQuotePayload struct {
	Symbol        string  `json:"symbol"`
	Name          string  `json:"name"`
	Market        string  `json:"market"`
	LastPrice     float64 `json:"last_price"`
	ChangeAmount  float64 `json:"change_amount"`
	ChangePercent float64 `json:"change_percent"`
	OpenPrice     float64 `json:"open_price"`
	HighPrice     float64 `json:"high_price"`
	LowPrice      float64 `json:"low_price"`
	PrevClose     float64 `json:"prev_close"`
	Volume        float64 `json:"volume"`
	Turnover      float64 `json:"turnover"`
	Source        string  `json:"source"`
}

type akshareBoardPayload struct {
	BoardType    string                           `json:"board_type"`
	Code         string                           `json:"code"`
	Name         string                           `json:"name"`
	Source       string                           `json:"source"`
	CompanyCount int                              `json:"company_count"`
	Constituents []akshareBoardConstituentPayload `json:"constituents"`
}

type akshareBoardConstituentPayload struct {
	Symbol         string  `json:"symbol"`
	Name           string  `json:"name"`
	TotalMarketCap float64 `json:"total_market_cap"`
	FloatMarketCap float64 `json:"float_market_cap"`
}

func fetchAKShareFullMarketQuotes(ctx context.Context, pythonPath, scriptPath string) ([]marketdata.Quote, error) {
	output, err := runAKShareScript(ctx, pythonPath, scriptPath, nil)
	if err != nil {
		return nil, err
	}

	var payload []akshareQuotePayload
	if err := json.Unmarshal(output, &payload); err != nil {
		return nil, fmt.Errorf("parse akshare full market payload failed: %w", err)
	}
	now := time.Now().Truncate(time.Minute)
	quotes := make([]marketdata.Quote, 0, len(payload))
	for _, item := range payload {
		if strings.TrimSpace(item.Symbol) == "" {
			continue
		}
		market := strings.TrimSpace(item.Market)
		if market == "" {
			market = "cn_stock"
		}
		source := strings.TrimSpace(item.Source)
		if source == "" {
			source = "akshare_sina"
		}
		quotes = append(quotes, marketdata.Quote{
			Symbol:        strings.TrimSpace(item.Symbol),
			Name:          strings.TrimSpace(item.Name),
			Market:        market,
			SnapshotTime:  now,
			LastPrice:     item.LastPrice,
			ChangeAmount:  item.ChangeAmount,
			ChangePercent: item.ChangePercent,
			OpenPrice:     item.OpenPrice,
			HighPrice:     item.HighPrice,
			LowPrice:      item.LowPrice,
			PrevClose:     item.PrevClose,
			Volume:        item.Volume,
			Turnover:      item.Turnover,
			Source:        source,
		})
	}
	return quotes, nil
}

func fetchAKShareBoardUniverse(ctx context.Context, pythonPath, scriptPath string, extraEnv []string) ([]akshareBoardPayload, error) {
	output, err := runAKShareScript(ctx, pythonPath, scriptPath, extraEnv)
	if err != nil {
		return nil, err
	}

	var payload []akshareBoardPayload
	if err := json.Unmarshal(output, &payload); err != nil {
		return nil, fmt.Errorf("parse akshare board payload failed: %w", err)
	}
	return payload, nil
}

func runAKShareScript(ctx context.Context, pythonPath, scriptPath string, extraEnv []string) ([]byte, error) {
	if strings.TrimSpace(scriptPath) == "" {
		return nil, fmt.Errorf("akshare script path is empty")
	}

	var (
		output  []byte
		lastErr error
	)
	for _, candidate := range pythonCandidates(pythonPath) {
		cmd := exec.CommandContext(ctx, candidate, scriptPath)
		if len(extraEnv) > 0 {
			cmd.Env = append(cmd.Env, extraEnv...)
		}
		var stderr strings.Builder
		cmd.Stderr = &stderr
		stdout, err := cmd.Output()
		if err == nil {
			output = stdout
			lastErr = nil
			break
		}
		message := strings.TrimSpace(stderr.String())
		if message != "" {
			lastErr = fmt.Errorf("run akshare script with %s failed: %v: %s", candidate, err, message)
		} else {
			lastErr = fmt.Errorf("run akshare script with %s failed: %w", candidate, err)
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return output, nil
}

func pythonCandidates(pythonPath string) []string {
	candidates := make([]string, 0, 3)
	if strings.TrimSpace(pythonPath) != "" {
		candidates = append(candidates, strings.TrimSpace(pythonPath))
	}
	if runtime.GOOS == "windows" {
		return append(candidates, "python")
	}
	return append(candidates, "python3", "python")
}

func (s *marketDataService) loadBoardConstituents(ctx context.Context) ([]model.MarketBoardConstituent, error) {
	if s.boardConstituentRepo == nil {
		return nil, fmt.Errorf("market board constituent repository is unavailable")
	}

	current, err := s.boardConstituentRepo.FindAll()
	if err != nil {
		return nil, err
	}
	latestSyncedAt, err := s.boardConstituentRepo.FindLatestSyncedAt()
	needsSync := err != nil || latestSyncedAt.IsZero() || time.Since(latestSyncedAt) >= boardUniverseSyncInterval || len(current) == 0
	if !needsSync {
		return current, nil
	}

	refreshed, syncErr := s.syncBoardConstituents(ctx)
	if syncErr == nil {
		return refreshed, nil
	}
	if len(current) > 0 {
		return current, nil
	}
	return nil, syncErr
}

func (s *marketDataService) syncBoardConstituents(ctx context.Context) ([]model.MarketBoardConstituent, error) {
	if s.akshareBoardFetcher == nil {
		return nil, fmt.Errorf("akshare board fetcher is unavailable")
	}
	extraEnv := []string{
		fmt.Sprintf("THS_PAGE_DELAY=%.3f", float64(s.marketConfig.THSPageDelayMS)/1000.0),
		fmt.Sprintf("THS_BOARD_DELAY=%.3f", float64(s.marketConfig.THSBoardDelayMS)/1000.0),
		fmt.Sprintf("THS_RETRY_BASE_DELAY=%.3f", float64(s.marketConfig.THSRetryBaseDelayMS)/1000.0),
	}
	payload, err := s.akshareBoardFetcher(ctx, s.marketConfig.AKSharePythonPath, s.marketConfig.AKShareBoardScriptPath, extraEnv)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	items := make([]model.MarketBoardConstituent, 0, len(payload)*32)
	for _, board := range payload {
		boardType := strings.TrimSpace(board.BoardType)
		boardCode := strings.TrimSpace(board.Code)
		boardName := strings.TrimSpace(board.Name)
		if boardType == "" || boardCode == "" || boardName == "" {
			continue
		}
		source := strings.TrimSpace(board.Source)
		if source == "" {
			source = "akshare_sector"
		}
		seen := make(map[string]struct{}, len(board.Constituents))
		for _, constituent := range board.Constituents {
			symbol := normalizeSymbol(constituent.Symbol)
			if symbol == "" {
				continue
			}
			if _, ok := seen[symbol]; ok {
				continue
			}
			seen[symbol] = struct{}{}
			items = append(items, model.MarketBoardConstituent{
				BoardType:      boardType,
				BoardCode:      boardCode,
				BoardName:      boardName,
				Symbol:         symbol,
				StockName:      strings.TrimSpace(constituent.Name),
				TotalMarketCap: decimal.NewFromFloat(constituent.TotalMarketCap),
				FloatMarketCap: decimal.NewFromFloat(constituent.FloatMarketCap),
				Source:         source,
				SyncedAt:       now,
			})
		}
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("no akshare board constituents parsed")
	}
	if err := s.boardConstituentRepo.ReplaceAll(items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *marketDataService) loadLatestUsableFullMarketSnapshots() (string, []model.MarketSnapshot, error) {
	if s.snapshotRepo == nil {
		return "", nil, fmt.Errorf("market snapshot repository is unavailable")
	}

	batchNo, err := s.snapshotRepo.FindLatestBatchNo()
	if err != nil {
		return "", nil, err
	}
	snapshots, err := s.snapshotRepo.FindByBatchNo(batchNo)
	if err != nil {
		return "", nil, err
	}
	if len(snapshots) >= minFullMarketCoverage {
		return batchNo, snapshots, nil
	}

	batchNos, err := s.snapshotRepo.FindRecentBatchNos(recentFullMarketBatchLimit)
	if err != nil {
		return batchNo, snapshots, err
	}
	for _, candidate := range batchNos {
		if candidate == "" || candidate == batchNo {
			continue
		}
		candidateSnapshots, candidateErr := s.snapshotRepo.FindByBatchNo(candidate)
		if candidateErr != nil {
			continue
		}
		if len(candidateSnapshots) >= minFullMarketCoverage {
			return candidate, candidateSnapshots, nil
		}
	}
	return batchNo, snapshots, nil
}

type boardSnapshotAggregate struct {
	boardType        string
	boardCode        string
	boardName        string
	source           string
	stockCount       int
	matchedCount     int
	lastPriceSum     decimal.Decimal
	changeAmountSum  decimal.Decimal
	changePercentSum decimal.Decimal
	volumeSum        decimal.Decimal
	turnoverSum      decimal.Decimal
	totalMarketCap   decimal.Decimal
	floatMarketCap   decimal.Decimal
	riseCount        int
	fallCount        int
	flatCount        int
}

func buildBoardSnapshotsFromConstituents(constituents []model.MarketBoardConstituent, snapshots []model.MarketSnapshot, batchNo string) []model.MarketBoardSnapshot {
	if len(constituents) == 0 || len(snapshots) == 0 {
		return []model.MarketBoardSnapshot{}
	}

	snapshotBySymbol := make(map[string]model.MarketSnapshot, len(snapshots))
	latestTime := snapshots[0].SnapshotTime
	latestSource := strings.TrimSpace(snapshots[0].Source)
	for _, snapshot := range snapshots {
		snapshotBySymbol[snapshot.Symbol] = snapshot
		if snapshot.SnapshotTime.After(latestTime) {
			latestTime = snapshot.SnapshotTime
		}
	}
	if latestSource == "" {
		latestSource = "market"
	}

	aggregates := make(map[string]*boardSnapshotAggregate, 320)
	seenByBoard := make(map[string]map[string]struct{}, 320)
	for _, item := range constituents {
		boardType := strings.TrimSpace(item.BoardType)
		boardCode := strings.TrimSpace(item.BoardCode)
		if boardType == "" || boardCode == "" {
			continue
		}
		key := boardType + "|" + boardCode
		aggregate := aggregates[key]
		if aggregate == nil {
			aggregate = &boardSnapshotAggregate{
				boardType: boardType,
				boardCode: boardCode,
				boardName: strings.TrimSpace(item.BoardName),
				source:    strings.TrimSpace(item.Source),
			}
			if aggregate.source == "" {
				aggregate.source = "akshare_sector"
			}
			aggregates[key] = aggregate
			seenByBoard[key] = make(map[string]struct{}, 64)
		}
		symbol := strings.TrimSpace(item.Symbol)
		if symbol == "" {
			continue
		}
		if _, ok := seenByBoard[key][symbol]; ok {
			continue
		}
		seenByBoard[key][symbol] = struct{}{}
		aggregate.stockCount++
		aggregate.totalMarketCap = aggregate.totalMarketCap.Add(item.TotalMarketCap)
		aggregate.floatMarketCap = aggregate.floatMarketCap.Add(item.FloatMarketCap)

		snapshot, ok := snapshotBySymbol[symbol]
		if !ok {
			continue
		}
		aggregate.matchedCount++
		aggregate.lastPriceSum = aggregate.lastPriceSum.Add(snapshot.LastPrice)
		aggregate.changeAmountSum = aggregate.changeAmountSum.Add(snapshot.ChangeAmount)
		aggregate.changePercentSum = aggregate.changePercentSum.Add(snapshot.ChangePercent)
		aggregate.volumeSum = aggregate.volumeSum.Add(snapshot.Volume)
		aggregate.turnoverSum = aggregate.turnoverSum.Add(snapshot.Turnover)
		switch {
		case snapshot.ChangePercent.GreaterThan(modelDecimalZero()):
			aggregate.riseCount++
		case snapshot.ChangePercent.LessThan(modelDecimalZero()):
			aggregate.fallCount++
		default:
			aggregate.flatCount++
		}
	}

	result := make([]model.MarketBoardSnapshot, 0, len(aggregates))
	for _, aggregate := range aggregates {
		if aggregate == nil || aggregate.stockCount == 0 {
			continue
		}
		denominator := aggregate.matchedCount
		if denominator == 0 {
			denominator = aggregate.stockCount
		}
		divisor := modelDecimalFromInt(denominator)
		result = append(result, model.MarketBoardSnapshot{
			BoardType:       aggregate.boardType,
			Code:            aggregate.boardCode,
			Name:            aggregate.boardName,
			LastPrice:       aggregate.lastPriceSum.Div(divisor),
			ChangeAmount:    aggregate.changeAmountSum.Div(divisor),
			ChangePercent:   aggregate.changePercentSum.Div(divisor),
			Volume:          aggregate.volumeSum,
			Turnover:        aggregate.turnoverSum,
			TotalMarketCap:  aggregate.totalMarketCap,
			FloatMarketCap:  aggregate.floatMarketCap,
			StockCount:      aggregate.stockCount,
			RiseCount:       aggregate.riseCount,
			FallCount:       aggregate.fallCount,
			FlatCount:       aggregate.flatCount,
			Source:          compactBoardSnapshotSource(aggregate.source, latestSource),
			BatchNo:         batchNo,
			SnapshotTime:    latestTime,
			ConstituentNode: aggregate.boardCode,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		left := result[i]
		right := result[j]
		if left.BoardType != right.BoardType {
			return left.BoardType < right.BoardType
		}
		if !left.ChangePercent.Equal(right.ChangePercent) {
			return left.ChangePercent.GreaterThan(right.ChangePercent)
		}
		if !left.Turnover.Equal(right.Turnover) {
			return left.Turnover.GreaterThan(right.Turnover)
		}
		return left.Name < right.Name
	})
	return result
}

func compactBoardSnapshotSource(constituentSource string, latestSource string) string {
	parts := make([]string, 0, 4)
	seen := make(map[string]struct{}, 4)
	appendTag := func(tag string) {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			return
		}
		if _, ok := seen[tag]; ok {
			return
		}
		seen[tag] = struct{}{}
		parts = append(parts, tag)
	}
	classify := func(raw string) string {
		raw = strings.ToLower(strings.TrimSpace(raw))
		switch {
		case raw == "":
			return ""
		case strings.Contains(raw, "taxonomy"):
			return "taxonomy"
		case strings.Contains(raw, "akshare"):
			return "akshare"
		case strings.Contains(raw, "sina"):
			return "sina"
		case strings.Contains(raw, "tencent"):
			return "tencent"
		case strings.Contains(raw, "eastmoney"):
			return "eastmoney"
		case strings.Contains(raw, "manual"):
			return "manual"
		case strings.Contains(raw, "board"):
			return "board"
		case strings.Contains(raw, "market"):
			return "market"
		default:
			return raw
		}
	}
	for _, item := range strings.Split(constituentSource, "+") {
		appendTag(classify(item))
	}
	for _, item := range strings.Split(latestSource, "+") {
		appendTag(classify(item))
	}
	if len(parts) == 0 {
		return "market"
	}
	joined := strings.Join(parts, "+")
	if len(joined) <= 32 {
		return joined
	}
	trimmed := make([]string, 0, len(parts))
	current := 0
	for _, part := range parts {
		partLen := len(part)
		if current > 0 {
			partLen++
		}
		if current+partLen > 32 {
			break
		}
		trimmed = append(trimmed, part)
		current += partLen
	}
	if len(trimmed) == 0 {
		return joined[:32]
	}
	return strings.Join(trimmed, "+")
}

func (s *marketDataService) symbols() []string {
	return normalizeSymbols(strings.Split(s.marketConfig.Symbols, ","))
}

func normalizeSymbols(symbols []string) []string {
	result := make([]string, 0, len(symbols))
	seen := make(map[string]struct{})
	for _, part := range symbols {
		symbol := normalizeSymbol(part)
		if symbol == "" {
			continue
		}
		if _, ok := seen[symbol]; ok {
			continue
		}
		seen[symbol] = struct{}{}
		result = append(result, symbol)
	}
	return result
}
