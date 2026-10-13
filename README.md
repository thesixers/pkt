# pkt (Project Kit)

A cross-platform project manager and dependency tracker for **JavaScript, Python, Go, and Rust** projects.

**pkt (Project Kit)** organizes your projects, tracks dependencies, and abstracts package manager differences — all from a single, unified CLI.

## Features

- 🤖 **Autonomous AI Coding Agent** — Fully integrated conversational terminal agent (`pkt chat`) running native filesystem execution tools natively synced locally to Groq, Gemini, and OpenAI!
- 📁 **Project-Aware Filesystem Engine** — Cross-project filesystem navigation and operations (`pkt ls`, `pkt cat`, `pkt cp`, `pkt mv`) with `@project` prefix routing, path traversal security, and cross-device move fallback
- 🔍 **Scriptable Project Introspection** — Extract granular project metadata (`pkt info -l`, `--pm`, `-p`, `-n`, `-i`, `-s`, `-d`) for shell scripts, aliases, and automation pipelines
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

| Command                       | Description                                            |
| ----------------------------- | ------------------------------------------------------ |
| `pkt create <name>`           | Create a new project in workspace                      |
| `pkt create <name> -l <lang>` | Create project with specified language                 |
| `pkt init [path]`             | Initialize existing project (auto-detects language)    |
| `pkt list`                    | List all tracked projects                              |
| `pkt list -l <lang>`          | List projects filtered by language                     |
| `pkt clone <url>`             | Clone repo and auto-track ⭐ NEW                       |
| `pkt open <project>`          | Open project in configured editor                      |
| `pkt delete <project>`        | Delete project from filesystem and database            |
| `pkt rename <project>`        | Rename a tracked project                               |
| `pkt info <project>`          | Display project details and docs (supports field flags) |
| `pkt stats`                   | Show footprint analytics covering your root domains    |
| `pkt status`                  | Display dynamic git uncommitted states deeply          |
| `pkt clean`                   | Prune localized heavy `.venv` / `node_modules` caches  |
| `pkt ls [path]`               | List filesystem paths or `@project` paths               |
| `pkt cat <file...>`           | Print text files from paths or `@project` paths         |
| `pkt cp <src> <dst>`          | Copy files or directories with `@project` routing      |
| `pkt mv <src> <dst>`          | Move files or directories with `@project` routing      |
| `pkt upgrade`                 | Self-update pkt to the latest GitHub release ⭐ NEW    |
| `pkt version`                 | Print pkt version number                               |

### Filesystem Commands

`pkt` provides project-aware filesystem commands (`ls`, `cat`, `cp`, and `mv`) that seamlessly accept standard filesystem paths **or** `@project` references.

#### The `@` Project Prefix

Prefixing any path with `@<project-name>` routes the operation directly to the tracked project root:
- `@api` refers to the root directory of the project named `api`
- `@api/src/main.go` refers to `src/main.go` inside the `api` project
- Path traversal escapes (e.g., `@api/../secret`) are strictly rejected for safety

#### `pkt ls` — list files and directories

```bash
pkt ls                        # Current directory
pkt ls ./src                  # Relative path
pkt ls @api                   # Root of tracked project "api"
pkt ls @api/src               # Subdirectory inside "api"
pkt ls -la @api/src           # Long format + hidden files
pkt ls -r --sort modified .   # Recursive, sorted by modification time
pkt ls --reverse --sort size  # Largest files last
pkt ls -f @api                # Files only
pkt ls -d @api                # Directories only
```

| Flag | Description |
|---|---|
| `-a, --all` | Include hidden files and directories |
| `-l, --long` | Show permissions, size, and modification time |
| `-r, --recursive` | Walk subdirectories |
| `-d, --dirs` | Directories only |
| `-f, --files` | Files only |
| `--sort <name\|size\|modified>` | Sort field (default: `name`) |
| `--reverse` | Reverse the sort order |

#### `pkt cat` — print text files

