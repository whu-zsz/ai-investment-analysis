package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	responsedto "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"

	"github.com/shopspring/decimal"
)

type structuredRecommendationReportInput struct {
	ReportTitle      string                                  `json:"report_title"`
	SummaryText      string                                  `json:"summary_text"`
	RiskAnalysis     string                                  `json:"risk_analysis"`
	Recommendations  []string                                `json:"recommendations"`
	Items            []aiRecommendationItem                  `json:"items"`
	RawAIOutput      string                                  `json:"raw_ai_output"`
	ToolTraceSummary []responsedto.ChatToolTraceStepResponse `json:"tool_trace_summary"`
}

func toToolResultSnapshots(results []chatToolResult) []responsedto.ChatToolResultSnapshotResponse {
	if len(results) == 0 {
		return nil
	}
	snapshots := make([]responsedto.ChatToolResultSnapshotResponse, 0, len(results))
	for _, item := range results {
		snapshots = append(snapshots, responsedto.ChatToolResultSnapshotResponse{
			ToolName: item.ToolName,
			Status:   item.Status,
			Summary:  item.Summary,
			Payload:  item.Payload,
			Error:    item.Error,
		})
	}
	return snapshots
}

func boardNamesFromMemberships(items []responsedto.StockBoardMembershipResponse) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			continue
		}
		result = append(result, name)
	}
	return result
}

func (s *recommendationService) recommendationFinalPrompt() string {
	return "你是一位专业、克制的 A 股推荐分析助手。你只能根据工具返回的用户画像、持仓关注、板块信息、近期指标和可信新闻回答。必须使用简体中文与 Markdown 输出，不得编造新闻、财务或板块信息。如果用户偏好、投资周期、可接受波动、是否接受外部潜力股等信息仍不充分，或者还没有先分析候选股票的近期指标与新闻，你必须先继续追问，不能直接给推荐结论。只有当信息充分、且你已经决定给出正式推荐时，必须先显式调用 save_recommendation_report 工具保存结构化报告，再在正文中说明报告已生成；如果你没有调用该工具，就不能声称报告已经生成。正式回答按 ## 结论、## 为什么关注、## 风险提醒、## 后续观察 输出。"
}

func (s *recommendationService) recommendationFinalUserPrompt(toolCtx *recommendationToolContext, results []chatToolResult) string {
	payloads := make([]string, 0, len(results))
	for _, item := range results {
		payload, _ := json.Marshal(item.Payload)
		payloads = append(payloads, fmt.Sprintf("- tool=%s\nsummary=%s\npayload=%s", item.ToolName, item.Summary, string(payload)))
	}
	if len(payloads) == 0 {
		payloads = append(payloads, "- 暂无工具结果")
	}
	insightLines := make([]string, 0, len(toolCtx.insights))
	for symbol, insight := range toolCtx.insights {
		indicatorsPayload, _ := json.Marshal(insight.Indicators)
		insightLines = append(insightLines, fmt.Sprintf("- %s %s\ntrend=%s\nindicators=%s\nnews=%s", symbol, fallbackString(insight.Detail.Name, symbol), insight.TrendSummary, string(indicatorsPayload), safeNewsCoverage(insight.News)))
	}
	if len(insightLines) == 0 {
		insightLines = append(insightLines, "- 暂无已分析的个股指标与新闻")
	}
	return fmt.Sprintf("用户投资偏好：%s\n用户风险承受：%s\n用户累计盈亏：%s\n\n历史对话：\n%s\n\n用户本轮问题：\n%s\n\n工具结果：\n%s\n\n已分析的个股指标与新闻：\n%s\n\n输出要求：\\n1. 先判断信息是否足够。\\n2. 如果用户意见仍不足，先提出 2-4 个简洁、具体的问题，不要直接推荐。\\n3. 推荐流程必须是：先从候选池中选出最符合用户需求的 10 个标的，逐一参考近期指标、板块与可信新闻完成调研，再从中收敛出最合适的 5 个。\\n4. 如果要推荐股票，必须明确说明该股票的近期指标和相关新闻分别说明了什么。\\n5. 明确区分已有关注标的和新闻新发现标的。\\n6. 推荐理由必须绑定工具结果。\\n7. 正式推荐阶段应基于已调研完成的 10 个候选，最终只输出最合适的 5 个，不要扩展成大而全清单。\\n8. 如果你准备给出正式推荐结论，必须先调用 save_recommendation_report，并在调用参数里写入 report_title、summary_text、risk_analysis、recommendations、items、raw_ai_output。\\n9. 如果还没有调用 save_recommendation_report，就不要声称报告已生成。\\n10. 不要输出 JSON。", toolCtx.user.InvestmentPreference, toolCtx.user.RiskTolerance, toolCtx.user.TotalProfit.StringFixed(2), conversationHistoryText(toolCtx.messages), toolCtx.question, strings.Join(payloads, "\n\n"), strings.Join(insightLines, "\n\n"))
}

