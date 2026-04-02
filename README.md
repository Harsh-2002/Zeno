# Zeno

Your terminal in the browser. Single binary, zero dependencies.

## Quick Start

```bash
./run 8080
```

Builds frontend + backend, opens browser at `http://localhost:8080`.

## Features

- Split panes (vertical & horizontal)
- Multiple tabs with drag-to-reorder
- 5 themes (Dark, Dracula, Solarized, Nord, Monokai)
- Search through scrollback (100K lines)
- Right-click context menu
- Settings panel with live config
- TLS with auto-generated self-signed certs
- Password protection
- SSH proxy mode
- JetBrains Mono font
- Config via `~/.zeno.toml`

## Usage

```bash
./zeno                                  # start on port 8080
./zeno --port 9090                      # custom port
./zeno --secret mysecret                # require auth
./zeno --tls                            # https with auto cert
./zeno --tls --cert my.crt --key my.key # custom cert
./zeno --config ./zeno.toml             # custom config path
./zeno ssh user@host                    # ssh proxy mode
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
| `Cmd+1-9` | Switch tab |

## Config

All settings are stored in `~/.zeno.toml` and editable from the Settings panel in the browser.

```toml
theme = "dark"
fontSize = 14
cursorStyle = "block"
cursorBlink = true
lineHeight = 1.1
scrollback = 100000
copyOnSelect = false
port = 8080
shell = "/bin/zsh"
```

## Build from Source

Requires Go 1.21+ and Node.js 18+.

```bash
git clone https://github.com/AmanVarshney01/zeno.git
cd zeno
./run 8080
```

## Architecture

```
frontend/          Svelte 5 + Vite
backend/           Go (net/http + gorilla/websocket + creack/pty)
run                Build script (frontend → backend/static → binary)
```

Single binary embeds all frontend assets. No runtime dependencies.

## License

MIT
