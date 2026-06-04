# Configuration

By default, mcmanager reads its config from `~/.config/mcmanager/config.toml` (respects `$XDG_CONFIG_HOME`). You can also pass a path directly:

```
mcmanager /path/to/config.toml
```

---

## Format

The config file is [TOML](https://toml.io). Each server is defined as an entry in the `[[servers]]` array.

```toml
[[servers]]
name        = "Survival"
directory   = "/home/user/servers/survival"
start_script = "start.sh"
version     = "Paper 1.21.4"
autostart   = true

[[servers]]
name        = "Creative"
directory   = "/home/user/servers/creative"
start_script = "start.sh"
version     = "Vanilla 1.20.6"
autostart   = false
```

---

## Fields

| Field         | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `name`        | string | yes      | Display name shown in the overview |
| `directory`   | string | yes      | Absolute path to the server's root folder |
| `start_script`| string | yes      | Path to the start script, relative to `directory` |
| `version`     | string | no       | Displayed in the overview (e.g. `"Paper 1.21.4"`) |
| `autostart`   | bool   | no       | If `true`, the server starts automatically when mcmanager launches. Defaults to `false` |

---

## Notes

- `start_script` is executed with `bash`, so it must be a valid shell script. A typical script looks like:
  ```sh
  #!/bin/bash
  java -Xms1G -Xmx4G -jar server.jar --nogui
  ```
- `directory` is used as the working directory when running the script, so relative paths inside the script resolve from there.
- mcmanager reads `logs/latest.log` inside `directory` to display server output and track online players.
