package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/benjith17/mcmanager/server"
)

var (
	styleStopped  = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	styleStarting = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	styleRunning  = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	styleSelected = lipgloss.NewStyle().Bold(true).Reverse(true)
	styleHeader   = lipgloss.NewStyle().Bold(true).Underline(true)
	styleFaint    = lipgloss.NewStyle().Faint(true)
)

const (
	colName    = 20
	colStatus  = 12
	colPlayers = 9
	colRAM     = 10
	colCPU     = 8
	colUptime  = 10
)

type Overview struct {
	rows     []overviewRow
	selected int
	idle     bool
}

type overviewRow struct {
	name    string
	status  server.Status
	players int
	ramMB   float64
	cpuPct  float64
	uptime  time.Duration
	version string
}

func NewOverview() Overview {
	return Overview{}
}

func (o Overview) refresh(servers []*server.Server, selected int, idle bool) Overview {
	o.selected = selected
	o.idle = idle
	o.rows = make([]overviewRow, len(servers))
	for i, srv := range servers {
		st := srv.GetStats()
		o.rows[i] = overviewRow{
			name:    srv.Config.Name,
			status:  srv.GetStatus(),
			players: len(srv.GetPlayers()),
			ramMB:   st.MemoryMB,
			cpuPct:  st.CPUPercent,
			uptime:  srv.GetUptime(),
			version: srv.Config.Version,
		}
	}
	return o
}

func (o Overview) render(width, height int) string {
	var sb strings.Builder

	header := padRight("NAME", colName) +
		padRight("STATUS", colStatus) +
		padRight("PLAYERS", colPlayers) +
		padRight("RAM", colRAM) +
		padRight("CPU", colCPU) +
		padRight("UPTIME", colUptime) +
		"VERSION"
	sb.WriteString(styleHeader.Render(header))
	sb.WriteString("\n")

	for i, row := range o.rows {
		badge := statusBadge(row.status)
		if row.status == server.Stopped {
			badge = "● Stopped"
		}
		badgeWidth := lipgloss.Width(badge)
		statusPad := colStatus - badgeWidth
		if statusPad < 0 {
			statusPad = 0
		}

		line := padRight(row.name, colName) +
			badge + strings.Repeat(" ", statusPad) +
			padRight(fmt.Sprintf("%d", row.players), colPlayers) +
			padRight(fmt.Sprintf("%.0f MB", row.ramMB), colRAM) +
			padRight(fmt.Sprintf("%.1f%%", row.cpuPct), colCPU) +
			padRight(formatUptime(row.uptime), colUptime) +
			row.version

		if row.status == server.Stopped {
			line = styleStopped.Render(line)
		}
		if i == o.selected && !o.idle {
			line = styleSelected.Render(line)
		}
		sb.WriteString(line + "\n")
	}

	if !o.idle {
		sb.WriteString("\n")
		sb.WriteString(styleFaint.Render("↑↓/jk select  s start  x stop  r restart  Enter detail  q quit"))
	}
	return sb.String()
}

func statusBadge(s server.Status) string {
	switch s {
	case server.Stopped:
		return styleStopped.Render("● Stopped")
	case server.Starting:
		return styleStarting.Render("● Starting")
	case server.Running:
		return styleRunning.Render("● Running")
	}
	return ""
}

func padRight(s string, n int) string {
	w := lipgloss.Width(s)
	if w >= n {
		return s
	}
	return s + strings.Repeat(" ", n-w)
}

func formatUptime(d time.Duration) string {
	if d == 0 {
		return "-"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	sec := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, sec)
}