func (s *recommendationService) generateStructuredRecommendationReport(ctx context.Context, toolCtx *recommendationToolContext, reply string, trace []responsedto.ChatToolTraceStepResponse) (structuredRecommendationReportInput, error) {
	lines := make([]string, 0, len(toolCtx.scored))
	for _, item := range toolCtx.scored {
		lines = append(lines, fmt.Sprintf("- %s %s：动作=%s，评分=%s，最新价=%s，涨跌幅=%s，理由=%s，风险=%s", item.candidate.symbol, fallbackString(item.candidate.assetName, item.candidate.symbol), item.action, item.score.StringFixed(2), decimalToString(item.candidate.latestPrice, 4), decimalToString(item.candidate.changePercent, 2), item.matchReason, item.riskNote))
	}
	tracePayload, _ := json.Marshal(trace)
	prompt := "你是一位严格的推荐报告结构化助手。根据已有 Markdown 结论和候选明细，输出 JSON，不要输出任何多余文本。"
	userPrompt := fmt.Sprintf("用户问题：%s\n\n最终回答：\n%s\n\n候选明细：\n%s\n\n步骤轨迹：%s\n\n输出 JSON：{\"report_title\":string,\"summary_text\":string,\"risk_analysis\":string,\"recommendations\":[string],\"items\":[{\"symbol\":string,\"action\":string,\"match_reason\":string,\"risk_note\":string}],\"raw_ai_output\":string}", toolCtx.question, reply, strings.Join(lines, "\n"), string(tracePayload))
	content, err := s.llmProvider.GetContent(ctx, prompt, userPrompt)
	if err != nil {
		return structuredRecommendationReportInput{}, err
	}
	content = extractJSONObject(content)
	var parsed structuredRecommendationReportInput
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return structuredRecommendationReportInput{}, err
	}
	if strings.TrimSpace(parsed.SummaryText) == "" {
		parsed.SummaryText = reply
	}
	if strings.TrimSpace(parsed.RawAIOutput) == "" {
		parsed.RawAIOutput = reply
	}
	parsed.ToolTraceSummary = trace
	return parsed, nil
}