```bash
pkt cat README.md                           # Single file
pkt cat @api/README.md                      # File inside tracked project
pkt cat @api/src/main.go @web/package.json  # Multiple files with headers
pkt cat -n @api/src/main.go                 # Show line numbers
pkt cat --start 20 --end 50 @api/README.md  # Lines 20–50
pkt cat --head 15 server.go                 # First 15 lines
pkt cat --tail 20 app.log                   # Last 20 lines
pkt cat --start 5 --head 10 main.go         # 10 lines starting at line 5
pkt cat --no-header a.txt b.txt             # Suppress headers between multiple files
pkt cat --force image.png                   # Force-print binary content
```

| Flag | Description |
|---|---|
| `-n, --number` | Show line numbers |
| `--start <line>` | Start at a one-based line number |
| `--end <line>` | Stop at a one-based line number (inclusive) |
| `--head <N>` | First N lines (can combine with `--start`) |
| `--tail <N>` | Last N lines (mutually exclusive with `--start`/`--end`/`--head`) |
| `--no-header` | Suppress filename headers when printing multiple files |
| `--force` | Print even if pkt detects binary content |

#### `pkt cp` — copy files and directories

Copy files or directory trees across projects or between the local filesystem and tracked projects.

```bash
pkt cp @api-server/config.json @frontend/config.json  # Project to project
pkt cp ./schema.sql @backend/migrations/              # Local file to project directory
pkt cp -r @legacy-app/data ./data                     # Recursive copy to local folder
```

- **Files to files:** Byte-for-byte stream preserving file mode permissions.
- **Files to directories:** Automatically appends the source filename so it lands inside the destination directory.
- **Directories to directories:** Recursively recreates nested directory structures and copies all contents.

#### `pkt mv` — move files and directories

Move files or directories across projects or between the local filesystem and tracked projects.

```bash
pkt mv ./old-script.sh @scripts/bin/                  # Move local file into project
pkt mv @api/v1 @api/legacy-v1                         # Move/rename inside project
pkt mv @backend/shared.go @frontend/shared.go         # Move between projects
```

- **Atomic rename:** Instantaneous OS-level rename when source and destination reside on the same filesystem.
- **Cross-device fallback:** Automatically falls back to a buffered copy followed by source removal when crossing disk partitions or mount points (`EXDEV`).

> **Path safety:** `@project/../other` is always rejected to prevent escaping a project's root.

### Project Information & Scripting

#### `pkt info` — inspect project details

Displays comprehensive metadata about a tracked project, including project ID, language, package manager, disk usage, absolute path, and a formatted README excerpt.

```bash
pkt info my-api               # Full project overview and documentation excerpt
```

**Field flags for scripting & automation:**

When any field flag is provided, `pkt info` switches to field mode, printing only the requested values (one per line) in canonical order. Ideal for shell scripts, aliases, and CI/CD pipelines:

```bash
pkt info my-api -l            # Print language (e.g., "go")
pkt info my-api --pm          # Print package manager (e.g., "go mod")
pkt info my-api -p            # Print absolute path
pkt info my-api -i            # Print project ID (ULID)
pkt info my-api -s            # Print disk size (e.g., "12.4 MB")
pkt info my-api -n            # Print project name
pkt info my-api -d            # Print rendered README documentation excerpt

# Scripting examples:
LANG=$(pkt info my-api -l)
PROJECT_DIR=$(pkt info my-api -p)
pkt info my-api -n -l --pm    # Prints name, language, and PM on separate lines
```

| Flag | Description |
|---|---|
| `-n, --name` | Print project name |
| `-i, --id` | Print unique project ID (ULID) |
| `-l, --lang` | Print project language (`javascript`, `python`, `go`, `rust`) |
| `--pm` | Print active package manager (`npm`, `pnpm`, `bun`, `pip`, `uv`, etc.) |
| `-p, --path` | Print absolute project path on disk |
| `-s, --size` | Print project disk footprint in human-readable units |
| `-d, --docs` | Print formatted README documentation excerpt |

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
| `pkt change lang <lang> [project]`                           | Change a project's language and auto-select package manager ⭐ NEW |
| `pkt change rebuild [project]`                               | Re-detect and rebuild language & package manager from disk ⭐ NEW|
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
# Language & Package Manager configuration
pkt change lang go my-project      # Switch project to Go (auto-detects go mod)
pkt change lang js                 # Change language for current directory's project
pkt change rebuild my-project      # Re-scan directory and rebuild language/PM configuration

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
