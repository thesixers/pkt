# pkt (Project Kit)

A cross-platform project manager and dependency tracker for **JavaScript, Python, Go, and Rust** projects.

**pkt (Project Kit)** organizes your projects, tracks dependencies, and abstracts package manager differences — all from a single, unified CLI.

## Features

- 🤖 **Autonomous AI Coding Agent** — Fully integrated conversational terminal agent (`pkt chat`) running native filesystem execution tools natively synced locally to Groq, Gemini, and OpenAI!
- 🌐 **Multi-language support** — JavaScript, Python, Go, and Rust
- 🗂️ **Centralized workspace** — All projects live in one configurable folder
- 🔄 **Package manager agnostic** — Works with npm, pnpm, bun, pip, poetry, uv, cargo, and go mod
- 📦 **Dependency tracking** — Database-backed tracking of all project dependencies
- 🆔 **Unique project IDs** — No more name conflicts, reference projects by ID
- 🚀 **Batch operations** — Add multiple packages in a single command (plus natural-language `pkt add --ai` querying)
- 🧹 **Deep System Maintenance** — Rapidly detect and prune `.venv`/`node_modules` caches recursively (`pkt clean`) and trace Global Git logic (`pkt status`)
- ⚡ **Zero setup** — Embedded SQLite database, no external dependencies
- 🐍 **Python venv** — Automatic virtual environment creation and management

## Prerequisites

Before using pkt, you need the following tools installed based on the languages you work with:

### Required for All Users

