package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	requestdto "stock-analysis-backend/internal/dto/request"
	responsedto "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
)

type recommendationChatExecutor struct {
	service     *recommendationService
	userID      uint64
	req         *requestdto.RecommendationChatRequest
	user        *model.User
	question    string
	messages    []stockChatMessage
	reportID    uint64
	report      *responsedto.AnalysisRecommendationsResponse
	stepContext *responsedto.ChatStepContextResponse
	newsItems   []newsCandidateItem
	trace       []responsedto.ChatToolTraceStepResponse

	candidates []recommendationCandidate
	scored     []scoredCandidate
	insights   map[string]recommendationStockInsight
}

type stockChatExecutor struct {
	service  *stockChatService
	req      *requestdto.StockChatRequest
	question string
	messages []stockChatMessage
	symbol   string
	detail   *responsedto.MarketStockDetailResponse
	profile  *responsedto.StockProfileResponse
	kline    *responsedto.MarketStockKlineResponse
	news     *StockNewsContext
}

type boardChatExecutor struct {
	service   *boardChatService
	req       *requestdto.BoardChatRequest
	question  string
	messages  []stockChatMessage
	boardType string
	code      string
	board     *responsedto.MarketBoardDetailResponse
	news      *StockNewsContext
}

func newRecommendationChatExecutor(service *recommendationService, userID uint64, req *requestdto.RecommendationChatRequest) (*recommendationChatExecutor, error) {
	if req == nil || strings.TrimSpace(req.Question) == "" {
		return nil, fmt.Errorf("question is required")
	}
	user, err := service.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	executor := &recommendationChatExecutor{
		service:  service,
		userID:   userID,
		req:      req,
		user:     user,
		question: strings.TrimSpace(req.Question),
		messages: normalizeStockChatMessages(req.Messages),
		reportID: req.ReportID,
		trace:    make([]responsedto.ChatToolTraceStepResponse, 0, 8),
		insights: make(map[string]recommendationStockInsight),
	}
	if service.contextService != nil && req.ContextID > 0 {
		if metadata, metaErr := service.contextService.loadMetadata(userID, req.ContextID); metaErr == nil && metadata != nil {
			executor.candidates = restoreRecommendationCandidates(metadata.Candidates)
			executor.newsItems = restoreNewsCandidateItems(metadata.NewsItems)
			executor.insights = restoreRecommendationInsights(metadata.Insights)
		}
	}
	return executor, nil
}

func newStockChatExecutor(service *stockChatService, req *requestdto.StockChatRequest) (*stockChatExecutor, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	symbol := normalizeSymbol(req.Symbol)
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	question := strings.TrimSpace(req.Question)
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}
	return &stockChatExecutor{
		service:  service,
		req:      req,
		question: question,
		messages: normalizeStockChatMessages(req.Messages),
		symbol:   symbol,
		kline:    &responsedto.MarketStockKlineResponse{Items: []responsedto.MarketKlineBarResponse{}},
	}, nil
}

func newBoardChatExecutor(service *boardChatService, req *requestdto.BoardChatRequest) (*boardChatExecutor, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}
	boardType := strings.TrimSpace(req.BoardType)
	code := strings.TrimSpace(req.Code)
	question := strings.TrimSpace(req.Question)
	if boardType == "" || code == "" {
		return nil, fmt.Errorf("board_type and code are required")
	}
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}
	return &boardChatExecutor{
		service:   service,
		req:       req,
		question:  question,
		messages:  normalizeStockChatMessages(req.Messages),
		boardType: boardType,
		code:      code,
	}, nil
}

func commonToolDefinitions(includeProfile, includeHoldings, includeRanking, includeSave bool) []chatToolDefinition {
	defs := make([]chatToolDefinition, 0, 8)
	if includeProfile {
		defs = append(defs, chatToolDefinition{Name: "get_user_investment_profile", Description: "读取用户投资偏好、风险承受和累计盈亏摘要"})
	}
	if includeHoldings {
		defs = append(defs, chatToolDefinition{Name: "get_user_positions_and_watch_history", Description: "读取用户持仓和历史交易关注标的，并补充基础行情"})
	}
	defs = append(defs,
		chatToolDefinition{Name: "get_recent_market_news_candidates", Description: "读取近期可信新闻热点候选或个股/板块相关新闻"},
		chatToolDefinition{Name: "get_stock_quote_and_trend", Description: "读取单只股票的最新行情、近端 K 线和趋势摘要"},
		chatToolDefinition{Name: "get_stock_profile_and_boards", Description: "读取股票简介、行业、概念、所属板块"},
		chatToolDefinition{Name: "get_board_heat_and_constituents", Description: "读取板块热度、宽度和成分股表现"},
	)
	if includeRanking {
		defs = append(defs, chatToolDefinition{Name: "rank_recommendation_candidates", Description: "基于用户偏好和候选池进行程序化排序"})
	}
	if includeSave {
		defs = append(defs, chatToolDefinition{Name: "save_recommendation_report", Description: "保存结构化推荐报告并返回 report_id"})
	}
	return defs
}

func (e *recommendationChatExecutor) Definitions() []chatToolDefinition {
	return []chatToolDefinition{
		{Name: "get_user_investment_profile", Description: "读取用户投资偏好、风险承受、累计盈亏等画像摘要"},
		{Name: "get_user_positions_and_watch_history", Description: "读取用户持仓、交易关注标的，并形成初始候选池"},
		{Name: "search_relevant_boards", Description: "根据用户问题、新闻摘要或主题描述，检索最相关的行业/概念板块"},
		{Name: "list_market_boards", Description: "读取当前可用的行业板块与概念板块名称列表"},
		{Name: "search_board_stocks", Description: "根据板块类型和板块代码，读取该板块的成分股与热度摘要"},
		{Name: "get_recent_market_news_candidates", Description: "读取近期可信新闻热点候选，补充潜力股来源"},
		{Name: "get_stock_indicators_and_news", Description: "读取单只股票近期指标摘要与相关新闻，在推荐前必须优先使用"},
		{Name: "rank_recommendation_candidates", Description: "基于用户偏好和候选池进行程序化排序"},
		{Name: "save_recommendation_report", Description: "保存结构化推荐报告并返回 report_id"},
	}
}
func (e *stockChatExecutor) Definitions() []chatToolDefinition {
	return commonToolDefinitions(false, false, false, false)
}
func (e *boardChatExecutor) Definitions() []chatToolDefinition {
	return commonToolDefinitions(false, false, false, false)
}

