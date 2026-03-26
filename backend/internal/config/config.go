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

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
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
	if cfg.Upload.Path == "" {
		cfg.Upload.Path = "./uploads"
	}
	if cfg.Upload.MaxUploadSize == 0 {
		cfg.Upload.MaxUploadSize = 10485760 // 10MB
	}

	return &cfg, nil
}
