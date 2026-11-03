package ui

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/erdajt/indexof/internal/config"
	"github.com/erdajt/indexof/internal/download"
	"github.com/erdajt/indexof/internal/listing"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-10)
		return m, nil

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd

	case tickMsg:
		for _, job := range m.manager.Jobs() {
			st, _, _, _, _, openAfter, dest := job.Snapshot()
			if st == download.StateDone {
				m.downloaded[dest] = struct{}{}
				if openAfter {
					job.ClearOpenAfter()
					return m, tea.Batch(
						tickCmd(),
						func() tea.Msg { return readyToOpenMsg{path: dest} },
					)
				}
			}
		}
		return m, tickCmd()

	case fetchedMsg:
		m.loading = false
		if msg.err != nil {
			m.err = msg.err.Error()
			return m, nil
		}
		m.err = ""
		m.base = msg.base
		items := make([]list.Item, 0, len(msg.entries))
		for _, e := range msg.entries {
			items = append(items, entryItem{entry: e})
		}
		m.list.SetItems(items)
		m.list.ResetFilter() // important: clear any stale filter
		m.list.Title = truncate(m.base.String(), m.width-10)
		if len(items) == 0 {
			m.status = "0 items — page had no parseable links"
		} else {
			m.status = fmt.Sprintf("%d items", len(items))
		}
		return m, nil

	case openDoneMsg:
		if msg.err != nil {
			m.status = msg.err.Error()
		} else {
			m.status = "editor closed"
		}
		return m, nil

	case readyToOpenMsg:
		c := exec.Command(m.cfg.Editor, msg.path)
		return m, tea.ExecProcess(c, func(err error) tea.Msg {
			return openDoneMsg{err: err}
		})

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := keyString(msg)

	if m.confirm != nil {
		return m.handleConfirm(key)
	}
	if m.loading {
		return m, nil
	}
	if key == keyForceQuit {
		m.cleanup()
		m.quitting = true
		return m, tea.Quit
	}
	if m.settingsView {
		return m.updateSettings(key)
	}
	if m.dlView {
		return m.updateDownloads(key)
	}

	// count prefix
	if (len(key) == 1 && key[0] >= '1' && key[0] <= '9') || (key == "0" && m.countBuf != "") {
		m.countBuf += key
		return m, nil
	}

	if m.visual {
		return m.handleVisual(key)
	}

	if m.list.FilterState() == list.Filtering {
		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		return m, cmd
	}

	return m.handleNormal(key)
}

func (m model) handleConfirm(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "y", "Y", keyEnter:
		e := *m.confirm
		openAfter := m.confirmAction == "open"
		m.confirm = nil
		m.startDownload(e, openAfter)
		m.status = "download started"
	case "n", "N", keyEsc, keyQuit:
		m.confirm = nil
		m.status = "cancelled"
	}
	return m, nil
}

func (m model) handleVisual(key string) (tea.Model, tea.Cmd) {
	switch key {
	case keyEsc:
		m.visual = false
		m.status = "normal"
		m.clearCount()
	case keyYank:
		start, end := m.visualStart, m.list.Index()
		if start > end {
			start, end = end, start
		}
		var lines []string
		items := m.list.Items()
		for i := start; i <= end && i < len(items); i++ {
			if it, ok := items[i].(entryItem); ok {
				lines = append(lines, it.entry.URL)
			}
		}
		_ = clipboard.WriteAll(strings.Join(lines, "\n"))
		m.visual = false
		m.status = fmt.Sprintf("yanked %d item(s)", len(lines))
		m.clearCount()
	case keyDown, "down":
		for i := 0; i < m.getCount(); i++ {
			m.list.CursorDown()
		}
		m.clearCount()
	case keyUp, "up":
		for i := 0; i < m.getCount(); i++ {
			m.list.CursorUp()
		}
		m.clearCount()
	}
	return m, nil
}

func (m model) handleNormal(key string) (tea.Model, tea.Cmd) {
	switch key {
	case keyQuit:
		m.cleanup()
		m.quitting = true
		return m, tea.Quit
	case keySettings:
		m.settingsView = true
		m.settingsCursor = 0
		return m, nil
	case keyDownloads:
		m.dlView = true
		return m, nil
	case keyVisual:
		m.visual = true
		m.visualStart = m.list.Index()
		m.status = "VISUAL"
		return m, nil
	case keyTop:
		m.list.Select(0)
		m.clearCount()
		return m, nil
	case keyBottom, keyEnd:
		if n := len(m.list.Items()); n > 0 {
			m.list.Select(n - 1)
		}
		m.clearCount()
		return m, nil
	case keyDown, "down":
		for i := 0; i < m.getCount(); i++ {
			m.list.CursorDown()
		}
		m.clearCount()
		return m, nil
	case keyUp, "up":
		for i := 0; i < m.getCount(); i++ {
			m.list.CursorUp()
		}
		m.clearCount()
		return m, nil
	case keyLeft, "left", keyBackspace:
		if len(m.history) > 1 {
			m.history = m.history[:len(m.history)-1]
			m.base = m.history[len(m.history)-1]
			m.loading = true
			m.clearCount()
			return m, m.fetch(m.base)
		}
	case keyRight, "right", keyEnter:
		return m.enterSelected()
	case keyDownload:
		return m.downloadSelected()
	case keyReload:
		m.loading = true
		m.err = ""
		return m, m.fetch(m.base)
	case keyToggleCwd:
		m.moveToCwd = !m.moveToCwd
		m.cfg.MoveToCwd = m.moveToCwd
		_ = config.Save(m.cfg)
		m.refreshDelegate()
		if m.moveToCwd {
			m.status = "target: current directory"
		} else {
			m.status = "target: " + m.cfg.DownloadDir
		}
	case keyToggleDel:
		m.deleteOnExit = !m.deleteOnExit
		m.cfg.DeleteOnExit = m.deleteOnExit
		_ = config.Save(m.cfg)
		if m.deleteOnExit {
			m.status = "delete on exit: on"
		} else {
			m.status = "delete on exit: off"
		}
	}
	return m, nil
}

