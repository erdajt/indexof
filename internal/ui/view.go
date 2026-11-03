package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m model) View() string {
	if m.quitting {
		return ""
	}

	if m.confirm != nil {
		return m.viewConfirm()
	}
	if m.settingsView {
		return m.viewSettings()
	}
	if m.dlView {
		return m.viewDownloads()
	}

	var b strings.Builder
	header := m.styles.Title.Render(" indexof ") + "  " +
		lipgloss.NewStyle().Foreground(m.styles.Subtle).Render(truncate(m.base.String(), m.width-24))
	if m.visual {
		header += "  " + lipgloss.NewStyle().Bold(true).Foreground(m.styles.Warning).Render("VISUAL")
	}
	if m.countBuf != "" {
		header += "  " + lipgloss.NewStyle().Foreground(m.styles.Accent).Render(m.countBuf)
	}
	b.WriteString(header + "\n\n")

	if m.loading {
		b.WriteString("  " + m.spinner.View() + "  fetching…\n")
		return b.String()
	}
	if m.err != "" {
		b.WriteString(m.styles.Error.Render("  "+m.err) + "\n\n")
		b.WriteString(m.styles.Help.Render("  r retry   q quit") + "\n")
		return b.String()
	}

	b.WriteString(m.styles.Frame.Width(m.width - 2).Render(m.list.View()))
	b.WriteString("\n")

	active := m.manager.ActiveCount()
	del := "on"
	if !m.deleteOnExit {
		del = "off"
	}
	tgt := "tmp"
	if m.moveToCwd {
		tgt = "cwd"
	}
	mode := lipgloss.NewStyle().Foreground(m.styles.Accent).Render(
		fmt.Sprintf("[del:%s  %s  %s]", del, tgt, m.cfg.Theme))

	dlInfo := ""
	if active > 0 {
		dlInfo = lipgloss.NewStyle().Foreground(m.styles.Warning).Render(fmt.Sprintf("  ↓%d", active))
	}

	b.WriteString("  " + m.styles.Status.Render(m.status) + "  " + mode + dlInfo + "\n")
	b.WriteString(m.styles.Help.Render("  hjkl  gg/G  10j  v y  d  p  s settings  r  /  q"))
	return b.String()
}

func (m model) viewConfirm() string {
	e := m.confirm
	size := e.Size
	if size == "" {
		size = humanBytes(e.SizeBytes)
	}
	msg := fmt.Sprintf("⚠  Big file: %s (%s)\n\n   Threshold: %d MB\n\n   Download?  [y] yes   [n] no",
		e.Name, size, m.cfg.BigFileThresholdMB)
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.styles.Warning).
		Padding(1, 3).
		Render(msg)
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, box)
}
