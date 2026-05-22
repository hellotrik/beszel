//go:build windows

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/getlantern/systray"
	"github.com/henrygd/beszel/agent"
	"github.com/henrygd/beszel/agent/trayconfig"
	"github.com/henrygd/beszel/agent/utils"
	"github.com/spf13/pflag"
	"golang.org/x/sys/windows"
)

var winSupervisor = newSupervisor()

func main() {
	if handleCLISubcommand() {
		return
	}
	if hasFlag("-service") || hasFlag("--service") {
		runWindowsServiceMode()
		return
	}
	if release, ok := acquireSingleInstance(); !ok {
		showInfoMessage("Beszel Agent", "Beszel Agent is already running.")
		os.Exit(0)
	} else {
		defer release()
	}
	systray.Run(onTrayReady, onTrayExit)
}

func runWindowsServiceMode() {
	agent.SupervisedMode = false
	cfg, err := trayconfig.Load()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.IsComplete() {
		if err := cfg.ApplyToEnv(); err != nil {
			log.Fatal(err)
		}
	} else {
		runWindowsServiceWithFlags()
		return
	}
	if err := startHeadlessFromConfig(cfg); err != nil {
		log.Fatal(err)
	}
}

func runWindowsServiceWithFlags() {
	var opts cmdOptions
	if parseWindowsCLIFlags(&opts) {
		return
	}
	if opts.hubURL != "" {
		os.Setenv("HUB_URL", opts.hubURL)
	}
	if opts.token != "" {
		os.Setenv("TOKEN", opts.token)
	}
	rc := agent.RuntimeConfigFromEnv()
	if err := agent.ApplyRuntimeConfig(rc); err != nil {
		log.Fatal(err)
	}
	serverConfig, err := buildServerOptionsFromEnv(opts)
	if err != nil {
		log.Fatal(err)
	}
	a, err := agent.NewAgent()
	if err != nil {
		log.Fatal(err)
	}
	if err := a.Start(serverConfig); err != nil {
		log.Fatal(err)
	}
}

type cmdOptions struct {
	key    string
	listen string
	hubURL string
	token  string
}

func parseWindowsCLIFlags(opts *cmdOptions) bool {
	pflag.StringVarP(&opts.key, "key", "k", "", "Public key(s) for SSH authentication")
	pflag.StringVarP(&opts.listen, "listen", "l", "", "Address or port to listen on")
	pflag.StringVarP(&opts.hubURL, "url", "u", "", "URL of the Beszel hub")
	pflag.StringVarP(&opts.token, "token", "t", "", "Token to use for authentication")
	help := pflag.BoolP("help", "h", false, "Show help")
	pflag.Parse()
	if *help {
		printCLIUsage()
		return true
	}
	return false
}

func buildServerOptionsFromEnv(opts cmdOptions) (agent.ServerOptions, error) {
	var serverConfig agent.ServerOptions
	var err error
	if opts.key != "" {
		serverConfig.Keys, err = agent.ParseKeys(opts.key)
	} else if key, ok := utils.GetEnv("KEY"); ok && key != "" {
		serverConfig.Keys, err = agent.ParseKeys(key)
	} else {
		keyFile, ok := utils.GetEnv("KEY_FILE")
		if !ok {
			return serverConfig, fmt.Errorf("no KEY provided")
		}
		pubKey, readErr := os.ReadFile(keyFile)
		if readErr != nil {
			return serverConfig, readErr
		}
		serverConfig.Keys, err = agent.ParseKeys(string(pubKey))
	}
	if err != nil {
		return serverConfig, err
	}
	addr := agent.GetAddress(opts.listen)
	serverConfig.Addr = addr
	serverConfig.Network = agent.GetNetwork(addr)
	return serverConfig, nil
}

