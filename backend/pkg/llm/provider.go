package llm

import "context"

type Provider interface {
	GetContent(ctx context.Context, systemPrompt, userPrompt string) (string, error)
	GetContentStream(ctx context.Context, systemPrompt, userPrompt string, onToken func(string) error) (string, error)
	ModelName() string
}
