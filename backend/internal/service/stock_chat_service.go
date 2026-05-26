package service

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	requestdto "stock-analysis-backend/internal/dto/request"
	responsedto "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/repository"
	"stock-analysis-backend/pkg/llm"

	gozap "go.uber.org/zap"
)

type StockChatService interface {
	Chat(ctx context.Context, userID uint64, req *requestdto.StockChatRequest) (*responsedto.StockChatResponse, error)
	ChatStream(ctx context.Context, userID uint64, req *requestdto.StockChatRequest, emit func(responsedto.StockChatStreamEvent) error) error
}

type stockChatService struct {
	marketStockService MarketStockService
	newsService        NewsService
	llmProvider        llm.Provider
	logger             *gozap.Logger
	contextService     *chatContextService
}

type stockChatContext struct {
	contextID   uint64
	symbol      string
	question    string
	messages    []stockChatMessage
	detail      *responsedto.MarketStockDetailResponse
	kline       *responsedto.MarketStockKlineResponse
	newsContext *StockNewsContext
	prompt      string
}

func NewStockChatService(marketStockService MarketStockService, newsService NewsService, llmProvider llm.Provider, logger *gozap.Logger, contextRepo repository.ChatContextRepository) StockChatService {
	return &stockChatService{
		marketStockService: marketStockService,
		newsService:        newsService,
		llmProvider:        llmProvider,
		logger:             logger,
		contextService:     newChatContextService(contextRepo),
	}
}

func (s *stockChatService) Chat(ctx context.Context, userID uint64, req *requestdto.StockChatRequest) (*responsedto.StockChatResponse, error) {
	req = s.hydrateStockChatRequest(userID, req)
	executor, err := newStockChatExecutor(s, req)
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
	response, ok := result.(*responsedto.StockChatResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected stock chat response type")
	}
	if err := s.persistResponse(userID, response, req); err != nil && s.logger != nil {
		s.logger.Warn("failed to persist stock chat context", gozap.Error(err))
	}
	return response, nil
}

func (s *stockChatService) ChatStream(ctx context.Context, userID uint64, req *requestdto.StockChatRequest, emit func(responsedto.StockChatStreamEvent) error) error {
	req = s.hydrateStockChatRequest(userID, req)
	if emit == nil {
		return fmt.Errorf("stream emitter is required")
	}
	executor, err := newStockChatExecutor(s, req)
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
	if response, ok := result.(*responsedto.StockChatResponse); ok {
		if persistErr := s.persistResponse(userID, response, req); persistErr != nil && s.logger != nil {
			s.logger.Warn("failed to persist stock chat context", gozap.Error(persistErr))
		}
	}
	return emit(responsedto.StockChatStreamEvent{Type: "done", Stage: "done", Message: "回答生成完成", Data: result})
}

func (s *stockChatService) prepareChatContext(ctx context.Context, userID uint64, req *requestdto.StockChatRequest, emit func(responsedto.StockChatStreamEvent) error) (*stockChatContext, error) {
	_ = userID
	symbol := normalizeSymbol(req.Symbol)
	if symbol == "" {
		return nil, fmt.Errorf("symbol is required")
	}
	question := strings.TrimSpace(req.Question)
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}
	if s.marketStockService == nil {
		return nil, fmt.Errorf("market stock service is unavailable")
	}
	if s.newsService == nil {
		return nil, fmt.Errorf("news service is unavailable")
	}
	if s.llmProvider == nil {
		return nil, fmt.Errorf("llm provider is unavailable")
	}

	if emit != nil {
		if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: "market", Message: "正在获取最新行情和近 20 个交易日走势"}); err != nil {
			return nil, err
		}
	}
	detail, err := s.marketStockService.GetStockDetail(symbol, req.RefreshMarket)
	if err != nil {
		return nil, err
	}
	kline, klineErr := s.marketStockService.GetStockKlines(symbol, "day", "qfq", 20, req.RefreshMarket)
	if klineErr != nil && s.logger != nil {
		s.logger.Warn("stock chat failed to load kline", gozap.String("symbol", symbol), gozap.Error(klineErr))
	}

	if emit != nil {
		if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: "news", Message: "正在拉取并筛选最近相关新闻"}); err != nil {
			return nil, err
		}
	}
	newsContext, err := s.newsService.GetStockNews(ctx, symbol, detail.Name)
	if err != nil {
		return nil, err
	}
	if emit != nil {
		if err := emit(responsedto.StockChatStreamEvent{
			Type:    "context",
			Stage:   "news",
			Message: newsContext.Coverage,
			Data:    buildStockChatNewsItems(newsContext),
		}); err != nil {
			return nil, err
		}
		if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: "prompt", Message: "正在把你的问题、行情和新闻组装成分析上下文"}); err != nil {
			return nil, err
		}
	}

	messages := normalizeStockChatMessages(req.Messages)
	return &stockChatContext{
		contextID:   req.ContextID,
		symbol:      symbol,
		question:    question,
		messages:    messages,
		detail:      detail,
		kline:       kline,
		newsContext: newsContext,
		prompt:      s.buildChatPrompt(detail, kline, newsContext, messages, question),
	}, nil
}

