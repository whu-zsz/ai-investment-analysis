package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	responsedto "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/pkg/llm"

	"github.com/shopspring/decimal"
)

type RecommendationService interface {
	GetCandidates(userID uint64) (*responsedto.AnalysisCandidatesResponse, error)
	GetRecommendations(userID uint64) (*responsedto.AnalysisRecommendationsResponse, error)
}

type recommendationService struct {
	userRepo           repository.UserRepository
	transactionRepo    repository.TransactionRepository
	portfolioRepo      repository.PortfolioRepository
	marketSnapshotRepo repository.MarketSnapshotRepository
	marketDataService  MarketDataService
	llmProvider        llm.Provider
}

type candidateSource struct {
	typeName string
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

func NewRecommendationService(
	userRepo repository.UserRepository,
	transactionRepo repository.TransactionRepository,
	portfolioRepo repository.PortfolioRepository,
	marketSnapshotRepo repository.MarketSnapshotRepository,
	marketDataService MarketDataService,
	llmProvider llm.Provider,
) RecommendationService {
	return &recommendationService{
		userRepo:           userRepo,
		transactionRepo:    transactionRepo,
		portfolioRepo:      portfolioRepo,
		marketSnapshotRepo: marketSnapshotRepo,
		marketDataService:  marketDataService,
		llmProvider:        llmProvider,
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

	scored := scoreCandidates(candidates, user)
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
				symbol:    normalized,
				assetName: strings.TrimSpace(assetName),
				assetType: strings.TrimSpace(assetType),
				sourceSet: make(map[string]struct{}),
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

	symbols := make([]string, 0, len(candidateMap))
	missingSymbols := make([]string, 0)
	for symbol, candidate := range candidateMap {
		symbols = append(symbols, symbol)
		snapshot, snapshotErr := s.marketSnapshotRepo.FindLatestBySymbol(symbol)
		if snapshotErr == nil && snapshot != nil {
			candidate.market = snapshot.Market
			candidate.latestPrice = snapshot.LastPrice
			candidate.changePercent = snapshot.ChangePercent
			candidate.dataStatus = marketDataStatusComplete
			if candidate.assetName == "" {
				candidate.assetName = snapshot.Name
			}
			continue
		}
		missingSymbols = append(missingSymbols, symbol)
	}

	if len(missingSymbols) > 0 {
		if _, fetchErr := s.marketDataService.FetchAndStoreQuotesBySymbols(ctx, missingSymbols); fetchErr == nil {
			for _, symbol := range missingSymbols {
				candidate := candidateMap[symbol]
				snapshot, snapshotErr := s.marketSnapshotRepo.FindLatestBySymbol(symbol)
				if snapshotErr != nil || snapshot == nil {
					continue
				}
				candidate.market = snapshot.Market
				candidate.latestPrice = snapshot.LastPrice
				candidate.changePercent = snapshot.ChangePercent
				candidate.dataStatus = marketDataStatusFetchedLive
				if candidate.assetName == "" {
					candidate.assetName = snapshot.Name
				}
			}
		}
	}

	result := make([]recommendationCandidate, 0, len(candidateMap))
	for _, symbol := range symbols {
		candidate := candidateMap[symbol]
		if candidate.assetName == "" {
			candidate.assetName = symbol
		}
		result = append(result, *candidate)
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

func scoreCandidates(candidates []recommendationCandidate, user *model.User) []scoredCandidate {
	result := make([]scoredCandidate, 0, len(candidates))
	preference := normalizeInvestmentStyle(user.InvestmentPreference)
	riskTolerance := strings.ToLower(strings.TrimSpace(user.RiskTolerance))
	for _, candidate := range candidates {
		score := decimal.NewFromInt(50)
		if candidate.isHeld {
			score = score.Add(decimal.NewFromInt(12))
		}
		if candidate.tradeCount >= 3 {
			score = score.Add(decimal.NewFromInt(10))
		} else if candidate.tradeCount > 0 {
			score = score.Add(decimal.NewFromInt(5))
		}
		if candidate.dataStatus == marketDataStatusComplete || candidate.dataStatus == marketDataStatusFetchedLive {
			score = score.Add(decimal.NewFromInt(8))
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
	if len(result) > 6 {
		result = result[:6]
	}
	return result
}

func buildMatchReason(candidate recommendationCandidate, preference string) string {
	parts := make([]string, 0, 3)
	if candidate.isHeld {
		parts = append(parts, "当前已持仓，便于结合已有成本和仓位继续决策")
	}
	if candidate.tradeCount > 0 {
		parts = append(parts, fmt.Sprintf("最近交易记录中出现 %d 次，属于你的熟悉标的", candidate.tradeCount))
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
	if riskTolerance == "low" || riskTolerance == "conservative" {
		if change.GreaterThan(decimal.NewFromInt(6)) || change.LessThan(decimal.NewFromInt(-4)) {
			return "你的风险承受偏低，这个标的近期波动偏大，建议控制仓位。"
		}
		return "波动尚在可控区间，更适合低频跟踪。"
	}
	if change.LessThan(decimal.NewFromInt(-8)) {
		return "近期回撤较明显，若继续关注应先确认回撤原因。"
	}
	return "建议结合仓位集中度和后续趋势继续判断。"
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
	content, err := s.llmProvider.GetContent(context.Background(), "你是一位严谨的股票推荐解释助手，只能输出合法 JSON。", userPrompt)
	if err != nil {
		return nil, err
	}
	var parsed aiRecommendationEnvelope
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &parsed); err != nil {
		return nil, err
	}
	return &parsed, nil
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
