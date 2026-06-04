package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/benjith17/mcmanager/config"
	"github.com/benjith17/mcmanager/stats"
	"github.com/shirou/gopsutil/v4/process"
)

const (
	maxLogLines         = 500
	logPollInterval     = 500 * time.Millisecond
	statsPollInterval   = 2 * time.Second
	gracefulStopTimeout = 30 * time.Second
)

type Status int

const (
	Stopped  Status = iota
	Starting
	Running
)

type Server struct {
	Config config.ServerConfig

	mu        sync.Mutex
	status    Status
	players   map[string]struct{}
	logs      []string
	stats     stats.ProcessStats
	startedAt time.Time

	cmd   *exec.Cmd
	stdin io.WriteCloser

	cancel context.CancelFunc
	done   chan struct{}
	wg     sync.WaitGroup
}

func New(cfg config.ServerConfig) *Server {
	return &Server{
		Config:  cfg,
		players: make(map[string]struct{}),
		logs:    make([]string, 0, maxLogLines),
	}
}

func (s *Server) GetStatus() Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status
}

func (s *Server) GetPlayers() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0, len(s.players))
	for p := range s.players {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}

func (s *Server) GetLogs() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.logs))
	copy(out, s.logs)
	return out
}

func (s *Server) GetStats() stats.ProcessStats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stats
}

func (s *Server) GetUptime() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.status == Stopped || s.startedAt.IsZero() {
		return 0
	}
	return time.Since(s.startedAt)
}

func (s *Server) Start() error {
	s.mu.Lock()
	if s.status != Stopped {
		s.mu.Unlock()
		return errors.New("server is not stopped")
	}
	s.status = Starting
	s.mu.Unlock()

	scriptPath := filepath.Join(s.Config.Directory, s.Config.StartScript)
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, "bash", scriptPath)
	cmd.Dir = s.Config.Directory

	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		s.mu.Lock()
		s.status = Stopped
		s.mu.Unlock()
		return err
	}

	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard

	if err := cmd.Start(); err != nil {
		cancel()
		s.mu.Lock()
		s.status = Stopped
		s.mu.Unlock()
		return err
	}

	s.mu.Lock()
	s.cmd = cmd
	s.stdin = stdinPipe
	s.cancel = cancel
	s.startedAt = time.Now()
	s.done = make(chan struct{})
	s.logs = s.logs[:0]
	s.players = make(map[string]struct{})
	s.mu.Unlock()

	s.wg.Add(3)
	go s.watchProcess()
	go s.tailLog(ctx)
	go s.pollStats(ctx)

	go func() {
		s.wg.Wait()
		close(s.done)
	}()

	return nil
}

func (s *Server) watchProcess() {
	defer s.wg.Done()
	_ = s.cmd.Wait()

	s.mu.Lock()
	s.status = Stopped
	s.cmd = nil
	s.stdin = nil
	s.players = make(map[string]struct{})
	s.stats = stats.ProcessStats{}
	s.mu.Unlock()

	s.cancel()
}

func (s *Server) Stop() error {
	s.mu.Lock()
	if s.status == Stopped {
		s.mu.Unlock()
		return errors.New("server is already stopped")
	}
	stdin := s.stdin
	done := s.done
	cmd := s.cmd
	s.mu.Unlock()

	if stdin != nil {
		_, _ = io.WriteString(stdin, "stop\n")
		_ = stdin.Close()
	}

	select {
	case <-done:
		return nil
	case <-time.After(gracefulStopTimeout):
		if cmd != nil && cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	}
	<-done
	return nil
}

func (s *Server) Restart() error {
	if err := s.Stop(); err != nil {
		return err
	}
	return s.Start()
}

func (s *Server) SendCommand(cmd string) error {
	s.mu.Lock()
	stdin := s.stdin
	s.mu.Unlock()
	if stdin == nil {
		return errors.New("server is not running")
	}
	_, err := io.WriteString(stdin, cmd+"\n")
	return err
}

func (s *Server) tailLog(ctx context.Context) {
	defer s.wg.Done()

	logPath := filepath.Join(s.Config.Directory, "logs", "latest.log")

	var f *os.File
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		var err error
		f, err = os.Open(logPath)
		if err == nil {
			break
		}
		time.Sleep(logPollInterval)
	}
	defer f.Close()

	// Seek to end — ignore content from before this session
	_, _ = f.Seek(0, io.SeekEnd)
	reader := bufio.NewReader(f)

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(logPollInterval):
		}
		for {
			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				s.processLogLine(strings.TrimRight(line, "\r\n"))
			}
			if err != nil {
				break
			}
		}
	}
}

func (s *Server) processLogLine(line string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logs = append(s.logs, line)
	if len(s.logs) > maxLogLines {
		s.logs = s.logs[1:]
	}

	if strings.Contains(line, "joined the game") {
		if name := extractPlayerName(line, "joined the game"); name != "" {
			s.players[name] = struct{}{}
		}
	} else if strings.Contains(line, "left the game") {
		if name := extractPlayerName(line, "left the game"); name != "" {
			delete(s.players, name)
		}
	} else if strings.Contains(line, "lost connection") {
		if name := extractPlayerName(line, "lost connection"); name != "" {
			delete(s.players, name)
		}
	}

	if s.status == Starting && strings.Contains(line, "Done (") {
		s.status = Running
	}
}

func extractPlayerName(line, suffix string) string {
	idx := strings.Index(line, "]: ")
	if idx < 0 {
		return ""
	}
	msg := line[idx+3:]
	if !strings.HasSuffix(msg, " "+suffix) {
		return ""
	}
	name := strings.TrimSuffix(msg, " "+suffix)
	if strings.Contains(name, " ") {
		return ""
	}
	return name
}

func (s *Server) pollStats(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(statsPollInterval):
		}

		s.mu.Lock()
		cmd := s.cmd
		s.mu.Unlock()

		if cmd == nil || cmd.Process == nil {
			continue
		}

		pid := int32(cmd.Process.Pid)
		proc, err := process.NewProcess(pid)
		if err != nil {
			continue
		}

		target := proc
		if children, err := proc.Children(); err == nil && len(children) > 0 {
			target = children[0]
		}

		cpu, _ := target.CPUPercent()
		mem, _ := target.MemoryInfo()
		if mem == nil {
			continue
		}

		s.mu.Lock()
		s.stats = stats.ProcessStats{
			CPUPercent: cpu,
			MemoryMB:   float64(mem.RSS) / 1024 / 1024,
		}
		s.mu.Unlock()
	}
}
