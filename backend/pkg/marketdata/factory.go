package marketdata

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"stock-analysis-backend/internal/config"
)

func NewProvider(cfg config.MarketConfig) (Provider, error) {
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" {
		return nil, fmt.Errorf("market provider is required")
	}

	eastmoneyClient := &http.Client{Timeout: timeout}
	tencentClient := &http.Client{Timeout: timeout}

	switch provider {
	case "mock":
		return NewMockProvider(), nil
	case "eastmoney":
		return NewEastmoneyProvider(cfg.EastmoneyBaseURL, cfg.EastmoneyUserAgent, cfg.EastmoneyReferer, eastmoneyClient), nil
	case "tencent":
		return NewTencentKlineProvider(cfg.TencentBaseURL, cfg.TencentUserAgent, tencentClient), nil
	case "hybrid":
		realtime := NewEastmoneyProvider(cfg.EastmoneyBaseURL, cfg.EastmoneyUserAgent, cfg.EastmoneyReferer, eastmoneyClient)
		history := NewTencentKlineProvider(cfg.TencentBaseURL, cfg.TencentUserAgent, tencentClient)
		return NewHybridProvider(realtime, history), nil
	default:
		return nil, fmt.Errorf("unsupported market provider: %s", cfg.Provider)
	}
}
