# pkt (Project Kit)

A cross-platform project manager and dependency tracker for JavaScript/Node.js projects.

**pkt (Project Kit)** organizes your projects, tracks dependencies, and abstracts package manager differences — all from a single, unified CLI.

## Features

- 🗂️ **Centralized workspace** — All projects live in one configurable folder
- 🔄 **Package manager agnostic** — Works with pnpm, npm, and bun seamlessly
- 📦 **Dependency tracking** — Database-backed tracking of all project dependencies
- 🆔 **Unique project IDs** — No more name conflicts, reference projects by ID
- 🚀 **Batch operations** — Add multiple packages in a single command

## Quick Start

```bash
# First-time setup (required)
pkt start

# Create a new project
pkt create my-app

# Or initialize an existing project
pkt init /path/to/existing-project

# Add dependencies
cd my-app
pkt add react react-dom
pkt add -D typescript eslint

# Install all dependencies from package.json
pkt install
```

## Installation

```bash
# Build from source
go build -o pkt .

# Move to PATH
sudo mv pkt /usr/local/bin/
```

**Requirements:**

- Go 1.24+
- PostgreSQL
- One of: pnpm, npm, or bun

## Commands

### Setup

| Command     | Description                               |
| ----------- | ----------------------------------------- |
| `pkt start` | Initialize pkt configuration and database |

### Project Management

| Command                | Description                                     |
| ---------------------- | ----------------------------------------------- |
| `pkt create <name>`    | Create a new empty project in workspace         |
| `pkt init [path]`      | Initialize an existing project for pkt tracking |
| `pkt list`             | List all tracked projects                       |
| `pkt open <project>`   | Open project in configured editor               |
| `pkt delete <project>` | Delete project from filesystem and database     |

### Dependency Management

> **Note:** These commands must be run inside a tracked project folder.

| Command               | Description                                |
| --------------------- | ------------------------------------------ |
| `pkt add <pkg...>`    | Add one or more dependencies               |
| `pkt add -D <pkg...>` | Add as dev dependencies                    |
| `pkt remove <pkg>`    | Remove a dependency                        |
| `pkt install`         | Install all dependencies from package.json |
| `pkt deps [project]`  | List project dependencies                  |

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

### `pkt start`

First-time setup wizard. Configures:

- Projects root folder (default: `~/Documents/workspace`)
- Default package manager (pnpm)
- Code editor command (e.g., `code`, `cursor`)

Creates `~/.pkt/config.json` and initializes the database.

### `pkt init [path]`

Initialize an existing project for pkt management.

```bash
pkt init .                        # Current directory
pkt init /path/to/my-project      # Specific project
```

**Behavior:**

- Reads `package.json` for project name
- Auto-detects package manager from lockfiles
- Moves project to workspace if outside it
- Syncs all dependencies to database

### `pkt install`

Install all dependencies from `package.json`.

```bash
cd my-project
pkt install
```

**Behavior:**

- Separates prod and dev dependencies
- Uses project's configured package manager
- Syncs installed versions to database

### `pkt add <packages...>`

Add one or more packages to the current project.

```bash
pkt add axios                     # Single package
pkt add axios lodash express      # Multiple packages
pkt add -D typescript eslint      # Dev dependencies
```

### Project Resolution

Commands that accept `<project>` support multiple formats:

```bash
pkt open my-app                   # By name
pkt open 01HJ9KJQ8D0XQZP2M3N4K5   # By ID
pkt open .                        # Current directory
```

If multiple projects share the same name, pkt prompts you to choose.

## Supported Package Managers

| PM   | Add           | Remove          | Init           |
| ---- | ------------- | --------------- | -------------- |
| pnpm | `pnpm add`    | `pnpm remove`   | `pnpm init -y` |
| npm  | `npm install` | `npm uninstall` | `npm init -y`  |
| bun  | `bun add`     | `bun remove`    | `bun init`     |

## Configuration

Configuration is stored in `~/.pkt/config.json`:

```json
{
  "projects_root": "~/Documents/workspace",
  "default_pm": "pnpm",
  "editor": "code",
  "initialized": true,
  "db_user": "pkt",
  "db_name": "pkt_db"
}
```

## Database

pkt uses PostgreSQL to track:

- Projects (ID, name, path, package manager)
- Dependencies (name, version, type per project)

The database is always synced from `package.json` — the source of truth.

## Architecture

- **Language:** Go
- **CLI Framework:** Cobra
- **Database:** PostgreSQL
- **ID Generation:** ULID

```
pkt/
├── cmd/          # CLI commands
├── internal/
│   ├── config/   # Configuration management
│   ├── db/       # Database operations
│   ├── pm/       # Package manager abstraction
│   └── utils/    # Utilities (fs, package.json, etc.)
└── main.go
```

## Safety

- pkt **only modifies** `package.json`
- Never touches source files or configs
- Database always syncs from filesystem
- Duplicate names resolved interactively

## License

MIT