func (e *recommendationChatExecutor) PlanningPrompt() string {
	return "你是推荐对话的工具规划助手。你只能输出 JSON 工具计划，不能直接给股票结论。"
}

func (e *recommendationChatExecutor) PlanActionPrompt() string {
	return "你是推荐对话的下一步动作决策助手。你不能直接回答股票结论，也不能输出工具名。你只能输出一个 JSON 对象，字段仅允许 next_action、target、reason。next_action 只能是 ask_user / load_profile / load_holdings / discover_news / explore_boards / rank_candidates / analyze_candidate / save_report / finish。target 只在 analyze_candidate 时可填写 top_ranked_stock / second_ranked_stock / best_news_candidate。若用户明确提到行业、赛道、板块、概念、主题偏好，或问题里出现了新闻事件、产业链描述、催化方向，你必须优先选择 explore_boards，让系统先检索最相关板块，再围绕该板块候选展开分析，不能忽略该偏好继续泛化推荐。正式推荐前，必须至少完成持仓候选、新闻候选、候选排序，并且至少完成 1 只候选股票的指标与新闻分析；如果用户给出了明确行业偏好，至少应优先分析 1 到 2 只与该偏好直接相关的股票。若信息仍不足以正式推荐，应优先选择 ask_user。若已经具备正式推荐条件，且准备生成正式推荐结果，则必须先选择 save_report，而不是直接 finish。禁止输出 Markdown、解释文字或代码块。"
}

func (e *stockChatExecutor) PlanningPrompt() string {
	return "你是个股对话的工具规划助手。你必须按需调用个股、新闻、板块相关工具，不要直接回答结论，只输出 JSON 工具计划。"
}

func (e *boardChatExecutor) PlanningPrompt() string {
	return "你是板块对话的工具规划助手。你必须按需调用板块、新闻、代表个股相关工具，不要直接回答结论，只输出 JSON 工具计划。"
}

func (e *recommendationChatExecutor) FinalPrompt() string {
	return e.service.recommendationFinalPrompt()
}
func (e *stockChatExecutor) FinalPrompt() string { return stockChatSystemPrompt }
func (e *boardChatExecutor) FinalPrompt() string { return boardChatSystemPrompt }

func (e *recommendationChatExecutor) Question() string { return e.question }
func (e *stockChatExecutor) Question() string          { return e.question }
func (e *boardChatExecutor) Question() string          { return e.question }

func (e *recommendationChatExecutor) ConversationHistory() string {
	return conversationHistoryText(e.messages)
}

func (e *recommendationChatExecutor) BuildActionPlanPrompt(results []chatToolResult) string {
	resultLines := make([]string, 0, len(results))
	for _, result := range results {
		payloadSummary := summarizeToolResultPayload(result)
		resultLines = append(resultLines, fmt.Sprintf("- tool=%s status=%s summary=%s payload_summary=%s error=%s", result.ToolName, result.Status, result.Summary, payloadSummary, result.Error))
	}
	if len(resultLines) == 0 {
		resultLines = append(resultLines, "- 暂无已执行工具结果")
	}
	return fmt.Sprintf("用户问题:\n%s\n\n历史对话:\n%s\n\n已执行结果摘要:\n%s\n\n当前状态:\n- 候选数=%d\n- 已分析股票数=%d\n- 是否具备推荐基础=%t\n\n请只输出 JSON：{\"next_action\": string, \"target\": string, \"reason\": string}", e.Question(), e.ConversationHistory(), strings.Join(resultLines, "\n"), len(e.candidates), len(e.insights), e.readyForRecommendation())
}

func (e *recommendationChatExecutor) MapActionPlan(plan recommendationActionPlan, results []chatToolResult) (*chatToolPlan, error) {
	action := strings.TrimSpace(plan.NextAction)
	target := strings.TrimSpace(plan.Target)
	normalizedAction, normalizedTarget := e.normalizeRecommendationAction(action, target, results)
	switch normalizedAction {
	case "ask_user", "finish":
		return &chatToolPlan{NeedMoreTools: false}, nil
	case "load_profile":
		return &chatToolPlan{NeedMoreTools: true, Calls: []chatToolCall{{ToolName: "get_user_investment_profile"}}}, nil
	case "load_holdings":
		return &chatToolPlan{NeedMoreTools: true, Calls: []chatToolCall{{ToolName: "get_user_positions_and_watch_history"}}}, nil
	case "discover_news":
		return &chatToolPlan{NeedMoreTools: true, Calls: []chatToolCall{{ToolName: "get_recent_market_news_candidates"}}}, nil
	case "explore_boards":
		if e.hasSuccessfulToolResult(results, "search_relevant_boards") {
			if boardType, code, ok := e.preferredBoardSelection(results); ok {
				args := json.RawMessage([]byte(fmt.Sprintf(`{"board_type":"%s","code":"%s","limit":20}`, boardType, code)))
				return &chatToolPlan{NeedMoreTools: true, Calls: []chatToolCall{{ToolName: "search_board_stocks", Args: args}}}, nil
			}
		}
		if boardType, code, ok := e.preferredBoardSelection(results); ok {
			args := json.RawMessage([]byte(fmt.Sprintf(`{"board_type":"%s","code":"%s","limit":20}`, boardType, code)))
			return &chatToolPlan{NeedMoreTools: true, Calls: []chatToolCall{{ToolName: "search_board_stocks", Args: args}}}, nil
		}
		args := json.RawMessage([]byte(fmt.Sprintf(`{"query":%q,"limit":8}`, recommendationPreferenceText(e.question, e.messages))))
		return &chatToolPlan{NeedMoreTools: true, Calls: []chatToolCall{{ToolName: "search_relevant_boards", Args: args}}}, nil
	case "rank_candidates":
		return &chatToolPlan{NeedMoreTools: true, Calls: []chatToolCall{{ToolName: "rank_recommendation_candidates"}}}, nil
	case "analyze_candidate":
		symbol := e.symbolForActionTarget(normalizedTarget)
		if symbol == "" {
			fallbackAction, fallbackTarget := e.recommendationFallbackAction(results)
			if fallbackAction == "analyze_candidate" {
				symbol = e.symbolForActionTarget(fallbackTarget)
			}
			if symbol == "" {
				return &chatToolPlan{NeedMoreTools: false}, nil
			}
		}
		args := json.RawMessage([]byte(fmt.Sprintf(`{"symbol":"%s"}`, symbol)))
		return &chatToolPlan{NeedMoreTools: true, Calls: []chatToolCall{{ToolName: "get_stock_indicators_and_news", Args: args}}}, nil
	case "save_report":
		return &chatToolPlan{NeedMoreTools: false}, nil
	default:
		return &chatToolPlan{NeedMoreTools: false}, nil
	}
}

