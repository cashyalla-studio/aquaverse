package config

import (
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server      ServerConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	Auth        AuthConfig
	Storage     StorageConfig
	AI          AIConfig
	Crawler     CrawlerConfig
	Notification NotificationConfig
}

type ServerConfig struct {
	Port        int    `mapstructure:"port"`
	Env         string `mapstructure:"env"`
	AllowOrigins []string `mapstructure:"allow_origins"`
}

type DatabaseConfig struct {
	DSN         string `mapstructure:"dsn"`
	MaxOpenConn int    `mapstructure:"max_open_conn"`
	MaxIdleConn int    `mapstructure:"max_idle_conn"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AuthConfig struct {
	JWTSecret          string `mapstructure:"jwt_secret"`
	AccessTokenExpiry  int    `mapstructure:"access_token_expiry_min"`  // minutes
	RefreshTokenExpiry int    `mapstructure:"refresh_token_expiry_day"` // days
}

type StorageConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

type AIConfig struct {
	Provider  string `mapstructure:"provider"` // "anthropic" or "openai"
	APIKey    string `mapstructure:"api_key"`
	Model     string `mapstructure:"model"`
	MaxTokens int    `mapstructure:"max_tokens"`
}

type CrawlerConfig struct {
	FishBaseAPIBase   string `mapstructure:"fishbase_api_base"`
	WikipediaAPIBase  string `mapstructure:"wikipedia_api_base"`
	GBIFAPIBase       string `mapstructure:"gbif_api_base"`
	RequestsPerMinute int    `mapstructure:"requests_per_minute"`
	UserAgent         string `mapstructure:"user_agent"`
}

type NotificationConfig struct {
	FCMServerKey string `mapstructure:"fcm_server_key"`
	SendGridKey  string `mapstructure:"sendgrid_key"`
	FromEmail    string `mapstructure:"from_email"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// ENV 오버라이드: AV_ 접두사 (AquaVerse)
	viper.SetEnvPrefix("AV")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 기본값
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.env", "development")
	viper.SetDefault("database.max_open_conn", 25)
	viper.SetDefault("database.max_idle_conn", 5)
	viper.SetDefault("auth.access_token_expiry_min", 60)
	viper.SetDefault("auth.refresh_token_expiry_day", 30)
	viper.SetDefault("ai.provider", "anthropic")
	viper.SetDefault("ai.model", "claude-haiku-4-5-20251001")
	viper.SetDefault("ai.max_tokens", 2048)
	viper.SetDefault("crawler.fishbase_api_base", "https://fishbase.ropensci.org")
	viper.SetDefault("crawler.wikipedia_api_base", "https://en.wikipedia.org/api/rest_v1")
	viper.SetDefault("crawler.gbif_api_base", "https://api.gbif.org/v1")
	viper.SetDefault("crawler.requests_per_minute", 10)
	viper.SetDefault("crawler.user_agent", "AquaVerse/1.0 (contact@aquaverse.app)")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		slog.Info("config file not found, using env/defaults")
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
