# pkt v1.2.0 Release Notes

## About pkt

**pkt (Project Kit)** is a cross-platform project manager and dependency tracker for JavaScript, Python, Go, and Rust projects. It organizes your projects, tracks dependencies, and abstracts package manager differences — all from a single, unified CLI.

---

## What's New in v1.2.0 (The Autonomous AI Upgrade)

This massive release transforms `pkt` from a standard project manager into a hyper-localized, natively-integrated Autonomous Coding Agent!

### 🤖 The Autonomous AI Studio

- **`pkt chat`**: Launch an active, infinite Terminal REPL session allowing a constant two-way conversation with a fully integrated Coding Agent! The AI physically uses local Context Tools (`list_dir`, `read_file`, `write_file`, `make_dir`, `delete_file`, `run_command`, `get_project_info`) to autonomously examine directories, parse your scripts, and perfectly execute bash commands across your active codebase!
- **Universal Provider Engine**: Connect explicitly to any top LLM ecosystem without limits. Natively hooks into **OpenAI, Google Gemini, and Groq** right out of the box via `pkt config set-ai <provider> <api_key>`.
- **Dynamic Cross-Provider Flagging**: Swap AI intelligence layers directly on any command simultaneously using the `-p` / `--provider` flag (e.g. `pkt generate "React hook" -p gemini`)!
- **Custom Model Binders**: Fine-tune your AI explicitly with `pkt config set-model <provider> <model_name>`, ensuring you can always tap precisely into Llama-3, GPT-4o, or Gemini 1.5 Flash natively.

### ⚡ Context-Aware Intelligence Commands

- **`pkt ask "question"`**: Instantly ask the AI specific logic queries. The CLI seamlessly pipes your project's custom `README.md` and Framework environments natively into the Context Window so it answers directly regarding your active language and Package Manager!
- **`pkt generate "feature"`**: Scaffold deep boilerplate templates, scripts, or components automatically tailored natively to your active environment.
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