func (s *recommendationService) persistRecommendationReportWithStructuredData(toolCtx *recommendationToolContext, input structuredRecommendationReportInput) (*responsedto.AnalysisRecommendationsResponse, error) {
	if s.analysisReportRepo == nil || s.analysisReportItemRepo == nil {
		return nil, fmt.Errorf("analysis report repository is unavailable")
	}
	finalScored := finalRecommendationCandidates(toolCtx.scored)
	candidates := convertScoredCandidates(finalScored, input.Items)
	items := make([]model.AnalysisReportItem, 0, len(finalScored))
	recommendations := append([]string(nil), input.Recommendations...)
	if len(recommendations) == 0 {
		for _, item := range finalScored {
			recommendations = append(recommendations, fmt.Sprintf("%s：%s", fallbackString(item.candidate.assetName, item.candidate.symbol), item.action))
		}
	}
	if len(recommendations) > recommendationResultLimit {
		recommendations = recommendations[:recommendationResultLimit]
	}
	for _, item := range finalScored {
		rawSourceTags, _ := json.Marshal(recommendationSourceTags(item.candidate))
		items = append(items, model.AnalysisReportItem{
			UserID:               toolCtx.user.ID,
			Symbol:               item.candidate.symbol,
			AssetName:            fallbackString(item.candidate.assetName, item.candidate.symbol),
			Market:               item.candidate.market,
			TradeCount:           item.candidate.tradeCount,
			BuyCount:             0,
			SellCount:            0,
			AnalysisText:         item.matchReason,
			Recommendation:       item.action,
			KeyPoints:            stringPointerIfNotEmpty(string(rawSourceTags)),
			RawAIOutput:          stringPointerIfNotEmpty(item.riskNote),
			LatestPrice:          item.candidate.latestPrice,
			PeriodPriceChangePct: item.candidate.changePercent,
			MarketDataStatus:     item.candidate.dataStatus,
			RiskLevel:            normalizeRiskLevel(toolCtx.user.RiskTolerance),
			InvestmentStyle:      stringPointerIfNotEmpty(toolCtx.user.InvestmentPreference),
		})
	}
	now := time.Now()
	tracePayload, _ := json.Marshal(input.ToolTraceSummary)
	recommendationsJSON := marshalJSONArray(recommendations)
	reportTitle := strings.TrimSpace(input.ReportTitle)
	if reportTitle == "" {
		reportTitle = fmt.Sprintf("AI 推荐报告 (%s)", now.Format("2006-01-02 15:04"))
	}
	report := &model.AnalysisReport{
		UserID:              toolCtx.user.ID,
		ReportType:          "recommendation",
		ReportTitle:         reportTitle,
		AnalysisPeriodStart: now,
		AnalysisPeriodEnd:   now,
		SymbolsCount:        len(items),
		TotalInvestment:     decimal.Zero,
		TotalProfit:         toolCtx.user.TotalProfit,
		ProfitRate:          decimal.Zero,
		RiskLevel:           normalizeRiskLevel(toolCtx.user.RiskTolerance),
		MarketDataStatus:    summarizeRecommendationDataStatus(finalScored),
		InvestmentStyle:     stringPointerIfNotEmpty(toolCtx.user.InvestmentPreference),
		SummaryText:         fallbackString(strings.TrimSpace(input.SummaryText), input.RawAIOutput),
		RiskAnalysis:        stringPointerIfNotEmpty(fallbackString(strings.TrimSpace(input.RiskAnalysis), "推荐报告基于用户偏好、历史关注和近期新闻热点自动生成，请结合自身仓位与风险承受能力判断。")),
		PatternInsights:     stringPointerIfNotEmpty(string(tracePayload)),
		PredictionText:      stringPointerIfNotEmpty("推荐结论需结合后续新闻兑现和市场风格切换持续验证。"),
		Recommendations:     stringPointerIfNotEmpty(recommendationsJSON),
		RawAIOutput:         stringPointerIfNotEmpty(fallbackString(strings.TrimSpace(input.RawAIOutput), input.SummaryText)),
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
		SummaryText: fallbackString(strings.TrimSpace(input.SummaryText), input.RawAIOutput),
		Candidates:  candidates,
	}, nil
}

func (s *stockChatService) buildToolDrivenFinalPrompt(detail *responsedto.MarketStockDetailResponse, profile *responsedto.StockProfileResponse, kline *responsedto.MarketStockKlineResponse, newsContext *StockNewsContext, messages []stockChatMessage, question string, results []chatToolResult) string {
	if detail == nil {
		return question
	}
	newsLines := make([]string, 0)
	if newsContext != nil {
		for _, item := range newsContext.Items {
			publishedAt := ""
			if !item.PublishedAt.IsZero() {
				publishedAt = item.PublishedAt.Format("2006-01-02 15:04")
			}
			newsLines = append(newsLines, fmt.Sprintf("- 标题：%s；来源：%s；时间：%s；摘要：%s；链接：%s", item.Title, item.Source, publishedAt, item.Summary, item.URL))
		}
	}
	boardText := "暂无板块补充"
	if profile != nil && len(profile.Boards) > 0 {
		boardText = strings.Join(boardNamesFromMemberships(profile.Boards), "、")
	}
	return fmt.Sprintf("当前股票：%s %s\n市场：%s\n最新行情：最新价=%s，涨跌幅=%s%%，最高=%s，最低=%s，成交量=%s，成交额=%s，数据源=%s，时间=%s\n趋势摘要：%s\n行业：%s\n概念：%s\n板块：%s\n新闻覆盖：%s\n新闻摘要：%s\n相关新闻：\n%s\n\n历史对话：\n%s\n\n用户本轮问题：%s", detail.Symbol, detail.Name, detail.Market, detail.LastPrice, detail.ChangePercent, detail.HighPrice, detail.LowPrice, detail.Volume, detail.Turnover, detail.Source, detail.FetchedAt, buildTrendSummary(detail, kline), safeProfileIndustry(profile), strings.Join(safeProfileConcepts(profile), "、"), boardText, safeNewsCoverage(newsContext), safeNewsSummary(newsContext), strings.Join(newsLines, "\n"), conversationHistoryText(messages), question)
}

