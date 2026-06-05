package tui

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/benjith17/mcmanager/config"
	"github.com/benjith17/mcmanager/server"
)

type view int

const (
	viewOverview view = iota
	viewDetail
)

type tickMsg time.Time

type switchViewMsg struct {
	target    view
	serverIdx int
}

const idleTimeout = 5 * time.Second

type App struct {
	servers  []*server.Server
	view     view
	selected int
	width    int
	height   int
	lastKey  time.Time
	overview Overview
	detail   Detail
}

func NewApp(cfg *config.Config) App {
	servers := make([]*server.Server, len(cfg.Servers))
	for i, s := range cfg.Servers {
		servers[i] = server.New(s)
	}
	return App{
		servers:  servers,
		view:     viewOverview,
		lastKey:  time.Now(),
		overview: NewOverview(),
		detail:   NewDetail(),
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (a App) Init() tea.Cmd {
	cmds := []tea.Cmd{tickCmd()}
	for _, srv := range a.servers {
		if srv.Config.Autostart {
			srv := srv
			cmds = append(cmds, func() tea.Msg {
				_ = srv.Start()
				return tickMsg(time.Now())
			})
		}
	}
	return tea.Batch(cmds...)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.detail = a.detail.resize(msg.Width, msg.Height)
		return a, nil

	case tickMsg:
		a.overview = a.overview.refresh(a.servers, a.selected, time.Since(a.lastKey) > idleTimeout)
		if a.view == viewDetail && len(a.servers) > 0 {
			a.detail = a.detail.refresh(a.servers[a.selected])
		}
		return a, tickCmd()

	case switchViewMsg:
		a.view = msg.target
		if msg.target == viewDetail && len(a.servers) > 0 {
			a.selected = msg.serverIdx
			a.detail = a.detail.reset(a.servers[a.selected], a.width, a.height)
		}
		return a, nil

	case tea.KeyMsg:
		a.lastKey = time.Now()
		if msg.String() == "ctrl+c" {
			return a, tea.Quit
		}
		switch a.view {
		case viewOverview:
			return a.updateOverview(msg)
		case viewDetail:
			return a.updateDetail(msg)
		}
	}
	return a, nil
}

func (a App) updateOverview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	n := len(a.servers)
	if n == 0 {
		if msg.String() == "q" {
			return a, tea.Quit
		}
		return a, nil
	}

	switch msg.String() {
	case "up", "k":
		if a.selected > 0 {
			a.selected--
		}
	case "down", "j":
		if a.selected < n-1 {
			a.selected++
		}
	case "s":
		srv := a.servers[a.selected]
		return a, func() tea.Msg {
			_ = srv.Start()
			return tickMsg(time.Now())
		}
	case "x":
		srv := a.servers[a.selected]
		return a, func() tea.Msg {
			_ = srv.Stop()
			return tickMsg(time.Now())
		}
	case "r":
		srv := a.servers[a.selected]
		return a, func() tea.Msg {
			_ = srv.Restart()
			return tickMsg(time.Now())
		}
	case "enter":
		idx := a.selected
		return a, func() tea.Msg {
			return switchViewMsg{target: viewDetail, serverIdx: idx}
		}
	case "q":
		return a, tea.Quit
	}

	a.overview = a.overview.refresh(a.servers, a.selected, time.Since(a.lastKey) > idleTimeout)
	return a, nil
}

func (a App) updateDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "esc" {
		return a, func() tea.Msg {
			return switchViewMsg{target: viewOverview}
		}
	}
	var cmd tea.Cmd
	a.detail, cmd = a.detail.update(msg, a.servers[a.selected])
	return a, cmd
}

func (a App) View() string {
	switch a.view {
	case viewDetail:
		return a.detail.render()
	default:
		return a.overview.render(a.width, a.height)
	}
}
