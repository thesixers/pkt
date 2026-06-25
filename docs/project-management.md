# Project Management

`pkt` tracks your repositories in an embedded SQLite database, giving you global visibility and control over all your codebases from anywhere on your machine.

---

## 1. Creating and Initializing

### Creating New Projects (`pkt create`)
**Command:** `pkt create <name> [-l <language>]`

This command scaffolds a new project directory inside your globally configured `projects_root` workspace.
- If you don't pass `-l`, an interactive prompt asks you to select between JavaScript, Python, Go, and Rust.
- It automatically creates the core manifest file (`package.json`, `requirements.txt`, `go.mod`, or `Cargo.toml`).
- It generates a ULID, stores the absolute path in the SQLite database, and begins tracking it.

### Tracking Existing Projects (`pkt init`)
**Command:** `pkt init [path]`

Converts an existing, unmanaged directory into a `pkt`-tracked project. 
- **Auto-Detection**: It parses the directory contents to automatically identify the Language and Package Manager (e.g., if it finds a `go.mod`, it sets it to Go/`go mod`).
- **Safety Guards**: `pkt init` is hard-coded to **never** initialize in critical system directories. It explicitly checks the absolute path and aborts if you attempt to run it in `/`, `~` (Home), `/usr`, `/var`, `/etc`, or `/bin`.

### Cloning Repositories (`pkt clone`)
**Command:** `pkt clone <url>`

A wrapper around `git clone`. It executes the clone into your global workspace root, parses the directory name from the URL, automatically navigates into it, and runs `pkt init .` to seamlessly bring it under `pkt` tracking in a single action.

---

## 2. Navigating and Modifying

### Opening Projects (`pkt open`)
**Command:** `pkt open <project-name>`

Instantly opens the target project in your configured editor (e.g., `code`, `cursor`, `vim`), regardless of where you are in the terminal. If you have multiple projects with the same name, an interactive list prompts you to pick the specific ULID.

### Renaming Projects (`pkt rename`)
**Command:** `pkt rename <old-name> <new-name>`

Renames the project internally in the SQLite database. Note: This does *not* rename the physical folder on disk, it only changes how `pkt` refers to it in terminal outputs and lookup commands.

### Deleting Projects (`pkt delete`)
**Command:** `pkt delete <project-name...>`

**DANGER:** This completely removes the project from the SQLite database **AND** recursively deletes the physical folder and all its contents from your hard drive (`os.RemoveAll`). It supports batch deletion (`pkt delete app1 app2 app3`).

---

## 3. Workspace Intelligence

### Listing Projects (`pkt list`)
**Command:** `pkt list [-l <language>] [-a]`

Uses `go-pretty/table` to render a PM2-style rounded ASCII table of all your tracked projects.
- By default, it shows the Project Name, Language, and Absolute Path.
- The `-l` flag filters the table by language.
- The `-a` (all) flag expands the table to include the ULID, active Package Manager, and the **Total Disk Size** of the project directory.

### Checking Git States (`pkt status`)
**Command:** `pkt status [project-name]`

Traverses the `.git` tree of your projects to report their synchronization state.
- If run without arguments, it scans **every single tracked project** in the database.
- It renders a table displaying the Project Name, current Branch, and Status (`clean`, `uncommitted changes`, or `ahead/behind origin`).
- You can pass a specific project name to only check that repository.

### Workspace Analytics (`pkt stats`)
**Command:** `pkt stats`

Aggregates metadata from the SQLite database and the filesystem to provide high-level analytics:
- Total number of tracked projects.
- Total disk space consumed by the entire workspace.
- A breakdown table showing the exact number of projects and total disk space utilized *per language*.

### Pruning Caches (`pkt clean`)
**Command:** `pkt clean`

Recursively traverses all tracked projects to find and permanently delete massive dependency cache folders (`node_modules` and `.venv`). This is incredibly useful for instantly reclaiming gigabytes of disk space across stale projects.