func (s *stockChatService) buildToolChatResponse(question string, messages []stockChatMessage, detail *responsedto.MarketStockDetailResponse, kline *responsedto.MarketStockKlineResponse, newsContext *StockNewsContext, reply string, trace []responsedto.ChatToolTraceStepResponse, results []chatToolResult) *responsedto.StockChatResponse {
	responseMessages := make([]responsedto.StockChatMessageResponse, 0, len(messages)+2)
	for _, item := range messages {
		responseMessages = append(responseMessages, responsedto.StockChatMessageResponse{Role: item.Role, Content: item.Content})
	}
	responseMessages = append(responseMessages, responsedto.StockChatMessageResponse{Role: "user", Content: question}, responsedto.StockChatMessageResponse{Role: "assistant", Content: reply})
	return &responsedto.StockChatResponse{
		Symbol:       detail.Symbol,
		AssetName:    detail.Name,
		Market:       detail.Market,
		Reply:        reply,
		AIModel:      s.llmProvider.ModelName(),
		GeneratedAt:  time.Now().Format("2006-01-02 15:04:05"),
		NewsStatus:   safeNewsStatus(newsContext),
		NewsSummary:  safeNewsSummary(newsContext),
		NewsCoverage: safeNewsCoverage(newsContext),
		NewsItems:    buildStockChatNewsItems(newsContext),
		Snapshot: responsedto.StockChatSnapshotResponse{
			LastPrice:     detail.LastPrice,
			ChangePercent: detail.ChangePercent,
			HighPrice:     detail.HighPrice,
			LowPrice:      detail.LowPrice,
			Volume:        detail.Volume,
			Turnover:      detail.Turnover,
			Source:        detail.Source,
			FetchedAt:     detail.FetchedAt,
			Period:        "day",
			TrendSummary:  buildTrendSummary(detail, kline),
		},
		Messages:  responseMessages,
		ToolTrace: trace,
	}
}

func (s *boardChatService) buildToolDrivenFinalPrompt(detail *responsedto.MarketBoardDetailResponse, newsContext *StockNewsContext, messages []stockChatMessage, question string, results []chatToolResult) string {
	if detail == nil {
		return question
	}
	newsLines := make([]string, 0)
	if newsContext != nil {
		for _, item := range newsContext.Items {
			publishedAt := ""
			if !item.PublishedAt.IsZero() {
				publishedAt = item.PublishedAt.Format("2006-01-02 15:04")
			}
			newsLines = append(newsLines, fmt.Sprintf("- 标题：%s；来源：%s；时间：%s；摘要：%s；链接：%s", item.Title, item.Source, publishedAt, item.Summary, item.URL))
		}
	}
	return fmt.Sprintf("当前板块：%s (%s/%s)\n板块行情：涨跌幅=%s%%，成交额=%s，成分股数量=%d，上涨=%d，下跌=%d，平盘=%d，快照时间=%s\n领涨成分：%s\n拖累成分：%s\n成交额前排：%s\n新闻覆盖：%s\n新闻摘要：%s\n相关新闻：\n%s\n\n历史对话：\n%s\n\n用户本轮问题：%s", detail.Board.Name, detail.Board.BoardType, detail.Board.Code, detail.Board.ChangePercent, detail.Board.Turnover, detail.Board.StockCount, detail.Board.RiseCount, detail.Board.FallCount, detail.Board.FlatCount, detail.SnapshotTime, summarizeBoardConstituents(detail.TopGainers), summarizeBoardConstituents(detail.TopLosers), summarizeBoardConstituents(detail.TopTurnover), safeNewsCoverage(newsContext), safeNewsSummary(newsContext), strings.Join(newsLines, "\n"), conversationHistoryText(messages), question)
}

