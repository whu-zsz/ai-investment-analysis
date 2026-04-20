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

	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "", "mock":
		return NewMockProvider(), nil
	case "eastmoney":
		return NewEastmoneyProvider(cfg.EastmoneyBaseURL, cfg.EastmoneyUserAgent, cfg.EastmoneyReferer, &http.Client{Timeout: timeout}), nil
	default:
		return nil, fmt.Errorf("unsupported market provider: %s", cfg.Provider)
	}
}
