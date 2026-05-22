package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	LLM      LLMConfig
	Deepseek DeepseekConfig
	Doubao   DoubaoConfig
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

var weakJWTSecrets = map[string]struct{}{
	"your_jwt_secret_key_change_this":               {},
	"your_jwt_secret_key_change_this_in_production": {},
	"change_this":                                  {},
	"default":                                      {},
	"secret":                                       {},
}

type LLMConfig struct {
	Provider string `mapstructure:"LLM_PROVIDER"`
}

type DeepseekConfig struct {
	APIKey string `mapstructure:"DEEPSEEK_API_KEY"`
	APIURL string `mapstructure:"DEEPSEEK_API_URL"`
	Model  string `mapstructure:"DEEPSEEK_MODEL"`
}

type DoubaoConfig struct {
	APIKey string `mapstructure:"DOUBAO_API_KEY"`
	APIURL string `mapstructure:"DOUBAO_API_URL"`
	Model  string `mapstructure:"DOUBAO_MODEL"`
}

type MarketConfig struct {
	Provider           string `mapstructure:"MARKET_PROVIDER"`
	Symbols            string `mapstructure:"MARKET_SYMBOLS"`
	SnapshotInterval   int    `mapstructure:"MARKET_SNAPSHOT_INTERVAL"`
	Enabled            bool   `mapstructure:"MARKET_ENABLED"`
	TimeoutSeconds     int    `mapstructure:"MARKET_TIMEOUT_SECONDS"`
	EastmoneyBaseURL   string `mapstructure:"MARKET_EASTMONEY_BASE_URL"`
	EastmoneyUserAgent string `mapstructure:"MARKET_EASTMONEY_USER_AGENT"`
	EastmoneyReferer   string `mapstructure:"MARKET_EASTMONEY_REFERER"`
}

type UploadConfig struct {
	Path          string `mapstructure:"UPLOAD_PATH"`
	MaxUploadSize int64  `mapstructure:"MAX_UPLOAD_SIZE"`
}

func LoadConfig() (*Config, error) {
	v := viper.New()
	v.AutomaticEnv()

	configFile, err := resolveConfigFile()
	if err != nil {
		return nil, err
	}
	if configFile != "" {
		v.SetConfigFile(configFile)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
		}
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: v.GetString("SERVER_PORT"),
		},
		Database: DatabaseConfig{
			Host:     v.GetString("DB_HOST"),
			Port:     v.GetString("DB_PORT"),
			User:     v.GetString("DB_USER"),
			Password: v.GetString("DB_PASSWORD"),
			DBName:   v.GetString("DB_NAME"),
		},
		JWT: JWTConfig{
			Secret:      v.GetString("JWT_SECRET"),
			ExpireHours: v.GetInt("JWT_EXPIRE_HOURS"),
		},
		LLM: LLMConfig{
			Provider: v.GetString("LLM_PROVIDER"),
		},
		Deepseek: DeepseekConfig{
			APIKey: v.GetString("DEEPSEEK_API_KEY"),
			APIURL: v.GetString("DEEPSEEK_API_URL"),
			Model:  v.GetString("DEEPSEEK_MODEL"),
		},
		Doubao: DoubaoConfig{
			APIKey: v.GetString("DOUBAO_API_KEY"),
			APIURL: v.GetString("DOUBAO_API_URL"),
			Model:  v.GetString("DOUBAO_MODEL"),
		},
		Market: MarketConfig{
			Provider:           v.GetString("MARKET_PROVIDER"),
			Symbols:            v.GetString("MARKET_SYMBOLS"),
			SnapshotInterval:   v.GetInt("MARKET_SNAPSHOT_INTERVAL"),
			Enabled:            v.GetBool("MARKET_ENABLED"),
			TimeoutSeconds:     v.GetInt("MARKET_TIMEOUT_SECONDS"),
			EastmoneyBaseURL:   v.GetString("MARKET_EASTMONEY_BASE_URL"),
			EastmoneyUserAgent: v.GetString("MARKET_EASTMONEY_USER_AGENT"),
			EastmoneyReferer:   v.GetString("MARKET_EASTMONEY_REFERER"),
		},
		Upload: UploadConfig{
			Path:          v.GetString("UPLOAD_PATH"),
			MaxUploadSize: v.GetInt64("MAX_UPLOAD_SIZE"),
		},
	}

	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}
	if cfg.JWT.ExpireHours == 0 {
		cfg.JWT.ExpireHours = 24
	}
	if cfg.LLM.Provider == "" {
		cfg.LLM.Provider = "deepseek"
	}
	if cfg.Deepseek.Model == "" {
		cfg.Deepseek.Model = "deepseek-chat"
	}
	if cfg.Deepseek.APIURL == "" {
		cfg.Deepseek.APIURL = "https://api.deepseek.com"
	}
	if cfg.Doubao.APIURL == "" {
		cfg.Doubao.APIURL = "https://ark.cn-beijing.volces.com"
	}
	cfg.Market.Provider = strings.ToLower(strings.TrimSpace(cfg.Market.Provider))
	if !cfg.Market.Enabled {
		cfg.Market.Provider = ""
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
		cfg.Upload.MaxUploadSize = 10485760
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func resolveConfigFile() (string, error) {
	candidates := make([]string, 0, 3)
	if envFile := strings.TrimSpace(os.Getenv("ENV_FILE")); envFile != "" {
		candidates = append(candidates, envFile)
	}
	candidates = append(candidates, ".env", filepath.Join("backend", ".env"))

	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		} else if !os.IsNotExist(err) {
			return "", fmt.Errorf("failed to access config file %s: %w", candidate, err)
		}
	}

	return "", nil
}

func validateConfig(cfg *Config) error {
	missing := make([]string, 0, 5)
	if strings.TrimSpace(cfg.Database.Host) == "" {
		missing = append(missing, "DB_HOST")
	}
	if strings.TrimSpace(cfg.Database.Port) == "" {
		missing = append(missing, "DB_PORT")
	}
	if strings.TrimSpace(cfg.Database.User) == "" {
		missing = append(missing, "DB_USER")
	}
	if strings.TrimSpace(cfg.Database.Password) == "" {
		missing = append(missing, "DB_PASSWORD")
	}
	if strings.TrimSpace(cfg.Database.DBName) == "" {
		missing = append(missing, "DB_NAME")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required database configuration: %s", strings.Join(missing, ", "))
	}

	jwtSecret := strings.TrimSpace(cfg.JWT.Secret)
	if jwtSecret == "" {
		return fmt.Errorf("missing required configuration: JWT_SECRET")
	}
	if len(jwtSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	if _, weak := weakJWTSecrets[strings.ToLower(jwtSecret)]; weak {
		return fmt.Errorf("JWT_SECRET uses an insecure default value")
	}

	if !cfg.Market.Enabled {
		return nil
	}
	if cfg.Market.Provider == "" {
		return fmt.Errorf("missing required configuration: MARKET_PROVIDER")
	}
	if cfg.Market.Provider != "mock" && cfg.Market.Provider != "eastmoney" {
		return fmt.Errorf("unsupported market provider: %s", cfg.Market.Provider)
	}
	if cfg.Market.Provider == "eastmoney" {
		if strings.TrimSpace(cfg.Market.EastmoneyBaseURL) == "" {
			return fmt.Errorf("missing required configuration: MARKET_EASTMONEY_BASE_URL")
		}
		if strings.TrimSpace(cfg.Market.EastmoneyReferer) == "" {
			return fmt.Errorf("missing required configuration: MARKET_EASTMONEY_REFERER")
		}
	}

	return nil
}
