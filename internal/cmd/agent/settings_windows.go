//go:build windows

package main

import (
	"github.com/henrygd/beszel/agent/trayconfig"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// showSettingsDialog opens a native form to edit tray agent settings on the walk UI thread.
func showSettingsDialog(owner walk.Form, initial trayconfig.Config) (trayconfig.Config, bool) {
	var dlg *walk.Dialog
	var acceptPB, cancelPB *walk.PushButton
	var hubEdit, tokenEdit, portEdit, listenEdit, logEdit *walk.LineEdit
	var keyEdit *walk.TextEdit
	var committed trayconfig.Config
	var didCommit bool

	cfg := initial
	cfg.Normalize()

	readForm := func() trayconfig.Config {
		out := trayconfig.Config{
			HubURL:   hubEdit.Text(),
			Token:    tokenEdit.Text(),
			Key:      keyEdit.Text(),
			Port:     portEdit.Text(),
			Listen:   listenEdit.Text(),
			LogLevel: logEdit.Text(),
		}
		out.Normalize()
		return out
	}

	msgOwner := func() walk.Form {
		if dlg != nil {
			return dlg
		}
		return owner
	}

	commitAndAccept := func() {
		out := readForm()
		if err := out.Validate(); err != nil {
			walk.MsgBox(msgOwner(), "配置无效", err.Error(), walk.MsgBoxIconWarning)
			return
		}
		committed = out
		didCommit = true
		dlg.Accept()
	}

	cmd, err := Dialog{
		AssignTo:      &dlg,
		Title:         "Beszel Agent 设置",
		MinSize:       Size{Width: 480, Height: 420},
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		Layout:        VBox{Margins: Margins{Left: 12, Top: 12, Right: 12, Bottom: 12}},
		Children: []Widget{
			Label{Text: "Hub 地址 (HUB_URL)"},
			LineEdit{AssignTo: &hubEdit, Text: cfg.HubURL},
			Label{Text: "令牌 (TOKEN)"},
			LineEdit{AssignTo: &tokenEdit, Text: cfg.Token, PasswordMode: true},
			Label{Text: "SSH 公钥 (KEY，单行粘贴即可)"},
			TextEdit{
				AssignTo: &keyEdit,
				Text:     cfg.Key,
				MinSize:  Size{Width: 440, Height: 72},
				VScroll:  true,
			},
			Label{Text: "SSH 端口 (PORT，默认 45876)"},
			LineEdit{AssignTo: &portEdit, Text: cfg.Port},
			Label{Text: "监听地址 (LISTEN，可选，覆盖 PORT)"},
			LineEdit{AssignTo: &listenEdit, Text: cfg.Listen},
			Label{Text: "日志级别 (LOG_LEVEL，可选: debug / warn / error)"},
			LineEdit{AssignTo: &logEdit, Text: cfg.LogLevel},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo:  &acceptPB,
						Text:      "保存",
						OnClicked: commitAndAccept,
					},
					PushButton{
						AssignTo:  &cancelPB,
						Text:      "取消",
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Run(owner)

	if err != nil || cmd != walk.DlgCmdOK {
		return initial, false
	}

	// Enter 可能走系统默认 OK，未经过 OnClicked 时再读一次表单。
	if !didCommit {
		out := readForm()
		if err := out.Validate(); err != nil {
			walk.MsgBox(msgOwner(), "配置无效", err.Error(), walk.MsgBoxIconWarning)
			return initial, false
		}
		committed = out
	}

	if hubEdit == nil || keyEdit == nil {
		return initial, false
	}
	return committed, true
}
