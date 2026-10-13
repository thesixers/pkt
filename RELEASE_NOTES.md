# pkt Release Notes

## About pkt

**pkt (Project Kit)** is a cross-platform project manager and dependency tracker for JavaScript, Python, Go, and Rust projects. It organizes your projects, tracks dependencies, and abstracts package manager differences — all from a single, unified CLI.

---

## What's New in v1.3.0 (Filesystem Routing & Scriptability Upgrade)

This major release introduces full **Project-Aware Filesystem Routing**, a powerful **Scriptable Project Introspection** engine, unified **Configuration & Language Migration**, and interactive CLI tooling.

### 📁 Project-Aware Filesystem Routing (`@project`)

No more calculating fragile relative paths (`../../`) or jumping across directories. `pkt` now treats tracked projects as first-class path destinations:

- **The `@` Resolution Engine**: Prefix any path argument with `@<project-name>` (e.g., `@api/src/main.go` or `@frontend/config.json`) to automatically resolve paths against the project's root in the SQLite database.
- **Path Traversal Security**: Strict directory traversal prevention blocks escape sequences (`@project/../other` is safely rejected).
- **`pkt ls [path]`**:
  - List files and directories across local paths or `@project` references.
  - Granular control with `-a` (include hidden), `-l` (long listing with permissions, human-readable size, and modification timestamp), `-r` (recursive traversal), `-d` (directories only), and `-f` (files only).
  - Flexible sorting: `--sort name|size|modified` with `--reverse`.
- **`pkt cat <file...>`**:
  - Print text files from standard filesystem paths or `@project` locations.
  - Line slicing controls: `--start <line>`, `--end <line>`, `--head <N>`, and `--tail <N>`, plus `-n` for line numbering.
  - Multi-file output with headers (`==> filename <==`) and `--no-header` suppression.
  - Automatic binary content detection with safety refusal and `--force` override.
- **`pkt cp <source> <destination>`**:
  - Copy individual files or recursive directory trees across projects or between the local filesystem and tracked workspaces.
  - Preserves exact file permissions (`os.FileMode`) and directory hierarchies.
- **`pkt mv <source> <destination>`**:
  - Move files or directories across projects with instant atomic OS-level renames.
  - Intelligent cross-device link (`EXDEV`) fallback: seamlessly falls back to buffered copy and source deletion when moving across disk partitions or mount boundaries.

### 🔍 Scriptable Project Introspection (`pkt info`)

`pkt info` is now deeply composable with shell scripts, aliases, and CI/CD pipelines:

- **Field Output Flags**: Extract single values without table borders or decorative formatting:
  - `-l, --lang` — project language (e.g., `go`, `javascript`, `python`, `rust`)
  - `--pm` — package manager (e.g., `go mod`, `pnpm`, `uv`, `cargo`)
  - `-p, --path` — absolute filesystem path
  - `-i, --id` — unique ULID
  - `-s, --size` — human-readable disk footprint
  - `-n, --name` — project name
  - `-d, --docs` — parsed README / markdown documentation excerpt
- **Predictable Canonical Order**: Combining multiple flags prints values in a deterministic, one-per-line sequence (`name`, `id`, `lang`, `pm`, `path`, `size`, `docs`).
- **Enhanced Full Display**: Full `pkt info` now displays both `Language` and `PM` directly in the project overview header.

### ⚙️ Language Management & Rebuild

- **`pkt change lang <language> [project]`**: Change a project's language on the fly and automatically detect/select the best available package manager for that language.
- **`pkt change rebuild [project]`**: Re-scan a project directory to automatically re-detect and synchronize its primary language and package manager in the database.

### 🤖 Multi-Provider AI Registry & Agent Enhancements

- **Universal Registry (`pkt config`)**: Inspect configured settings, global defaults, and all registered AI endpoints in a unified view.
- **Local & Self-Hosted Providers**: First-class support for Ollama, LocalAI, vLLM, and custom endpoints via `pkt change ai add provider <name> --url <url>`.
- **Agent Resilience**: Added 429 exponential backoff retries, sliding context window trimming, terminal markdown rendering, and enhanced tool calling.

### 🚀 CLI Tooling & Safety

- **`pkt upgrade`**: Automatic in-place self-updater checking GitHub releases.
- **PM2-Style Tables**: Modern rounded ASCII tables for `pkt list` and `pkt outdated` via `go-pretty/table`.
- **`pkt version`**: Native version reporting command.

---

## What's New in v1.2.0 (The Autonomous AI Upgrade)

This massive release transforms `pkt` from a standard project manager into a hyper-localized, natively-integrated Autonomous Coding Agent!

### 🤖 The Autonomous AI Studio

- **`pkt chat`**: Launch an active, infinite Terminal REPL session allowing a constant two-way conversation with a fully integrated Coding Agent! The AI physically uses local Context Tools (`list_dir`, `read_file`, `write_file`, `make_dir`, `delete_file`, `run_command`, `get_project_info`) to autonomously examine directories, parse your scripts, and perfectly execute bash commands across your active codebase!
- **Universal Provider Engine**: Connect explicitly to any top LLM ecosystem without limits. Natively hooks into **OpenAI, Google Gemini, and Groq** right out of the box via `pkt change ai add provider <name>` and `pkt change ai add key <key> --provider <name>`.
- **Custom Model Binders**: Fine-tune your AI explicitly with `pkt change ai add model <model> --provider <name>`, ensuring you can always tap precisely into Llama-3, GPT-4o, or Gemini 1.5 Flash natively.

### ⚡ Context-Aware Intelligence Commands

- **`pkt debug`**: Native pipeline to debug server traces. Aggressively pipes broken error text or raw `stdin` chains straight to the AI (e.g. `cat log.txt | pkt debug`).
- **`pkt add --ai <prompt>`**: Talk to the CLI naturally to query package dependencies. Tell the AI what you want to achieve, and it physically translates it into hard CLI dependency parameters and organically routes it directly to your package manager!

### 🧹 Advanced Workspace Upgrades

- **`pkt clean`**: Interactively detect and entirely prune out massive blackhole caching folders across your disks (e.g. `node_modules`, `.venv`, `.next`) to instantly reclaim hundreds of Gigabytes!
- **`pkt status`**: Crawl standard `git status` commands synchronously across every single project inside your `projects_root` dynamically.
- **`pkt stats`**: Run an architectural scan to summarize language footprints across your hard drive and display aggressive physical byte-sizes!

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