func (e *recommendationChatExecutor) normalizeRecommendationAction(action string, target string, results []chatToolResult) (string, string) {
	action = strings.TrimSpace(action)
	target = strings.TrimSpace(target)
	if action == "rank_candidates" {
		if len(e.candidates) == 0 {
			return e.recommendationFallbackAction(results)
		}
		if e.preferenceMentionsBoardTheme() && !e.hasSuccessfulToolResult(results, "search_board_stocks") {
			return e.recommendationFallbackAction(results)
		}
	}
	switch action {
	case "load_profile", "load_holdings", "discover_news", "explore_boards", "rank_candidates", "analyze_candidate", "save_report", "ask_user", "finish":
		if e.recommendationCanFinish(action) {
			return action, target
		}
		return e.recommendationFallbackAction(results)
	default:
		return e.recommendationFallbackAction(results)
	}
}

func (e *recommendationChatExecutor) recommendationCanFinish(action string) bool {
	switch action {
	case "ask_user":
		return !recommendationPreferenceReady(e.user, e.question, e.messages) && len(e.insights) > 0
	case "finish", "save_report":
		return e.readyForRecommendation()
	default:
		return true
	}
}

func (e *recommendationChatExecutor) recommendationFallbackAction(results []chatToolResult) (string, string) {
	if !e.hasSuccessfulToolResult(results, "get_user_investment_profile") {
		return "load_profile", ""
	}
	if !e.hasSuccessfulToolResult(results, "get_user_positions_and_watch_history") && len(cachedUserCandidates(e.candidates)) == 0 {
		return "load_holdings", ""
	}
	if !e.hasSuccessfulToolResult(results, "get_recent_market_news_candidates") && len(cachedNewsCandidates(e.candidates)) == 0 {
		return "discover_news", ""
	}
	if e.preferenceMentionsBoardTheme() && !e.hasSuccessfulToolResult(results, "search_relevant_boards") {
		return "explore_boards", ""
	}
	if e.preferenceMentionsBoardTheme() && (e.hasSuccessfulToolResult(results, "search_relevant_boards") || e.hasSuccessfulToolResult(results, "list_market_boards")) && !e.hasSuccessfulToolResult(results, "search_board_stocks") {
		return "explore_boards", ""
	}
	if len(e.candidates) == 0 {
		return "load_holdings", ""
	}
	if len(e.scored) == 0 {
		return "rank_candidates", ""
	}
	for i := 0; i < len(e.scored) && i < recommendationShortlistLimit; i++ {
		if !e.hasInsightForRankedCandidate(i) {
			if i == 1 {
				return "analyze_candidate", "second_ranked_stock"
			}
			return "analyze_candidate", "top_ranked_stock"
		}
	}
	if !recommendationPreferenceReady(e.user, e.question, e.messages) {
		return "ask_user", ""
	}
	if e.readyForRecommendation() {
		return "save_report", ""
	}
	return "finish", ""
}

func (e *recommendationChatExecutor) hasSuccessfulToolResult(results []chatToolResult, toolName string) bool {
	for _, result := range results {
		if result.ToolName == toolName && result.Status == "ok" {
			return true
		}
	}
	return false
}

func (e *recommendationChatExecutor) hasInsightForRankedCandidate(index int) bool {
	if index < 0 || index >= len(e.scored) {
		return false
	}
	_, ok := e.insights[e.scored[index].candidate.symbol]
	return ok
}