func (s *boardChatService) buildToolChatResponse(question string, messages []stockChatMessage, boardType, code string, detail *responsedto.MarketBoardDetailResponse, newsContext *StockNewsContext, reply string, trace []responsedto.ChatToolTraceStepResponse, results []chatToolResult) *responsedto.BoardChatResponse {
	responseMessages := make([]responsedto.StockChatMessageResponse, 0, len(messages)+2)
	for _, item := range messages {
		responseMessages = append(responseMessages, responsedto.StockChatMessageResponse{Role: item.Role, Content: item.Content})
	}
	responseMessages = append(responseMessages, responsedto.StockChatMessageResponse{Role: "user", Content: question}, responsedto.StockChatMessageResponse{Role: "assistant", Content: reply})
	return &responsedto.BoardChatResponse{
		BoardType:    boardType,
		Code:         code,
		AssetName:    detail.Board.Name,
		Market:       "board",
		Reply:        reply,
		AIModel:      s.llmProvider.ModelName(),
		GeneratedAt:  time.Now().Format("2006-01-02 15:04:05"),
		NewsStatus:   safeNewsStatus(newsContext),
		NewsSummary:  safeNewsSummary(newsContext),
		NewsCoverage: safeNewsCoverage(newsContext),
		NewsItems:    buildStockChatNewsItems(newsContext),
		Snapshot: responsedto.StockChatSnapshotResponse{
			LastPrice:     "0",
			ChangePercent: detail.Board.ChangePercent,
			Volume:        detail.Board.Volume,
			Turnover:      detail.Board.Turnover,
			Source:        detail.Source,
			FetchedAt:     detail.RefreshedAt,
			Period:        "board",
			TrendSummary:  buildBoardTrendSummary(detail),
		},
		Messages:    responseMessages,
		ToolTrace:   trace,
		ToolResults: toToolResultSnapshots(results),
	}
}

func safeProfileIndustry(profile *responsedto.StockProfileResponse) string {
	if profile == nil {
		return ""
	}
	return profile.Industry
}

func safeProfileConcepts(profile *responsedto.StockProfileResponse) []string {
	if profile == nil {
		return nil
	}
	return profile.Concepts
}

func safeNewsStatus(newsContext *StockNewsContext) string {
	if newsContext == nil {
		return "unavailable"
	}
	return newsContext.Status
}

func safeNewsCoverage(newsContext *StockNewsContext) string {
	if newsContext == nil {
		return "新闻未覆盖"
	}
	return newsContext.Coverage
}

func safeNewsSummary(newsContext *StockNewsContext) string {
	if newsContext == nil {
		return "暂无新闻摘要"
	}
	return newsContext.Summary
}

func recommendationBoardCatalog(boardType string, items []responsedto.MarketBoardItemResponse) []map[string]string {
	result := make([]map[string]string, 0, len(items))
	for _, item := range items {
		name := strings.TrimSpace(item.Name)
		code := strings.TrimSpace(item.Code)
		if name == "" || code == "" {
			continue
		}
		result = append(result, map[string]string{
			"board_type":     boardType,
			"code":           code,
			"name":           name,
			"change_percent": strings.TrimSpace(item.ChangePercent),
		})
	}
	return result
}

