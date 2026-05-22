//go:build windows

package main

import (
	"context"
	"log/slog"
	"sync"

	"github.com/henrygd/beszel/agent"
	"github.com/henrygd/beszel/agent/trayconfig"
)

type agentSupervisor struct {
	mu     sync.Mutex
	agent  *agent.Agent
	cancel context.CancelFunc
	cfg    trayconfig.Config
	errMsg string
}

func newSupervisor() *agentSupervisor {
	return &agentSupervisor{}
}

func (s *agentSupervisor) StatusLabel() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.errMsg != "" {
		return "Status: Error — " + s.errMsg
	}
	if !s.cfg.IsComplete() {
		return "Status: Missing configuration"
	}
	if s.agent == nil {
		return "Status: Stopped"
	}
	switch s.agent.ConnectionState() {
	case agent.WebSocketConnected:
		return "Status: Connected (WebSocket)"
	case agent.SSHConnected:
		return "Status: Connected (SSH)"
	default:
		return "Status: Disconnected"
	}
}

func (s *agentSupervisor) LoadConfig() (trayconfig.Config, error) {
	cfg, err := trayconfig.Load()
	if err != nil {
		return cfg, err
	}
	s.mu.Lock()
	s.cfg = cfg
	s.mu.Unlock()
	return cfg, nil
}

func (s *agentSupervisor) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cancel != nil {
		s.cancel()
		s.cancel = nil
	}
	if s.agent != nil {
		_ = s.agent.Stop()
		s.agent = nil
	}
	s.errMsg = ""
}

func (s *agentSupervisor) Start() error {
	s.Stop()
	cfg, err := trayconfig.Load()
	if err != nil {
		s.mu.Lock()
		s.errMsg = err.Error()
		s.mu.Unlock()
		return err
	}
	if !cfg.IsComplete() {
		s.mu.Lock()
		s.cfg = cfg
		s.errMsg = "HUB_URL, TOKEN, and KEY are required in config.json"
		s.mu.Unlock()
		return nil
	}
	if err := cfg.ApplyToEnv(); err != nil {
		s.mu.Lock()
		s.errMsg = err.Error()
		s.mu.Unlock()
		return err
	}
	keys, err := agent.ParseRuntimeKeys(cfg.RuntimeConfig())
	if err != nil {
		s.mu.Lock()
		s.errMsg = err.Error()
		s.mu.Unlock()
		return err
	}
	a, err := agent.NewAgent()
	if err != nil {
		s.mu.Lock()
		s.errMsg = err.Error()
		s.mu.Unlock()
		return err
	}
	addr := agent.GetAddress(cfg.Listen)
	if cfg.Listen == "" {
		addr = agent.GetAddress(cfg.Port)
	}
	opts := agent.ServerOptions{
		Addr:    addr,
		Network: agent.GetNetwork(addr),
		Keys:    keys,
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.agent = a
	s.cancel = cancel
	s.cfg = cfg
	s.errMsg = ""
	s.mu.Unlock()

	agent.SupervisedMode = true
	go func() {
		err := a.Run(ctx, opts)
		agent.SupervisedMode = false
		s.mu.Lock()
		if err != nil && context.Cause(ctx) != context.Canceled {
			s.errMsg = err.Error()
			slog.Error("Agent stopped", "err", err)
		}
		s.agent = nil
		s.cancel = nil
		s.mu.Unlock()
	}()
	return nil
}

func (s *agentSupervisor) Reload() error {
	s.Stop()
	return s.Start()
}
