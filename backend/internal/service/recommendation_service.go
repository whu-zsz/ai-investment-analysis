package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	requestdto "stock-analysis-backend/internal/dto/request"
	responsedto "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/pkg/llm"
	"stock-analysis-backend/pkg/news"

	"github.com/shopspring/decimal"
)

const (
	recommendationLLMTimeout     = 8 * time.Second
	recommendationCandidateLimit = 18
	recommendationShortlistLimit = 10
	recommendationResultLimit    = 5
	recommendationInsightLimit   = 10
)

type RecommendationService interface {
	GetCandidates(userID uint64) (*responsedto.AnalysisCandidatesResponse, error)
	GetRecommendations(userID uint64) (*responsedto.AnalysisRecommendationsResponse, error)
	RecommendationChat(ctx context.Context, userID uint64, req *requestdto.RecommendationChatRequest) (*responsedto.RecommendationChatResponse, error)
	RecommendationChatStream(ctx context.Context, userID uint64, req *requestdto.RecommendationChatRequest, emit func(responsedto.StockChatStreamEvent) error) error
	GetRecommendationChatContext(userID, contextID uint64) (*responsedto.RecommendationChatContextSnapshotResponse, error)
}

type recommendationService struct {
	userRepo               repository.UserRepository
	transactionRepo        repository.TransactionRepository
	portfolioRepo          repository.PortfolioRepository
	marketSnapshotRepo     repository.MarketSnapshotRepository
	analysisReportRepo     repository.AnalysisReportRepository
	analysisReportItemRepo repository.AnalysisReportItemRepository
	marketDataService      MarketDataService
	marketStockService     MarketStockService
	marketSnapshotService  MarketSnapshotService
	newsService            NewsService
	llmProvider            llm.Provider
	contextService         *chatContextService
}

type candidateSource struct {
	typeName    string
	headline    string
	summary     string
	provider    string
	publishedAt time.Time
}

type recommendationCandidate struct {
	symbol        string
	assetName     string
	assetType     string
	market        string
	sources       []candidateSource
	sourceSet     map[string]struct{}
	isHeld        bool
	tradeCount    int
	latestPrice   decimal.Decimal
	changePercent decimal.Decimal
	dataStatus    string
}

type recommendationToolContext struct {
	user       *model.User
	question   string
	messages   []stockChatMessage
	candidates []recommendationCandidate
	scored     []scoredCandidate
	newsItems  []newsCandidateItem
	contextID  uint64
	reportID   uint64
	insights   map[string]recommendationStockInsight
}

type newsCandidateItem struct {
	Symbol        string `json:"symbol"`
	AssetName     string `json:"asset_name"`
	Headline      string `json:"headline"`
	Summary       string `json:"summary"`
	BoardName     string `json:"board_name"`
	ChangePercent string `json:"change_percent"`
	Provider      string `json:"provider"`
	PublishedAt   string `json:"published_at"`
}

type aiRecommendationEnvelope struct {
	SummaryText string                 `json:"summary_text"`
	Items       []aiRecommendationItem `json:"items"`
}

type aiRecommendationItem struct {
	Symbol      string `json:"symbol"`
	Action      string `json:"action"`
	MatchReason string `json:"match_reason"`
	RiskNote    string `json:"risk_note"`
}

type recommendationStockInsight struct {
	Symbol       string
	Detail       *responsedto.MarketStockDetailResponse
	Profile      *responsedto.StockProfileResponse
	Kline        *responsedto.MarketStockKlineResponse
	News         *StockNewsContext
	Indicators   map[string]string
	TrendSummary string
}

type recommendationReportEnvelope struct {
	ReportTitle     string                 `json:"report_title"`
	SummaryText     string                 `json:"summary_text"`
	RiskAnalysis    string                 `json:"risk_analysis"`
	Recommendations []string               `json:"recommendations"`
	Items           []aiRecommendationItem `json:"items"`
}

type recommendationChatContextMetadata struct {
	ToolTrace   []responsedto.ChatToolTraceStepResponse      `json:"tool_trace,omitempty"`
	ToolResults []responsedto.ChatToolResultSnapshotResponse `json:"tool_results,omitempty"`
	StepContext *responsedto.ChatStepContextResponse         `json:"step_context,omitempty"`
	NewsItems   []responsedto.StockChatNewsItemResponse      `json:"news_items,omitempty"`
	GeneratedAt string                                       `json:"generated_at,omitempty"`
	ReportTitle string                                       `json:"report_title,omitempty"`
	Candidates  []persistedRecommendationCandidate           `json:"candidates,omitempty"`
	Insights    map[string]persistedRecommendationInsight    `json:"insights,omitempty"`
}

type persistedRecommendationCandidateSource struct {
	TypeName    string `json:"type_name"`
	Headline    string `json:"headline,omitempty"`
	Summary     string `json:"summary,omitempty"`
	Provider    string `json:"provider,omitempty"`
	PublishedAt string `json:"published_at,omitempty"`
}

type persistedRecommendationCandidate struct {
	Symbol        string                                 `json:"symbol"`
	AssetName     string                                 `json:"asset_name"`
	AssetType     string                                 `json:"asset_type"`
	Market        string                                 `json:"market"`
	Sources       []persistedRecommendationCandidateSource `json:"sources,omitempty"`
	IsHeld        bool                                   `json:"is_held"`
	TradeCount    int                                    `json:"trade_count"`
	LatestPrice   string                                 `json:"latest_price,omitempty"`
	ChangePercent string                                 `json:"change_percent,omitempty"`
	DataStatus    string                                 `json:"data_status,omitempty"`
}

type persistedRecommendationInsight struct {
	Symbol       string                                   `json:"symbol"`
	Name         string                                   `json:"name,omitempty"`
	Market       string                                   `json:"market,omitempty"`
	Industry     string                                   `json:"industry,omitempty"`
	Region       string                                   `json:"region,omitempty"`
	Concepts     []string                                 `json:"concepts,omitempty"`
	Boards       []responsedto.StockBoardMembershipResponse `json:"boards,omitempty"`
	TrendSummary string                                   `json:"trend_summary,omitempty"`
	Indicators   map[string]string                        `json:"indicators,omitempty"`
	NewsStatus   string                                   `json:"news_status,omitempty"`
	NewsSummary  string                                   `json:"news_summary,omitempty"`
	NewsCoverage string                                   `json:"news_coverage,omitempty"`
}

func NewRecommendationService(
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	portfolioRepo repository.PortfolioRepository,
	marketSnapshotRepo repository.MarketSnapshotRepository,
	analysisReportRepo repository.AnalysisReportRepository,
	analysisReportItemRepo repository.AnalysisReportItemRepository,
	marketDataService MarketDataService,
	marketStockService MarketStockService,
	marketSnapshotService MarketSnapshotService,
	newsService NewsService,
	llmProvider llm.Provider,
	contextRepo repository.ChatContextRepository,
) RecommendationService {
	return &recommendationService{
		userRepo:               userRepo,
		transactionRepo:        transactionRepo,
		portfolioRepo:          portfolioRepo,
		marketSnapshotRepo:     marketSnapshotRepo,
		analysisReportRepo:     analysisReportRepo,
		analysisReportItemRepo: analysisReportItemRepo,
		marketDataService:      marketDataService,
		marketStockService:     marketStockService,
		marketSnapshotService:  marketSnapshotService,
		newsService:            newsService,
		llmProvider:            llmProvider,
		contextService:         newChatContextService(contextRepo),
	}
}

func (s *recommendationService) GetCandidates(userID uint64) (*responsedto.AnalysisCandidatesResponse, error) {
	candidates, err := s.buildCandidates(context.Background(), userID)
	if err != nil {
		return nil, err
	}

	result := &responsedto.AnalysisCandidatesResponse{
		Candidates: make([]responsedto.AnalysisCandidateResponse, 0, len(candidates)),
	}
	if len(candidates) > 0 {
		result.DefaultSymbol = candidates[0].symbol
	}

	for _, candidate := range candidates {
		sources := make([]responsedto.AnalysisCandidateSource, 0, len(candidate.sources))
		for _, source := range candidate.sources {
			sources = append(sources, responsedto.AnalysisCandidateSource{Type: source.typeName})
		}
		result.Candidates = append(result.Candidates, responsedto.AnalysisCandidateResponse{
			Symbol:        candidate.symbol,
			AssetName:     fallbackString(candidate.assetName, candidate.symbol),
			AssetType:     candidate.assetType,
			Market:        candidate.market,
			Sources:       sources,
			IsHeld:        candidate.isHeld,
			TradeCount:    candidate.tradeCount,
			LastPrice:     decimalToString(candidate.latestPrice, 4),
			ChangePercent: decimalToString(candidate.changePercent, 2),
		})
	}

	return result, nil
}

