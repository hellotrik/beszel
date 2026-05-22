package agent

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/henrygd/beszel/agent/utils"
	"golang.org/x/crypto/ssh"
)

// RuntimeConfig holds agent connection settings applied to the process environment.
type RuntimeConfig struct {
	HubURL   string
	Token    string
	Key      string
	Listen   string
	Port     string
	LogLevel string
}

// SupervisedMode indicates the agent runs under a supervisor (e.g. Windows tray) and must not os.Exit on recoverable errors.
var SupervisedMode bool

// ApplyRuntimeConfig validates settings and sets environment variables used by the agent.
func ApplyRuntimeConfig(c RuntimeConfig) error {
	c.HubURL = strings.TrimSpace(c.HubURL)
	c.Token = strings.TrimSpace(c.Token)
	c.Key = strings.TrimSpace(c.Key)
	c.Listen = strings.TrimSpace(c.Listen)
	c.Port = strings.TrimSpace(c.Port)
	c.LogLevel = strings.TrimSpace(c.LogLevel)

	if c.HubURL == "" {
		return errors.New("HUB_URL is required")
	}
	u, err := url.Parse(c.HubURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return fmt.Errorf("invalid HUB_URL: %q", c.HubURL)
	}
	if c.Token == "" {
		return errors.New("TOKEN is required")
	}
	if len(c.Token) > 64 {
		return errors.New("TOKEN must be at most 64 characters")
	}
	if c.Key == "" {
		return errors.New("KEY is required")
	}

	_ = os.Setenv("HUB_URL", c.HubURL)
	_ = os.Setenv("TOKEN", c.Token)
	_ = os.Setenv("KEY", c.Key)
	if c.Listen != "" {
		_ = os.Setenv("LISTEN", c.Listen)
	} else {
		_ = os.Unsetenv("LISTEN")
	}
	if c.Port != "" {
		_ = os.Setenv("PORT", c.Port)
	} else {
		_ = os.Unsetenv("PORT")
	}
	if c.LogLevel != "" {
		_ = os.Setenv("LOG_LEVEL", c.LogLevel)
	} else {
		_ = os.Unsetenv("LOG_LEVEL")
	}
	return nil
}

// RuntimeConfigFromEnv builds RuntimeConfig from current environment variables.
func RuntimeConfigFromEnv() RuntimeConfig {
	c := RuntimeConfig{}
	if v, ok := utils.GetEnv("HUB_URL"); ok {
		c.HubURL = v
	}
	if v, ok := utils.GetEnv("TOKEN"); ok {
		c.Token = v
	}
	if v, ok := utils.GetEnv("KEY"); ok {
		c.Key = v
	}
	if v, ok := utils.GetEnv("LISTEN"); ok {
		c.Listen = v
	}
	if v, ok := utils.GetEnv("PORT"); ok {
		c.Port = v
	}
	if v, ok := utils.GetEnv("LOG_LEVEL"); ok {
		c.LogLevel = v
	}
	return c
}

// ParseRuntimeKeys parses SSH public keys from RuntimeConfig.Key.
func ParseRuntimeKeys(c RuntimeConfig) ([]ssh.PublicKey, error) {
	return ParseKeys(c.Key)
}

// Run starts the agent with the given context and server options. Cancel ctx to stop.
func (a *Agent) Run(ctx context.Context, serverOptions ServerOptions) error {
	a.keys = serverOptions.Keys
	return a.connectionManager.StartWithContext(ctx, serverOptions)
}

// Stop shuts down active connections. Call after Run returns or to stop a running agent.
func (a *Agent) Stop() error {
	if a.connectionManager == nil {
		return nil
	}
	return a.connectionManager.stop()
}

// ConnectionState returns the current hub connection state.
func (a *Agent) ConnectionState() ConnectionState {
	if a.connectionManager == nil {
		return Disconnected
	}
	return a.connectionManager.State
}