var stockChatSystemPrompt = `你是一位专业、克制的 A 股个股分析助手。
必须只依据输入中提供的真实行情、趋势数据和新闻上下文回答，不得编造新闻、公告、研报、价格或财务数据。
回答必须使用简体中文，优先回应用户本轮问题，语气自然、明确，不要写成制式模板。
请使用 Markdown 输出，并尽量遵循下面的表达方式：
- 开头先直接回答用户问题，用 1-2 句话给出判断。
- 然后给出一个“## 重点提示”小节，列出 2-4 条最关键的观察，不要重复。
- 如有必要，再补充“## 新闻影响”、“## 走势观察”、“## 风险”、“## 操作建议”等小节。
- 不是每次都必须把所有小节写满；重点是先回答问题，再补充依据。
- 如果新闻覆盖不完整，必须明确说明信息可能不足。
- 不要输出 JSON，不要写“1. 2. 3.”式模板，不要堆砌空话。
- 如果提到走势判断，尽量结合输入里的价格、涨跌幅、区间趋势和量价信息来支撑。`

type stockChatMessage struct {
	Role    string
	Content string
}

func normalizeStockChatMessages(messages []requestdto.StockChatMessageRequest) []stockChatMessage {
	result := make([]stockChatMessage, 0, len(messages))
	for _, item := range messages {
		role := strings.ToLower(strings.TrimSpace(item.Role))
		if role != "user" && role != "assistant" {
			continue
		}
		content := normalizeNarrativeText(item.Content)
		if content == "" {
			continue
		}
		result = append(result, stockChatMessage{Role: role, Content: content})
	}
	if len(result) > 8 {
		result = result[len(result)-8:]
	}
	return result
}

func (s *stockChatService) buildChatPrompt(detail *responsedto.MarketStockDetailResponse, kline *responsedto.MarketStockKlineResponse, newsContext *StockNewsContext, messages []stockChatMessage, question string) string {
	newsLines := make([]string, 0, len(newsContext.Items))
	for _, item := range newsContext.Items {
		publishedAt := ""
		if !item.PublishedAt.IsZero() {
			publishedAt = item.PublishedAt.Format("2006-01-02 15:04")
		}
		newsLines = append(newsLines, fmt.Sprintf("- 标题：%s；来源：%s；时间：%s；摘要：%s；链接：%s", item.Title, item.Source, publishedAt, item.Summary, item.URL))
	}
	conversation := make([]string, 0, len(messages))
	for _, item := range messages {
		conversation = append(conversation, fmt.Sprintf("%s: %s", item.Role, item.Content))
	}
	return fmt.Sprintf(`当前股票：%s %s
市场：%s
最新行情：最新价=%s，涨跌幅=%s%%，最高=%s，最低=%s，成交量=%s，成交额=%s，数据源=%s，行情时间=%s
趋势摘要：%s
新闻状态：%s
新闻覆盖：%s
新闻摘要：%s
相关新闻：
%s

历史对话：
%s

用户本轮问题：%s

回答要求：
1. 优先回答用户问题，不要泛泛复述所有材料。
2. 明确区分新闻面与价格走势面。
3. 如果新闻与价格信号冲突，要指出冲突点。
4. 重点提示必须具体、可落地。`, detail.Symbol, detail.Name, detail.Market, detail.LastPrice, detail.ChangePercent, detail.HighPrice, detail.LowPrice, detail.Volume, detail.Turnover, detail.Source, detail.FetchedAt, buildTrendSummary(detail, kline), newsContext.Status, newsContext.Coverage, newsContext.Summary, strings.Join(newsLines, "\n"), strings.Join(conversation, "\n"), question)
}

