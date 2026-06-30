package config

import (
	"fmt"
	"os"
	"time"

	"github.com/goccy/go-yaml"
)

const (
	defaultAddr      = "0.0.0.0"
	defaultPort      = 30078
	defaultTimezone  = "Asia/Shanghai"
	defaultProvider  = "mock"
	defaultBatchSize = 500
	defaultLogLevel  = "info"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Auth     AuthConfig     `yaml:"auth"`
	Market   MarketConfig   `yaml:"market"`
	Tushare  TushareConfig  `yaml:"tushare"`
	Log      LogConfig      `yaml:"log"`
}

type ServerConfig struct {
	Addr     string `yaml:"addr"`
	Port     int    `yaml:"port"`
	Timezone string `yaml:"timezone"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type AuthConfig struct {
	AdminToken string `yaml:"admin_token"`
}

type MarketConfig struct {
	Provider  string `yaml:"provider"`
	BatchSize int    `yaml:"batch_size"`
}

type TushareConfig struct {
	BaseURL string `yaml:"base_url"`
	Token   string `yaml:"token"`
}

type LogConfig struct {
	Level string `yaml:"level"`
}

func Load(path string) (Config, error) {
	cfg := defaultConfig()

	if err := loadYAML(path, &cfg); err != nil {
		return Config{}, err
	}
	if err := validate(cfg); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		Server: ServerConfig{
			Addr:     defaultAddr,
			Port:     defaultPort,
			Timezone: defaultTimezone,
		},
		Database: DatabaseConfig{
			Driver: "mysql",
		},
		Market: MarketConfig{
			Provider:  defaultProvider,
			BatchSize: defaultBatchSize,
		},
		Tushare: TushareConfig{
			BaseURL: "https://api.tushare.pro",
		},
		Log: LogConfig{
			Level: defaultLogLevel,
		},
	}
}

func loadYAML(path string, cfg *Config) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}
	if err := yaml.Unmarshal([]byte(os.ExpandEnv(string(content))), cfg); err != nil {
		return fmt.Errorf("parse config %s: %w", path, err)
	}
	return nil
}

func validate(cfg Config) error {
	if cfg.Server.Port <= 0 {
		return fmt.Errorf("server port must be positive")
	}
	if _, err := time.LoadLocation(cfg.Server.Timezone); err != nil {
		return fmt.Errorf("invalid server timezone %q: %w", cfg.Server.Timezone, err)
	}
	if cfg.Database.Driver != "mysql" {
		return fmt.Errorf("database driver must be mysql")
	}
	if cfg.Database.DSN == "" {
		return fmt.Errorf("database dsn is required")
	}
	if cfg.Auth.AdminToken == "" {
		return fmt.Errorf("admin token is required")
	}
	if cfg.Market.Provider != "mock" && cfg.Market.Provider != "tushare" {
		return fmt.Errorf("market provider must be mock or tushare")
	}
	if cfg.Market.BatchSize <= 0 {
		return fmt.Errorf("market batch size must be positive")
	}
	if cfg.Market.Provider == "tushare" && cfg.Tushare.Token == "" {
		return fmt.Errorf("tushare token is required when market provider is tushare")
	}
	switch cfg.Log.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("log level must be debug, info, warn, or error")
	}
	return nil
}
