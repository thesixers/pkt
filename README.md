# pkt (Project Kit)

A cross-platform project manager and dependency tracker for **JavaScript, Python, Go, and Rust** projects.

**pkt (Project Kit)** organizes your projects, tracks dependencies, and abstracts package manager differences — all from a single, unified CLI.

## Features

- 🌐 **Multi-language support** — JavaScript, Python, Go, and Rust
- 🗂️ **Centralized workspace** — All projects live in one configurable folder
- 🔄 **Package manager agnostic** — Works with npm, pnpm, bun, pip, poetry, uv, cargo, and go mod
- 📦 **Dependency tracking** — Database-backed tracking of all project dependencies
- 🆔 **Unique project IDs** — No more name conflicts, reference projects by ID
- 🚀 **Batch operations** — Add multiple packages in a single command
- ⚡ **Zero setup** — Embedded SQLite database, no external dependencies

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

# Install all dependencies
pkt install
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

### Requirements

| Language       | Requirements                  |
| -------------- | ----------------------------- |
| **JavaScript** | Node.js + (npm, pnpm, or bun) |
| **Python**     | Python + (pip, poetry, or uv) |
| **Go**         | Go 1.18+                      |
| **Rust**       | Rust + Cargo                  |

> **Zero setup database!** pkt uses an embedded SQLite database that's created automatically.

### Cross-Platform Support

pkt compiles to a single binary with no external dependencies:

| Platform | Binary                  |
| -------- | ----------------------- |
| Linux    | `pkt-linux-amd64`       |
| macOS    | `pkt-darwin-amd64`      |
| Windows  | `pkt-windows-amd64.exe` |

## Commands

### Setup

| Command     | Description                               |
| ----------- | ----------------------------------------- |
| `pkt start` | Initialize pkt configuration and database |

### Project Management

| Command                       | Description                                         |
| ----------------------------- | --------------------------------------------------- |
| `pkt create <name>`           | Create a new project in workspace                   |
| `pkt create <name> -l <lang>` | Create project with specified language              |
| `pkt init [path]`             | Initialize existing project (auto-detects language) |
| `pkt list`                    | List all tracked projects                           |
| `pkt list -l <lang>`          | List projects filtered by language                  |
| `pkt open <project>`          | Open project in configured editor                   |
| `pkt delete <project>`        | Delete project from filesystem and database         |
| `pkt rename <project>`        | Rename a tracked project                            |
| `pkt search <query>`          | Search through tracked projects                     |

### Dependency Management

> **Note:** These commands must be run inside a tracked project folder.

| Command               | Description                  |
| --------------------- | ---------------------------- |
| `pkt add <pkg...>`    | Add one or more dependencies |
| `pkt add -D <pkg...>` | Add as dev dependencies      |
| `pkt remove <pkg...>` | Remove dependencies          |
| `pkt install`         | Install all dependencies     |
| `pkt deps [project]`  | List project dependencies    |

### Package Manager

| Command                     | Description                          |
| --------------------------- | ------------------------------------ |
| `pkt pm set <pm> <project>` | Change package manager for a project |

### Configuration

| Command                   | Description                        |
| ------------------------- | ---------------------------------- |
| `pkt config`              | Show current configuration         |
| `pkt config editor <cmd>` | Change editor (e.g., code, cursor) |
| `pkt config pm <pm>`      | Change default package manager     |

## Command Details

### `pkt create <name>`

Create a new project with interactive language selection:

```bash
pkt create my-app              # Prompts for language
pkt create my-app -l javascript  # JavaScript project
pkt create my-app -l python      # Python project
pkt create my-app -l go          # Go project
pkt create my-app -l rust        # Rust project
```

### `pkt init [path]`

Initialize an existing project. Language is auto-detected from manifest files:

```bash
pkt init .                        # Current directory
pkt init /path/to/my-project      # Specific project
```

**Auto-detection:**

- `package.json` → JavaScript
- `pyproject.toml` / `requirements.txt` → Python
- `go.mod` → Go
- `Cargo.toml` → Rust

### `pkt list`

List all projects with filtering options:

```bash
pkt list              # All projects
pkt list -l python    # Only Python projects
pkt list -l javascript # Only JavaScript projects
```

Output shows language and package manager for each project:

```
NAME          LANGUAGE     PM      ID                           PATH
my-rust-app   rust         cargo   01HJ9KJQ8D0XQZP2M3N4K5       /home/user/workspace/my-rust-app
my-py-app     python       pip     01HJ9KJR2A1BQZP3M4N5K6       /home/user/workspace/my-py-app
my-js-app     javascript   pnpm    01HJ9KJS5C2DRZQ4N5P6L7       /home/user/workspace/my-js-app
```

### Project Resolution

Commands that accept `<project>` support multiple formats:

```bash
pkt open my-app                   # By name
pkt open 01HJ9KJQ8D0XQZP2M3N4K5   # By ID
pkt open .                        # Current directory
```

## Package Manager Commands

### JavaScript

| PM   | Add           | Remove          | Install        | Init          |
| ---- | ------------- | --------------- | -------------- | ------------- |
| pnpm | `pnpm add`    | `pnpm remove`   | `pnpm install` | `pnpm init`   |
| npm  | `npm install` | `npm uninstall` | `npm install`  | `npm init -y` |
| bun  | `bun add`     | `bun remove`    | `bun install`  | `bun init -y` |

### Python

| PM     | Add           | Remove          | Install                           | Init                       |
| ------ | ------------- | --------------- | --------------------------------- | -------------------------- |
| uv     | `uv add`      | `uv remove`     | `uv sync`                         | `uv init`                  |
| pip    | `pip install` | `pip uninstall` | `pip install -r requirements.txt` | (creates requirements.txt) |
| poetry | `poetry add`  | `poetry remove` | `poetry install`                  | `poetry init -n`           |

### Go

| PM  | Add      | Remove        | Install           | Init          |
| --- | -------- | ------------- | ----------------- | ------------- |
| go  | `go get` | `go mod tidy` | `go mod download` | `go mod init` |

### Rust

| PM    | Add         | Remove         | Install       | Init         |
| ----- | ----------- | -------------- | ------------- | ------------ |
| cargo | `cargo add` | `cargo remove` | `cargo build` | `cargo init` |

## Configuration

Configuration is stored in `~/.pkt/config.json`:

```json
{
  "projects_root": "~/Documents/workspace",
  "default_pm": "pnpm",
  "editor": "code",
  "initialized": true
}
```

## Database

pkt uses an embedded SQLite database at `~/.pkt/pkt2.db` to track:

- **Projects** — ID, name, path, language, package manager
- **Dependencies** — name, version, type (prod/dev) per project

The database is always synced from manifest files — the source of truth.

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
