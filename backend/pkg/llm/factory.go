package llm

import (
	"fmt"
	"strings"

	"stock-analysis-backend/internal/config"
	"stock-analysis-backend/pkg/deepseek"
	"stock-analysis-backend/pkg/doubao"
)

func NewProvider(cfg *config.Config) (Provider, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.LLM.Provider)) {
	case "", "deepseek":
		if strings.TrimSpace(cfg.Deepseek.APIKey) == "" {
			return nil, fmt.Errorf("deepseek api key is required")
		}
		return deepseek.NewClient(cfg.Deepseek.APIKey, cfg.Deepseek.APIURL, cfg.Deepseek.Model), nil
	case "doubao", "ark":
		if strings.TrimSpace(cfg.Doubao.APIKey) == "" {
			return nil, fmt.Errorf("doubao api key is required")
		}
		if strings.TrimSpace(cfg.Doubao.Model) == "" {
			return nil, fmt.Errorf("doubao model is required")
		}
		return doubao.NewClient(cfg.Doubao.APIKey, cfg.Doubao.APIURL, cfg.Doubao.Model), nil
	default:
		return nil, fmt.Errorf("unsupported llm provider: %s", cfg.LLM.Provider)
	}
}
