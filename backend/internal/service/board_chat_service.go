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

type BoardChatService interface {
	Chat(ctx context.Context, userID uint64, req *requestdto.BoardChatRequest) (*responsedto.BoardChatResponse, error)
	ChatStream(ctx context.Context, userID uint64, req *requestdto.BoardChatRequest, emit func(responsedto.StockChatStreamEvent) error) error
}

type boardChatService struct {
	marketSnapshotService MarketSnapshotService
	newsService           NewsService
	llmProvider           llm.Provider
	logger                *gozap.Logger
	contextService        *chatContextService
}

type boardChatContext struct {
	contextID   uint64
	boardType   string
	code        string
	question    string
	messages    []stockChatMessage
	boardDetail *responsedto.MarketBoardDetailResponse
	newsContext *StockNewsContext
	prompt      string
}

func NewBoardChatService(marketSnapshotService MarketSnapshotService, newsService NewsService, llmProvider llm.Provider, logger *gozap.Logger, contextRepo repository.ChatContextRepository) BoardChatService {
	return &boardChatService{
		marketSnapshotService: marketSnapshotService,
		newsService:           newsService,
		llmProvider:           llmProvider,
		logger:                logger,
		contextService:        newChatContextService(contextRepo),
	}
}

var boardChatSystemPrompt = `你是一位专业、克制的 A 股板块分析助手。
必须只依据输入中提供的真实板块数据、成分股表现和新闻上下文回答，不得编造成分、新闻、公告、资金流或价格数据。
回答必须使用简体中文，优先直接回应用户问题，语气自然、清晰，不要写成生硬的制式报告。
请使用 Markdown 输出，并尽量遵循下面的表达方式：
- 开头先给一个简短结论，用 1-2 句话直接回答。
- 然后给出一个“## 重点提示”小节，列出 2-4 条真正重要的信号，每条尽量短，不要重复。
- 如有必要，再用“## 板块表现”、“## 新闻催化”、“## 风险”、“## 操作建议”这些二级标题展开。
- 不是每次都必须把所有小节写满；信息不足时可以少写，但要把关键判断讲清楚。
- 如果新闻覆盖不完整，必须明确说明信息可能不足。
- 不要输出 JSON，不要写“1. 2. 3.”式模板，不要堆砌空话。
- 如果提到数据结论，尽量结合输入里的涨跌幅、上涨家数、成交额、领涨/拖累成分来支撑。`

func (s *boardChatService) Chat(ctx context.Context, userID uint64, req *requestdto.BoardChatRequest) (*responsedto.BoardChatResponse, error) {
	req = s.hydrateBoardChatRequest(userID, req)
	executor, err := newBoardChatExecutor(s, req)
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
	response, ok := result.(*responsedto.BoardChatResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected board chat response type")
	}
	if err := s.persistResponse(userID, response, req); err != nil && s.logger != nil {
		s.logger.Warn("failed to persist board chat context", gozap.Error(err))
	}
	return response, nil
}

func (s *boardChatService) ChatStream(ctx context.Context, userID uint64, req *requestdto.BoardChatRequest, emit func(responsedto.StockChatStreamEvent) error) error {
	req = s.hydrateBoardChatRequest(userID, req)
	if emit == nil {
		return fmt.Errorf("stream emitter is required")
	}
	executor, err := newBoardChatExecutor(s, req)
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
	if response, ok := result.(*responsedto.BoardChatResponse); ok {
		if persistErr := s.persistResponse(userID, response, req); persistErr != nil && s.logger != nil {
			s.logger.Warn("failed to persist board chat context", gozap.Error(persistErr))
		}
	}
	return emit(responsedto.StockChatStreamEvent{Type: "done", Stage: "done", Message: "回答生成完成", Data: result})
}

func (s *boardChatService) prepareChatContext(ctx context.Context, userID uint64, req *requestdto.BoardChatRequest, emit func(responsedto.StockChatStreamEvent) error) (*boardChatContext, error) {
	_ = userID
	boardType := strings.TrimSpace(req.BoardType)
	code := strings.TrimSpace(req.Code)
	question := strings.TrimSpace(req.Question)
	if boardType == "" || code == "" {
		return nil, fmt.Errorf("board_type and code are required")
	}
	if question == "" {
		return nil, fmt.Errorf("question is required")
	}
	if s.marketSnapshotService == nil {
		return nil, fmt.Errorf("market snapshot service is unavailable")
	}
	if s.newsService == nil {
		return nil, fmt.Errorf("news service is unavailable")
	}
	if s.llmProvider == nil {
		return nil, fmt.Errorf("llm provider is unavailable")
	}
	if emit != nil {
		if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: "market", Message: "正在获取板块表现、成分股和宽度数据"}); err != nil {
			return nil, err
		}
	}
	boardDetail, err := s.marketSnapshotService.GetBoardDetail(ctx, boardType, code, 20)
	if err != nil {
		return nil, err
	}
	if emit != nil {
		if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: "news", Message: "正在拉取并筛选板块相关新闻"}); err != nil {
			return nil, err
		}
	}
	newsContext, err := s.newsService.GetTopicNews(ctx, boardDetail.Board.Name)
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
		if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: "prompt", Message: "正在把你的问题、板块数据和新闻组装成分析上下文"}); err != nil {
			return nil, err
		}
	}

	messages := normalizeStockChatMessages(req.Messages)
	return &boardChatContext{
		contextID:   req.ContextID,
		boardType:   boardType,
		code:        code,
		question:    question,
		messages:    messages,
		boardDetail: boardDetail,
		newsContext: newsContext,
		prompt:      s.buildChatPrompt(boardDetail, newsContext, messages, question),
	}, nil
}