func (s *recommendationService) GetRecommendations(userID uint64) (*responsedto.AnalysisRecommendationsResponse, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	candidates, err := s.buildCandidates(context.Background(), userID)
	if err != nil {
		return nil, err
	}

	response := &responsedto.AnalysisRecommendationsResponse{
		GeneratedAt: time.Now().Format("2006-01-02 15:04:05"),
		ProfileSummary: responsedto.RecommendationProfileSummary{
			InvestmentPreference: user.InvestmentPreference,
			RiskTolerance:        user.RiskTolerance,
			TotalProfit:          user.TotalProfit.StringFixed(2),
			HeldPositions:        countHeldCandidates(candidates),
			CandidateCount:       len(candidates),
		},
		Candidates: []responsedto.RecommendationItemResponse{},
	}

	if len(candidates) == 0 {
		response.SummaryText = "当前还没有可用于推荐的用户相关股票池，请先上传交易记录或形成持仓后再查看推荐。"
		return response, nil
	}

	preferenceText := recommendationPreferenceText("", nil)
	scored := finalRecommendationCandidates(shortlistScoredCandidates(scoreCandidates(candidates, user, preferenceText)))
	aiOutput, err := s.generateRecommendationExplanation(user, scored)
	if err != nil {
		response.SummaryText = buildFallbackRecommendationSummary(user, scored)
		response.Candidates = convertScoredCandidates(scored, nil)
		return response, nil
	}

	response.SummaryText = fallbackString(aiOutput.SummaryText, buildFallbackRecommendationSummary(user, scored))
	response.Candidates = convertScoredCandidates(scored, aiOutput.Items)
	return response, nil
}

func (s *recommendationService) RecommendationChat(ctx context.Context, userID uint64, req *requestdto.RecommendationChatRequest) (*responsedto.RecommendationChatResponse, error) {
	req = s.hydrateRecommendationChatRequest(userID, req)
	executor, err := newRecommendationChatExecutor(s, userID, req)
	if err != nil {
		return nil, err
	}
	orchestrator := newChatOrchestrator(s.llmProvider)
	reply, results, trace, err := orchestrator.Run(ctx, executor, nil)
	if err != nil {
		return nil, err
	}
	result, err := executor.BuildDoneResponse(reply, trace, results)
	if err != nil {
		return nil, err
	}
	response, ok := result.(*responsedto.RecommendationChatResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected recommendation response type")
	}
	if err := s.persistRecommendationChat(userID, response, req, executor.toToolContext()); err != nil {
		return nil, err
	}
	return response, nil
}

func (s *recommendationService) RecommendationChatStream(ctx context.Context, userID uint64, req *requestdto.RecommendationChatRequest, emit func(responsedto.StockChatStreamEvent) error) error {
	if emit == nil {
		return fmt.Errorf("stream emitter is required")
	}
	req = s.hydrateRecommendationChatRequest(userID, req)
	executor, err := newRecommendationChatExecutor(s, userID, req)
	if err != nil {
		_ = emit(responsedto.StockChatStreamEvent{Type: "error", Message: err.Error()})
		return err
	}
	orchestrator := newChatOrchestrator(s.llmProvider)
	reply, results, trace, err := orchestrator.Run(ctx, executor, emit)
	if err != nil {
		_ = emit(responsedto.StockChatStreamEvent{Type: "error", Stage: "report", Message: err.Error()})
		return err
	}
	result, err := executor.BuildDoneResponse(reply, trace, results)
	if err != nil {
		_ = emit(responsedto.StockChatStreamEvent{Type: "error", Stage: "done", Message: err.Error()})
		return err
	}
	message := "推荐分析已完成，尚未生成报告"
	if response, ok := result.(*responsedto.RecommendationChatResponse); ok {
		if persistErr := s.persistRecommendationChat(userID, response, req, executor.toToolContext()); persistErr != nil {
			_ = emit(responsedto.StockChatStreamEvent{Type: "error", Stage: "done", Message: persistErr.Error()})
			return persistErr
		}
		if response.ReportID > 0 {
			message = "推荐报告已生成"
		}
	}
	return emit(responsedto.StockChatStreamEvent{Type: "done", Stage: "done", Message: message, Data: result})
}

type scoredCandidate struct {
	candidate   recommendationCandidate
	score       decimal.Decimal
	action      string
	matchReason string
	riskNote    string
}

func (s *recommendationService) buildCandidates(ctx context.Context, userID uint64) ([]recommendationCandidate, error) {
	portfolios, err := s.portfolioRepo.FindByUserID(userID)
	if err != nil {
		return nil, err
	}
	transactions, _, err := s.transactionRepo.FindByUserID(userID, 200, 0)
	if err != nil {
		return nil, err
	}

	candidateMap := make(map[string]*recommendationCandidate)
	addCandidate := func(symbol, assetName, assetType string, isHeld bool, sourceType string) {
		normalized := normalizeSymbol(symbol)
		if normalized == "" {
			return
		}
		candidate, ok := candidateMap[normalized]
		if !ok {
			candidate = &recommendationCandidate{
				symbol:     normalized,
				assetName:  strings.TrimSpace(assetName),
				assetType:  strings.TrimSpace(assetType),
				sourceSet:  make(map[string]struct{}),
				dataStatus: marketDataStatusUnavailable,
			}
			candidateMap[normalized] = candidate
		}
		if strings.TrimSpace(candidate.assetName) == "" && strings.TrimSpace(assetName) != "" {
			candidate.assetName = strings.TrimSpace(assetName)
		}
		if strings.TrimSpace(candidate.assetType) == "" && strings.TrimSpace(assetType) != "" {
			candidate.assetType = strings.TrimSpace(assetType)
		}
		if isHeld {
			candidate.isHeld = true
		}
		if sourceType != "" {
			if _, exists := candidate.sourceSet[sourceType]; !exists {
				candidate.sourceSet[sourceType] = struct{}{}
				candidate.sources = append(candidate.sources, candidateSource{typeName: sourceType})
			}
		}
	}

	for _, portfolio := range portfolios {
		addCandidate(portfolio.AssetCode, portfolio.AssetName, portfolio.AssetType, true, "portfolio")
	}
	for _, tx := range transactions {
		addCandidate(tx.AssetCode, tx.AssetName, tx.AssetType, false, "transactions")
		candidateMap[normalizeSymbol(tx.AssetCode)].tradeCount++
	}

	if len(candidateMap) == 0 {
		return []recommendationCandidate{}, nil
	}

	result := make([]recommendationCandidate, 0, len(candidateMap))
	for _, candidate := range candidateMap {
		result = append(result, *candidate)
	}
	result = s.hydrateRecommendationCandidates(ctx, result, true)
	for i := range result {
		if result[i].assetName == "" {
			result[i].assetName = result[i].symbol
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].isHeld != result[j].isHeld {
			return result[i].isHeld
		}
		if result[i].tradeCount != result[j].tradeCount {
			return result[i].tradeCount > result[j].tradeCount
		}
		return result[i].symbol < result[j].symbol
	})
	return result, nil
}

