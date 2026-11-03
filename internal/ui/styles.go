package ui

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/erdajt/indexof/internal/theme"
)

type styles struct {
	Title   lipgloss.Style
	Status  lipgloss.Style
	Help    lipgloss.Style
	Error   lipgloss.Style
	Ok      lipgloss.Style
	Frame   lipgloss.Style
	Accent  lipgloss.Color
	Subtle  lipgloss.Color
	Text    lipgloss.Color
	Dir     lipgloss.Color
	File    lipgloss.Color
	SelBg   lipgloss.Color
	SelFg   lipgloss.Color
	Warning lipgloss.Color
}

func newStyles(t theme.Theme) styles {
	return styles{
		Accent:  t.Accent,
		Subtle:  t.Subtle,
		Text:    t.Text,
		Dir:     t.Dir,
		File:    t.File,
		SelBg:   t.SelBg,
		SelFg:   t.SelFg,
		Warning: t.Warning,
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("230")).
			Background(t.Accent).
			Padding(0, 1),
		Status: lipgloss.NewStyle().Foreground(t.Subtle),
		Help:   lipgloss.NewStyle().Foreground(t.Subtle),
		Error:  lipgloss.NewStyle().Foreground(t.Error).Bold(true),
		Ok:     lipgloss.NewStyle().Foreground(t.Ok),
		Frame: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(t.Border).
			Padding(0, 1),
	}
}
