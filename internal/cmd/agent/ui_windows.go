//go:build windows

package main

import (
	"log/slog"
	"runtime"

	"github.com/henrygd/beszel/agent/trayconfig"
	"github.com/lxn/walk"
)

// runOnWalkUI runs fn on a dedicated OS thread for modal walk dialogs (no background host window).
func runOnWalkUI(fn func()) {
	done := make(chan struct{})
	go func() {
		runtime.LockOSThread()
		defer close(done)
		defer runtime.UnlockOSThread()
		defer func() {
			if r := recover(); r != nil {
				slog.Error("walk UI 异常", "err", r)
				showInfoMessage("Beszel Agent", "设置窗口打开失败，请稍后重试或编辑 config.json。")
			}
		}()
		walk.SetPanicOnError(false)
		fn()
	}()
	<-done
}

// openSettingsUI shows the settings dialog (lazy; does not start until user opens settings).
func openSettingsUI() (trayconfig.Config, bool) {
	var out trayconfig.Config
	var ok bool
	runOnWalkUI(func() {
		cfg, _ := trayconfig.Load()
		out, ok = showSettingsDialog(nil, cfg)
	})
	return out, ok
}

// promptFirstRunSettings notifies the user when config.json is incomplete (no modal walk at startup).
func promptFirstRunSettings() {
	cfg, err := trayconfig.Load()
	if err != nil {
		showInfoMessage("Beszel Agent", "读取配置失败: "+err.Error())
		return
	}
	if cfg.IsComplete() {
		return
	}
	showInfoMessage("Beszel Agent", "尚未配置。\n请右键托盘图标，选择「设置…」填写 Hub 地址、令牌和 SSH 公钥。")
}

// applyTrayConfig writes config.json, reloads the agent, and reports errors to the user.
func applyTrayConfig(cfg trayconfig.Config) {
	if err := trayconfig.Save(cfg); err != nil {
		showInfoMessage("Beszel Agent", "保存配置失败:\n"+err.Error())
		return
	}
	if err := winSupervisor.Reload(); err != nil {
		showInfoMessage("Beszel Agent", "配置已写入，但 Agent 启动失败:\n"+err.Error())
		return
	}
	if path, err := trayconfig.ConfigPath(); err == nil {
		showInfoMessage("Beszel Agent", "配置已保存:\n"+path)
	}
}
