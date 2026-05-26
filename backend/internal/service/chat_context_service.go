package service

import (
	"encoding/json"
	"fmt"
	"strings"

	requestdto "stock-analysis-backend/internal/dto/request"
	responsedto "stock-analysis-backend/internal/dto/response"
	"stock-analysis-backend/internal/model"
	"stock-analysis-backend/internal/repository"
)

type ChatContextService interface {
	GetSnapshot(userID, contextID uint64) (*responsedto.ChatContextSnapshotResponse, error)
	GetSnapshotByReport(userID, reportID uint64, contextType string) (*responsedto.ChatContextSnapshotResponse, error)
}

type chatContextService struct {
	repo repository.ChatContextRepository
}

type persistedChatContextMetadata struct {
	ToolTrace   []responsedto.ChatToolTraceStepResponse      `json:"tool_trace,omitempty"`
	ToolResults []responsedto.ChatToolResultSnapshotResponse `json:"tool_results,omitempty"`
	StepContext *responsedto.ChatStepContextResponse         `json:"step_context,omitempty"`
	NewsItems   []responsedto.StockChatNewsItemResponse      `json:"news_items,omitempty"`
	GeneratedAt string                                       `json:"generated_at,omitempty"`
	ReportTitle string                                       `json:"report_title,omitempty"`
	Candidates  []persistedRecommendationCandidate           `json:"candidates,omitempty"`
	Insights    map[string]persistedRecommendationInsight    `json:"insights,omitempty"`
}

func newChatContextService(repo repository.ChatContextRepository) *chatContextService {
	if repo == nil {
		return nil
	}
	return &chatContextService{repo: repo}
}

func NewChatContextService(repo repository.ChatContextRepository) ChatContextService {
	return newChatContextService(repo)
}

