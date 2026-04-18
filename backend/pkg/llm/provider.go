package llm

import "context"

// Provider 定义统一的大模型文本生成能力。
type Provider interface {
	GetContent(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	ModelName() string
}