func (s *boardChatService) buildChatPrompt(detail *responsedto.MarketBoardDetailResponse, newsContext *StockNewsContext, messages []stockChatMessage, question string) string {
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
	gainers := summarizeBoardConstituents(detail.TopGainers)
	losers := summarizeBoardConstituents(detail.TopLosers)
	turnover := summarizeBoardConstituents(detail.TopTurnover)
	return fmt.Sprintf(`当前板块：%s (%s/%s)
板块行情：涨跌幅=%s%%，成交额=%s，成分股数量=%d，上涨=%d，下跌=%d，平盘=%d，快照时间=%s
领涨成分：%s
拖累成分：%s
成交额前排：%s
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
2. 明确区分板块面与新闻面。
3. 指出板块强弱是否由少数龙头驱动，还是整体扩散。
4. 重点提示必须具体、可落地。`,
		detail.Board.Name,
		detail.Board.BoardType,
		detail.Board.Code,
		detail.Board.ChangePercent,
		detail.Board.Turnover,
		detail.Board.StockCount,
		detail.Board.RiseCount,
		detail.Board.FallCount,
		detail.Board.FlatCount,
		detail.SnapshotTime,
		gainers,
		losers,
		turnover,
		newsContext.Status,
		newsContext.Coverage,
		newsContext.Summary,
		strings.Join(newsLines, "\n"),
		strings.Join(conversation, "\n"),
		question,
	)
}

func (s *boardChatService) buildResponse(chatCtx *boardChatContext, reply string) *responsedto.BoardChatResponse {
	responseMessages := make([]responsedto.StockChatMessageResponse, 0, len(chatCtx.messages)+2)
	for _, item := range chatCtx.messages {
		responseMessages = append(responseMessages, responsedto.StockChatMessageResponse{Role: item.Role, Content: item.Content})
	}
	responseMessages = append(responseMessages,
		responsedto.StockChatMessageResponse{Role: "user", Content: chatCtx.question},
		responsedto.StockChatMessageResponse{Role: "assistant", Content: reply},
	)

	return &responsedto.BoardChatResponse{
		ContextID:    chatCtx.contextID,
		BoardType:    chatCtx.boardType,
		Code:         chatCtx.code,
		AssetName:    chatCtx.boardDetail.Board.Name,
		Market:       "board",
		Reply:        reply,
		AIModel:      s.llmProvider.ModelName(),
		GeneratedAt:  time.Now().Format("2006-01-02 15:04:05"),
		NewsStatus:   chatCtx.newsContext.Status,
		NewsSummary:  chatCtx.newsContext.Summary,
		NewsCoverage: chatCtx.newsContext.Coverage,
		NewsItems:    buildStockChatNewsItems(chatCtx.newsContext),
		Snapshot: responsedto.StockChatSnapshotResponse{
			LastPrice:     "0",
			ChangePercent: chatCtx.boardDetail.Board.ChangePercent,
			HighPrice:     "",
			LowPrice:      "",
			Volume:        chatCtx.boardDetail.Board.Volume,
			Turnover:      chatCtx.boardDetail.Board.Turnover,
			Source:        chatCtx.boardDetail.Source,
			FetchedAt:     chatCtx.boardDetail.RefreshedAt,
			Period:        "board",
			TrendSummary:  buildBoardTrendSummary(chatCtx.boardDetail),
		},
		Messages: responseMessages,
	}
}

func (s *boardChatService) hydrateBoardChatRequest(userID uint64, req *requestdto.BoardChatRequest) *requestdto.BoardChatRequest {
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

func (s *boardChatService) persistResponse(userID uint64, resp *responsedto.BoardChatResponse, req *requestdto.BoardChatRequest) error {
	if s.contextService == nil || resp == nil || req == nil {
		return nil
	}
	targetKey := strings.ToLower(strings.TrimSpace(req.BoardType)) + ":" + strings.TrimSpace(req.Code)
	metadata := persistedChatContextMetadata{
		ToolTrace:   resp.ToolTrace,
		ToolResults: resp.ToolResults,
		NewsItems:   resp.NewsItems,
		GeneratedAt: resp.GeneratedAt,
	}
	contextID, err := s.contextService.saveContext(
		userID,
		"board",
		targetKey,
		fallbackString(resp.AssetName, targetKey),
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

func summarizeBoardConstituents(items []responsedto.BoardConstituentResponse) string {
	if len(items) == 0 {
		return "暂无"
	}
	parts := make([]string, 0, boardChatMinInt(len(items), 5))
	for index, item := range items {
		if index >= 5 {
			break
		}
		parts = append(parts, fmt.Sprintf("%s(%s%%)", item.Name, item.ChangePercent))
	}
	return strings.Join(parts, "、")
}

func buildBoardTrendSummary(detail *responsedto.MarketBoardDetailResponse) string {
	if detail == nil {
		return ""
	}
	return fmt.Sprintf("板块当前涨跌幅 %s%%，上涨 %d 家、下跌 %d 家、平盘 %d 家，成交额 %s。", detail.Board.ChangePercent, detail.Board.RiseCount, detail.Board.FallCount, detail.Board.FlatCount, detail.Board.Turnover)
}

func boardChatMinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func parseBoardFloat(value string) float64 {
	parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil {
		return 0
	}
	return parsed
}