func recommendationPreferenceReady(user *model.User, question string, messages []stockChatMessage) bool {
	text := strings.ToLower(strings.TrimSpace(question + "\n" + conversationHistoryText(messages)))
	if strings.TrimSpace(user.InvestmentPreference) != "" && strings.TrimSpace(user.RiskTolerance) != "" {
		if containsAny(text, "短线", "中线", "长线", "一周", "一个月", "三个月", "半年", "一年", "高波动", "低波动", "成长", "价值", "外部潜力股", "新闻热点") {
			return true
		}
	}
	return containsAny(text, "风险", "周期", "偏好", "波动", "持有多久", "接受", "仓位")
}

func containsAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(text, strings.ToLower(strings.TrimSpace(needle))) {
			return true
		}
	}
	return false
}

func (s *recommendationService) buildRecommendationStockInsight(ctx context.Context, symbol string) (recommendationStockInsight, error) {
	detail, err := s.marketStockService.GetStockDetail(symbol, true)
	if err != nil {
		return recommendationStockInsight{}, err
	}
	profile, _ := s.marketStockService.GetStockProfile(symbol)
	kline, _ := s.marketStockService.GetStockKlines(symbol, "day", "qfq", 20, true)
	newsContext, err := s.newsService.GetStockNews(ctx, symbol, fallbackString(detail.Name, symbol))
	if err != nil {
		return recommendationStockInsight{}, err
	}
	indicators, trendSummary := summarizeRecommendationIndicators(detail, kline)
	return recommendationStockInsight{
		Symbol:       symbol,
		Detail:       detail,
		Profile:      profile,
		Kline:        kline,
		News:         newsContext,
		Indicators:   indicators,
		TrendSummary: trendSummary,
	}, nil
}

func summarizeRecommendationIndicators(detail *responsedto.MarketStockDetailResponse, kline *responsedto.MarketStockKlineResponse) (map[string]string, string) {
	result := map[string]string{
		"last_price":     strings.TrimSpace(detail.LastPrice),
		"change_percent": strings.TrimSpace(detail.ChangePercent),
		"turnover_rate":  strings.TrimSpace(detail.TurnoverRate),
		"volume_ratio":   strings.TrimSpace(detail.VolumeRatio),
		"amplitude":      strings.TrimSpace(detail.Amplitude),
	}
	if kline == nil || len(kline.Items) == 0 {
		return result, fmt.Sprintf("%s 当前涨跌幅 %s，暂无足够 K 线补充区间趋势。", fallbackString(detail.Name, detail.Symbol), fallbackString(detail.ChangePercent, "0"))
	}
	closes := make([]float64, 0, len(kline.Items))
	highs := make([]float64, 0, len(kline.Items))
	lows := make([]float64, 0, len(kline.Items))
	for _, item := range kline.Items {
		closes = append(closes, parseFloat(item.ClosePrice))
		highs = append(highs, parseFloat(item.HighPrice))
		lows = append(lows, parseFloat(item.LowPrice))
	}
	lastClose := closes[len(closes)-1]
	prevClose := closes[maxInt(0, len(closes)-6)]
	windowLow := minFloatSlice(lows)
	windowHigh := maxFloatSlice(highs)
	rangePct := 0.0
	if windowHigh > windowLow {
		rangePct = (lastClose - windowLow) / (windowHigh - windowLow) * 100
	}
	change5d := 0.0
	if prevClose > 0 {
		change5d = (lastClose - prevClose) / prevClose * 100
	}
	result["change_5d"] = formatFloat(change5d)
	result["ma5"] = formatFloat(averageLast(closes, 5))
	result["ma10"] = formatFloat(averageLast(closes, 10))
	result["ma20"] = formatFloat(averageLast(closes, 20))
	result["range_position_pct"] = formatFloat(rangePct)
	trend := fmt.Sprintf("%s 近 5 日涨跌 %s%%，MA5 %s、MA10 %s、MA20 %s，区间位置 %.0f%%。", fallbackString(detail.Name, detail.Symbol), formatFloat(change5d), result["ma5"], result["ma10"], result["ma20"], rangePct)
	return result, trend
}

