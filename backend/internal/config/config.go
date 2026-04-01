package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Deepseek DeepseekConfig
	Market   MarketConfig
	Upload   UploadConfig
}

type ServerConfig struct {
	Port string `mapstructure:"SERVER_PORT"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"DB_HOST"`
	Port     string `mapstructure:"DB_PORT"`
	User     string `mapstructure:"DB_USER"`
	Password string `mapstructure:"DB_PASSWORD"`
	DBName   string `mapstructure:"DB_NAME"`
}

type JWTConfig struct {
	Secret      string `mapstructure:"JWT_SECRET"`
	ExpireHours int    `mapstructure:"JWT_EXPIRE_HOURS"`
}

type DeepseekConfig struct {
	APIKey  string `mapstructure:"DEEPSEEK_API_KEY"`
	APIURL  string `mapstructure:"DEEPSEEK_API_URL"`
	Model   string `mapstructure:"DEEPSEEK_MODEL"`
}

type MarketConfig struct {
	Provider         string `mapstructure:"MARKET_PROVIDER"`
	Symbols          string `mapstructure:"MARKET_SYMBOLS"`
	SnapshotInterval int    `mapstructure:"MARKET_SNAPSHOT_INTERVAL"`
	Enabled          bool   `mapstructure:"MARKET_ENABLED"`
	TimeoutSeconds   int    `mapstructure:"MARKET_TIMEOUT_SECONDS"`
	EastmoneyBaseURL string `mapstructure:"MARKET_EASTMONEY_BASE_URL"`
	EastmoneyUserAgent string `mapstructure:"MARKET_EASTMONEY_USER_AGENT"`
	EastmoneyReferer string `mapstructure:"MARKET_EASTMONEY_REFERER"`
}

type UploadConfig struct {
	Path          string `mapstructure:"UPLOAD_PATH"`
	MaxUploadSize int64  `mapstructure:"MAX_UPLOAD_SIZE"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigFile(".env")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("DB_HOST"),
			Port:     viper.GetString("DB_PORT"),
			User:     viper.GetString("DB_USER"),
			Password: viper.GetString("DB_PASSWORD"),
			DBName:   viper.GetString("DB_NAME"),
		},
		JWT: JWTConfig{
			Secret:      viper.GetString("JWT_SECRET"),
			ExpireHours: viper.GetInt("JWT_EXPIRE_HOURS"),
		},
		Deepseek: DeepseekConfig{
			APIKey: viper.GetString("DEEPSEEK_API_KEY"),
			APIURL: viper.GetString("DEEPSEEK_API_URL"),
			Model:  viper.GetString("DEEPSEEK_MODEL"),
		},
		Market: MarketConfig{
			Provider:           viper.GetString("MARKET_PROVIDER"),
			Symbols:            viper.GetString("MARKET_SYMBOLS"),
			SnapshotInterval:   viper.GetInt("MARKET_SNAPSHOT_INTERVAL"),
			Enabled:            viper.GetBool("MARKET_ENABLED"),
			TimeoutSeconds:     viper.GetInt("MARKET_TIMEOUT_SECONDS"),
			EastmoneyBaseURL:   viper.GetString("MARKET_EASTMONEY_BASE_URL"),
			EastmoneyUserAgent: viper.GetString("MARKET_EASTMONEY_USER_AGENT"),
			EastmoneyReferer:   viper.GetString("MARKET_EASTMONEY_REFERER"),
		},
		Upload: UploadConfig{
			Path:          viper.GetString("UPLOAD_PATH"),
			MaxUploadSize: viper.GetInt64("MAX_UPLOAD_SIZE"),
		},
	}

	// 设置默认值
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.JWT.ExpireHours == 0 {
		cfg.JWT.ExpireHours = 24
	}
	if cfg.Deepseek.Model == "" {
		cfg.Deepseek.Model = "deepseek-chat"
	}
	if cfg.Market.Provider == "" {
		cfg.Market.Provider = "mock"
	}
	if cfg.Market.Symbols == "" {
		cfg.Market.Symbols = "000001.SH,399001.SZ,399006.SZ,000300.SH"
	}
	if cfg.Market.SnapshotInterval == 0 {
		cfg.Market.SnapshotInterval = 60
	}
	if cfg.Market.TimeoutSeconds == 0 {
		cfg.Market.TimeoutSeconds = 5
	}
	if cfg.Market.EastmoneyBaseURL == "" {
		cfg.Market.EastmoneyBaseURL = "https://push2.eastmoney.com/api/qt/ulist.np/get"
	}
	if cfg.Market.EastmoneyUserAgent == "" {
		cfg.Market.EastmoneyUserAgent = "Mozilla/5.0"
	}
	if cfg.Market.EastmoneyReferer == "" {
		cfg.Market.EastmoneyReferer = "https://quote.eastmoney.com/center/gridlist.html"
	}
	if cfg.Upload.Path == "" {
		cfg.Upload.Path = "./uploads"
	}
	if cfg.Upload.MaxUploadSize == 0 {
		cfg.Upload.MaxUploadSize = 10485760 // 10MB
	}

	return cfg, nil
}