func (s *recommendationService) hydrateRecommendationCandidates(ctx context.Context, candidates []recommendationCandidate, fetchMissing bool) []recommendationCandidate {
	if len(candidates) == 0 || s.marketSnapshotRepo == nil {
		return candidates
	}
	missingSymbols := make([]string, 0)
	for i := range candidates {
		symbol := normalizeSymbol(candidates[i].symbol)
		if symbol == "" {
			continue
		}
		snapshot, snapshotErr := s.marketSnapshotRepo.FindLatestBySymbol(symbol)
		if snapshotErr == nil && snapshot != nil {
			candidates[i].market = snapshot.Market
			candidates[i].latestPrice = snapshot.LastPrice
			candidates[i].changePercent = snapshot.ChangePercent
			candidates[i].dataStatus = marketDataStatusComplete
			if strings.TrimSpace(candidates[i].assetName) == "" {
				candidates[i].assetName = snapshot.Name
			}
			continue
		}
		missingSymbols = append(missingSymbols, symbol)
	}
	if !fetchMissing || len(missingSymbols) == 0 || s.marketDataService == nil {
		return candidates
	}
	if _, fetchErr := s.marketDataService.FetchAndStoreQuotesBySymbols(ctx, missingSymbols); fetchErr != nil {
		return candidates
	}
	for i := range candidates {
		symbol := normalizeSymbol(candidates[i].symbol)
		if symbol == "" || (candidates[i].dataStatus == marketDataStatusComplete || candidates[i].dataStatus == marketDataStatusFetchedLive) {
			continue
		}
		snapshot, snapshotErr := s.marketSnapshotRepo.FindLatestBySymbol(symbol)
		if snapshotErr != nil || snapshot == nil {
			continue
		}
		candidates[i].market = snapshot.Market
		candidates[i].latestPrice = snapshot.LastPrice
		candidates[i].changePercent = snapshot.ChangePercent
		candidates[i].dataStatus = marketDataStatusFetchedLive
		if strings.TrimSpace(candidates[i].assetName) == "" {
			candidates[i].assetName = snapshot.Name
		}
	}
	return candidates
}

func scoreCandidates(candidates []recommendationCandidate, user *model.User, preferenceText string) []scoredCandidate {
	result := make([]scoredCandidate, 0, len(candidates))
	preference := normalizeInvestmentStyle(user.InvestmentPreference)
	riskTolerance := strings.ToLower(strings.TrimSpace(user.RiskTolerance))
	preferenceThemes := extractRecommendationThemes(preferenceText)
	for _, candidate := range candidates {
		score := decimal.NewFromInt(50)
		if candidate.isHeld {
			score = score.Add(decimal.NewFromInt(4))
		}
		if candidate.tradeCount >= 3 {
			score = score.Add(decimal.NewFromInt(3))
		} else if candidate.tradeCount > 0 {
			score = score.Add(decimal.NewFromInt(2))
		}
		if candidate.dataStatus == marketDataStatusComplete || candidate.dataStatus == marketDataStatusFetchedLive {
			score = score.Add(decimal.NewFromInt(8))
		}
		if hasCandidateSource(candidate, "news_discovery") {
			if candidate.changePercent.GreaterThan(decimal.Zero) && candidate.changePercent.LessThanOrEqual(decimal.NewFromInt(6)) {
				score = score.Add(decimal.NewFromInt(12))
			} else if candidate.changePercent.GreaterThan(decimal.NewFromInt(8)) {
				score = score.Sub(decimal.NewFromInt(10))
			} else {
				score = score.Add(decimal.NewFromInt(6))
			}
		}
		if hasCandidateSource(candidate, "board_theme") {
			score = score.Add(decimal.NewFromInt(8))
		}
		if candidateMatchesThemes(candidate, preferenceThemes) {
			score = score.Add(decimal.NewFromInt(16))
		}

		change := candidate.changePercent
		switch preference {
		case "conservative":
			if change.GreaterThan(decimal.NewFromInt(0)) && change.LessThanOrEqual(decimal.NewFromInt(5)) {
				score = score.Add(decimal.NewFromInt(10))
			}
			if change.GreaterThan(decimal.NewFromInt(8)) || change.LessThan(decimal.NewFromInt(-5)) {
				score = score.Sub(decimal.NewFromInt(8))
			}
		case "aggressive":
			if change.GreaterThan(decimal.NewFromInt(2)) {
				score = score.Add(decimal.NewFromInt(10))
			}
			if change.LessThan(decimal.NewFromInt(-6)) {
				score = score.Sub(decimal.NewFromInt(4))
			}
		default:
			if change.GreaterThan(decimal.NewFromFloat(0.5)) && change.LessThanOrEqual(decimal.NewFromInt(6)) {
				score = score.Add(decimal.NewFromInt(8))
			}
		}

		action := "observe"
		if candidate.isHeld {
			action = "hold"
		}
		if score.GreaterThanOrEqual(decimal.NewFromInt(75)) {
			action = "buy"
		} else if candidate.isHeld && change.LessThan(decimal.NewFromInt(-6)) {
			action = "reduce"
		} else if !candidate.isHeld && score.GreaterThanOrEqual(decimal.NewFromInt(62)) {
			action = "observe"
		}

		matchReason := buildMatchReason(candidate, preference)
		if candidateMatchesThemes(candidate, preferenceThemes) {
			matchReason = fmt.Sprintf("命中你明确提到的偏好主题（%s）；%s", strings.Join(preferenceThemes, "、"), matchReason)
		}
		riskNote := buildRiskNote(candidate, riskTolerance)
		result = append(result, scoredCandidate{
			candidate:   candidate,
			score:       score,
			action:      action,
			matchReason: matchReason,
			riskNote:    riskNote,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		if !result[i].score.Equal(result[j].score) {
			return result[i].score.GreaterThan(result[j].score)
		}
		if result[i].candidate.isHeld != result[j].candidate.isHeld {
			return result[i].candidate.isHeld
		}
		return result[i].candidate.symbol < result[j].candidate.symbol
	})
	if len(result) > recommendationCandidateLimit {
		result = result[:recommendationCandidateLimit]
	}
	return result
}

func shortlistScoredCandidates(items []scoredCandidate) []scoredCandidate {
	if len(items) <= recommendationShortlistLimit {
		return items
	}
	return append([]scoredCandidate(nil), items[:recommendationShortlistLimit]...)
}

func finalRecommendationCandidates(items []scoredCandidate) []scoredCandidate {
	if len(items) <= recommendationResultLimit {
		return items
	}
	return append([]scoredCandidate(nil), items[:recommendationResultLimit]...)
}

func recommendationPreferenceText(question string, messages []stockChatMessage) string {
	parts := make([]string, 0, len(messages)+1)
	if trimmed := strings.TrimSpace(question); trimmed != "" {
		parts = append(parts, trimmed)
	}
	for _, message := range messages {
		if trimmed := strings.TrimSpace(message.Content); trimmed != "" {
			parts = append(parts, trimmed)
		}
	}
	return strings.Join(parts, "\n")
}

func extractRecommendationThemes(text string) []string {
	normalized := strings.ToLower(strings.TrimSpace(text))
	if normalized == "" {
		return nil
	}
	themes := []string{
		"白酒", "酿酒", "消费", "半导体", "芯片", "银行", "创新药", "新能源", "军工", "医药", "算力", "人工智能", "券商",
	}
	matched := make([]string, 0, 4)
	for _, theme := range themes {
		if strings.Contains(normalized, strings.ToLower(theme)) {
			matched = append(matched, theme)
		}
	}
	return uniqueStrings(matched)
}

func candidateMatchesThemes(candidate recommendationCandidate, themes []string) bool {
	if len(themes) == 0 {
		return false
	}
	contentParts := []string{candidate.symbol, candidate.assetName}
	for _, source := range candidate.sources {
		contentParts = append(contentParts, source.typeName, source.headline, source.summary)
	}
	content := strings.ToLower(strings.Join(contentParts, "\n"))
	for _, theme := range themes {
		if strings.Contains(content, strings.ToLower(strings.TrimSpace(theme))) {
			return true
		}
	}
	return false
}

func buildMatchReason(candidate recommendationCandidate, preference string) string {
	parts := make([]string, 0, 4)
	if candidate.isHeld {
		parts = append(parts, "当前已持仓，便于结合已有成本和仓位继续决策")
	}
	if candidate.tradeCount > 0 {
		parts = append(parts, fmt.Sprintf("最近交易记录中出现 %d 次，属于你的熟悉标的", candidate.tradeCount))
	}
	if hasCandidateSource(candidate, "news_discovery") {
		source := recommendationPrimarySource(candidate, "news_discovery")
		if strings.TrimSpace(source.headline) != "" {
			parts = append(parts, fmt.Sprintf("近期出现可追溯新闻催化：%s", strings.TrimSpace(source.headline)))
		} else {
			parts = append(parts, "近期有可信新闻催化，具备事件驱动观察价值")
		}
	}
	if candidate.dataStatus == marketDataStatusComplete || candidate.dataStatus == marketDataStatusFetchedLive {
		parts = append(parts, fmt.Sprintf("最新涨跌幅 %s，适合%s风格下继续观察", decimalToString(candidate.changePercent, 2), preferenceLabel(preference)))
		return strings.Join(parts, "；")
	}
	parts = append(parts, "当前市场快照不足，建议结合更多数据再决定")
	return strings.Join(parts, "；")
}

func buildRiskNote(candidate recommendationCandidate, riskTolerance string) string {
	change := candidate.changePercent
	if candidate.dataStatus == marketDataStatusUnavailable {
		return "当前缺少可靠市场快照，推荐结论仅能作为候选观察参考。"
	}
	if hasCandidateSource(candidate, "news_discovery") && change.GreaterThan(decimal.NewFromFloat(7.5)) {
		return "虽然存在新闻催化，但短线涨幅已经偏大，更适合等待回踩或继续确认催化持续性。"
	}
	if riskTolerance == "low" || riskTolerance == "conservative" {
		if change.GreaterThan(decimal.NewFromInt(6)) || change.LessThan(decimal.NewFromInt(-4)) {
			return "你的风险承受偏低，这个标的近期波动偏大，建议控制仓位。"
		}
		return "波动尚在可控区间，更适合低频跟踪。"
	}
	if change.LessThan(decimal.NewFromInt(-8)) {
		return "近期回撤较明显，若继续关注应先确认回撤原因。"
	}
	return "建议结合仓位集中度、催化兑现节奏和后续趋势继续判断。"
}

func (s *recommendationService) buildRecommendationToolContext(ctx context.Context, userID uint64, req *requestdto.RecommendationChatRequest, emit func(responsedto.StockChatStreamEvent) error) (*recommendationToolContext, error) {
	if req == nil || strings.TrimSpace(req.Question) == "" {
		return nil, fmt.Errorf("question is required")
	}
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if emit != nil {
		if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: "market", Message: "正在读取用户画像和已有持仓/关注标的"}); err != nil {
			return nil, err
		}
	}
	candidates, err := s.buildCandidates(ctx, userID)
	if err != nil {
		return nil, err
	}
	if emit != nil {
		if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: "news", Message: "正在从近期新闻中发现潜力股"}); err != nil {
			return nil, err
		}
	}
	discovered, err := s.discoverNewsCandidates(ctx)
	if err == nil {
		candidates = mergeRecommendationCandidates(candidates, discovered)
	}
	scored := shortlistScoredCandidates(scoreCandidates(candidates, user, recommendationPreferenceText(strings.TrimSpace(req.Question), normalizeStockChatMessages(req.Messages))))
	if emit != nil {
		if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: "prompt", Message: "正在整合偏好、候选池和新闻依据"}); err != nil {
			return nil, err
		}
	}
	return &recommendationToolContext{
		user:       user,
		question:   strings.TrimSpace(req.Question),
		messages:   normalizeStockChatMessages(req.Messages),
		candidates: candidates,
		scored:     scored,
		newsItems:  buildNewsCandidateItems(scored),
		reportID:   req.ReportID,
	}, nil
}