func (m model) enterSelected() (tea.Model, tea.Cmd) {
	i, ok := m.list.SelectedItem().(entryItem)
	if !ok {
		return m, nil
	}
	if i.entry.IsDir {
		u, err := url.Parse(i.entry.URL)
		if err != nil {
			return m, nil
		}
		m.history = append(m.history, u)
		m.base = u
		m.loading = true
		return m, m.fetch(u)
	}

	name := filepath.Base(i.entry.Name)
	if name == "" || name == "." || name == ".." {
		name = "file"
	}
	dest := m.manager.DestPath(m.cfg.DownloadDir, m.moveToCwd, name)

	// Already on disk → just open
	if st, err := os.Stat(dest); err == nil && !st.IsDir() && st.Size() > 0 {
		m.downloaded[dest] = struct{}{}
		return m, func() tea.Msg { return readyToOpenMsg{path: dest} }
	}

	if m.isBig(i.entry) {
		m.confirm = &i.entry
		m.confirmAction = "open"
		return m, nil
	}

	m.startDownload(i.entry, true)
	m.status = "downloading…"
	return m, nil
}

func (m model) downloadSelected() (tea.Model, tea.Cmd) {
	i, ok := m.list.SelectedItem().(entryItem)
	if !ok || i.entry.IsDir {
		return m, nil
	}

	name := filepath.Base(i.entry.Name)
	dest := m.manager.DestPath(m.cfg.DownloadDir, m.moveToCwd, name)

	// Already downloaded → don't retry
	if st, err := os.Stat(dest); err == nil && !st.IsDir() && st.Size() > 0 {
		m.downloaded[dest] = struct{}{}
		m.status = "already downloaded: " + dest
		return m, nil
	}

	if m.isBig(i.entry) {
		m.confirm = &i.entry
		m.confirmAction = "download"
		return m, nil
	}

	m.startDownload(i.entry, false)
	m.status = "download started"
	return m, nil
}

func (m model) fetch(u *url.URL) tea.Cmd {
	return func() tea.Msg {
		if u == nil {
			return fetchedMsg{err: fmt.Errorf("invalid url")}
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(m.cfg.TimeoutSec)*time.Second)
		defer cancel()
		body, code, err := m.client.Get(ctx, u.String())
		if err != nil {
			return fetchedMsg{err: err}
		}
		if code == 404 {
			return fetchedMsg{err: fmt.Errorf("404 not found")}
		}
		if code >= 400 {
			return fetchedMsg{err: fmt.Errorf("http %d", code)}
		}
		entries, err := listing.Parse(u, body, m.cfg.ShowHidden)
		return fetchedMsg{entries: entries, err: err, base: u}
	}
}

func (m *model) startDownload(e listing.Entry, openAfter bool) {
	name := filepath.Base(e.Name)
	if name == "" || name == "." || name == ".." {
		name = "file"
	}
	dest := m.manager.DestPath(m.cfg.DownloadDir, m.moveToCwd, name)
	m.manager.Start(e.URL, name, dest, e.SizeBytes, openAfter)
}

func (m model) isBig(e listing.Entry) bool {
	if e.SizeBytes <= 0 {
		return false
	}
	return e.SizeBytes > int64(m.cfg.BigFileThresholdMB)*1024*1024
}

func (m model) getCount() int {
	if m.countBuf == "" {
		return 1
	}
	n, err := strconv.Atoi(m.countBuf)
	if err != nil || n < 1 {
		return 1
	}
	if n > 9999 {
		n = 9999
	}
	return n
}

func (m *model) clearCount() { m.countBuf = "" }

func (m *model) refreshDelegate() {
	m.list.SetDelegate(itemDelegate{
		styles:      m.styles,
		moveToCwd:   m.moveToCwd,
		downloadDir: m.cfg.DownloadDir,
		cwd:         m.cwd,
	})
}

func (m model) cleanup() {
	if !m.deleteOnExit {
		return
	}
	for path := range m.downloaded {
		_ = os.Remove(path)
	}
}
