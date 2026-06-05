# MCServerCLI

A terminal UI for managing Minecraft servers. Written in Go using Bubble Tea.

## Architecture

The app has three layers: config parsing, server process management, and the TUI.

### Config (`config/`)

Loads a TOML file. Each server entry maps to a `ServerConfig` struct. The config path defaults to `~/.config/mcmanager/config.toml` but can be overridden via a CLI argument.

### Server management (`server/`)

`Server` is a self-contained active object. Calling `Start()` execs `bash <start_script>` in the server's directory and launches three goroutines:

- **`watchProcess`** — blocks on `cmd.Wait()`. When the process exits (for any reason), it resets all state and calls `cancel()` to stop the other goroutines.
- **`tailLog`** — polls `logs/latest.log` every 500ms. Seeks to the end of the file on first open so pre-existing content is ignored. Parses new lines for player join/leave events and the "Done (" string that signals the server has finished starting. Buffers the last 500 lines for the TUI.
- **`pollStats`** — queries gopsutil every 2s for CPU% and RSS memory. Targets the Java subprocess if bash has children (i.e. the start script did not use `exec`), otherwise uses the bash process itself.

Graceful stop sends `"stop\n"` to the process's stdin, waits up to 30 seconds for exit, then kills.

All mutable state is protected by a single `sync.Mutex`. The TUI reads state via getters (`GetStatus`, `GetPlayers`, `GetLogs`, `GetStats`, `GetUptime`).

### TUI (`tui/`)

Built with Bubble Tea. The root model (`App`) routes between two views.

**Overview** — a table showing all servers with columns: Name, Status, Players, RAM, CPU, Uptime, Version. A 1-second tick drives refreshes by calling the server getters and re-rendering. Start/stop/restart are wrapped in `tea.Cmd` closures so they don't block the UI event loop. Stopped servers render the entire row in grey. After 5 seconds of no keypresses the selection highlight and key hint footer are hidden; any keypress restores them.

**Detail** — full-screen view for a single server. Contains a `bubbles/viewport` for scrollable log output and a `bubbles/textinput` for sending commands to the server's stdin. Auto-scrolls to the bottom unless the user has manually scrolled up.

Keys: `↑↓`/`jk` to navigate, `s` start, `x` stop, `r` restart, `Enter` to open detail, `Esc` to go back, `q`/`Ctrl+C` to quit.