func (e *recommendationChatExecutor) preferenceMentionsBoardTheme() bool {
	content := strings.ToLower(strings.TrimSpace(e.question + "\n" + e.ConversationHistory()))
	for _, keyword := range []string{"白酒", "酿酒", "消费", "半导体", "芯片", "银行", "创新药", "新能源", "军工", "医药", "算力", "人工智能", "券商"} {
		if strings.Contains(content, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

func (e *recommendationChatExecutor) symbolForActionTarget(target string) string {
	if len(e.scored) == 0 && len(e.candidates) > 0 {
		e.scored = scoreCandidates(e.candidates, e.user, recommendationPreferenceText(e.question, e.messages))
	}
	preferenceThemes := extractRecommendationThemes(recommendationPreferenceText(e.question, e.messages))
	if target == "top_ranked_stock" {
		for i := 0; i < len(e.scored) && i < recommendationShortlistLimit; i++ {
			symbol := e.scored[i].candidate.symbol
			if _, ok := e.insights[symbol]; !ok {
				return symbol
			}
		}
	}
	switch target {
	case "second_ranked_stock":
		for i := 1; i < len(e.scored) && i < recommendationShortlistLimit; i++ {
			symbol := e.scored[i].candidate.symbol
			if _, ok := e.insights[symbol]; !ok {
				return symbol
			}
		}
	case "best_news_candidate":
		for _, item := range e.scored {
			if hasCandidateSource(item.candidate, "news_discovery") {
				if _, ok := e.insights[item.candidate.symbol]; ok {
					continue
				}
				return item.candidate.symbol
			}
		}
	}
	if len(preferenceThemes) > 0 {
		for _, item := range e.scored {
			if candidateMatchesThemes(item.candidate, preferenceThemes) {
				return item.candidate.symbol
			}
		}
	}
	if len(e.scored) > 0 {
		return e.scored[0].candidate.symbol
	}
	return ""
}

func (e *recommendationChatExecutor) preferredBoardSelection(results []chatToolResult) (string, string, bool) {
	themes := extractRecommendationThemes(recommendationPreferenceText(e.question, e.messages))
	for i := len(results) - 1; i >= 0; i-- {
		result := results[i]
		if result.Status != "ok" {
			continue
		}
		switch result.ToolName {
		case "search_relevant_boards":
			payload, ok := result.Payload.(map[string]any)
			if !ok {
				continue
			}
			boardType, code, matched := matchBoardFromPayload(payload["boards"], themes)
			if matched {
				return boardType, code, true
			}
			boardType, code, matched = matchBoardFromPayload(payload["boards"], nil)
			if matched {
				return boardType, code, true
			}
		case "list_market_boards":
			if len(themes) == 0 {
				continue
			}
			payload, ok := result.Payload.(map[string]any)
			if !ok {
				continue
			}
			for _, key := range []string{"concept_boards", "industry_boards"} {
				boardType, code, matched := matchBoardFromPayload(payload[key], themes)
				if matched {
					return boardType, code, true
				}
			}
		}
	}
	return "", "", false
}

func matchBoardFromPayload(raw any, themes []string) (string, string, bool) {
	pickBoard := func(name string, code string, boardType string) (string, string, bool) {
		name = strings.TrimSpace(name)
		code = strings.TrimSpace(code)
		boardType = strings.TrimSpace(boardType)
		if name == "" || code == "" || boardType == "" {
			return "", "", false
		}
		if len(themes) == 0 {
			return boardType, code, true
		}
		lowerName := strings.ToLower(name)
		for _, theme := range themes {
			if strings.Contains(lowerName, strings.ToLower(strings.TrimSpace(theme))) {
				return boardType, code, true
			}
		}
		return "", "", false
	}
	tryTypedBoards := func(items []responsedto.MarketBoardItemResponse) (string, string, bool) {
		for _, entry := range items {
			if boardType, code, ok := pickBoard(entry.Name, entry.Code, entry.BoardType); ok {
				return boardType, code, true
			}
		}
		return "", "", false
	}
	switch items := raw.(type) {
	case []any:
		for _, item := range items {
			entry, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if boardType, code, ok := pickBoard(fmt.Sprintf("%v", entry["name"]), fmt.Sprintf("%v", entry["code"]), fmt.Sprintf("%v", entry["board_type"])); ok {
				return boardType, code, true
			}
		}
	case []map[string]any:
		for _, entry := range items {
			if boardType, code, ok := pickBoard(fmt.Sprintf("%v", entry["name"]), fmt.Sprintf("%v", entry["code"]), fmt.Sprintf("%v", entry["board_type"])); ok {
				return boardType, code, true
			}
		}
	case []responsedto.MarketBoardItemResponse:
		return tryTypedBoards(items)
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return "", "", false
	}
	var typed []responsedto.MarketBoardItemResponse
	if err := json.Unmarshal(encoded, &typed); err != nil {
		return "", "", false
	}
	return tryTypedBoards(typed)
}

func (e *stockChatExecutor) ConversationHistory() string { return conversationHistoryText(e.messages) }
func (e *boardChatExecutor) ConversationHistory() string { return conversationHistoryText(e.messages) }

func conversationHistoryText(messages []stockChatMessage) string {
	if len(messages) == 0 {
		return "暂无历史对话"
	}
	lines := make([]string, 0, len(messages))
	for _, item := range messages {
		lines = append(lines, fmt.Sprintf("%s: %s", item.Role, item.Content))
	}
	return strings.Join(lines, "\n")
}

func (e *recommendationChatExecutor) ExecuteTool(ctx context.Context, call chatToolCall) (*chatToolResult, *responsedto.ChatStepContextResponse, error) {
	switch call.ToolName {
	case "get_user_investment_profile":
		profile := responsedto.RecommendationProfileSummary{
			InvestmentPreference: e.user.InvestmentPreference,
			RiskTolerance:        e.user.RiskTolerance,
			TotalProfit:          e.user.TotalProfit.StringFixed(2),
			HeldPositions:        countHeldCandidates(e.candidates),
			CandidateCount:       len(e.candidates),
		}
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: fmt.Sprintf("已读取用户画像：%s / %s", e.user.InvestmentPreference, e.user.RiskTolerance), Payload: profile}, &responsedto.ChatStepContextResponse{Stage: "profile", Label: "用户画像", Summary: fmt.Sprintf("偏好 %s，风险承受 %s，累计盈亏 %s", e.user.InvestmentPreference, e.user.RiskTolerance, e.user.TotalProfit.StringFixed(2)), ProfileSummary: &profile}, nil
	case "get_user_positions_and_watch_history":
		candidates := cachedUserCandidates(e.candidates)
		if len(candidates) == 0 {
			var err error
			candidates, err = e.service.buildCandidates(ctx, e.userID)
			if err != nil {
				return nil, nil, err
			}
			e.candidates = mergeRecommendationCandidates(e.candidates, candidates)
		}
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: fmt.Sprintf("已读取 %d 个用户相关标的", len(candidates)), Payload: e.service.candidatePayload(candidates)}, &responsedto.ChatStepContextResponse{Stage: "holdings", Label: "持仓与关注", Summary: fmt.Sprintf("共 %d 个候选，已持仓 %d 个", len(candidates), countHeldCandidates(candidates)), CandidateCount: len(candidates), HeldCount: countHeldCandidates(candidates), ReferenceSymbols: candidateSymbols(candidates)}, nil
	case "get_recent_market_news_candidates":
		discovered := cachedNewsCandidates(e.candidates)
		if len(discovered) == 0 {
			var err error
			discovered, err = e.service.discoverNewsCandidates(ctx)
			if err != nil {
				return nil, nil, err
			}
			if len(e.candidates) == 0 {
				e.candidates = discovered
			} else {
				e.candidates = mergeRecommendationCandidates(e.candidates, discovered)
			}
		}
		if len(e.candidates) > 0 {
			e.scored = scoreCandidates(e.candidates, e.user, recommendationPreferenceText(e.question, e.messages))
			e.newsItems = buildNewsCandidateItems(e.scored)
		}
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: fmt.Sprintf("已发现 %d 个新闻潜力股候选", len(discovered)), Payload: recommendationNewsPayload(buildNewsCandidateItems(scoreCandidates(discovered, e.user, recommendationPreferenceText(e.question, e.messages))))}, &responsedto.ChatStepContextResponse{Stage: "news", Label: "新闻热点", Summary: fmt.Sprintf("新增新闻潜力股 %d 个", len(discovered)), DiscoveryCount: len(discovered), NewsItems: recommendationNewsItems(e.newsItems)}, nil
	case "list_market_boards":
		// Recommendation needs a much wider board catalog than the homepage preview.
		breadth, err := e.service.marketSnapshotService.GetDashboardMarketBreadth(ctx, 120)
		if err != nil {
			return nil, nil, err
		}
		industryBoards := recommendationBoardCatalog("industry", breadth.Sectors)
		conceptBoards := recommendationBoardCatalog("concept", breadth.Concepts)
		payload := map[string]any{
			"industry_boards": industryBoards,
			"concept_boards":  conceptBoards,
			"snapshot_time":   breadth.SnapshotTime,
			"source":          breadth.Source,
			"is_partial":      breadth.IsPartial,
		}
		boardNames := make([]string, 0, len(industryBoards)+len(conceptBoards))
		for _, item := range industryBoards {
			boardNames = append(boardNames, item["name"])
		}
		for _, item := range conceptBoards {
			boardNames = append(boardNames, item["name"])
		}
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: fmt.Sprintf("已读取 %d 个行业板块、%d 个概念板块", len(industryBoards), len(conceptBoards)), Payload: payload}, &responsedto.ChatStepContextResponse{Stage: "board", Label: "板块目录", Summary: fmt.Sprintf("已加载行业 %d 个、概念 %d 个板块名称", len(industryBoards), len(conceptBoards)), ReferenceBoards: uniqueStrings(boardNames)}, nil
	case "search_relevant_boards":
		var args struct {
			Query string `json:"query"`
			Limit int    `json:"limit"`
		}
		if len(call.Args) > 0 {
			_ = json.Unmarshal(call.Args, &args)
		}
		query := strings.TrimSpace(args.Query)
		if query == "" {
			query = recommendationPreferenceText(e.question, e.messages)
		}
		boards, err := e.service.marketSnapshotService.SearchRelevantBoards(ctx, query, args.Limit)
		if err != nil {
			return nil, nil, err
		}
		payloadBoards := make([]map[string]string, 0, len(boards))
		referenceBoards := make([]string, 0, len(boards))
		for _, item := range boards {
			payloadBoards = append(payloadBoards, map[string]string{
				"board_type":     item.BoardType,
				"code":           item.Code,
				"name":           item.Name,
				"change_percent": item.ChangePercent,
			})
			referenceBoards = append(referenceBoards, item.Name)
		}
		payload := map[string]any{
			"query":  query,
			"boards": payloadBoards,
		}
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: fmt.Sprintf("已检索到 %d 个与当前问题最相关的板块", len(payloadBoards)), Payload: payload}, &responsedto.ChatStepContextResponse{Stage: "board", Label: "板块检索", Summary: fmt.Sprintf("已根据问题与新闻摘要检索到 %d 个候选板块", len(payloadBoards)), ReferenceBoards: uniqueStrings(referenceBoards)}, nil
	case "search_board_stocks":
		var args struct {
			BoardType string `json:"board_type"`
			Code      string `json:"code"`
			Limit     int    `json:"limit"`
		}
		if len(call.Args) > 0 {
			_ = json.Unmarshal(call.Args, &args)
		}
		boardType := strings.TrimSpace(args.BoardType)
		code := strings.TrimSpace(args.Code)
		if boardType == "" || code == "" {
			return &chatToolResult{
					ToolName: call.ToolName,
					Status:   "incomplete",
					Summary:  "缺少板块参数，需先提供 board_type 和 code，或先调用 list_market_boards 获取可用板块。",
					Payload: map[string]any{
						"requires": []string{"board_type", "code"},
					},
				}, &responsedto.ChatStepContextResponse{
					Stage:   "board",
					Label:   "板块成分",
					Summary: "本次未执行板块成分检索，等待先确定板块类型和板块代码。",
				}, nil
		}
		board, err := e.service.marketSnapshotService.GetBoardDetail(ctx, boardType, code, args.Limit)
		if err != nil {
			return nil, nil, err
		}
		leaders := make([]map[string]string, 0, limitInt(len(board.TopGainers), 8))
		referenceSymbols := make([]string, 0, limitInt(len(board.Constituents), 8))
		for _, item := range board.TopGainers {
			leaders = append(leaders, map[string]string{
				"symbol":         item.Symbol,
				"name":           item.Name,
				"change_percent": item.ChangePercent,
				"turnover":       item.Turnover,
			})
			referenceSymbols = append(referenceSymbols, normalizeSymbol(item.Symbol))
		}
		payload := map[string]any{
			"board_type":     board.Board.BoardType,
			"board_code":     board.Board.Code,
			"board_name":     board.Board.Name,
			"change_percent": board.Board.ChangePercent,
			"stock_count":    board.Board.StockCount,
			"rise_count":     board.Board.RiseCount,
			"fall_count":     board.Board.FallCount,
			"leaders":        leaders,
		}
		if len(board.Constituents) > 0 {
			themedCandidates := make([]recommendationCandidate, 0, limitInt(len(board.Constituents), 20))
			for _, item := range board.Constituents {
				symbol := normalizeSymbol(item.Symbol)
				if symbol == "" {
					continue
				}
				candidate := recommendationCandidate{
					symbol:        symbol,
					assetName:     strings.TrimSpace(item.Name),
					assetType:     "stock",
					market:        strings.TrimSpace(item.Market),
					latestPrice:   parseDecimalOrZero(item.LastPrice),
					changePercent: parseDecimalOrZero(item.ChangePercent),
					dataStatus:    marketDataStatusUnavailable,
					sourceSet:     map[string]struct{}{"board_theme": {}},
					sources: []candidateSource{{
						typeName: "board_theme",
						headline: board.Board.Name,
						summary:  fmt.Sprintf("来自偏好板块 %s", board.Board.Name),
					}},
				}
				if strings.TrimSpace(item.LastPrice) != "" && !candidate.latestPrice.IsZero() {
					candidate.dataStatus = marketDataStatusComplete
				}
				themedCandidates = append(themedCandidates, candidate)
			}
			if len(themedCandidates) > 0 {
				themedCandidates = e.service.hydrateRecommendationCandidates(ctx, themedCandidates, true)
				e.candidates = mergeRecommendationCandidates(e.candidates, themedCandidates)
				e.scored = scoreCandidates(e.candidates, e.user, recommendationPreferenceText(e.question, e.messages))
				e.newsItems = buildNewsCandidateItems(e.scored)
			}
		}
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: fmt.Sprintf("已读取板块 %s 的 %d 只成分股摘要", board.Board.Name, len(board.Constituents)), Payload: payload}, &responsedto.ChatStepContextResponse{Stage: "board", Label: "板块成分", Summary: fmt.Sprintf("%s 涨跌幅 %s，成分股 %d 只", board.Board.Name, board.Board.ChangePercent, board.Board.StockCount), ReferenceBoards: []string{board.Board.Name}, ReferenceSymbols: uniqueStrings(referenceSymbols)}, nil
	case "get_stock_indicators_and_news":
		var args struct {
			Symbol string `json:"symbol"`
		}
		if len(call.Args) > 0 {
			_ = json.Unmarshal(call.Args, &args)
		}
		symbol := normalizeSymbol(args.Symbol)
		if symbol == "" {
			return &chatToolResult{
					ToolName: call.ToolName,
					Status:   "incomplete",
					Summary:  "缺少股票代码，需先提供 symbol，或先从候选排序结果中选择一只股票再分析。",
					Payload: map[string]any{
						"requires": []string{"symbol"},
					},
				}, &responsedto.ChatStepContextResponse{
					Stage:   "trend",
					Label:   "指标与新闻",
					Summary: "本次未执行个股指标与新闻分析，等待先确定 symbol。",
				}, nil
		}
		if _, ok := e.insights[symbol]; !ok && len(e.insights) >= recommendationInsightLimit {
			return &chatToolResult{
					ToolName: call.ToolName,
					Status:   "incomplete",
					Summary:  fmt.Sprintf("本轮深度分析已达到上限 %d，请先基于已分析标的继续追问或生成结论。", recommendationInsightLimit),
					Payload: map[string]any{
						"limit":            recommendationInsightLimit,
						"analyzed_symbols": uniqueStrings(mapsKeys(e.insights)),
					},
				}, &responsedto.ChatStepContextResponse{
					Stage:            "trend",
					Label:            "指标与新闻",
					Summary:          fmt.Sprintf("本轮最多深度分析 %d 只股票，当前已完成 %d 只。", recommendationInsightLimit, len(e.insights)),
					ReferenceSymbols: uniqueStrings(mapsKeys(e.insights)),
				}, nil
		}
		insight, err := e.service.buildRecommendationStockInsight(ctx, symbol)
		if err != nil {
			return nil, nil, err
		}
		e.insights[symbol] = insight
		newsItems := buildStockChatNewsItems(insight.News)
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: fmt.Sprintf("已分析 %s 的近期指标与新闻", fallbackString(insight.Detail.Name, symbol)), Payload: insight.Payload()}, &responsedto.ChatStepContextResponse{Stage: "trend", Label: "指标与新闻", Summary: insight.StepSummary(), NewsItems: newsItems, ReferenceSymbols: []string{symbol}, ReferenceBoards: insight.ReferenceBoards()}, nil
	case "rank_recommendation_candidates":
		if len(e.candidates) == 0 {
			return nil, nil, fmt.Errorf("candidates are unavailable")
		}
		e.scored = scoreCandidates(e.candidates, e.user, recommendationPreferenceText(e.question, e.messages))
		e.newsItems = buildNewsCandidateItems(e.scored)
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: fmt.Sprintf("已完成候选排序，保留前 %d 个", len(e.scored)), Payload: convertScoredCandidates(e.scored, nil)}, &responsedto.ChatStepContextResponse{Stage: "ranking", Label: "候选排序", Summary: fmt.Sprintf("候选池 %d 个，排序后保留 %d 个", len(e.candidates), len(e.scored)), CandidateCount: len(e.candidates), FocusSummary: "排序已结合用户偏好、持仓、交易次数和行情变化。"}, nil
	case "get_board_heat_and_constituents":
		if len(e.scored) == 0 {
			if len(e.candidates) == 0 {
				return nil, nil, fmt.Errorf("candidates are unavailable")
			}
			e.scored = scoreCandidates(e.candidates, e.user, recommendationPreferenceText(e.question, e.messages))
		}
		boardNames := make([]string, 0, 8)
		for _, item := range e.scored {
			profile, err := e.service.marketStockService.GetStockProfile(item.candidate.symbol)
			if err != nil {
				continue
			}
			for _, board := range profile.Boards {
				name := strings.TrimSpace(board.Name)
				if name == "" {
					continue
				}
				boardNames = append(boardNames, name)
			}
			if len(boardNames) >= 5 {
				break
			}
		}
		boardNames = uniqueStrings(boardNames)
		if len(boardNames) == 0 {
			return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: "未补充到有效板块信息", Payload: map[string]any{"boards": []string{}}}, &responsedto.ChatStepContextResponse{Stage: "board", Label: "板块热度", Summary: "候选标的未补充到有效板块信息"}, nil
		}
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: fmt.Sprintf("已补充 %d 个关联板块", len(boardNames)), Payload: map[string]any{"boards": boardNames}}, &responsedto.ChatStepContextResponse{Stage: "board", Label: "板块热度", Summary: fmt.Sprintf("已识别关联板块：%s", strings.Join(boardNames, "、")), ReferenceBoards: boardNames}, nil
	case "save_recommendation_report":
		var args struct {
			ReportTitle      string                                  `json:"report_title"`
			SummaryText      string                                  `json:"summary_text"`
			RiskAnalysis     string                                  `json:"risk_analysis"`
			Recommendations  []string                                `json:"recommendations"`
			Items            []aiRecommendationItem                  `json:"items"`
			RawAIOutput      string                                  `json:"raw_ai_output"`
			ToolTraceSummary []responsedto.ChatToolTraceStepResponse `json:"tool_trace_summary"`
		}
		if len(call.Args) > 0 {
			_ = json.Unmarshal(call.Args, &args)
		}
		if strings.TrimSpace(args.SummaryText) == "" {
			return &chatToolResult{
					ToolName: call.ToolName,
					Status:   "incomplete",
					Summary:  "缺少 summary_text，需先整理正式推荐结论后再保存报告。",
					Payload: map[string]any{
						"requires": []string{"summary_text", "risk_analysis", "recommendations", "items", "raw_ai_output"},
					},
				}, &responsedto.ChatStepContextResponse{
					Stage:   "report",
					Label:   "保存报告",
					Summary: "报告尚未保存，等待先生成完整结论与结构化字段。",
				}, nil
		}
		report, err := e.service.persistRecommendationReportWithStructuredData(e.toToolContext(), args)
		if err != nil {
			return nil, nil, err
		}
		e.report = report
		e.reportID = report.ReportID
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: fmt.Sprintf("推荐报告已保存，report_id=%d", report.ReportID), Payload: report}, &responsedto.ChatStepContextResponse{Stage: "report", Label: "保存报告", Summary: fmt.Sprintf("报告 %d 已保存", report.ReportID)}, nil
	default:
		return nil, nil, fmt.Errorf("unsupported tool: %s", call.ToolName)
	}
}

