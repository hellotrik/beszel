//go:build windows

package main

import (
	"github.com/henrygd/beszel/agent/trayconfig"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// showSettingsDialog opens a native form to edit tray agent settings. Returns saved config and whether user saved.
func showSettingsDialog(initial trayconfig.Config) (trayconfig.Config, bool) {
	var dlg *walk.Dialog
	var acceptPB, cancelPB *walk.PushButton
	var hubEdit, tokenEdit, keyEdit, portEdit, listenEdit, logEdit *walk.LineEdit
	cfg := initial
	cfg.Normalize()

	cmd, err := Dialog{
		AssignTo:      &dlg,
		Title:         "Beszel Agent Settings",
		MinSize:       Size{Width: 440, Height: 340},
		DefaultButton: &acceptPB,
		CancelButton:  &cancelPB,
		Layout:        VBox{Margins: Margins{Left: 10, Top: 10, Right: 10, Bottom: 10}},
		Children: []Widget{
			Label{Text: "HUB_URL"},
			LineEdit{AssignTo: &hubEdit, Text: cfg.HubURL},
			Label{Text: "TOKEN"},
			LineEdit{AssignTo: &tokenEdit, Text: cfg.Token},
			Label{Text: "KEY (SSH public key)"},
			LineEdit{AssignTo: &keyEdit, Text: cfg.Key},
			Label{Text: "PORT (SSH fallback, default 45876)"},
			LineEdit{AssignTo: &portEdit, Text: cfg.Port},
			Label{Text: "LISTEN (optional, overrides PORT)"},
			LineEdit{AssignTo: &listenEdit, Text: cfg.Listen},
			Label{Text: "LOG_LEVEL (optional: debug, warn, error)"},
			LineEdit{AssignTo: &logEdit, Text: cfg.LogLevel},
			Composite{
				Layout: HBox{},
				Children: []Widget{
					HSpacer{},
					PushButton{
						AssignTo:  &acceptPB,
						Text:      "OK",
						OnClicked: func() { dlg.Accept() },
					},
					PushButton{
						AssignTo:  &cancelPB,
						Text:      "Cancel",
						OnClicked: func() { dlg.Cancel() },
					},
				},
			},
		},
	}.Run(nil)

	if err != nil || cmd != walk.DlgCmdOK || hubEdit == nil {
		return initial, false
	}

	out := trayconfig.Config{
		HubURL:   hubEdit.Text(),
		Token:    tokenEdit.Text(),
		Key:      keyEdit.Text(),
		Port:     portEdit.Text(),
		Listen:   listenEdit.Text(),
		LogLevel: logEdit.Text(),
	}
	out.Normalize()
	if err := out.Validate(); err != nil {
		walk.MsgBox(nil, "Invalid settings", err.Error(), walk.MsgBoxIconWarning)
		return initial, false
	}
	return out, true
}
