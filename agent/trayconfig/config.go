// Package trayconfig persists Windows tray agent settings in config.json.
package trayconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/henrygd/beszel/agent"
	"github.com/henrygd/beszel/agent/utils"
)

const configFileName = "config.json"

// Config is stored beside agent data (e.g. %APPDATA%\beszel-agent\config.json).
type Config struct {
	HubURL   string `json:"hub_url"`
	Token    string `json:"token"`
	Key      string `json:"key"`
	Port     string `json:"port,omitempty"`
	Listen   string `json:"listen,omitempty"`
	LogLevel string `json:"log_level,omitempty"`
}

// DefaultPort is the default SSH listen port when unset.
const DefaultPort = "45876"

// ConfigPath returns the path to config.json under the agent data directory.
func ConfigPath() (string, error) {
	dataDir, err := agent.GetDataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dataDir, configFileName), nil
}

// Load reads config.json or migrates from environment variables if missing.
func Load() (Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := migrateFromEnv()
			if cfg.HubURL != "" && cfg.Token != "" && cfg.Key != "" {
				_ = Save(cfg)
			}
			return cfg, nil
		}
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	cfg.Normalize()
	return cfg, nil
}

func migrateFromEnv() Config {
	cfg := Config{Port: DefaultPort}
	rc := agent.RuntimeConfigFromEnv()
	cfg.HubURL = rc.HubURL
	cfg.Token = rc.Token
	cfg.Key = rc.Key
	cfg.Listen = rc.Listen
	if rc.Port != "" {
		cfg.Port = rc.Port
	}
	cfg.LogLevel = rc.LogLevel
	cfg.Normalize()
	return cfg
}

// Normalize trims fields and applies defaults (safe to call before save/validate).
func (c *Config) Normalize() {
	c.HubURL = strings.TrimSpace(c.HubURL)
	c.Token = strings.TrimSpace(c.Token)
	c.Key = strings.TrimSpace(c.Key)
	c.Listen = strings.TrimSpace(c.Listen)
	c.Port = strings.TrimSpace(c.Port)
	c.LogLevel = strings.TrimSpace(c.LogLevel)
	if c.Port == "" {
		c.Port = DefaultPort
	}
}

// Save writes config.json.
func Save(cfg Config) error {
	cfg.Normalize()
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// ApplyToEnv applies config via agent.ApplyRuntimeConfig.
func (c Config) ApplyToEnv() error {
	return agent.ApplyRuntimeConfig(c.RuntimeConfig())
}

// RuntimeConfig converts to agent.RuntimeConfig.
func (c Config) RuntimeConfig() agent.RuntimeConfig {
	return agent.RuntimeConfig{
		HubURL:   c.HubURL,
		Token:    c.Token,
		Key:      c.Key,
		Listen:   c.Listen,
		Port:     c.Port,
		LogLevel: c.LogLevel,
	}
}

// Validate checks required fields without applying env.
func (c Config) Validate() error {
	return c.ApplyToEnv()
}

// IsComplete returns true if required fields are non-empty.
func (c Config) IsComplete() bool {
	c.Normalize()
	return c.HubURL != "" && c.Token != "" && c.Key != ""
}

// EnvOverridesFromFile returns true if config.json exists (used for precedence messaging).
func EnvOverridesFromFile() bool {
	path, err := ConfigPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

// LoadOrEnv loads config.json when present; otherwise builds from environment only.
func LoadOrEnv() (Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		cfg := migrateFromEnv()
		if !cfg.IsComplete() {
			return cfg, nil
		}
		return cfg, nil
	}
	return Load()
}

// HasEnvConfig returns true when HUB_URL, TOKEN, and KEY are set in the environment.
func HasEnvConfig() bool {
	hub, _ := utils.GetEnv("HUB_URL")
	tok, _ := utils.GetEnv("TOKEN")
	key, _ := utils.GetEnv("KEY")
	return strings.TrimSpace(hub) != "" && strings.TrimSpace(tok) != "" && strings.TrimSpace(key) != ""
}