func (e *stockChatExecutor) ExecuteTool(ctx context.Context, call chatToolCall) (*chatToolResult, *responsedto.ChatStepContextResponse, error) {
	switch call.ToolName {
	case "get_stock_quote_and_trend":
		detail, err := e.service.marketStockService.GetStockDetail(e.symbol, true)
		if err != nil {
			return nil, nil, err
		}
		kline, _ := e.service.marketStockService.GetStockKlines(e.symbol, "day", "qfq", 20, true)
		e.detail = detail
		if kline != nil {
			e.kline = kline
		}
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: fmt.Sprintf("已读取 %s 最新行情和走势", detail.Name), Payload: map[string]interface{}{"detail": detail, "trend_summary": buildTrendSummary(detail, e.kline)}}, &responsedto.ChatStepContextResponse{Stage: "trend", Label: "行情走势", Summary: buildTrendSummary(detail, e.kline), ReferenceSymbols: []string{e.symbol}}, nil
	case "get_stock_profile_and_boards":
		profile, err := e.service.marketStockService.GetStockProfile(e.symbol)
		if err != nil {
			return nil, nil, err
		}
		e.profile = profile
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: fmt.Sprintf("已读取 %s 的行业和板块信息", profile.Name), Payload: profile}, &responsedto.ChatStepContextResponse{Stage: "quote", Label: "标的资料", Summary: fmt.Sprintf("行业 %s，概念 %d 个，板块 %d 个", profile.Industry, len(profile.Concepts), len(profile.Boards)), ReferenceSymbols: []string{e.symbol}, ReferenceBoards: boardNamesFromMemberships(profile.Boards)}, nil
	case "get_recent_market_news_candidates":
		name := e.symbol
		if e.detail != nil && strings.TrimSpace(e.detail.Name) != "" {
			name = e.detail.Name
		}
		newsContext, err := e.service.newsService.GetStockNews(ctx, e.symbol, name)
		if err != nil {
			return nil, nil, err
		}
		e.news = newsContext
		items := buildStockChatNewsItems(newsContext)
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: newsContext.Coverage, Payload: items}, &responsedto.ChatStepContextResponse{Stage: "news", Label: "相关新闻", Summary: newsContext.Coverage, NewsItems: items, ReferenceSymbols: []string{e.symbol}}, nil
	default:
		return nil, nil, fmt.Errorf("unsupported tool: %s", call.ToolName)
	}
}

