package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/benjith17/mcmanager/server"
)

type Detail struct {
	vp         viewport.Model
	ti         textinput.Model
	serverName string
	statusStr  string
	width      int
	height     int
	atBottom   bool
}

func NewDetail() Detail {
	ti := textinput.New()
	ti.Placeholder = "send command..."
	ti.Focus()

	return Detail{
		vp:       viewport.New(0, 0),
		ti:       ti,
		atBottom: true,
	}
}

func (d Detail) resize(width, height int) Detail {
	d.width = width
	d.height = height
	d.vp.Width = width
	d.vp.Height = height - 4 // header + separator + input + hints
	return d
}

func (d Detail) reset(srv *server.Server, width, height int) Detail {
	d = d.resize(width, height)
	d.serverName = srv.Config.Name
	d.ti.SetValue("")
	d.ti.Focus()
	d.atBottom = true
	d = d.refresh(srv)
	return d
}

func (d Detail) refresh(srv *server.Server) Detail {
	d.statusStr = statusBadge(srv.GetStatus())
	logs := srv.GetLogs()
	d.vp.SetContent(strings.Join(wrapLogLines(logs, d.vp.Width), "\n"))
	if d.atBottom {
		d.vp.GotoBottom()
	}
	return d
}

// wrapLogLines hard-wraps each line to width characters so the viewport doesn't clip them.
func wrapLogLines(lines []string, width int) []string {
	if width <= 0 {
		return lines
	}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		for len(line) > width {
			out = append(out, line[:width])
			line = line[width:]
		}
		out = append(out, line)
	}
	return out
}

func (d Detail) update(msg tea.KeyMsg, srv *server.Server) (Detail, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if cmd := d.ti.Value(); cmd != "" {
			_ = srv.SendCommand(cmd)
			d.ti.SetValue("")
		}
		return d, nil
	case "up":
		d.vp.LineUp(1)
		d.atBottom = false
		return d, nil
	case "down":
		d.vp.LineDown(1)
		d.atBottom = d.vp.AtBottom()
		return d, nil
	case "pgup":
		d.vp.HalfViewUp()
		d.atBottom = false
		return d, nil
	case "pgdown":
		d.vp.HalfViewDown()
		d.atBottom = d.vp.AtBottom()
		return d, nil
	default:
		var cmd tea.Cmd
		d.ti, cmd = d.ti.Update(msg)
		return d, cmd
	}
}

func (d Detail) render() string {
	header := lipgloss.NewStyle().Bold(true).Render(d.serverName) + "  " + d.statusStr
	separator := strings.Repeat("─", d.width)
	input := "> " + d.ti.View()
	hints := styleFaint.Render("Esc back  ↑↓/PgUp/PgDn scroll  Enter send")

	return strings.Join([]string{header, separator, d.vp.View(), input, hints}, "\n")
}
