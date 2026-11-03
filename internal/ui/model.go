package ui

import (
	"net/url"
	"os"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/erdajt/indexof/internal/config"
	"github.com/erdajt/indexof/internal/download"
	"github.com/erdajt/indexof/internal/netclient"
	"github.com/erdajt/indexof/internal/listing"
	"github.com/erdajt/indexof/internal/theme"
)

type model struct {
	cfg      config.Config
	client   *netclient.Client
	manager  *download.Manager
	styles   styles
	list     list.Model
	spinner  spinner.Model
	loading  bool
	err      string
	status   string
	base     *url.URL
	history  []*url.URL
	quitting bool
	width    int
	height   int

	deleteOnExit bool
	moveToCwd    bool
	downloaded   map[string]struct{}
	cwd          string

	dlView   bool
	dlCursor int

	confirm       *listing.Entry
	confirmAction string // "open" | "download"

	countBuf    string
	visual      bool
	visualStart int

	settingsView   bool
	settingsCursor int
}

type fetchedMsg struct {
	entries []listing.Entry
	err     error
	base    *url.URL
}

type openDoneMsg struct{ err error }
type readyToOpenMsg struct{ path string }
type tickMsg struct{}

func newModel(startURL string, cfg config.Config) (model, error) {
	th := theme.Get(cfg.Theme)
	st := newStyles(th)

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(st.Accent)

	cwd, _ := os.Getwd()
	client := netclient.New(cfg.Retries, cfg.TimeoutSec, cfg.UserAgent)
	mgr := download.NewManager(client)

	delegate := itemDelegate{
		styles:      st,
		moveToCwd:   cfg.MoveToCwd,
		downloadDir: cfg.DownloadDir,
		cwd:         cwd,
	}

	l := list.New([]list.Item{}, delegate, 0, 0)
	l.Title = "indexof"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = st.Title
	l.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(st.Accent)
	l.Styles.FilterCursor = lipgloss.NewStyle().Foreground(st.Accent)
	l.Styles.NoItems = lipgloss.NewStyle().Foreground(st.Subtle).Padding(0, 2)

	m := model{
		cfg:          cfg,
		client:       client,
		manager:      mgr,
		styles:       st,
		list:         l,
		spinner:      s,
		loading:      true,
		deleteOnExit: cfg.DeleteOnExit,
		moveToCwd:    cfg.MoveToCwd,
		downloaded:   make(map[string]struct{}),
		cwd:          cwd,
	}

	u, err := url.Parse(startURL)
	if err != nil {
		return m, err
	}
	m.base = u
	m.history = []*url.URL{u}
	return m, nil
}

func (m model) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.fetch(m.base), tickCmd())
}