func (e *boardChatExecutor) ExecuteTool(ctx context.Context, call chatToolCall) (*chatToolResult, *responsedto.ChatStepContextResponse, error) {
	switch call.ToolName {
	case "get_board_heat_and_constituents":
		board, err := e.service.marketSnapshotService.GetBoardDetail(ctx, e.boardType, e.code, 20)
		if err != nil {
			return nil, nil, err
		}
		e.board = board
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: fmt.Sprintf("已读取板块 %s 的热度和成分股", board.Board.Name), Payload: board}, &responsedto.ChatStepContextResponse{Stage: "board", Label: "板块热度", Summary: buildBoardTrendSummary(board), ReferenceBoards: []string{board.Board.Name}}, nil
	case "get_recent_market_news_candidates":
		keyword := e.code
		if e.board != nil && strings.TrimSpace(e.board.Board.Name) != "" {
			keyword = e.board.Board.Name
		}
		newsContext, err := e.service.newsService.GetTopicNews(ctx, keyword)
		if err != nil {
			return nil, nil, err
		}
		e.news = newsContext
		items := buildStockChatNewsItems(newsContext)
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: newsContext.Coverage, Payload: items}, &responsedto.ChatStepContextResponse{Stage: "news", Label: "板块新闻", Summary: newsContext.Coverage, NewsItems: items, ReferenceBoards: []string{keyword}}, nil
	case "get_stock_quote_and_trend":
		if e.board == nil || len(e.board.TopGainers) == 0 {
			return nil, nil, fmt.Errorf("board constituents are unavailable")
		}
		leader := e.board.TopGainers[0]
		symbol := normalizeSymbol(leader.Symbol)
		if symbol == "" {
			return nil, nil, fmt.Errorf("leader symbol is unavailable")
		}
		payload := map[string]string{
			"symbol":         symbol,
			"name":           leader.Name,
			"change_percent": leader.ChangePercent,
			"turnover":       leader.Turnover,
		}
		return &chatToolResult{ToolName: call.ToolName, Status: "ok", Summary: fmt.Sprintf("已补充代表个股 %s 的走势", leader.Name), Payload: payload}, &responsedto.ChatStepContextResponse{Stage: "trend", Label: "代表个股", Summary: fmt.Sprintf("代表个股 %s，涨跌幅 %s", leader.Name, leader.ChangePercent), ReferenceSymbols: []string{symbol}}, nil
	default:
		return nil, nil, fmt.Errorf("unsupported tool: %s", call.ToolName)
	}
}