func (s *recommendationService) discoverNewsCandidates(ctx context.Context) ([]recommendationCandidate, error) {
	if s.marketSnapshotService == nil {
		return nil, fmt.Errorf("market snapshot service is unavailable")
	}
	if s.newsService == nil {
		return nil, fmt.Errorf("news service is unavailable")
	}
	breadth, err := s.marketSnapshotService.GetDashboardMarketBreadth(ctx, 18)
	if err != nil {
		return nil, err
	}

	pool := make([]responsedto.MarketBreadthItemResponse, 0, len(breadth.TopTurnover)+len(breadth.TopGainers))
	pool = append(pool, breadth.TopTurnover...)
	for _, item := range breadth.TopGainers {
		change := parseDecimalOrZero(item.ChangePercent)
		if change.GreaterThan(decimal.NewFromInt(8)) || change.LessThanOrEqual(decimal.Zero) {
			continue
		}
		pool = append(pool, item)
	}

	seen := make(map[string]struct{}, len(pool))
	result := make([]recommendationCandidate, 0, 6)
	for _, item := range pool {
		symbol := normalizeSymbol(item.Symbol)
		if symbol == "" {
			continue
		}
		if _, ok := seen[symbol]; ok {
			continue
		}
		seen[symbol] = struct{}{}

		change := parseDecimalOrZero(item.ChangePercent)
		if change.GreaterThan(decimal.NewFromFloat(9.2)) || change.LessThan(decimal.NewFromFloat(-3.5)) {
			continue
		}
		turnoverRate := parseDecimalOrZero(item.TurnoverRate)
		if !turnoverRate.IsZero() && turnoverRate.GreaterThan(decimal.NewFromInt(18)) {
			continue
		}

		newsContext, newsErr := s.newsService.GetStockNews(ctx, symbol, strings.TrimSpace(item.Name))
		if newsErr != nil || newsContext == nil || len(newsContext.Items) == 0 {
			continue
		}
		newsItem := recommendationSelectCatalystNews(newsContext)
		if newsItem == nil {
			continue
		}
		if !recommendationHasCatalyst(newsItem) {
			continue
		}

		result = append(result, recommendationCandidate{
			symbol:    symbol,
			assetName: strings.TrimSpace(item.Name),
			assetType: "stock",
			market:    item.Market,
			sources: []candidateSource{{
				typeName:    "news_discovery",
				headline:    strings.TrimSpace(newsItem.Title),
				summary:     strings.TrimSpace(newsItem.Summary),
				provider:    fallbackString(strings.TrimSpace(newsItem.Source), strings.TrimSpace(newsItem.Provider)),
				publishedAt: newsItem.PublishedAt,
			}},
			sourceSet:     map[string]struct{}{"news_discovery": {}},
			isHeld:        false,
			tradeCount:    0,
			latestPrice:   parseDecimalOrZero(item.LastPrice),
			changePercent: change,
			dataStatus:    marketDataStatusComplete,
		})
		if len(result) >= 6 {
			break
		}
	}
	return result, nil
}

func mergeRecommendationCandidates(base []recommendationCandidate, discovered []recommendationCandidate) []recommendationCandidate {
	index := make(map[string]*recommendationCandidate, len(base))
	result := make([]recommendationCandidate, 0, len(base)+len(discovered))
	for i := range base {
		candidate := base[i]
		result = append(result, candidate)
		index[candidate.symbol] = &result[len(result)-1]
	}
	for _, candidate := range discovered {
		if existing, ok := index[candidate.symbol]; ok {
			if strings.TrimSpace(existing.assetName) == "" && strings.TrimSpace(candidate.assetName) != "" {
				existing.assetName = candidate.assetName
			}
			if strings.TrimSpace(existing.assetType) == "" && strings.TrimSpace(candidate.assetType) != "" {
				existing.assetType = candidate.assetType
			}
			if strings.TrimSpace(existing.market) == "" && strings.TrimSpace(candidate.market) != "" {
				existing.market = candidate.market
			}
			if existing.latestPrice.IsZero() && !candidate.latestPrice.IsZero() {
				existing.latestPrice = candidate.latestPrice
			}
			if existing.changePercent.IsZero() && !candidate.changePercent.IsZero() {
				existing.changePercent = candidate.changePercent
			}
			if existing.dataStatus == marketDataStatusUnavailable && candidate.dataStatus != "" {
				existing.dataStatus = candidate.dataStatus
			}
			if candidate.isHeld {
				existing.isHeld = true
			}
			if candidate.tradeCount > existing.tradeCount {
				existing.tradeCount = candidate.tradeCount
			}
			if existing.sourceSet == nil {
				existing.sourceSet = make(map[string]struct{})
			}
			for _, source := range candidate.sources {
				if source.typeName == "" {
					continue
				}
				if _, exists := existing.sourceSet[source.typeName]; exists {
					continue
				}
				existing.sourceSet[source.typeName] = struct{}{}
				existing.sources = append(existing.sources, source)
			}
			continue
		}
		result = append(result, candidate)
		index[candidate.symbol] = &result[len(result)-1]
	}
	return result
}

