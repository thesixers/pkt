# pkt v1.1.0 Release Notes

## About pkt

**pkt (Project Kit)** is a cross-platform project manager and dependency tracker for JavaScript, Python, Go, and Rust projects. It organizes your projects, tracks dependencies, and abstracts package manager differences — all from a single, unified CLI.

---

## What's New in v1.1.0

### Improvements

- **Improved path handling in `pkt start`** — When entering a folder name for the projects root, it now correctly resolves to `$HOME/Documents/<folder_name>` instead of a relative path
- **Cleaner console output** — Removed excessive emojis for a more professional CLI experience
- **CI/CD Pipeline** — Added GitHub Actions workflow with:
  - Automated build and test on push/PR
  - Code linting with golangci-lint
  - Cross-platform release builds (Linux, macOS, Windows)

### Bug Fixes

- Fixed path resolution issue that caused `pkt init` to fail with "destination is inside source" error when projects were outside the workspace

---

## Features

- **Multi-language support** — JavaScript, Python, Go, and Rust
- **Centralized workspace** — All projects in one configurable folder
- **Package manager agnostic** — Works with npm, pnpm, bun, pip, poetry, uv, cargo, and go mod
- **Dependency tracking** — Database-backed tracking of all project dependencies
- **Unique project IDs** — No more name conflicts
- **Batch operations** — Add multiple packages in a single command
- **Zero setup** — Embedded SQLite database, no external dependencies
- **Python venv** — Automatic virtual environment creation and management

---

## Installation

### Download Binary

Download the appropriate binary for your platform from the assets below:

| Platform              | Binary                  |
| --------------------- | ----------------------- |
| Linux (x64)           | `pkt-linux-amd64`       |
| Linux (ARM)           | `pkt-linux-arm64`       |
| macOS (Intel)         | `pkt-darwin-amd64`      |
| macOS (Apple Silicon) | `pkt-darwin-arm64`      |
| Windows (x64)         | `pkt-windows-amd64.exe` |
| Windows (ARM)         | `pkt-windows-arm64.exe` |

### Quick Install (Linux/macOS)

```bash
# Download (replace with your platform)
curl -LO https://github.com/thesixers/pkt/releases/download/v1.1.0/pkt-linux-amd64

# Make executable
chmod +x pkt-linux-amd64

# Move to PATH
sudo mv pkt-linux-amd64 /usr/local/bin/pkt
```

### Build from Source

```bash
git clone https://github.com/thesixers/pkt.git
cd pkt
go build -o pkt .
sudo mv pkt /usr/local/bin/
```

---

## Quick Start

```bash
# Initialize pkt
pkt start

# Create a new project
pkt create my-app

# Add dependencies
cd my-app
pkt add react react-dom

# Run scripts
pkt run dev
```

---

## Full Changelog

See all commits: https://github.com/thesixers/pkt/compare/v1.0.1...v1.1.0