func (e *recommendationChatExecutor) BuildFinalUserPrompt(results []chatToolResult) string {
	return e.service.recommendationFinalUserPrompt(e.toToolContext(), results)
}

func (e *stockChatExecutor) BuildFinalUserPrompt(results []chatToolResult) string {
	return e.service.buildToolDrivenFinalPrompt(e.detail, e.profile, e.kline, e.news, e.messages, e.question, results)
}

func (e *boardChatExecutor) BuildFinalUserPrompt(results []chatToolResult) string {
	return e.service.buildToolDrivenFinalPrompt(e.board, e.news, e.messages, e.question, results)
}

func (e *recommendationChatExecutor) BuildDoneResponse(reply string, trace []responsedto.ChatToolTraceStepResponse, results []chatToolResult) (interface{}, error) {
	toolCtx := e.toToolContext()
	response := e.service.buildRecommendationChatResponse(toolCtx, reply, e.report)
	response.ToolTrace = trace
	response.ToolResults = toToolResultSnapshots(results)
	response.StepContext = e.stepContext
	return response, nil
}

func (e *stockChatExecutor) BuildDoneResponse(reply string, trace []responsedto.ChatToolTraceStepResponse, results []chatToolResult) (interface{}, error) {
	if e.detail == nil {
		detail, err := e.service.marketStockService.GetStockDetail(e.symbol, e.req.RefreshMarket)
		if err != nil {
			return nil, fmt.Errorf("stock detail is unavailable")
		}
		e.detail = detail
		if e.kline == nil || len(e.kline.Items) == 0 {
			if kline, klineErr := e.service.marketStockService.GetStockKlines(e.symbol, "day", "qfq", 20, e.req.RefreshMarket); klineErr == nil && kline != nil {
				e.kline = kline
			}
		}
	}
	response := e.service.buildToolChatResponse(e.question, e.messages, e.detail, e.kline, e.news, reply, trace, results)
	return response, nil
}