func buildNewsCandidateItems(scored []scoredCandidate) []newsCandidateItem {
	items := make([]newsCandidateItem, 0, len(scored))
	for _, candidate := range scored {
		if !hasCandidateSource(candidate.candidate, "news_discovery") {
			continue
		}
		source := recommendationPrimarySource(candidate.candidate, "news_discovery")
		publishedAt := ""
		if !source.publishedAt.IsZero() {
			publishedAt = source.publishedAt.Format("2006-01-02 15:04:05")
		}
		headline := strings.TrimSpace(source.headline)
		if headline == "" {
			headline = fmt.Sprintf("%s 进入近期新闻催化候选", fallbackString(candidate.candidate.assetName, candidate.candidate.symbol))
		}
		summary := strings.TrimSpace(source.summary)
		if summary == "" {
			summary = candidate.matchReason
		}
		items = append(items, newsCandidateItem{
			Symbol:        candidate.candidate.symbol,
			AssetName:     fallbackString(candidate.candidate.assetName, candidate.candidate.symbol),
			Headline:      headline,
			Summary:       summary,
			ChangePercent: decimalToString(candidate.candidate.changePercent, 2),
			Provider:      strings.TrimSpace(source.provider),
			PublishedAt:   publishedAt,
		})
	}
	return items
}

func hasCandidateSource(candidate recommendationCandidate, sourceType string) bool {
	for _, source := range candidate.sources {
		if source.typeName == sourceType {
			return true
		}
	}
	return false
}

func recommendationSourceTags(candidate recommendationCandidate) []string {
	tags := make([]string, 0, len(candidate.sources))
	for _, source := range candidate.sources {
		tags = append(tags, source.typeName)
	}
	sort.Strings(tags)
	return tags
}

func recommendationPrimarySource(candidate recommendationCandidate, sourceType string) candidateSource {
	for _, source := range candidate.sources {
		if source.typeName == sourceType {
			return source
		}
	}
	return candidateSource{}
}

func recommendationSelectCatalystNews(newsContext *StockNewsContext) *news.Item {
	if newsContext == nil || len(newsContext.Items) == 0 {
		return nil
	}
	for i := range newsContext.Items {
		item := &newsContext.Items[i]
		if recommendationHasCatalyst(item) {
			return item
		}
	}
	return &newsContext.Items[0]
}

func recommendationHasCatalyst(item *news.Item) bool {
	if item == nil {
		return false
	}
	text := strings.ToLower(strings.TrimSpace(item.Title + " " + item.Summary))
	return containsAny(text,
		"业绩", "预增", "合同", "订单", "中标", "回购", "增持", "分红", "新品", "发布", "量产", "涨价", "突破",
		"景气", "扩产", "收购", "合作", "政策", "核准", "获批", "并购", "算力", "芯片", "ai", "出口", "关税",
		"earnings", "contract", "order", "approval", "launch", "acquisition", "policy", "growth", "guidance",
	)
}

func (s *recommendationService) generateRecommendationExplanation(user *model.User, candidates []scoredCandidate) (*aiRecommendationEnvelope, error) {
	if s.llmProvider == nil {
		return nil, fmt.Errorf("llm provider is nil")
	}
	lines := make([]string, 0, len(candidates))
	for _, item := range candidates {
		candidate := item.candidate
		lines = append(lines, fmt.Sprintf("- %s %s: 当前动作=%s, 评分=%s, 是否持仓=%t, 交易次数=%d, 最新价=%s, 涨跌幅=%s, 数据状态=%s", candidate.symbol, fallbackString(candidate.assetName, candidate.symbol), item.action, item.score.StringFixed(2), candidate.isHeld, candidate.tradeCount, decimalToString(candidate.latestPrice, 4), decimalToString(candidate.changePercent, 2), candidate.dataStatus))
	}
	userPrompt := fmt.Sprintf(`请基于以下用户状态和候选标的，输出一个合法 JSON 对象，帮助解释推荐理由。
用户投资偏好：%s
用户风险承受：%s
用户累计盈亏：%s
候选标的：
%s

输出 JSON 结构：
{
  "summary_text": "string",
  "items": [
    {
      "symbol": "string",
      "action": "buy|hold|reduce|sell|observe",
      "match_reason": "string",
      "risk_note": "string"
    }
  ]
}

要求：
1. 只能输出 JSON，不要 markdown。
2. 只能基于给定候选数据解释，不要编造外部消息。
3. summary_text 要直接概括用户当前更适合关注哪些类型标的。
4. items 里最多返回 %d 个候选。
`, user.InvestmentPreference, user.RiskTolerance, user.TotalProfit.StringFixed(2), strings.Join(lines, "\n"), len(candidates))
	ctx, cancel := context.WithTimeout(context.Background(), recommendationLLMTimeout)
	defer cancel()
	content, err := s.llmProvider.GetContent(ctx, "你是一位严谨的股票推荐解释助手，只能输出合法 JSON。", userPrompt)
	if err != nil {
		return nil, err
	}
	var parsed aiRecommendationEnvelope
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &parsed); err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (s *recommendationService) recommendationReportPrompt(toolCtx *recommendationToolContext) string {
	return "你是一位专业、克制的 A 股推荐分析助手。你只能根据用户画像、用户已有持仓/交易关注、新闻发现候选、行情变化和系统给出的候选池作出推荐。必须使用简体中文与 Markdown 输出，不得编造新闻、财务或板块信息。先直接回答用户问题，再给出“## 重点提示”“## 候选对比”“## 风险”“## 操作建议”这些小节。"
}

func (s *recommendationService) recommendationReportUserPrompt(toolCtx *recommendationToolContext) string {
	lines := make([]string, 0, len(toolCtx.scored))
	for _, item := range toolCtx.scored {
		sourceTags := strings.Join(recommendationSourceTags(item.candidate), ",")
		lines = append(lines, fmt.Sprintf("- %s %s：来源=%s，动作=%s，评分=%s，是否持仓=%t，交易次数=%d，最新价=%s，涨跌幅=%s，推荐理由=%s，风险=%s",
			item.candidate.symbol,
			fallbackString(item.candidate.assetName, item.candidate.symbol),
			sourceTags,
			item.action,
			item.score.StringFixed(2),
			item.candidate.isHeld,
			item.candidate.tradeCount,
			decimalToString(item.candidate.latestPrice, 4),
			decimalToString(item.candidate.changePercent, 2),
			item.matchReason,
			item.riskNote,
		))
	}
	history := make([]string, 0, len(toolCtx.messages))
	for _, item := range toolCtx.messages {
		history = append(history, fmt.Sprintf("%s: %s", item.Role, item.Content))
	}
	return fmt.Sprintf(`用户投资偏好：%s
用户风险承受：%s
用户累计盈亏：%s
用户当前候选池总数：%d
其中新闻潜力股数量：%d

候选标的：
%s

历史对话：
%s

用户本轮问题：
%s

输出要求：
1. 直接回答用户问题。
2. 推荐对象不必局限于当前持仓，允许从新闻潜力股中选出更值得关注的标的。
3. 必须明确区分“已有关注标的”和“新发现潜力股”。
4. 不要输出 JSON。`,
		toolCtx.user.InvestmentPreference,
		toolCtx.user.RiskTolerance,
		toolCtx.user.TotalProfit.StringFixed(2),
		len(toolCtx.candidates),
		len(toolCtx.newsItems),
		strings.Join(lines, "\n"),
		strings.Join(history, "\n"),
		toolCtx.question,
	)
}

func (s *recommendationService) generateRecommendationReport(ctx context.Context, toolCtx *recommendationToolContext) (string, *responsedto.AnalysisRecommendationsResponse, error) {
	reply, err := s.llmProvider.GetContent(ctx, s.recommendationReportPrompt(toolCtx), s.recommendationReportUserPrompt(toolCtx))
	if err != nil {
		return "", nil, err
	}
	reply = normalizeMarkdownNarrative(reply)
	report, err := s.persistRecommendationReport(toolCtx, reply)
	if err != nil {
		return "", nil, err
	}
	return reply, report, nil
}

