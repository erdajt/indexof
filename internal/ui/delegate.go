// SPDX-FileCopyrightText: 2026 erdajt <erdajtsopjani.tech@gmail.com>
// SPDX-License-Identifier: GPL-3.0-or-later

package ui

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/erdajt/indexof/internal/listing"
)

type entryItem struct {
	entry listing.Entry
}

func (i entryItem) Title() string       { return i.entry.Name }
func (i entryItem) Description() string { return i.entry.Size }
func (i entryItem) FilterValue() string { return i.entry.Name }

type itemDelegate struct {
	styles      styles
	moveToCwd   bool
	downloadDir string
	cwd         string
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) dest(name string) string {
	if d.moveToCwd {
		return filepath.Join(d.cwd, name)
	}
	return filepath.Join(d.downloadDir, name)
}

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(entryItem)
	if !ok {
		return
	}

	name := i.entry.Name
	icon := "📄  "
	fg := d.styles.File
	if i.entry.IsDir {
		icon = "📁  "
		fg = d.styles.Dir
		name += "/"
	}

	local := ""
	if !i.entry.IsDir {
		if st, err := os.Stat(d.dest(name)); err == nil && !st.IsDir() {
			local = "✓ "
		}
	}

	size := i.entry.Size
	if size == "" {
		size = "-"
	}

	left := local + icon + name
	right := size
	width := m.Width()
	gap := width - lipgloss.Width(left) - lipgloss.Width(right) - 4
	if gap < 1 {
		gap = 1
	}
	line := left + strings.Repeat(" ", gap) + right

	if index == m.Index() {
		line = lipgloss.NewStyle().
			Foreground(d.styles.SelFg).
			Background(d.styles.SelBg).
			Bold(true).
			Padding(0, 1).
			Render("› " + line)
	} else {
		line = lipgloss.NewStyle().
			Foreground(fg).
			Padding(0, 1).
			Render("  " + line)
	}
	fmt.Fprint(w, line)
}