| Tool    | Why                      | Install                             |
| ------- | ------------------------ | ----------------------------------- |
| **Git** | Required for `pkt clone` | [git-scm.com](https://git-scm.com/) |

### Per-Language Requirements

| Language       | Required Tools              | Install                               |
| -------------- | --------------------------- | ------------------------------------- |
| **JavaScript** | Node.js + npm (or pnpm/bun) | [nodejs.org](https://nodejs.org/)     |
| **Python**     | Python 3.8+                 | [python.org](https://www.python.org/) |
| **Go**         | Go 1.18+                    | [go.dev](https://go.dev/)             |
| **Rust**       | Rust + Cargo                | [rustup.rs](https://rustup.rs/)       |

### Optional Package Managers

| Package Manager | Language   | Install                                                             |
| --------------- | ---------- | ------------------------------------------------------------------- |
| pnpm            | JavaScript | `npm install -g pnpm`                                               |
| bun             | JavaScript | [bun.sh](https://bun.sh/)                                           |
| poetry          | Python     | `pip install poetry`                                                |
| uv              | Python     | `pip install uv` or [docs.astral.sh/uv](https://docs.astral.sh/uv/) |

> **Note:** pkt will use whatever package manager is available. For JavaScript, it prefers pnpm > bun > npm. For Python, it prefers uv > poetry > pip.

## Quick Start

```bash
# First-time setup
pkt start

# Create a new project (prompts for language)
pkt create my-app

# Or specify language directly
pkt create my-api -l python
pkt create my-cli -l go
pkt create my-lib -l rust

# Initialize an existing project (auto-detects language)
pkt init /path/to/existing-project

# Add dependencies (from within a project)
cd my-app
pkt add react react-dom        # JavaScript
pkt add requests flask         # Python
pkt add -D typescript eslint   # Dev dependencies

# Run scripts
pkt run dev                    # npm/pnpm run dev
pkt run test                   # Run tests for any language
```

## Supported Languages

| Language       | Package Managers | Manifest File                        | Lockfile(s)                                        |
| -------------- | ---------------- | ------------------------------------ | -------------------------------------------------- |
| **JavaScript** | npm, pnpm, bun   | `package.json`                       | `package-lock.json`, `pnpm-lock.yaml`, `bun.lockb` |
| **Python**     | pip, poetry, uv  | `requirements.txt`, `pyproject.toml` | `poetry.lock`, `uv.lock`                           |
| **Go**         | go mod           | `go.mod`                             | `go.sum`                                           |
| **Rust**       | cargo            | `Cargo.toml`                         | `Cargo.lock`                                       |

## Installation

### Download Binary

Download the latest binary for your platform from the [Releases](https://github.com/thesixers/pkt/releases) page.

### Build from Source

```bash
git clone https://github.com/thesixers/pkt.git
cd pkt
go build -o pkt .
sudo mv pkt /usr/local/bin/
```

### Cross-Platform Support

pkt compiles to a single binary with no external dependencies:

| Platform | Binary                  |
| -------- | ----------------------- |
| Linux    | `pkt-linux-amd64`       |
| macOS    | `pkt-darwin-amd64`      |
| Windows  | `pkt-windows-amd64.exe` |

## Commands

### 🤖 Intelligence & AI Agent (NEW)

| Command               | Description                                                                     |
| --------------------- | ------------------------------------------------------------------------------- |
| `pkt chat`            | Launch a highly-capable Autonomous AI Coding Agent REPL directly in your shell! |
| `pkt debug [log]`     | Pass error traces to deeply diagnose logic natively (supports `stdin` piping)   |
| `pkt add --ai <desc>` | Install local packages purely through natural-language descriptions!            |

### Setup

| Command     | Description                               |
| ----------- | ----------------------------------------- |
| `pkt start` | Initialize pkt configuration and database |

### Project Management

| Command                       | Description                                           |
| ----------------------------- | ----------------------------------------------------- |
| `pkt create <name>`           | Create a new project in workspace                     |
| `pkt create <name> -l <lang>` | Create project with specified language                |
| `pkt init [path]`             | Initialize existing project (auto-detects language)   |
| `pkt list`                    | List all tracked projects                             |
| `pkt list -l <lang>`          | List projects filtered by language                    |
| `pkt clone <url>`             | Clone repo and auto-track ⭐ NEW                      |
| `pkt open <project>`          | Open project in configured editor                     |
| `pkt delete <project>`        | Delete project from filesystem and database           |
| `pkt rename <project>`        | Rename a tracked project                              |
| `pkt stats`                   | Show footprint analytics covering your root domains   |
| `pkt status`                  | Display dynamic git uncommitted states deeply         |
| `pkt clean`                   | Prune localized heavy `.venv` / `node_modules` caches |

### Dependency Management

> **Note:** These commands must be run inside a tracked project folder.

| Command               | Description                        |
| --------------------- | ---------------------------------- |
| `pkt add <pkg...>`    | Add one or more dependencies       |
| `pkt add -D <pkg...>` | Add as dev dependencies            |
| `pkt remove <pkg...>` | Remove dependencies                |
| `pkt update [pkg...]` | Update dependencies ⭐ NEW         |
| `pkt outdated`        | Check for outdated packages ⭐ NEW |
| `pkt deps [project]`  | List project dependencies          |

### Running Scripts

| Command                    | Description                                              |
| -------------------------- | -------------------------------------------------------- |
| `pkt run <script>`         | Run a script from package.json or common commands ⭐ NEW |
| `pkt exec <project> <cmd>` | Run command in another project's context ⭐ NEW          |

**`pkt run` examples by language:**

| Language       | Commands                                               |
| -------------- | ------------------------------------------------------ |
| **JavaScript** | `pkt run dev`, `pkt run build`, `pkt run <any-script>` |
| **Python**     | `pkt run test` (pytest), `pkt run main.py`             |
| **Go**         | `pkt run run`, `pkt run test`, `pkt run build`         |
| **Rust**       | `pkt run run`, `pkt run test`, `pkt run build`         |

### Configuration

| Command                                                      | Description                                                      |
| ------------------------------------------------------------ | ---------------------------------------------------------------- |
| `pkt config`                                                 | Show current config and full provider registry                   |
| `pkt change editor <cmd>`                                    | Change editor (e.g., code, cursor)                               |
| `pkt change pm <pm>`                                         | Change current project's package manager                         |
| `pkt change pm <pm> -l <lang> -g`                            | Change a language's default package manager                      |
| `pkt change ai <provider>`                                   | Switch active AI provider                                        |
| `pkt change root <path>`                                     | Set projects root for newly created or cloned projects           |
| `pkt change ai add provider <name>`                          | Register a cloud or local provider                               |
| `pkt change ai add provider <name> --url <url>`              | Register a self-hosted provider at a custom URL                  |
| `pkt change ai add key <api_key> --provider <name>`          | Add an API key to a registered provider                          |
| `pkt change ai add model <model> --provider <name>`          | Add a model to a registered provider                             |
| `pkt change ai remove key <api_key> --provider <name>`       | Remove an API key from a provider                                |
| `pkt change ai remove model <model> --provider <name>`       | Remove a model from a provider                                   |
| `pkt change ai remove provider <name>`                       | Remove a registered provider entirely                            |

**Examples:**

```bash
# Cloud providers
pkt change ai add provider groq
pkt change ai add key sk-xxxx --provider groq
pkt change ai add key AIxxxx --provider gemini

# Local Ollama (no key needed)
pkt change ai add provider ollama
pkt change ai ollama

# Self-hosted at a custom URL
pkt change ai add provider local --url http://localhost:1234
pkt change ai add model phi3 --provider local
pkt change ai local

# Add multiple models
pkt change ai add model llama-3.3-70b-versatile --provider groq
pkt change ai add model llama-3.1-8b-instant --provider groq

# Choose where newly created projects are stored
pkt change root ~/Projects
```

## Python Virtual Environment

pkt automatically manages Python virtual environments:

- **pip projects**: Auto-creates `.venv/` on first `pkt add`
- **poetry projects**: Configures `virtualenvs.in-project = true`
- **uv projects**: Uses uv's built-in venv management

```bash
# For pip projects, pkt will:
# 1. Create .venv/ if it doesn't exist
# 2. Install packages in the venv
# 3. Update requirements.txt

pkt add requests
# → Creates .venv/ → pip install requests → pip freeze > requirements.txt
```

## Configuration

Configuration is stored in `~/.pkt/config.json`:

```json
{
  "projects_root": "~/Documents/workspace",
  "default_pms": {
    "javascript": "pnpm",
    "python": "uv",
    "go": "go",
    "rust": "cargo"
  },
  "editor": "code",
  "initialized": true
}
```

## Database

pkt uses an embedded SQLite database at `~/.pkt/pkt2.db` to track:

- **Projects** — ID, name, path, language, package manager
- **Dependencies** — name, version, type (prod/dev) per project

> **Zero setup** — The database is created automatically on first run.

## Architecture

| Component     | Technology                                        |
| ------------- | ------------------------------------------------- |
| Language      | Go 1.24+                                          |
| CLI Framework | Cobra                                             |
| Database      | SQLite (embedded, pure Go via modernc.org/sqlite) |
| ID Generation | ULID                                              |

```
pkt/
├── cmd/              # CLI commands
├── internal/
│   ├── config/       # Configuration management
│   ├── db/           # SQLite database operations
│   ├── lang/         # Language detection & abstraction
│   ├── pm/           # Package manager abstraction
│   └── utils/        # Utilities (fs, package.json, etc.)
└── main.go
```

## Safety

- pkt **only modifies** manifest files (package.json, requirements.txt, etc.)
- Never touches source files or configs
- Database always syncs from filesystem
- Duplicate names resolved interactively

## License

MIT