func (s *recommendationService) persistRecommendationReport(toolCtx *recommendationToolContext, reply string) (*responsedto.AnalysisRecommendationsResponse, error) {
	if s.analysisReportRepo == nil || s.analysisReportItemRepo == nil {
		return nil, fmt.Errorf("analysis report repository is unavailable")
	}
	candidates := convertScoredCandidates(toolCtx.scored, nil)
	recommendations := make([]string, 0, len(candidates))
	items := make([]model.AnalysisReportItem, 0, len(candidates))
	for _, item := range toolCtx.scored {
		recommendations = append(recommendations, fmt.Sprintf("%s：%s", fallbackString(item.candidate.assetName, item.candidate.symbol), item.action))
		rawSourceTags, _ := json.Marshal(recommendationSourceTags(item.candidate))
		items = append(items, model.AnalysisReportItem{
			UserID:               toolCtx.user.ID,
			Symbol:               item.candidate.symbol,
			AssetName:            fallbackString(item.candidate.assetName, item.candidate.symbol),
			Market:               item.candidate.market,
			TradeCount:           item.candidate.tradeCount,
			BuyCount:             0,
			SellCount:            0,
			BuyAmount:            decimal.Zero,
			SellAmount:           decimal.Zero,
			NetQuantity:          decimal.Zero,
			RealizedProfit:       decimal.Zero,
			RealizedProfitRate:   decimal.Zero,
			EndingPositionQty:    decimal.Zero,
			EndingAvgCost:        decimal.Zero,
			LatestPrice:          item.candidate.latestPrice,
			LatestMarketValue:    decimal.Zero,
			UnrealizedProfit:     decimal.Zero,
			TotalProfit:          decimal.Zero,
			ChangePercent7D:      decimal.Zero,
			PeriodPriceChangePct: item.candidate.changePercent,
			MarketDataStatus:     item.candidate.dataStatus,
			RiskLevel:            normalizeRiskLevel(toolCtx.user.RiskTolerance),
			InvestmentStyle:      stringPointerIfNotEmpty(toolCtx.user.InvestmentPreference),
			AnalysisText:         item.matchReason,
			Recommendation:       item.action,
			KeyPoints:            stringPointerIfNotEmpty(string(rawSourceTags)),
			RawAIOutput:          stringPointerIfNotEmpty(item.riskNote),
		})
	}
	now := time.Now()
	recommendationsJSON := marshalJSONArray(recommendations)
	raw := reply
	report := &model.AnalysisReport{
		UserID:              toolCtx.user.ID,
		ReportType:          "recommendation",
		ReportTitle:         fmt.Sprintf("AI 推荐报告 (%s)", now.Format("2006-01-02 15:04")),
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now,
		SymbolsCount:        len(items),
		WinningTrades:       0,
		LosingTrades:        0,
		TotalInvestment:     decimal.Zero,
		TotalProfit:         toolCtx.user.TotalProfit,
		ProfitRate:          decimal.Zero,
		RiskLevel:           normalizeRiskLevel(toolCtx.user.RiskTolerance),
		MarketDataStatus:    summarizeRecommendationDataStatus(toolCtx.scored),
		InvestmentStyle:     stringPointerIfNotEmpty(toolCtx.user.InvestmentPreference),
		SummaryText:         reply,
		RiskAnalysis:        stringPointerIfNotEmpty("推荐报告基于用户偏好、历史关注和近期新闻热点自动生成，请结合自身仓位与风险承受能力判断。"),
		PatternInsights:     stringPointerIfNotEmpty(fmt.Sprintf("候选池 %d 个，其中新闻潜力股 %d 个。", len(toolCtx.candidates), len(toolCtx.newsItems))),
		PredictionText:      stringPointerIfNotEmpty("推荐结论偏向当前候选池中的相对强势标的，但后续仍需结合新闻兑现和市场风格切换持续验证。"),
		ChartData:           nil,
		Recommendations:     stringPointerIfNotEmpty(recommendationsJSON),
		RawAIOutput:         stringPointerIfNotEmpty(raw),
		AIModel:             fallbackString(s.llmProvider.ModelName(), "unknown"),
	}
	if err := s.analysisReportRepo.CreateWithItems(report, items); err != nil {
		return nil, err
	}
	return &responsedto.AnalysisRecommendationsResponse{
		GeneratedAt: now.Format("2006-01-02 15:04:05"),
		ReportID:    report.ID,
		ProfileSummary: responsedto.RecommendationProfileSummary{
			InvestmentPreference: toolCtx.user.InvestmentPreference,
			RiskTolerance:        toolCtx.user.RiskTolerance,
			TotalProfit:          toolCtx.user.TotalProfit.StringFixed(2),
			HeldPositions:        countHeldCandidates(toolCtx.candidates),
			CandidateCount:       len(toolCtx.candidates),
		},
		SummaryText: reply,
		Candidates:  candidates,
	}, nil
}

func summarizeRecommendationDataStatus(candidates []scoredCandidate) string {
	statuses := make([]string, 0, len(candidates))
	for _, item := range candidates {
		statuses = append(statuses, item.candidate.dataStatus)
	}
	return summarizeMarketDataStatus(statuses)
}

func (s *recommendationService) buildRecommendationChatResponse(toolCtx *recommendationToolContext, reply string, report *responsedto.AnalysisRecommendationsResponse) *responsedto.RecommendationChatResponse {
	responseMessages := make([]responsedto.StockChatMessageResponse, 0, len(toolCtx.messages)+2)
	for _, item := range toolCtx.messages {
		responseMessages = append(responseMessages, responsedto.StockChatMessageResponse{Role: item.Role, Content: item.Content})
	}
	responseMessages = append(responseMessages,
		responsedto.StockChatMessageResponse{Role: "user", Content: toolCtx.question},
		responsedto.StockChatMessageResponse{Role: "assistant", Content: reply},
	)
	stepKey := "clarify"
	stepLabel := "继续补充偏好"
	focusSummary := "本轮仍在补充用户偏好、候选范围、新闻与指标分析，暂未生成正式报告。"
	reportID := uint64(0)
	reportTitle := ""
	candidates := []responsedto.RecommendationItemResponse{}
	profileSummary := responsedto.RecommendationProfileSummary{
		InvestmentPreference: toolCtx.user.InvestmentPreference,
		RiskTolerance:        toolCtx.user.RiskTolerance,
		TotalProfit:          toolCtx.user.TotalProfit.StringFixed(2),
		HeldPositions:        countHeldCandidates(toolCtx.candidates),
		CandidateCount:       len(toolCtx.candidates),
	}
	if report != nil {
		stepKey = "done"
		stepLabel = "报告生成完成"
		focusSummary = "本轮推荐已综合用户相关标的和近期新闻潜力股。"
		reportID = report.ReportID
		reportTitle = fmt.Sprintf("AI 推荐报告 (%s)", time.Now().Format("2006-01-02 15:04"))
		candidates = report.Candidates
		profileSummary = report.ProfileSummary
	}
	generatedAt := time.Now().Format("2006-01-02 15:04:05")
	return &responsedto.RecommendationChatResponse{
		ContextID:    toolCtx.contextID,
		AssetName:    "AI 推荐对话",
		Market:       "recommendation",
		Reply:        reply,
		AIModel:      fallbackString(s.llmProvider.ModelName(), "unknown"),
		GeneratedAt:  generatedAt,
		NewsStatus:   "partial",
		NewsSummary:  fmt.Sprintf("本轮纳入 %d 个新闻潜力股候选。", len(toolCtx.newsItems)),
		NewsCoverage: fmt.Sprintf("用户相关候选 %d 个，新闻潜力股 %d 个。", len(toolCtx.candidates)-len(toolCtx.newsItems), len(toolCtx.newsItems)),
		NewsItems:    recommendationNewsItems(toolCtx.newsItems),
		Snapshot: responsedto.StockChatSnapshotResponse{
			LastPrice:     "0",
			ChangePercent: "0",
			Period:        "recommendation",
			TrendSummary:  fmt.Sprintf("推荐候选池 %d 个，已持仓 %d 个，新闻潜力股 %d 个。", len(toolCtx.candidates), countHeldCandidates(toolCtx.candidates), len(toolCtx.newsItems)),
		},
		Messages: responseMessages,
		Context: responsedto.RecommendationChatContextResponse{
			StepKey:        stepKey,
			StepLabel:      stepLabel,
			ProfileSummary: profileSummary,
			CandidateCount: len(toolCtx.candidates),
			DiscoveryCount: len(toolCtx.newsItems),
			HeldCount:      countHeldCandidates(toolCtx.candidates),
			FocusSummary:   focusSummary,
		},
		ReportID:    reportID,
		ReportTitle: reportTitle,
		Candidates:  candidates,
	}
}

