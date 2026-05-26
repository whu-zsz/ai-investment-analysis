package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	responsedto "stock-analysis-backend/internal/dto/response"
)

const chatToolMaxRounds = 6
const chatToolRetryBudget = 2

type chatToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type chatToolResult struct {
	ToolName   string      `json:"tool_name"`
	ToolCallID string      `json:"tool_call_id"`
	Status     string      `json:"status"`
	Summary    string      `json:"summary"`
	Payload    interface{} `json:"payload,omitempty"`
	Error      string      `json:"error,omitempty"`
}

type chatToolCall struct {
	ToolName string          `json:"tool_name"`
	Args     json.RawMessage `json:"args,omitempty"`
}

type chatToolPlan struct {
	NeedMoreTools bool           `json:"need_more_tools"`
	AssistantNote string         `json:"assistant_note,omitempty"`
	Calls         []chatToolCall `json:"calls,omitempty"`
}

type recommendationActionPlan struct {
	NextAction string `json:"next_action"`
	Target     string `json:"target,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

type chatToolExecutor interface {
	Definitions() []chatToolDefinition
	PlanningPrompt() string
	FinalPrompt() string
	Question() string
	ConversationHistory() string
	ExecuteTool(ctx context.Context, call chatToolCall) (*chatToolResult, *responsedto.ChatStepContextResponse, error)
	BuildFinalUserPrompt(results []chatToolResult) string
	BuildDoneResponse(reply string, trace []responsedto.ChatToolTraceStepResponse, results []chatToolResult) (interface{}, error)
}

type recommendationActionPlanner interface {
	PlanActionPrompt() string
	BuildActionPlanPrompt(results []chatToolResult) string
	MapActionPlan(plan recommendationActionPlan, results []chatToolResult) (*chatToolPlan, error)
}

type chatOrchestrator struct {
	llmProvider interface {
		GetContent(ctx context.Context, systemPrompt, userPrompt string) (string, error)
		GetContentStream(ctx context.Context, systemPrompt, userPrompt string, onToken func(token string) error) (string, error)
		ModelName() string
	}
}

func newChatOrchestrator(provider interface {
	GetContent(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	GetContentStream(ctx context.Context, systemPrompt, userPrompt string, onToken func(token string) error) (string, error)
	ModelName() string
}) *chatOrchestrator {
	return &chatOrchestrator{llmProvider: provider}
}

func (o *chatOrchestrator) Run(ctx context.Context, executor chatToolExecutor, emit func(responsedto.StockChatStreamEvent) error) (string, []chatToolResult, []responsedto.ChatToolTraceStepResponse, error) {
	trace := make([]responsedto.ChatToolTraceStepResponse, 0, chatToolMaxRounds+2)
	results := make([]chatToolResult, 0, chatToolMaxRounds)
	retryBudget := chatToolRetryBudget
	if emit != nil {
		if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: "planning", Message: "AI 正在规划所需工具步骤"}); err != nil {
			return "", nil, trace, err
		}
	}
	for round := 0; round < chatToolMaxRounds; round++ {
		if round > 0 && emit != nil {
			if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: "planning", Message: fmt.Sprintf("AI 正在根据已有结果进行第 %d 轮规划", round+1)}); err != nil {
				return "", nil, trace, err
			}
		}
		plan, err := o.plan(ctx, executor, results)
		if err != nil {
			return "", results, trace, err
		}
		if !plan.NeedMoreTools || len(plan.Calls) == 0 {
			break
		}
		hadHardError := false
		for _, call := range plan.Calls {
			startedAt := time.Now()
			stage, label := stageMetaForTool(call.ToolName)
			argsSummary := strings.TrimSpace(string(call.Args))
			if argsSummary == "" {
				argsSummary = "{}"
			}
			step := responsedto.ChatToolTraceStepResponse{
				Stage:     stage,
				Label:     label,
				Status:    "process",
				ToolName:  call.ToolName,
				StartedAt: startedAt.Format("2006-01-02 15:04:05"),
			}
			if emit != nil {
				if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: stage, Message: fmt.Sprintf("正在执行 %s", label)}); err != nil {
					return "", results, trace, err
				}
			}
			result, contextData, err := executor.ExecuteTool(ctx, call)
			elapsed := time.Since(startedAt)
			if err != nil {
				fmt.Printf("[chat-tools] tool=%s stage=%s status=error elapsed=%s args=%s err=%s\n", call.ToolName, stage, elapsed.Round(time.Millisecond), argsSummary, err.Error())
				hadHardError = true
				step.Status = "error"
				step.Summary = err.Error()
				step.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
				trace = append(trace, step)
				if emit != nil {
					_ = emit(responsedto.StockChatStreamEvent{Type: "error", Stage: stage, Message: err.Error()})
				}
				results = append(results, chatToolResult{
					ToolName: call.ToolName,
					Status:   "error",
					Summary:  err.Error(),
					Error:    err.Error(),
				})
				continue
			}
			if result == nil {
				result = &chatToolResult{ToolName: call.ToolName, Status: "empty", Summary: "工具未返回结果"}
			}
			fmt.Printf("[chat-tools] tool=%s stage=%s status=%s elapsed=%s args=%s summary=%s\n", call.ToolName, stage, result.Status, elapsed.Round(time.Millisecond), argsSummary, strings.TrimSpace(result.Summary))
			if recommendationExecutor, ok := executor.(*recommendationChatExecutor); ok {
				recommendationExecutor.stepContext = contextData
			}
			if result.Status == "error" {
				hadHardError = true
			}
			step.Status = "finish"
			step.Summary = strings.TrimSpace(result.Summary)
			step.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
			trace = append(trace, step)
			results = append(results, *result)
			if emit != nil {
				message := result.Summary
				if message == "" {
					message = fmt.Sprintf("%s 已完成", label)
				}
				if err := emit(responsedto.StockChatStreamEvent{Type: "context", Stage: stage, Message: message, Data: contextData}); err != nil {
					return "", results, trace, err
				}
			}
		}
		if hadHardError && retryBudget > 0 {
			retryBudget--
			if emit != nil {
				if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: "planning", Message: fmt.Sprintf("工具调用出现错误，AI 将根据错误信息重新规划，剩余重试 %d 次", retryBudget)}); err != nil {
					return "", results, trace, err
				}
			}
			continue
		}
	}
	if emit != nil {
		if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: "report", Message: "AI 正在基于工具结果生成回答"}); err != nil {
			return "", results, trace, err
		}
	}
	finalPrompt := executor.BuildFinalUserPrompt(results)
	buffer := strings.Builder{}
	reply, err := o.llmProvider.GetContentStream(ctx, executor.FinalPrompt(), finalPrompt, func(token string) error {
		buffer.WriteString(token)
		if emit == nil {
			return nil
		}
		return emit(responsedto.StockChatStreamEvent{Type: "token", Stage: "report", Token: token})
	})
	if err != nil {
		return "", results, trace, err
	}
	rawReply := strings.TrimSpace(reply)
	if rawReply == "" {
		rawReply = strings.TrimSpace(buffer.String())
	}
	if toolCalls, ok := extractInlineToolCalls(rawReply); ok && len(toolCalls) > 0 {
		if emit != nil {
			if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: "report", Message: "AI 正在提交最终报告"}); err != nil {
				return "", results, trace, err
			}
		}
		for _, call := range toolCalls {
			startedAt := time.Now()
			stage, label := stageMetaForTool(call.ToolName)
			step := responsedto.ChatToolTraceStepResponse{
				Stage:     stage,
				Label:     label,
				Status:    "process",
				ToolName:  call.ToolName,
				StartedAt: startedAt.Format("2006-01-02 15:04:05"),
			}
			if emit != nil {
				if err := emit(responsedto.StockChatStreamEvent{Type: "step", Stage: stage, Message: fmt.Sprintf("正在执行 %s", label)}); err != nil {
					return "", results, trace, err
				}
			}
			result, contextData, execErr := executor.ExecuteTool(ctx, call)
			if execErr != nil {
				step.Status = "error"
				step.Summary = execErr.Error()
				step.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
				trace = append(trace, step)
				return "", results, trace, execErr
			}
			if result == nil {
				result = &chatToolResult{ToolName: call.ToolName, Status: "empty", Summary: "工具未返回结果"}
			}
			step.Status = "finish"
			step.Summary = strings.TrimSpace(result.Summary)
			step.FinishedAt = time.Now().Format("2006-01-02 15:04:05")
			trace = append(trace, step)
			results = append(results, *result)
			if emit != nil {
				message := result.Summary
				if message == "" {
					message = fmt.Sprintf("%s 已完成", label)
				}
				if err := emit(responsedto.StockChatStreamEvent{Type: "context", Stage: stage, Message: message, Data: contextData}); err != nil {
					return "", results, trace, err
				}
			}
		}
		reply = buildPostToolReply(results)
	} else {
		reply = normalizeMarkdownNarrative(rawReply)
		if strings.TrimSpace(reply) == "" {
			reply = normalizeMarkdownNarrative(buffer.String())
		}
	}
	return reply, results, trace, nil
}

func (o *chatOrchestrator) plan(ctx context.Context, executor chatToolExecutor, results []chatToolResult) (*chatToolPlan, error) {
	if actionPlanner, ok := executor.(recommendationActionPlanner); ok {
		return o.planRecommendationAction(ctx, actionPlanner, results)
	}
	userPrompt := buildToolPlanningUserPrompt(executor, results)
	content, err := o.llmProvider.GetContent(ctx, executor.PlanningPrompt(), userPrompt)
	if err != nil {
		return nil, err
	}
	rawContent := strings.TrimSpace(content)
	content = extractJSONObject(content)
	var plan chatToolPlan
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &plan); err != nil {
		fmt.Printf("[chat-tools] plan-parse-failed raw=%s extracted=%s err=%s\n", truncateForLog(rawContent, 1200), truncateForLog(strings.TrimSpace(content), 1200), err.Error())
		return nil, fmt.Errorf("failed to parse tool plan: %w", err)
	}
	return &plan, nil
}

func (o *chatOrchestrator) planRecommendationAction(ctx context.Context, planner recommendationActionPlanner, results []chatToolResult) (*chatToolPlan, error) {
	userPrompt := planner.BuildActionPlanPrompt(results)
	content, err := o.llmProvider.GetContent(ctx, planner.PlanActionPrompt(), userPrompt)
	if err != nil {
		return nil, err
	}
	rawContent := strings.TrimSpace(content)
	content = extractJSONObject(content)
	var actionPlan recommendationActionPlan
	if err := json.Unmarshal([]byte(strings.TrimSpace(content)), &actionPlan); err != nil {
		fmt.Printf("[chat-tools] action-plan-parse-failed raw=%s extracted=%s err=%s\n", truncateForLog(rawContent, 1200), truncateForLog(strings.TrimSpace(content), 1200), err.Error())
		return nil, fmt.Errorf("failed to parse action plan: %w", err)
	}
	return planner.MapActionPlan(actionPlan, results)
}

func buildToolPlanningUserPrompt(executor chatToolExecutor, results []chatToolResult) string {
	toolDefs, _ := json.Marshal(executor.Definitions())
	resultLines := make([]string, 0, len(results))
	for _, result := range results {
		payloadSummary := summarizeToolResultPayload(result)
		resultLines = append(resultLines, fmt.Sprintf("- tool=%s status=%s summary=%s payload_summary=%s error=%s", result.ToolName, result.Status, result.Summary, payloadSummary, result.Error))
	}
	if len(resultLines) == 0 {
		resultLines = append(resultLines, "- 暂无已执行工具结果")
	}
	return fmt.Sprintf("用户问题:\n%s\n\n历史对话:\n%s\n\n可用工具:\n%s\n\n已执行工具结果摘要:\n%s\n\n注意：上面只提供轻量摘要，禁止重复请求已经完成且信息充足的工具。请判断下一步是否需要继续调用工具。输出 JSON: {\"need_more_tools\": bool, \"assistant_note\": string, \"calls\": [{\"tool_name\": string, \"args\": object}]}。如果已有信息足够，请返回 need_more_tools=false 且 calls 为空。一次最多返回 2 个工具调用。", executor.Question(), executor.ConversationHistory(), string(toolDefs), strings.Join(resultLines, "\n"))
}

func summarizeToolResultPayload(result chatToolResult) string {
	if result.Payload == nil {
		return "{}"
	}
	compact := map[string]any{}
	switch payload := result.Payload.(type) {
	case map[string]any:
		for _, key := range []string{"symbol", "name", "board_type", "board_code", "board_name", "change_percent", "trend_summary", "news_coverage", "candidate_count", "held_count", "discovery_count", "report_id", "report_title", "action", "match_reason", "risk_note", "analyzed_symbols", "boards", "requires", "limit"} {
			if value, ok := payload[key]; ok {
				compact[key] = value
			}
		}
		if newsItems, ok := payload["news_items"].([]map[string]any); ok {
			compact["news_headlines"] = firstHeadlineMaps(newsItems, 2)
		}
		if newsItems, ok := payload["news_items"].([]any); ok {
			compact["news_headlines"] = firstHeadlineValues(newsItems, 2)
		}
		if items, ok := payload["items"].([]any); ok {
			compact["items_count"] = len(items)
		}
		if leaders, ok := payload["leaders"].([]any); ok {
			compact["leaders_count"] = len(leaders)
		}
		if constituents, ok := payload["constituents"].([]any); ok {
			compact["constituents_count"] = len(constituents)
		}
	default:
		bytes, _ := json.Marshal(payload)
		if len(bytes) > 320 {
			return string(bytes[:320]) + "..."
		}
		return string(bytes)
	}
	if len(compact) == 0 {
		bytes, _ := json.Marshal(result.Payload)
		if len(bytes) > 320 {
			return string(bytes[:320]) + "..."
		}
		return string(bytes)
	}
	bytes, _ := json.Marshal(compact)
	if len(bytes) > 600 {
		return string(bytes[:600]) + "..."
	}
	return string(bytes)
}

func firstHeadlineMaps(items []map[string]any, limit int) []string {
	result := make([]string, 0, limit)
	for _, item := range items {
		if len(result) >= limit {
			break
		}
		title, _ := item["title"].(string)
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}
		result = append(result, title)
	}
	return result
}

func firstHeadlineValues(items []any, limit int) []string {
	result := make([]string, 0, limit)
	for _, raw := range items {
		if len(result) >= limit {
			break
		}
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		title, _ := item["title"].(string)
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}
		result = append(result, title)
	}
	return result
}

func extractJSONObject(content string) string {
	trimmed := strings.TrimSpace(content)
	start := strings.Index(trimmed, "{")
	if start < 0 {
		return trimmed
	}
	depth := 0
	inString := false
	escaped := false
	for idx := start; idx < len(trimmed); idx++ {
		ch := trimmed[idx]
		if escaped {
			escaped = false
			continue
		}
		if ch == '\\' {
			escaped = true
			continue
		}
		if ch == '"' {
			inString = !inString
			continue
		}
		if inString {
			continue
		}
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return trimmed[start : idx+1]
			}
		}
	}
	return trimmed[start:]
}

func truncateForLog(value string, limit int) string {
	trimmed := strings.TrimSpace(value)
	if limit <= 0 {
		return trimmed
	}
	runes := []rune(trimmed)
	if len(runes) <= limit {
		return trimmed
	}
	return string(runes[:limit]) + "..."
}

func extractInlineToolCalls(reply string) ([]chatToolCall, bool) {
	reply = strings.TrimSpace(reply)
	if reply == "" {
		return nil, false
	}
	startMarker := "<|FunctionCallBegin|>"
	endMarker := "<|FunctionCallEnd|>"
	start := strings.Index(reply, startMarker)
	end := strings.LastIndex(reply, endMarker)
	if start < 0 || end < 0 || end <= start {
		return nil, false
	}
	payload := strings.TrimSpace(reply[start+len(startMarker) : end])
	if payload == "" {
		return nil, false
	}
	var rawCalls []struct {
		Name       string          `json:"name"`
		Parameters json.RawMessage `json:"parameters"`
	}
	if err := json.Unmarshal([]byte(payload), &rawCalls); err != nil {
		fmt.Printf("[chat-tools] final-inline-call-parse-failed payload=%s err=%s\n", truncateForLog(payload, 1200), err.Error())
		return nil, false
	}
	calls := make([]chatToolCall, 0, len(rawCalls))
	for _, item := range rawCalls {
		calls = append(calls, chatToolCall{ToolName: strings.TrimSpace(item.Name), Args: item.Parameters})
	}
	return calls, len(calls) > 0
}

func buildPostToolReply(results []chatToolResult) string {
	for i := len(results) - 1; i >= 0; i-- {
		item := results[i]
		if item.ToolName != "save_recommendation_report" || item.Status != "ok" {
			continue
		}
		switch payload := item.Payload.(type) {
		case map[string]interface{}:
			if reportID, exists := payload["report_id"]; exists {
				return fmt.Sprintf("## 报告已生成\n推荐报告已经保存，报告 ID：%v。你可以返回推荐报告页继续查看完整内容。", reportID)
			}
		case *responsedto.AnalysisRecommendationsResponse:
			return fmt.Sprintf("## 报告已生成\n推荐报告已经保存，报告 ID：%d。你可以返回推荐报告页继续查看完整内容。", payload.ReportID)
		}
		return "## 报告已生成\n推荐报告已经保存，你可以返回推荐报告页继续查看完整内容。"
	}
	return "## 分析完成\n本轮分析已完成，但尚未生成正式报告。"
}

func stageMetaForTool(toolName string) (string, string) {
	switch toolName {
	case "get_user_investment_profile":
		return "profile", "读取用户画像"
	case "get_user_positions_and_watch_history":
		return "holdings", "读取持仓与关注记录"
	case "get_recent_market_news_candidates":
		return "news", "筛选新闻热点"
	case "list_market_boards":
		return "board", "读取板块目录"
	case "search_relevant_boards":
		return "board", "检索相关板块"
	case "search_board_stocks":
		return "board", "检索板块成分股"
	case "get_stock_indicators_and_news":
		return "trend", "分析股票指标与新闻"
	case "get_stock_quote_and_trend":
		return "trend", "补充行情与趋势"
	case "get_stock_profile_and_boards":
		return "quote", "补充个股资料与板块"
	case "get_board_heat_and_constituents":
		return "board", "读取板块热度与成分股"
	case "rank_recommendation_candidates":
		return "ranking", "候选排序"
	case "save_recommendation_report":
		return "report", "保存推荐报告"
	default:
		return "planning", toolName
	}
}
