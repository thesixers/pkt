# Dependency Management

`pkt` acts as a unified abstraction layer over 8 different package managers (`npm`, `pnpm`, `bun`, `pip`, `poetry`, `uv`, `go mod`, `cargo`). 

> **Important Context**: Almost all commands in this document **must** be executed from within a tracked project directory (or a subdirectory of one). `pkt` traverses up the directory tree to identify the active project and its configured package manager before executing.

---

## 1. Adding and Removing Dependencies

### Adding Packages (`pkt add`)
**Command:** `pkt add <package...> [-D | --dev]`

Adds dependencies to the current project. 
- **Batching**: You can add multiple packages at once (`pkt add react react-dom`).
- **Dev Dependencies**: The `-D` flag routes to the specific PM's development installation flag (e.g., `npm i -D`, `poetry add --group dev`).
- **Native Streaming**: Instead of swallowing the terminal output, `pkt` connects the underlying package manager process directly to `os.Stdout` (`runCommandInteractive`). This means you see the native progress bars, warnings, and formatting exactly as if you had run the underlying tool.

### Removing Packages (`pkt remove`)
**Command:** `pkt remove <package...>`

Uninstalls the dependency from the project and removes its entry from the manifest (`package.json`, `Cargo.toml`, etc.) and the SQLite tracking database.

---

## 2. Installing and Updating

### Fresh Installations (`pkt install`)
**Command:** `pkt install`

Reads the existing manifest and lockfiles to execute a clean installation of all required dependencies. Used primarily after cloning a repository.

### Updating Packages (`pkt update`)
**Command:** `pkt update [package...]`

- If run without arguments, it triggers a global update of all dependencies in the project to their latest versions allowed by the manifest constraints.
- If passed specific package names, it only updates those dependencies.

### Checking Outdated Packages (`pkt outdated`)
**Command:** `pkt outdated`

Executes the native PM outdated check (e.g., `npm outdated`, `cargo outdated`) and parses the output to render a beautiful `go-pretty` rounded ASCII table showing the Current, Wanted, and Latest versions of lagging dependencies.

---

## 3. Python Virtual Environment Magic

Python dependency management is notoriously fragmented. `pkt` heavily abstracts this to provide a completely seamless "Node-like" experience.

### The Problem
Typically, running `pip install requests` installs the package globally, breaking your system Python. You must remember to `python -m venv .venv` and `source .venv/bin/activate` every time.

### The `pkt` Solution
When you run `pkt add requests` inside a Python project, `pkt` intercepts the command and executes the following autonomous workflow based on your configured PM:

#### 1. If using `pip`:
- It checks if a `.venv` directory exists in the project root.
- If missing, it silently executes `python3 -m venv .venv`.
- It dynamically rewrites your command to explicitly use the virtual environment executable: `.venv/bin/pip install requests`.
- Afterward, it automatically runs `.venv/bin/pip freeze > requirements.txt` to ensure your manifest is locked.

#### 2. If using `poetry`:
- It globally configures Poetry to force local virtual environments (`poetry config virtualenvs.in-project true`).
- It executes `poetry add requests`.

#### 3. If using `uv`:
- It leverages `uv`'s native blazing-fast venv integration (`uv pip install requests`).

*Result: You never have to manually activate a virtual environment or run `pip freeze` ever again.*
