package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/erdajt/indexof/internal/download"
)

func (m model) updateDownloads(key string) (tea.Model, tea.Cmd) {
	jobs := m.manager.Jobs()
	switch key {
	case keyDownloads, keyEsc:
		m.dlView = false
	case keyDown, "down":
		if m.dlCursor < len(jobs)-1 {
			m.dlCursor++
		}
	case keyUp, "up":
		if m.dlCursor > 0 {
			m.dlCursor--
		}
	case keyCancel:
		if m.dlCursor < len(jobs) {
			m.manager.Cancel(jobs[m.dlCursor])
		}
	case keyPause:
		if m.dlCursor < len(jobs) {
			job := jobs[m.dlCursor]
			st := job.GetState()
			if st == download.StateRunning {
				m.manager.Pause(job)
			} else if st == download.StatePaused {
				m.manager.Resume(job)
			}
		}
	case keyClear:
		m.manager.ClearFinished()
		if m.dlCursor >= len(m.manager.Jobs()) {
			m.dlCursor = len(m.manager.Jobs()) - 1
		}
		if m.dlCursor < 0 {
			m.dlCursor = 0
		}
	}
	return m, nil
}

func (m model) viewDownloads() string {
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(m.styles.Accent).Render("  Downloads") + "\n\n")

	jobs := m.manager.Jobs()
	if len(jobs) == 0 {
		b.WriteString(m.styles.Help.Render("  No downloads yet") + "\n")
	} else {
		for i, job := range jobs {
			st, written, total, name, errStr := job.Snapshot()

			cursor := "  "
			if i == m.dlCursor {
				cursor = "› "
			}

			pct := 0.0
			if total > 0 {
				pct = float64(written) / float64(total)
			}
			bar := progressBar(pct, 28, m.styles.Accent)
			info := fmt.Sprintf("%s  %s/%s  %s", bar, humanBytes(written), humanBytes(total), st.String())

			line := cursor + name
			if st == download.StateError {
				line += "  " + m.styles.Error.Render(errStr)
			}

			style := lipgloss.NewStyle()
			if i == m.dlCursor {
				style = style.Bold(true).Foreground(m.styles.SelFg)
			}
			b.WriteString(style.Render(line) + "\n")
			b.WriteString("    " + info + "\n\n")
		}
	}
	b.WriteString(m.styles.Help.Render("  j/k  space pause/resume  x cancel  c clear  p/esc back"))
	return b.String()
}

func progressBar(pct float64, width int, accent lipgloss.Color) string {
	if width < 8 {
		width = 8
	}
	filled := int(pct * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	bar := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	return lipgloss.NewStyle().Foreground(accent).Render(bar)
}
