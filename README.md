# Zeno

Your terminal in the browser. Single binary, zero dependencies.

## Install

**macOS / Linux:**
```bash
curl -fsSL https://raw.githubusercontent.com/Harsh-2002/Zeno/main/install.sh | bash
```

**Windows (PowerShell):**
```powershell
irm https://raw.githubusercontent.com/Harsh-2002/Zeno/main/install.ps1 | iex
```

Or download a binary from [Releases](https://github.com/Harsh-2002/Zeno/releases).

## Quick Start

```bash
zeno                    # start on port 8080
zeno --secret mytoken   # require authentication
zeno --tls              # enable HTTPS
```

## Features

- **Split panes** — vertical and horizontal, draggable dividers, double-click to equalize
- **Multiple tabs** — create, close, drag-to-reorder, rename, notification badges
- **5 themes** — Dark, Dracula, Solarized Dark, Nord, Monokai
- **Search** — regex, case-sensitive, whole-word, theme-aware highlights
- **File browser** — browse, download, upload, rename, delete
- **Session reconnect** — 256KB ring buffer, auto-reconnect on disconnect
- **Multiple fonts** — JetBrains Mono, Fira Code, Cascadia Code, Menlo, Monaco
- **Font ligatures** — toggle in settings
- **TLS** — auto self-signed cert or custom cert/key
- **Authentication** — secret token with secure session cookies
- **SSH proxy** — `zeno ssh user@host`
- **Settings panel** — live config, persisted to TOML file
- **Right-click context menu** — copy, paste, search, split, save output, browse files
- **100K line scrollback** with custom scrollbar
- **Responsive** — works on mobile and desktop

## Usage

```bash
zeno                                    # start on :8080
zeno --port 9090                        # custom port
zeno --secret mysecret                  # require auth
zeno --tls                              # HTTPS with auto cert
zeno --tls --cert my.crt --key my.key   # custom cert
zeno --config ./zeno.toml               # custom config path
zeno ssh user@host                      # SSH proxy mode
zeno ssh user@host -p 2222 -i ~/.ssh/key
```

## Keyboard Shortcuts

| Shortcut | Action |
|----------|--------|
| `Cmd+D` | Split vertical |
| `Cmd+Shift+D` | Split horizontal |
| `Cmd+W` | Close pane / tab |
| `Cmd+F` | Search |
| `Cmd+T` | New tab |
| `Cmd+,` | Settings |
| `Cmd+=/Cmd+-` | Font size |
| `Cmd+0` | Reset font size |
| `Cmd+1-9` | Switch tab |
| `Cmd+Shift+[/]` | Previous / next tab |

## Config

Settings are stored in `~/.zeno.toml` and editable from the Settings panel.

```toml
theme = "dark"
fontSize = 14
fontFamily = "JetBrains Mono"
fontLigatures = false
cursorStyle = "block"
cursorBlink = true
lineHeight = 1.1
scrollback = 100000
copyOnSelect = false
port = 8080
shell = "/bin/zsh"
```

Use `--config` to load from a custom path:

```bash
zeno --config ./zeno.toml
```

## Build from Source

Requires Go 1.21+ and Node.js 18+.

```bash
git clone https://github.com/Harsh-2002/Zeno.git
cd Zeno
./run 8080
```

## Architecture

```
frontend/    Svelte 5 + Vite
backend/     Go (net/http + gorilla/websocket + creack/pty)
run          Build script
```

Single binary embeds all frontend assets. No runtime dependencies.

## License

MIT