func (s *recommendationService) GetRecommendationChatContext(userID, contextID uint64) (*responsedto.RecommendationChatContextSnapshotResponse, error) {
	if contextID == 0 {
		return nil, fmt.Errorf("context id is required")
	}
	if s.contextService == nil {
		return nil, fmt.Errorf("context service is unavailable")
	}
	entity, err := s.contextService.loadEntity(userID, contextID)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, fmt.Errorf("chat context not found")
	}
	messages := make([]responsedto.StockChatMessageResponse, 0)
	if strings.TrimSpace(entity.MessagesJSON) != "" {
		var stored []requestdto.StockChatMessageRequest
		if err := json.Unmarshal([]byte(entity.MessagesJSON), &stored); err != nil {
			return nil, err
		}
		for _, item := range stored {
			messages = append(messages, responsedto.StockChatMessageResponse{Role: item.Role, Content: item.Content})
		}
	}
	metadata := recommendationChatContextMetadata{}
	if strings.TrimSpace(entity.MetadataJSON) != "" {
		if err := json.Unmarshal([]byte(entity.MetadataJSON), &metadata); err != nil {
			return nil, err
		}
		if metadata.StepContext == nil && len(metadata.ToolTrace) > 0 {
			last := metadata.ToolTrace[len(metadata.ToolTrace)-1]
			metadata.StepContext = &responsedto.ChatStepContextResponse{Stage: last.Stage, Label: last.Label, Summary: last.Summary, ToolName: last.ToolName}
		}
	}
	reportID := uint64(0)
	if entity.ReportID != nil {
		reportID = *entity.ReportID
	}
	return &responsedto.RecommendationChatContextSnapshotResponse{
		Messages:    messages,
		ToolTrace:   metadata.ToolTrace,
		ToolResults: metadata.ToolResults,
		StepContext: metadata.StepContext,
		NewsItems:   metadata.NewsItems,
		Reply:       entity.LastReply,
		GeneratedAt: metadata.GeneratedAt,
		ReportID:    reportID,
		ReportTitle: metadata.ReportTitle,
	}, nil
}

func (s *recommendationService) hydrateRecommendationChatRequest(userID uint64, req *requestdto.RecommendationChatRequest) *requestdto.RecommendationChatRequest {
	if req == nil || req.ContextID == 0 || len(req.Messages) > 0 {
		return req
	}
	if s.contextService == nil {
		return req
	}
	messages, err := s.contextService.loadMessages(userID, req.ContextID)
	if err != nil {
		return req
	}
	metadata, metaErr := s.contextService.loadMetadata(userID, req.ContextID)
	if metaErr == nil {
		if toolMessage := buildToolResultsContextMessage(metadata); strings.TrimSpace(toolMessage) != "" {
			messages = append(messages, requestdto.StockChatMessageRequest{Role: "system", Content: toolMessage})
		}
	}
	if len(messages) == 0 {
		return req
	}
	cloned := *req
	cloned.Messages = messages
	return &cloned
}

func (s *recommendationService) persistRecommendationChat(userID uint64, resp *responsedto.RecommendationChatResponse, req *requestdto.RecommendationChatRequest, toolCtx *recommendationToolContext) error {
	if s.contextService == nil || resp == nil || req == nil {
		return nil
	}
	metadata := recommendationChatContextMetadata{
		ToolTrace:   resp.ToolTrace,
		ToolResults: resp.ToolResults,
		StepContext: resp.StepContext,
		NewsItems:   resp.NewsItems,
		GeneratedAt: resp.GeneratedAt,
		ReportTitle: resp.ReportTitle,
	}
	if toolCtx != nil {
		metadata.Candidates = snapshotRecommendationCandidates(toolCtx.candidates)
		metadata.Insights = snapshotRecommendationInsights(toolCtx.insights)
	}
	contextID, err := s.contextService.saveContext(
		userID,
		"recommendation",
		fmt.Sprintf("recommendation:%d", resp.ReportID),
		fallbackString(resp.ReportTitle, "AI 推荐对话"),
		req.ContextID,
		resp.ReportID,
		resp.Messages,
		req.Question,
		resp.Reply,
		metadata,
	)
	if err != nil {
		return err
	}
	resp.ContextID = contextID
	return nil
}

func recommendationNewsItems(items []newsCandidateItem) []responsedto.StockChatNewsItemResponse {
	result := make([]responsedto.StockChatNewsItemResponse, 0, len(items))
	for _, item := range items {
		result = append(result, responsedto.StockChatNewsItemResponse{
			Title:       item.Headline,
			Summary:     item.Summary,
			Source:      "news_discovery",
			URL:         "",
			PublishedAt: "",
			Provider:    item.Symbol,
		})
	}
	return result
}

func recommendationNewsPayload(items []newsCandidateItem) map[string]any {
	payloadItems := make([]map[string]string, 0, minInt(len(items), 3))
	for _, item := range items {
		if len(payloadItems) >= 3 {
			break
		}
		payloadItems = append(payloadItems, map[string]string{
			"symbol":         item.Symbol,
			"asset_name":     item.AssetName,
			"headline":       item.Headline,
			"digest":         compactNewsDigest(item.Summary, 88),
			"change_percent": item.ChangePercent,
			"provider":       item.Provider,
			"published_at":   item.PublishedAt,
		})
	}
	return map[string]any{
		"discovery_count":      len(items),
		"top_headlines":        payloadItems,
		"news_signal_summary":  recommendationNewsSignalSummary(items),
	}
}

func recommendationNewsSignalSummary(items []newsCandidateItem) string {
	if len(items) == 0 {
		return "暂无可用新闻催化摘要"
	}
	names := make([]string, 0, minInt(len(items), 3))
	for _, item := range items {
		if len(names) >= 3 {
			break
		}
		name := strings.TrimSpace(item.AssetName)
		if name == "" {
			name = strings.TrimSpace(item.Symbol)
		}
		if name == "" {
			continue
		}
		names = append(names, name)
	}
	if len(names) == 0 {
		return fmt.Sprintf("近期共识别 %d 个新闻催化候选", len(items))
	}
	return fmt.Sprintf("近期共识别 %d 个新闻催化候选，当前优先关注：%s。", len(items), strings.Join(names, "、"))
}

func compactNewsDigest(value string, limit int) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" || limit <= 0 {
		return trimmed
	}
	runes := []rune(trimmed)
	if len(runes) <= limit {
		return trimmed
	}
	return strings.TrimSpace(string(runes[:limit])) + "..."
}

func convertScoredCandidates(candidates []scoredCandidate, aiItems []aiRecommendationItem) []responsedto.RecommendationItemResponse {
	aiMap := make(map[string]aiRecommendationItem, len(aiItems))
	for _, item := range aiItems {
		aiMap[normalizeSymbol(item.Symbol)] = item
	}
	result := make([]responsedto.RecommendationItemResponse, 0, len(candidates))
	for _, item := range candidates {
		override, ok := aiMap[item.candidate.symbol]
		action := item.action
		matchReason := item.matchReason
		riskNote := item.riskNote
		if ok {
			action = normalizeRecommendation(override.Action)
			matchReason = fallbackString(strings.TrimSpace(override.MatchReason), matchReason)
			riskNote = fallbackString(strings.TrimSpace(override.RiskNote), riskNote)
		}
		result = append(result, responsedto.RecommendationItemResponse{
			Symbol:        item.candidate.symbol,
			AssetName:     fallbackString(item.candidate.assetName, item.candidate.symbol),
			AssetType:     item.candidate.assetType,
			Market:        item.candidate.market,
			Action:        action,
			Score:         item.score.StringFixed(2),
			LatestPrice:   decimalToString(item.candidate.latestPrice, 4),
			ChangePercent: decimalToString(item.candidate.changePercent, 2),
			MatchReason:   matchReason,
			RiskNote:      riskNote,
			DataStatus:    item.candidate.dataStatus,
			IsHeld:        item.candidate.isHeld,
			TradeCount:    item.candidate.tradeCount,
		})
	}
	return result
}

func buildFallbackRecommendationSummary(user *model.User, candidates []scoredCandidate) string {
	if len(candidates) == 0 {
		return "当前没有足够的候选标的可供推荐。"
	}
	top := candidates[0]
	return fmt.Sprintf("基于你的%s投资偏好和%s风险承受能力，当前更适合优先关注 %s 等熟悉标的，先从已有交易经验和现有持仓中筛选机会，再决定是否继续加仓或观察。", preferenceLabel(normalizeInvestmentStyle(user.InvestmentPreference)), fallbackString(strings.TrimSpace(user.RiskTolerance), "当前"), top.candidate.symbol)
}