func (s *chatContextService) loadMessages(userID, contextID uint64) ([]requestdto.StockChatMessageRequest, error) {
	entity, err := s.loadEntity(userID, contextID)
	if err != nil || entity == nil {
		return nil, err
	}
	var messages []requestdto.StockChatMessageRequest
	if strings.TrimSpace(entity.MessagesJSON) == "" {
		return nil, nil
	}
	if err := json.Unmarshal([]byte(entity.MessagesJSON), &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *chatContextService) loadMetadata(userID, contextID uint64) (*persistedChatContextMetadata, error) {
	entity, err := s.loadEntity(userID, contextID)
	if err != nil || entity == nil {
		return nil, err
	}
	metadata := &persistedChatContextMetadata{}
	if strings.TrimSpace(entity.MetadataJSON) == "" {
		return metadata, nil
	}
	if err := json.Unmarshal([]byte(entity.MetadataJSON), metadata); err != nil {
		return nil, err
	}
	return metadata, nil
}

func buildToolResultsContextMessage(metadata *persistedChatContextMetadata) string {
	if metadata == nil {
		return ""
	}
	parts := make([]string, 0, len(metadata.ToolResults)+2)
	if metadata.StepContext != nil {
		parts = append(parts, fmt.Sprintf("当前进度：%s / %s。%s", strings.TrimSpace(metadata.StepContext.Stage), strings.TrimSpace(metadata.StepContext.Label), strings.TrimSpace(metadata.StepContext.Summary)))
	}
	for _, item := range metadata.ToolResults {
		payload := ""
		if item.Payload != nil {
			if raw, err := json.Marshal(item.Payload); err == nil {
				payload = string(raw)
			}
		}
		line := fmt.Sprintf("- tool=%s status=%s summary=%s", strings.TrimSpace(item.ToolName), strings.TrimSpace(item.Status), strings.TrimSpace(item.Summary))
		if strings.TrimSpace(payload) != "" {
			line += fmt.Sprintf(" payload=%s", payload)
		}
		if strings.TrimSpace(item.Error) != "" {
			line += fmt.Sprintf(" error=%s", strings.TrimSpace(item.Error))
		}
		parts = append(parts, line)
	}
	if len(parts) == 0 {
		return ""
	}
	return "以下是同一上下文中已经完成的工具调用与结果，请在本轮直接复用这些结果，不要忽略或重复向用户索取已查到的信息。\n" + strings.Join(parts, "\n")
}

func (s *chatContextService) loadEntity(userID, contextID uint64) (*model.ChatContext, error) {
	if s == nil || s.repo == nil || contextID == 0 {
		return nil, nil
	}
	entity, err := s.repo.FindByID(contextID)
	if err != nil {
		return nil, err
	}
	if entity.UserID != userID {
		return nil, fmt.Errorf("chat context does not belong to current user")
	}
	return entity, nil
}

func (s *chatContextService) saveContext(userID uint64, contextType, targetKey, title string, contextID, reportID uint64, messages []responsedto.StockChatMessageResponse, question, reply string, metadata interface{}) (uint64, error) {
	if s == nil || s.repo == nil {
		return 0, nil
	}
	storedMessages := make([]requestdto.StockChatMessageRequest, 0, len(messages))
	for _, item := range messages {
		storedMessages = append(storedMessages, requestdto.StockChatMessageRequest{
			Role:    item.Role,
			Content: item.Content,
		})
	}
	rawMessages, err := json.Marshal(storedMessages)
	if err != nil {
		return 0, err
	}
	rawMetadata := ""
	if metadata != nil {
		if payload, metaErr := json.Marshal(metadata); metaErr != nil {
			return 0, metaErr
		} else {
			rawMetadata = string(payload)
		}
	}
	var reportPtr *uint64
	if reportID > 0 {
		reportPtr = &reportID
	}
	if contextID > 0 {
		entity, err := s.loadEntity(userID, contextID)
		if err != nil {
			return 0, err
		}
		if entity == nil {
			return 0, nil
		}
		entity.Title = title
		entity.TargetKey = targetKey
		entity.ReportID = reportPtr
		entity.MessagesJSON = string(rawMessages)
		entity.MetadataJSON = rawMetadata
		entity.LastQuestion = question
		entity.LastReply = reply
		if err := s.repo.Update(entity); err != nil {
			return 0, err
		}
		return entity.ID, nil
	}
	entity := &model.ChatContext{
		UserID:       userID,
		ContextType:  contextType,
		TargetKey:    targetKey,
		Title:        title,
		ReportID:     reportPtr,
		MessagesJSON: string(rawMessages),
		MetadataJSON: rawMetadata,
		LastQuestion: question,
		LastReply:    reply,
	}
	if err := s.repo.Create(entity); err != nil {
		return 0, err
	}
	return entity.ID, nil
}


func (s *chatContextService) GetSnapshot(userID, contextID uint64) (*responsedto.ChatContextSnapshotResponse, error) {
	entity, err := s.loadEntity(userID, contextID)
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
	metadata := persistedChatContextMetadata{}
	if strings.TrimSpace(entity.MetadataJSON) != "" {
		if err := json.Unmarshal([]byte(entity.MetadataJSON), &metadata); err != nil {
			return nil, err
		}
	}
	reportID := uint64(0)
	if entity.ReportID != nil {
		reportID = *entity.ReportID
	}
	return &responsedto.ChatContextSnapshotResponse{
		ContextID:   entity.ID,
		ContextType: entity.ContextType,
		TargetKey:   entity.TargetKey,
		Title:       entity.Title,
		ReportID:    reportID,
		Messages:    messages,
		ToolTrace:   metadata.ToolTrace,
		ToolResults: metadata.ToolResults,
		StepContext: metadata.StepContext,
		NewsItems:   metadata.NewsItems,
		Reply:       entity.LastReply,
		GeneratedAt: metadata.GeneratedAt,
		ReportTitle: metadata.ReportTitle,
	}, nil
}

func (s *chatContextService) GetSnapshotByReport(userID, reportID uint64, contextType string) (*responsedto.ChatContextSnapshotResponse, error) {
	if reportID == 0 {
		return nil, fmt.Errorf("report id is required")
	}
	entity, err := s.repo.FindLatestByUserReport(userID, reportID, contextType)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, fmt.Errorf("chat context not found")
	}
	return s.GetSnapshot(userID, entity.ID)
}