func (i recommendationStockInsight) Payload() map[string]any {
	payload := map[string]any{
		"symbol":              i.Symbol,
		"name":                fallbackString(i.Detail.Name, i.Symbol),
		"trend_summary":       i.TrendSummary,
		"indicators":          i.Indicators,
		"news_status":         safeNewsStatus(i.News),
		"news_coverage":       safeNewsCoverage(i.News),
		"news_count":          safeNewsCount(i.News),
		"top_headlines":       recommendationHeadlineSummaries(i.News, 3),
		"news_digest":         recommendationHeadlineDigests(i.News, 3, 88),
		"news_signal_summary": recommendationNewsSignalText(i.News),
	}
	if i.Profile != nil {
		payload["industry"] = safeProfileIndustry(i.Profile)
		payload["concepts"] = safeProfileConcepts(i.Profile)
		payload["boards"] = boardNamesFromMemberships(i.Profile.Boards)
	}
	return payload
}

func recommendationHeadlineSummaries(newsContext *StockNewsContext, limit int) []map[string]string {
	if newsContext == nil || len(newsContext.Items) == 0 || limit <= 0 {
		return nil
	}
	result := make([]map[string]string, 0, limit)
	for _, item := range newsContext.Items {
		if len(result) >= limit {
			break
		}
		title := strings.TrimSpace(item.Title)
		if title == "" {
			continue
		}
		entry := map[string]string{
			"title": title,
			"source": strings.TrimSpace(item.Source),
		}
		if !item.PublishedAt.IsZero() {
			entry["published_at"] = item.PublishedAt.Format("2006-01-02 15:04:05")
		}
		result = append(result, entry)
	}
	return result
}

func recommendationHeadlineDigests(newsContext *StockNewsContext, limit, digestLimit int) []map[string]string {
	if newsContext == nil || len(newsContext.Items) == 0 || limit <= 0 {
		return nil
	}
	result := make([]map[string]string, 0, limit)
	for _, item := range newsContext.Items {
		if len(result) >= limit {
			break
		}
		title := strings.TrimSpace(item.Title)
		if title == "" {
			continue
		}
		digest := compactNewsDigest(strings.TrimSpace(item.Summary), digestLimit)
		if digest == "" {
			digest = compactNewsDigest(title, digestLimit)
		}
		entry := map[string]string{
			"title": title,
			"digest": digest,
			"source": strings.TrimSpace(item.Source),
		}
		if !item.PublishedAt.IsZero() {
			entry["published_at"] = item.PublishedAt.Format("2006-01-02 15:04:05")
		}
		result = append(result, entry)
	}
	return result
}

func recommendationNewsSignalText(newsContext *StockNewsContext) string {
	if newsContext == nil {
		return "暂无可用新闻信号"
	}
	summary := compactNewsDigest(strings.TrimSpace(newsContext.Summary), 120)
	if summary != "" {
		return summary
	}
	return safeNewsCoverage(newsContext)
}

func safeNewsCount(newsContext *StockNewsContext) int {
	if newsContext == nil {
		return 0
	}
	return len(newsContext.Items)
}

func (i recommendationStockInsight) StepSummary() string {
	return fmt.Sprintf("%s；新闻覆盖：%s", i.TrendSummary, safeNewsCoverage(i.News))
}

func (i recommendationStockInsight) ReferenceBoards() []string {
	if i.Profile == nil {
		return nil
	}
	return uniqueStrings(append([]string{strings.TrimSpace(i.Profile.Industry)}, boardNamesFromMemberships(i.Profile.Boards)...))
}

func parseFloat(value string) float64 {
	parsed, _ := decimal.NewFromString(strings.TrimSpace(value))
	f, _ := parsed.Float64()
	return f
}

func averageLast(values []float64, n int) float64 {
	if len(values) == 0 {
		return 0
	}
	if n <= 0 || n > len(values) {
		n = len(values)
	}
	sum := 0.0
	for _, item := range values[len(values)-n:] {
		sum += item
	}
	return sum / float64(n)
}

func minFloatSlice(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	result := values[0]
	for _, item := range values[1:] {
		result = math.Min(result, item)
	}
	return result
}

func maxFloatSlice(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	result := values[0]
	for _, item := range values[1:] {
		result = math.Max(result, item)
	}
	return result
}

func formatFloat(value float64) string {
	return decimal.NewFromFloat(value).StringFixed(2)
}

func limitInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
