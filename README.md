<div align="center">
  <pre>
╔══════════════════════════════════════════╗
║        █████  ██   ██                   ║
║       ██   ██  ██ ██                    ║
║       ███████   ███                     ║
║       ██   ██  ██ ██                    ║
║       ██   ██ ██   ██                   ║
║                                          ║
║     TUI API Client v2                    ║
║     Terminal-native HTTP client          ║
╚══════════════════════════════════════════╝
  </pre>
  <h1>ax — TUI API Client</h1>
  <p>
    <strong>A fast, beautiful terminal-based HTTP client</strong><br>
    <em>Inspired by Postman/Insomnia, powered by Bubble Tea, built for the terminal</em>
  </p>
  <p>
    <a href="https://github.com/zaidejjo/ax/releases"><img src="https://img.shields.io/github/v/release/zaidejjo/ax?style=flat&label=release&color=7C3AED" alt="Release"></a>
    <a href="https://github.com/zaidejjo/ax/actions"><img src="https://img.shields.io/github/actions/workflow/status/zaidejjo/ax/ci.yml?branch=main&style=flat&label=CI&color=10B981" alt="CI"></a>
    <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go&logoColor=white" alt="Go"></a>
    <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue?style=flat" alt="License"></a>
  </p>
</div>

---

**ax** is a terminal-native HTTP client for building, debugging, and
replaying API requests — no GUI, no Electron, no bloat. It combines the
polish of modern API clients with the speed and scriptability of the terminal.

## ✨ Features

| Feature | Description |
|---------|-------------|
| **Three-Pane TUI** | Sidebar (history), request builder, response viewer — Tab to cycle |
| **xh/httpie Syntax** | `:8080/api/users name==John` — shorthand for URL, headers, JSON, and form data |
| **Syntax Highlighting** | Chroma-powered JSON highlighting with the Catppuccin Macchiato theme |
| **Persistent History** | SQLite-backed request history with WAL mode — survives restarts |
| **Clipboard Integration** | `Ctrl+Y` copies response body — works on X11, Wayland, and macOS |
| **Smart Method Detection** | Type `name==John` → auto-detects POST; prefix `DELETE` for explicit method |
| **Keybinding Help** | Press `?` for a complete overlay reference |
| **Shell Completions** | `ax completion [bash\|zsh\|fish]` for tab-completion |
| **Cross-Platform** | Linux, macOS, Windows — statically linked, zero CGo |
| **No Electron** | Single 23 MB binary, minimal memory footprint |

## 🚀 Installation

### Homebrew (macOS / Linux)

```bash
brew install zaidejjo/tap/ax
```

### Go Install

```bash
go install github.com/zaidejjo/ax/cmd/ax@latest
```

### Direct Binary

Download the archive for your platform from the
[releases page](https://github.com/zaidejjo/ax/releases):

| Platform | Architecture | Format |
|----------|-------------|--------|
| Linux    | amd64       | `ax_{{ version }}_linux_amd64.tar.gz` |
| Linux    | arm64       | `ax_{{ version }}_linux_arm64.tar.gz` |
| Linux    | amd64       | `ax_{{ version }}_x86_64.deb` |
| Linux    | arm64       | `ax_{{ version }}_aarch64.deb` |
| macOS    | amd64       | `ax_{{ version }}_macOS_amd64.tar.gz` |
| macOS    | arm64       | `ax_{{ version }}_macOS_arm64.tar.gz` |
| Windows  | amd64       | `ax_{{ version }}_windows_amd64.zip` |

```bash
# Extract and run:
tar xzf ax_*.tar.gz
./ax
```

### Build from Source

```bash
git clone https://github.com/zaidejjo/ax.git
cd ax
make build
```

## 🎮 Usage

### Launch the TUI

```bash
ax
```

### Smart Syntax (xh / httpie style)

ax understands httpie/xh-style shorthand directly in the URL input:

```bash
# Start typing in the URL field at launch:
# Press Ctrl+R to send

# GET localhost:8080/api/users
#    Type: :8080/api/users

# POST with JSON body to localhost
#    Type: POST :8080/api/users name=="John" age=="30"

# GET with custom header
#    Type: https://api.example.com/data Authorization:Bearer\ntoken123

# PATCH with JSON field
#    Type: PATCH example.com/resource/1 name=="Updated"

# Form-encoded POST
#    Type: https://example.com/login username=john password=secret

# DELETE with shorthand
#    Type: DELETE :8080/api/users/42
```

### Keybindings

| Key | Action |
|-----|--------|
| `Tab` / `Shift+Tab` | Cycle focus between panes |
| `Ctrl+R` | Execute the current request |
| `m` | Cycle HTTP method |
| `Ctrl+Y` | Copy response body to clipboard |
| `Enter` | (sidebar) Load selected history entry |
| `Ctrl+D` | (sidebar) Delete selected entry |
| `↑/↓` / `PgUp/PgDn` | Scroll response body |
| `?` | Toggle help overlay |
| `q` / `Ctrl+C` | Quit |

### Shell Completions

```bash
# Bash
source <(ax completion bash)

# Zsh
source <(ax completion zsh)

# Fish
ax completion fish > ~/.config/fish/completions/ax.fish
```

## 🏗 Architecture

```
ax/
├── cmd/ax/              # Entry point, CLI flags, completions
├── internal/
│   ├── client/          # Pure Go HTTP client (net/http wrapper)
│   │   ├── http.go      # Request/Response types, Do(), options
│   │   └── parser.go    # xh/httpie-style single-line parser
│   ├── history/         # SQLite-backed persistent request store
│   │   └── store.go     # CRUD operations, WAL mode, schema migration
│   └── tui/             # Bubble Tea v2 terminal UI
│       ├── model.go     # Root model, update loop, message dispatch
│       ├── styles.go    # Lip Gloss theme (violet palette)
│       ├── clipboard.go # Platform clipboard (xclip/wl-copy/pbcopy)
│       ├── pane_help.go # Keybinding reference overlay
│       ├── pane_request.go  # URL input + method cycling
│       ├── pane_response.go # Viewport + chroma highlighting
│       └── pane_sidebar.go  # History list (bubbles/list)
├── .goreleaser.yaml     # Cross-compilation & distribution
└── Makefile             # Build, test, lint, run
```

**Stack:** [Bubble Tea v2](https://github.com/charmbracelet/bubbletea) ·
[Bubbles v2](https://github.com/charmbracelet/bubbles) ·
[Lip Gloss v2](https://github.com/charmbracelet/lipgloss) ·
[Chroma v2](https://github.com/alecthomas/chroma) ·
[modernc.org/sqlite](https://gitlab.com/cznic/sqlite) (pure Go, zero CGo)

## 🧪 Testing

```bash
# Run all tests
make test

# Run with race detector
go test ./... -race -count=1

# Run specific package tests
go test ./internal/client/... -v
go test ./internal/history/... -v
```

Currently **71 tests** across the client, parser, and history packages.

## 🤝 Contributing

Contributions are welcome! Please open an issue first to discuss what you'd
like to change.

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Run tests (`make test`)
4. Commit with conventional commits (`feat:`, `fix:`, `docs:`, etc.)
5. Open a pull request

## 📄 License

[MIT](LICENSE) © Zaid Ejjo

---

<div align="center">
  <sub>Built with ❤️ and Go — no JavaScript was harmed in the making of this API client</sub>
</div>
