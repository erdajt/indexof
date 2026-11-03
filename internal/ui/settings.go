package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/erdajt/indexof/internal/config"
	"github.com/erdajt/indexof/internal/theme"
)

func (m model) updateSettings(key string) (tea.Model, tea.Cmd) {
	switch key {
	case keyEsc, keyQuit, keySettings:
		m.settingsView = false
		return m, nil
	case keyDown, "down":
		if m.settingsCursor < 8 {
			m.settingsCursor++
		}
	case keyUp, "up":
		if m.settingsCursor > 0 {
			m.settingsCursor--
		}
	case keyEnter, keyRight:
		switch m.settingsCursor {
		case 0:
			editors := []string{"nvim", "vim", "nano", "code", "hx"}
			for i, e := range editors {
				if e == m.cfg.Editor {
					m.cfg.Editor = editors[(i+1)%len(editors)]
					break
				}
			}
		case 1:
			names := theme.Names()
			for i, n := range names {
				if n == m.cfg.Theme {
					m.cfg.Theme = names[(i+1)%len(names)]
					break
				}
			}
			m.styles = newStyles(theme.Get(m.cfg.Theme))
			m.refreshDelegate()
		case 2:
			m.cfg.Retries = m.cfg.Retries%10 + 1
		case 3:
			m.cfg.TimeoutSec = ((m.cfg.TimeoutSec/5)%12 + 1) * 5
		case 4:
			m.cfg.ShowHidden = !m.cfg.ShowHidden
		case 5:
			m.cfg.DeleteOnExit = !m.cfg.DeleteOnExit
			m.deleteOnExit = m.cfg.DeleteOnExit
		case 6:
			m.cfg.MoveToCwd = !m.cfg.MoveToCwd
			m.moveToCwd = m.cfg.MoveToCwd
			m.refreshDelegate()
		case 7:
			m.cfg.BigFileThresholdMB = ((m.cfg.BigFileThresholdMB/10)%20 + 1) * 10
		}
		_ = config.Save(m.cfg)
		m.status = "config saved"
	}
	return m, nil
}

func (m model) viewSettings() string {
	var b strings.Builder
	b.WriteString(m.styles.Title.Render(" settings "))
	b.WriteString("\n\n")

	items := []struct{ name, val string }{
		{"Editor", m.cfg.Editor},
		{"Theme", m.cfg.Theme},
		{"Retries", fmt.Sprintf("%d", m.cfg.Retries)},
		{"Timeout (sec)", fmt.Sprintf("%d", m.cfg.TimeoutSec)},
		{"Show hidden", fmt.Sprintf("%v", m.cfg.ShowHidden)},
		{"Delete on exit", fmt.Sprintf("%v", m.cfg.DeleteOnExit)},
		{"Move to cwd", fmt.Sprintf("%v", m.cfg.MoveToCwd)},
		{"Big file threshold (MB)", fmt.Sprintf("%d", m.cfg.BigFileThresholdMB)},
		{"Download dir", m.cfg.DownloadDir},
	}

	for i, it := range items {
		cursor := "  "
		style := lipgloss.NewStyle().Foreground(m.styles.Text)
		if i == m.settingsCursor {
			cursor = "› "
			style = style.Bold(true).Foreground(m.styles.SelFg).Background(m.styles.SelBg)
		}
		b.WriteString(style.Render(fmt.Sprintf("%s%-28s %s", cursor, it.name, it.val)))
		b.WriteByte('\n')
	}
	b.WriteByte('\n')
	b.WriteString(m.styles.Help.Render("  j/k move   enter cycle/toggle   s/esc back   (auto-saves)"))
	return b.String()
}