func preferenceLabel(value string) string {
	switch normalizeInvestmentStyle(value) {
	case "conservative":
		return "保守型"
	case "aggressive":
		return "激进型"
	default:
		return "稳健型"
	}
}

func countHeldCandidates(candidates []recommendationCandidate) int {
	count := 0
	for _, candidate := range candidates {
		if candidate.isHeld {
			count++
		}
	}
	return count
}

func decimalToString(value decimal.Decimal, scale int32) string {
	if value.IsZero() {
		return value.StringFixed(scale)
	}
	return value.StringFixed(scale)
}

func snapshotRecommendationCandidates(candidates []recommendationCandidate) []persistedRecommendationCandidate {
	result := make([]persistedRecommendationCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		sources := make([]persistedRecommendationCandidateSource, 0, len(candidate.sources))
		for _, source := range candidate.sources {
			sources = append(sources, persistedRecommendationCandidateSource{
				TypeName:    source.typeName,
				Headline:    source.headline,
				Summary:     source.summary,
				Provider:    source.provider,
				PublishedAt: source.publishedAt.Format(time.RFC3339),
			})
		}
		result = append(result, persistedRecommendationCandidate{
			Symbol:        candidate.symbol,
			AssetName:     candidate.assetName,
			AssetType:     candidate.assetType,
			Market:        candidate.market,
			Sources:       sources,
			IsHeld:        candidate.isHeld,
			TradeCount:    candidate.tradeCount,
			LatestPrice:   decimalToString(candidate.latestPrice, 4),
			ChangePercent: decimalToString(candidate.changePercent, 2),
			DataStatus:    candidate.dataStatus,
		})
	}
	return result
}

func restoreRecommendationCandidates(items []persistedRecommendationCandidate) []recommendationCandidate {
	result := make([]recommendationCandidate, 0, len(items))
	for _, item := range items {
		candidate := recommendationCandidate{
			symbol:        normalizeSymbol(item.Symbol),
			assetName:     strings.TrimSpace(item.AssetName),
			assetType:     strings.TrimSpace(item.AssetType),
			market:        strings.TrimSpace(item.Market),
			isHeld:        item.IsHeld,
			tradeCount:    item.TradeCount,
			dataStatus:    strings.TrimSpace(item.DataStatus),
			sourceSet:     make(map[string]struct{}),
			latestPrice:   decimal.Zero,
			changePercent: decimal.Zero,
		}
		if value, err := decimal.NewFromString(strings.TrimSpace(item.LatestPrice)); err == nil {
			candidate.latestPrice = value
		}
		if value, err := decimal.NewFromString(strings.TrimSpace(item.ChangePercent)); err == nil {
			candidate.changePercent = value
		}
		for _, source := range item.Sources {
			typeName := strings.TrimSpace(source.TypeName)
			candidate.sources = append(candidate.sources, candidateSource{
				typeName:    typeName,
				headline:    strings.TrimSpace(source.Headline),
				summary:     strings.TrimSpace(source.Summary),
				provider:    strings.TrimSpace(source.Provider),
				publishedAt: func() time.Time {
					value, err := time.Parse(time.RFC3339, strings.TrimSpace(source.PublishedAt))
					if err != nil {
						return time.Time{}
					}
					return value
				}(),
			})
			if typeName != "" {
				candidate.sourceSet[typeName] = struct{}{}
			}
		}
		result = append(result, candidate)
	}
	return result
}

func cachedUserCandidates(candidates []recommendationCandidate) []recommendationCandidate {
	result := make([]recommendationCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if hasCandidateSource(candidate, "news_discovery") || hasCandidateSource(candidate, "board_theme") {
			continue
		}
		result = append(result, candidate)
	}
	return result
}

func cachedNewsCandidates(candidates []recommendationCandidate) []recommendationCandidate {
	result := make([]recommendationCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if hasCandidateSource(candidate, "news_discovery") {
			result = append(result, candidate)
		}
	}
	return result
}

func snapshotRecommendationInsights(insights map[string]recommendationStockInsight) map[string]persistedRecommendationInsight {
	if len(insights) == 0 {
		return nil
	}
	result := make(map[string]persistedRecommendationInsight, len(insights))
	for symbol, insight := range insights {
		entry := persistedRecommendationInsight{
			Symbol:       symbol,
			TrendSummary: insight.TrendSummary,
			Indicators:   insight.Indicators,
			NewsStatus:   safeNewsStatus(insight.News),
			NewsSummary:  safeNewsSummary(insight.News),
			NewsCoverage: safeNewsCoverage(insight.News),
		}
		if insight.Detail != nil {
			entry.Name = insight.Detail.Name
			entry.Market = insight.Detail.Market
			entry.Industry = insight.Detail.Industry
			entry.Region = insight.Detail.Region
			if len(insight.Detail.Concepts) > 0 {
				entry.Concepts = append([]string(nil), insight.Detail.Concepts...)
			}
		}
		if insight.Profile != nil {
			if strings.TrimSpace(entry.Name) == "" {
				entry.Name = insight.Profile.Name
			}
			if strings.TrimSpace(entry.Market) == "" {
				entry.Market = insight.Profile.Market
			}
			if strings.TrimSpace(entry.Industry) == "" {
				entry.Industry = insight.Profile.Industry
			}
			if strings.TrimSpace(entry.Region) == "" {
				entry.Region = insight.Profile.Region
			}
			if len(entry.Concepts) == 0 && len(insight.Profile.Concepts) > 0 {
				entry.Concepts = append([]string(nil), insight.Profile.Concepts...)
			}
			if len(insight.Profile.Boards) > 0 {
				entry.Boards = append([]responsedto.StockBoardMembershipResponse(nil), insight.Profile.Boards...)
			}
		}
		result[symbol] = entry
	}
	return result
}

func restoreRecommendationInsights(items map[string]persistedRecommendationInsight) map[string]recommendationStockInsight {
	if len(items) == 0 {
		return map[string]recommendationStockInsight{}
	}
	result := make(map[string]recommendationStockInsight, len(items))
	for symbol, item := range items {
		normalized := normalizeSymbol(symbol)
		if normalized == "" {
			normalized = normalizeSymbol(item.Symbol)
		}
		if normalized == "" {
			continue
		}
		var profile *responsedto.StockProfileResponse
		if strings.TrimSpace(item.Industry) != "" || strings.TrimSpace(item.Region) != "" || len(item.Concepts) > 0 || len(item.Boards) > 0 {
			profile = &responsedto.StockProfileResponse{
				Symbol:   normalized,
				Name:     strings.TrimSpace(item.Name),
				Market:   strings.TrimSpace(item.Market),
				Industry: strings.TrimSpace(item.Industry),
				Region:   strings.TrimSpace(item.Region),
				Concepts: append([]string(nil), item.Concepts...),
				Boards:   append([]responsedto.StockBoardMembershipResponse(nil), item.Boards...),
			}
		}
		result[normalized] = recommendationStockInsight{
			Symbol: normalized,
			Detail: &responsedto.MarketStockDetailResponse{
				Symbol:   normalized,
				Name:     strings.TrimSpace(item.Name),
				Market:   strings.TrimSpace(item.Market),
				Industry: strings.TrimSpace(item.Industry),
				Region:   strings.TrimSpace(item.Region),
				Concepts: append([]string(nil), item.Concepts...),
			},
			Profile: profile,
			News: &StockNewsContext{
				Status:   strings.TrimSpace(item.NewsStatus),
				Summary:  strings.TrimSpace(item.NewsSummary),
				Coverage: strings.TrimSpace(item.NewsCoverage),
			},
			Indicators:   item.Indicators,
			TrendSummary: strings.TrimSpace(item.TrendSummary),
		}
	}
	return result
}

func restoreNewsCandidateItems(items []responsedto.StockChatNewsItemResponse) []newsCandidateItem {
	result := make([]newsCandidateItem, 0, len(items))
	for _, item := range items {
		result = append(result, newsCandidateItem{
			Headline:    strings.TrimSpace(item.Title),
			Summary:     strings.TrimSpace(item.Summary),
			Provider:    strings.TrimSpace(item.Provider),
			PublishedAt: strings.TrimSpace(item.PublishedAt),
		})
	}
	return result
}
