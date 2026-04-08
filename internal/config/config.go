package config

import (
	"encoding/json"
	"os"
	"time"
)

type Config struct {
	HTTPAddr         string
	SessionTTL       time.Duration
	SessionSecret    string
	NovelAIAPIBase   string
	NovelAIImageBase string
	NovelAITextBase  string
	LogLevel         string
}

type fileConfig struct {
	HTTPAddr         string `json:"http_addr"`
	SessionTTL       string `json:"session_ttl"`
	SessionSecret    string `json:"session_secret"`
	NovelAIAPIBase   string `json:"novelai_api_base"`
	NovelAIImageBase string `json:"novelai_image_base"`
	NovelAITextBase  string `json:"novelai_text_base"`
	LogLevel         string `json:"log_level"`
}

func Load() Config {
	path := defaultString(os.Getenv("CONFIG_PATH"), "config.json")
	fc := loadFileConfig(path)
	return loadWithFile(os.Getenv, fc)
}

func LoadFromEnv(getenv func(string) string) Config {
	return loadWithFile(getenv, fileConfig{})
}

func loadWithFile(getenv func(string) string, fc fileConfig) Config {
	defaultTTL := defaultDuration(fc.SessionTTL, 24*time.Hour)
	return Config{
		HTTPAddr:         defaultString(getenv("HTTP_ADDR"), defaultString(fc.HTTPAddr, ":8080")),
		SessionTTL:       defaultDuration(getenv("SESSION_TTL"), defaultTTL),
		SessionSecret:    defaultString(getenv("SESSION_SECRET"), defaultString(fc.SessionSecret, "dev-secret-change-me")),
		NovelAIAPIBase:   defaultString(getenv("NOVELAI_API_BASE"), defaultString(fc.NovelAIAPIBase, "https://api.novelai.net")),
		NovelAIImageBase: defaultString(getenv("NOVELAI_IMAGE_BASE"), defaultString(fc.NovelAIImageBase, "https://image.novelai.net")),
		NovelAITextBase:  defaultString(getenv("NOVELAI_TEXT_BASE"), defaultString(fc.NovelAITextBase, "https://text.novelai.net")),
		LogLevel:         defaultString(getenv("LOG_LEVEL"), defaultString(fc.LogLevel, "info")),
	}
}

func loadFileConfig(path string) fileConfig {
	raw, err := os.ReadFile(path)
	if err != nil {
		return fileConfig{}
	}
	var cfg fileConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return fileConfig{}
	}
	return cfg
}

func defaultString(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func defaultDuration(v string, fallback time.Duration) time.Duration {
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return fallback
	}
	return d
}