func (s *stockChatService) buildResponse(chatCtx *stockChatContext, reply string) *responsedto.StockChatResponse {
	responseMessages := make([]responsedto.StockChatMessageResponse, 0, len(chatCtx.messages)+2)
	for _, item := range chatCtx.messages {
		responseMessages = append(responseMessages, responsedto.StockChatMessageResponse{Role: item.Role, Content: item.Content})
	}
	responseMessages = append(responseMessages,
		responsedto.StockChatMessageResponse{Role: "user", Content: chatCtx.question},
		responsedto.StockChatMessageResponse{Role: "assistant", Content: reply},
	)

	return &responsedto.StockChatResponse{
		ContextID:    chatCtx.contextID,
		Symbol:       chatCtx.symbol,
		AssetName:    chatCtx.detail.Name,
		Market:       chatCtx.detail.Market,
		Reply:        reply,
		AIModel:      s.llmProvider.ModelName(),
		GeneratedAt:  time.Now().Format("2006-01-02 15:04:05"),
		NewsStatus:   chatCtx.newsContext.Status,
		NewsSummary:  chatCtx.newsContext.Summary,
		NewsCoverage: chatCtx.newsContext.Coverage,
		NewsItems:    buildStockChatNewsItems(chatCtx.newsContext),
		Snapshot: responsedto.StockChatSnapshotResponse{
			LastPrice:     chatCtx.detail.LastPrice,
			ChangePercent: chatCtx.detail.ChangePercent,
			HighPrice:     chatCtx.detail.HighPrice,
			LowPrice:      chatCtx.detail.LowPrice,
			Volume:        chatCtx.detail.Volume,
			Turnover:      chatCtx.detail.Turnover,
			Source:        chatCtx.detail.Source,
			FetchedAt:     chatCtx.detail.FetchedAt,
			Period:        "day",
			TrendSummary:  buildTrendSummary(chatCtx.detail, chatCtx.kline),
		},
		Messages: responseMessages,
	}
}

func (s *stockChatService) hydrateStockChatRequest(userID uint64, req *requestdto.StockChatRequest) *requestdto.StockChatRequest {
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

func (s *stockChatService) persistResponse(userID uint64, resp *responsedto.StockChatResponse, req *requestdto.StockChatRequest) error {
	if s.contextService == nil || resp == nil || req == nil {
		return nil
	}
	metadata := persistedChatContextMetadata{
		ToolTrace:   resp.ToolTrace,
		ToolResults: resp.ToolResults,
		NewsItems:   resp.NewsItems,
		GeneratedAt: resp.GeneratedAt,
	}
	contextID, err := s.contextService.saveContext(
		userID,
		"stock",
		normalizeSymbol(req.Symbol),
		fallbackString(resp.AssetName, normalizeSymbol(req.Symbol)),
		req.ContextID,
		0,
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

func buildStockChatNewsItems(newsContext *StockNewsContext) []responsedto.StockChatNewsItemResponse {
	if newsContext == nil {
		return nil
	}
	newsItems := make([]responsedto.StockChatNewsItemResponse, 0, len(newsContext.Items))
	for _, item := range newsContext.Items {
		publishedAt := ""
		if !item.PublishedAt.IsZero() {
			publishedAt = item.PublishedAt.Format("2006-01-02 15:04:05")
		}
		newsItems = append(newsItems, responsedto.StockChatNewsItemResponse{
			Title:       item.Title,
			Summary:     item.Summary,
			Source:      item.Source,
			URL:         item.URL,
			PublishedAt: publishedAt,
			Provider:    item.Provider,
		})
	}
	return newsItems
}

func buildTrendSummary(detail *responsedto.MarketStockDetailResponse, kline *responsedto.MarketStockKlineResponse) string {
	if kline == nil || len(kline.Items) == 0 {
		return fmt.Sprintf("当前仅有快照数据，最新价 %s，日内涨跌幅 %s%%。", detail.LastPrice, detail.ChangePercent)
	}
	first := kline.Items[0]
	last := kline.Items[len(kline.Items)-1]
	firstClose := parseStockFloat(first.ClosePrice)
	lastClose := parseStockFloat(last.ClosePrice)
	highPrice := parseStockFloat(detail.HighPrice)
	lowPrice := parseStockFloat(detail.LowPrice)
	changePct := 0.0
	if firstClose != 0 {
		changePct = (lastClose - firstClose) / firstClose * 100
	}
	return fmt.Sprintf("近 %d 个交易周期收盘价从 %s 变动到 %s，区间变化 %.2f%%；当前日内高低区间 %s - %s，最新日涨跌幅 %s%%。", len(kline.Items), first.ClosePrice, last.ClosePrice, changePct, formatStockFloat(lowPrice), formatStockFloat(highPrice), detail.ChangePercent)
}

func parseStockFloat(value string) float64 {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0
	}
	return parsed
}

func formatStockFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', 4, 64)
}
