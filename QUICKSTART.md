# pkt - Quick Start Guide

## Prerequisites

Before using pkt, ensure you have:

1. **PostgreSQL** installed and running
2. At least one package manager installed: **pnpm**, **npm**, or **bun**
3. The pkt binary built (see below)

---

## Building pkt

```bash
cd /home/genesix/Documents/Workspace/pkt

# Build the binary
make build

# Or install system-wide
make install
```

This creates an executable at `bin/pkt` (or installs to `/usr/local/bin`).

---

## First-Time Setup

Run this command first:

```bash
./bin/pkt start
```

You'll be prompted for:

- **Projects root folder** (default: `~/Documents/workspace`)
- **Default package manager** (pnpm, npm, or bun)
- **Editor command** (e.g., `code`, `vim`)

This creates:

- Config file: `~/.pkt/config.json`
- PostgreSQL database: `pkt_db`

---

## Usage Examples

### Create a New Project

```bash
./bin/pkt create my-awesome-app
```

Creates folder and tracks it with a unique ID.

### List All Projects

```bash
./bin/pkt list
```

Shows all tracked projects with their IDs, package managers, and paths.

### Open Project in Editor

```bash
# By name
./bin/pkt open my-awesome-app

# By ID
./bin/pkt open 01HJ9KJQ8D...

# Current directory
cd ~/Documents/workspace/my-awesome-app
../../pkt open .
```

### Add Dependencies

```bash
cd ~/Documents/workspace/my-awesome-app

# Production dependency
../../pkt add axios

# Dev dependency
../../pkt add -D typescript
```

### List Dependencies

```bash
# From inside project folder
../../pkt deps

# Or from anywhere
./bin/pkt deps my-awesome-app
```

### Remove Dependencies

```bash
cd ~/Documents/workspace/my-awesome-app
../../pkt remove axios
```

### Change Package Manager

```bash
# Switch to npm
./bin/pkt pm set npm my-awesome-app

# Or from inside project
cd ~/Documents/workspace/my-awesome-app
../../pkt pm set bun .
```

This automatically rewrites package.json scripts to use the new PM.

### Delete Project

```bash
./bin/pkt delete my-awesome-app
```

Prompts for confirmation before deleting folder and database record.

---

## Configuration

### Database Connection

Default connection: `localhost:5432` as user `postgres` to database `pkt_db`.

Override with environment variables:

```bash
export PKT_DB_HOST=localhost
export PKT_DB_PORT=5432
export PKT_DB_USER=postgres
export PKT_DB_PASSWORD=yourpassword  # optional
export PKT_DB_NAME=pkt_db
```

### Config File Location

`~/.pkt/config.json`

Example:

```json
{
  "projects_root": "/home/genesix/Documents/workspace",
  "default_pm": "pnpm",
  "editor_command": "code"
}
```

---

## All Commands

```
pkt start                           # Initialize pkt
pkt create <name>                   # Create new project
pkt list                            # List all projects
pkt open <project|id|.>             # Open in editor
pkt delete <project|id>             # Delete project
pkt add <package> [-D]              # Add dependency
pkt remove <package>                # Remove dependency
pkt deps [project|id|.]             # List dependencies
pkt pm set <pm> <project|id|.>      # Change package manager
```

Use `--help` with any command for details.

---

## Tips

1. **Duplicate Names**: If multiple projects have the same name, pkt prompts you to select which one.

2. **Current Directory**: Use `.` to reference the project in your current folder:

   ```bash
   pkt open .
   pkt pm set npm .
   ```

3. **Project IDs**: Every project has a unique ULID. Use it when names are ambiguous:

   ```bash
   pkt open 01HJ9KJQ8D...
   ```

4. **Dependencies**: The `add` and `remove` commands must be run inside a project folder.

5. **Package Manager Switching**: When you change PMs, pkt automatically updates scripts in package.json.

---

## Troubleshooting

**"pkt not initialized"**
→ Run `pkt start` first

**"database connection failed"**
→ Ensure PostgreSQL is running and credentials are correct

**"not in a tracked project"**
→ Navigate to a project folder created with `pkt create`, or run from inside one

**"package manager not available"**
→ Install the package manager (pnpm/npm/bun) or choose a different one

---

## Development

```bash
# Run tests
make test

# Build for development
make build

# Clean build artifacts
make clean
```

---

**For full documentation, see [README.md](file:///home/genesix/Documents/Workspace/pkt/README.md)**
