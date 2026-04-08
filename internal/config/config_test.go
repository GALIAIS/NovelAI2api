package config

import (
	"testing"
	"time"
)

func TestLoadUsesDefaults(t *testing.T) {
	cfg := LoadFromEnv(func(string) string { return "" })

	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr = %q, want :8080", cfg.HTTPAddr)
	}
	if cfg.SessionTTL != 24*time.Hour {
		t.Fatalf("SessionTTL = %s, want 24h", cfg.SessionTTL)
	}
	if cfg.NovelAIAPIBase != "https://api.novelai.net" {
		t.Fatalf("NovelAIAPIBase = %q", cfg.NovelAIAPIBase)
	}
}

func TestLoadParsesSessionTTLFromEnv(t *testing.T) {
	cfg := LoadFromEnv(func(k string) string {
		if k == "SESSION_TTL" {
			return "48h"
		}
		return ""
	})

	if cfg.SessionTTL != 48*time.Hour {
		t.Fatalf("SessionTTL = %s, want 48h", cfg.SessionTTL)
	}
}

func TestLoadWithFileConfig(t *testing.T) {
	cfg := loadWithFile(func(string) string { return "" }, fileConfig{
		HTTPAddr:   ":8787",
		SessionTTL: "36h",
		LogLevel:   "debug",
	})
	if cfg.HTTPAddr != ":8787" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
	if cfg.SessionTTL != 36*time.Hour {
		t.Fatalf("SessionTTL = %s", cfg.SessionTTL)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("LogLevel = %q", cfg.LogLevel)
	}
}

func TestEnvOverridesFileConfig(t *testing.T) {
	cfg := loadWithFile(func(k string) string {
		if k == "HTTP_ADDR" {
			return ":9999"
		}
		return ""
	}, fileConfig{HTTPAddr: ":8787"})
	if cfg.HTTPAddr != ":9999" {
		t.Fatalf("HTTPAddr = %q", cfg.HTTPAddr)
	}
}
