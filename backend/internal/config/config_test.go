package config

import (
	"strings"
	"testing"
)

func validConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     "127.0.0.1",
			Port:     "3306",
			User:     "root",
			Password: "db-password",
			DBName:   "stock_analysis",
		},
		JWT: JWTConfig{
			Secret:      "1234567890abcdef1234567890abcdef",
			ExpireHours: 24,
		},
		Market: MarketConfig{
			Provider:         "mock",
			Symbols:          "000001.SH",
			SnapshotInterval: 60,
			Enabled:          true,
			TimeoutSeconds:   5,
		},
	}
}

func assertValidateConfigErrorContains(t *testing.T, cfg *Config, want string) {
	t.Helper()

	err := validateConfig(cfg)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("expected error containing %q, got %v", want, err)
	}
}

func TestValidateConfigRejectsWeakJWTSecret(t *testing.T) {
	cfg := validConfig()
	cfg.JWT.Secret = "your_jwt_secret_key_change_this_in_production"

	assertValidateConfigErrorContains(t, cfg, "insecure default value")
}

func TestValidateConfigRejectsShortJWTSecret(t *testing.T) {
	cfg := validConfig()
	cfg.JWT.Secret = "too-short-secret"

	assertValidateConfigErrorContains(t, cfg, "at least 32 characters")
}

func TestValidateConfigRequiresMarketProviderWhenEnabled(t *testing.T) {
	cfg := validConfig()
	cfg.Market.Provider = ""

	assertValidateConfigErrorContains(t, cfg, "MARKET_PROVIDER")
}

func TestValidateConfigAllowsMissingMarketProviderWhenDisabled(t *testing.T) {
	cfg := validConfig()
	cfg.Market.Enabled = false
	cfg.Market.Provider = ""

	if err := validateConfig(cfg); err != nil {
		t.Fatalf("expected disabled market config to pass validation, got %v", err)
	}
}

func TestValidateConfigRejectsUnsupportedMarketProvider(t *testing.T) {
	cfg := validConfig()
	cfg.Market.Provider = "unknown"

	assertValidateConfigErrorContains(t, cfg, "unsupported market provider")
}

func TestValidateConfigRequiresEastmoneyFields(t *testing.T) {
	cfg := validConfig()
	cfg.Market.Provider = "eastmoney"
	cfg.Market.EastmoneyBaseURL = ""
	cfg.Market.EastmoneyReferer = ""

	assertValidateConfigErrorContains(t, cfg, "MARKET_EASTMONEY_BASE_URL")
}
