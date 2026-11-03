package ui

import "github.com/charmbracelet/bubbletea"

const (
	keyQuit      = "q"
	keyForceQuit = "ctrl+c"
	keyReload    = "r"
	keyDownloads = "p"
	keySettings  = "s"
	keyVisual    = "v"
	keyYank      = "y"
	keyEsc       = "esc"
	keyEnter     = "enter"
	keyDown      = "j"
	keyUp        = "k"
	keyLeft      = "h"
	keyRight     = "l"
	keyTop       = "gg"
	keyBottom    = "G"
	keyEnd       = "$"
	keyDownload  = "d"
	keyToggleCwd = "c"
	keyToggleDel = "K"
	keyPause     = " "
	keyCancel    = "x"
	keyClear     = "c"
	keyBackspace = "backspace"
)

func keyString(msg tea.KeyMsg) string {
	return msg.String()
}