func startHeadlessFromConfig(cfg trayconfig.Config) error {
	keys, err := agent.ParseRuntimeKeys(cfg.RuntimeConfig())
	if err != nil {
		return err
	}
	a, err := agent.NewAgent()
	if err != nil {
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
	return a.Start(opts)
}

func onTrayReady() {
	systray.SetIcon(trayIconData)
	systray.SetTitle("Beszel Agent")
	systray.SetTooltip("Beszel Agent " + getBeszelVersion())

	mStatus := systray.AddMenuItem("Status: Starting…", "Connection status")
	mStatus.Disable()
	systray.AddSeparator()
	mStart := systray.AddMenuItem("Start", "Start agent")
	mStop := systray.AddMenuItem("Stop", "Stop agent")
	mSettings := systray.AddMenuItem("Settings…", "Edit configuration")
	mOpenCfg := systray.AddMenuItem("Open config.json", "Open in Notepad")
	mReload := systray.AddMenuItem("Reload configuration", "Reload config and restart")
	mOpenLogs := systray.AddMenuItem("Open logs folder", "Open log directory")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Exit Beszel Agent")

	go func() {
		_, _ = winSupervisor.LoadConfig()
		updateStatusMenu(mStatus)
		_ = winSupervisor.Start()
		updateStatusMenu(mStatus)
	}()

	go menuLoop(mStatus, mStart, mStop, mSettings, mOpenCfg, mReload, mOpenLogs, mQuit)
}

func menuLoop(mStatus *systray.MenuItem, mStart, mStop, mSettings, mOpenCfg, mReload, mOpenLogs, mQuit *systray.MenuItem) {
	for {
		select {
		case <-mStart.ClickedCh:
			_ = winSupervisor.Start()
			updateStatusMenu(mStatus)
		case <-mStop.ClickedCh:
			winSupervisor.Stop()
			updateStatusMenu(mStatus)
		case <-mSettings.ClickedCh:
			cfg, _ := trayconfig.Load()
			if saved, ok := showSettingsDialog(cfg); ok {
				_ = trayconfig.Save(saved)
				_ = winSupervisor.Reload()
				updateStatusMenu(mStatus)
			}
		case <-mOpenCfg.ClickedCh:
			openConfigFile()
		case <-mReload.ClickedCh:
			_ = winSupervisor.Reload()
			updateStatusMenu(mStatus)
		case <-mOpenLogs.ClickedCh:
			openLogsFolder()
		case <-mQuit.ClickedCh:
			winSupervisor.Stop()
			systray.Quit()
			return
		}
	}
}

func updateStatusMenu(mStatus *systray.MenuItem) {
	mStatus.SetTitle(winSupervisor.StatusLabel())
}

func onTrayExit() {
	winSupervisor.Stop()
}

func openConfigFile() {
	path, err := trayconfig.ConfigPath()
	if err != nil {
		showInfoMessage("Beszel Agent", err.Error())
		return
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg := trayconfig.Config{Port: trayconfig.DefaultPort}
		_ = trayconfig.Save(cfg)
	}
	_ = exec.Command("notepad.exe", path).Start()
}

func openLogsFolder() {
	dir := os.Getenv("ProgramData")
	if dir == "" {
		dir = `C:\ProgramData`
	}
	logDir := filepath.Join(dir, "beszel-agent", "logs")
	_ = os.MkdirAll(logDir, 0o755)
	_ = exec.Command("explorer.exe", logDir).Start()
}

func acquireSingleInstance() (release func(), ok bool) {
	name, _ := windows.UTF16PtrFromString(`Global\BeszelAgentTray`)
	h, err := windows.CreateMutex(nil, false, name)
	if err != nil {
		return func() {}, true
	}
	if windows.GetLastError() == windows.ERROR_ALREADY_EXISTS {
		windows.CloseHandle(h)
		return nil, false
	}
	return func() { windows.CloseHandle(h) }, true
}

func showInfoMessage(title, body string) {
	t, _ := windows.UTF16PtrFromString(title)
	b, _ := windows.UTF16PtrFromString(body)
	windows.MessageBox(0, b, t, windows.MB_OK|windows.MB_ICONINFORMATION)
}