func (e *boardChatExecutor) BuildDoneResponse(reply string, trace []responsedto.ChatToolTraceStepResponse, results []chatToolResult) (interface{}, error) {
	if e.board == nil {
		board, err := e.service.marketSnapshotService.GetBoardDetail(context.Background(), e.boardType, e.code, 20)
		if err != nil {
			return nil, fmt.Errorf("board detail is unavailable")
		}
		e.board = board
	}
	response := e.service.buildToolChatResponse(e.question, e.messages, e.boardType, e.code, e.board, e.news, reply, trace, results)
	return response, nil
}

func (e *recommendationChatExecutor) toToolContext() *recommendationToolContext {
	if len(e.scored) == 0 && len(e.candidates) > 0 {
		e.scored = scoreCandidates(e.candidates, e.user, recommendationPreferenceText(e.question, e.messages))
	}
	if len(e.newsItems) == 0 && len(e.scored) > 0 {
		e.newsItems = buildNewsCandidateItems(e.scored)
	}
	return &recommendationToolContext{
		user:       e.user,
		question:   e.question,
		messages:   e.messages,
		candidates: e.candidates,
		scored:     e.scored,
		newsItems:  e.newsItems,
		contextID:  e.req.ContextID,
		reportID:   e.reportID,
		insights:   e.insights,
	}
}

func (e *recommendationChatExecutor) readyForRecommendation() bool {
	if len(e.messages) == 0 {
		return false
	}
	if len(e.scored) == 0 {
		return false
	}
	if !recommendationPreferenceReady(e.user, e.question, e.messages) {
		return false
	}
	required := len(e.scored)
	if required > recommendationShortlistLimit {
		required = recommendationShortlistLimit
	}
	if required == 0 || len(e.insights) < required {
		return false
	}
	validCount := 0
	for i, item := range e.scored {
		if i >= recommendationShortlistLimit {
			break
		}
		insight, ok := e.insights[item.candidate.symbol]
		if !ok {
			continue
		}
		if insight.Detail == nil {
			continue
		}
		if insight.News == nil {
			continue
		}
		validCount++
	}
	return validCount >= required
}

func mapsKeys[V any](input map[string]V) []string {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func candidateSymbols(candidates []recommendationCandidate) []string {
	result := make([]string, 0, len(candidates))
	for _, item := range candidates {
		result = append(result, item.symbol)
	}
	sort.Strings(result)
	if len(result) > 8 {
		result = result[:8]
	}
	return result
}

func uniqueStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	result := make([]string, 0, len(items))
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
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

func (s *recommendationService) candidatePayload(candidates []recommendationCandidate) []responsedto.AnalysisCandidateResponse {
	result := make([]responsedto.AnalysisCandidateResponse, 0, len(candidates))
	for _, candidate := range candidates {
		sources := make([]responsedto.AnalysisCandidateSource, 0, len(candidate.sources))
		for _, source := range candidate.sources {
			sources = append(sources, responsedto.AnalysisCandidateSource{Type: source.typeName})
		}
		result = append(result, responsedto.AnalysisCandidateResponse{
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
	return result
}
