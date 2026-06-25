# pkt: Exhaustive Manual

Welcome to the exhaustive technical manual for **pkt** (Project Kit).

`pkt` is a cross-platform project manager, dependency tracker, and AI-assisted development hub for **JavaScript, Python, Go, and Rust** projects. It abstracts away the complexity of juggling different package managers (`npm`, `pnpm`, `bun`, `pip`, `poetry`, `uv`, `cargo`, `go mod`) and centralizes all your work into a highly optimized database-backed CLI.

---

## 🏗️ Architecture & Internals

Understanding how `pkt` works under the hood is critical to leveraging its full power.

### 1. The Embedded Database
Unlike traditional CLI tools that rely on parsing massive JSON or YAML files on every command, `pkt` ships with an **embedded, pure-Go SQLite database**.
- **Location**: `~/.pkt/pkt2.db`
- **Purpose**: Tracks every project's Absolute Path, Name, Language, Package Manager, and Unique ID. It also maintains a ledger of all your dependencies across all projects.
- **Why?**: This allows `pkt` to perform instantaneous lookups, power intelligent cross-project commands (like `pkt cp` and `pkt exec`), and provide lightning-fast shell auto-completion without having to scan your hard drive.

### 2. ULID Project Tracking
Every project tracked by `pkt` is assigned a Unique Lexicographically Sortable Identifier (ULID).
- **Format**: A 26-character alphanumeric string (e.g., `01J1E7XXXXX`).
- **Conflict Resolution**: If you have two projects named `api` in different folders, `pkt` uses the ULID to uniquely identify them, ensuring commands like `pkt run` or `pkt open` never target the wrong codebase.

### 3. Language & Package Manager Abstraction
`pkt` uses an internal "Provider Interface" (`internal/pm`) to abstract package management logic.
When you run `pkt add react`, the core engine routes the command to the currently active project's Provider.
- If it's a Node project set to `pnpm`, it silently translates the command to `pnpm install react`.
- If it's a Python project set to `uv`, it creates a `.venv`, activates it, and runs `uv pip install react`.

---

## 🛡️ Safety & Data Guarantees

`pkt` is designed to be highly destructive to caches and dependencies, but **absolutely safe** for your source code.

### 1. Read-Only Source Code
No `pkt` command (other than the AI autonomous agent `pkt chat`) will ever modify your application source code (`.js`, `.go`, `.py`, `.rs`). `pkt` exclusively manages manifest files (`package.json`, `requirements.txt`, etc.).

### 2. Initialization Guards
The `pkt init` command contains strict safety guards to prevent you from accidentally tracking your entire hard drive. It explicitly blocks initialization in:
- The system root (`/`)
- Your User Home Directory (`~`)
- System folders (`/etc`, `/usr`, `/var`)

### 3. Cross-Device Fallbacks
Commands like `pkt mv` attempt to use native, instantaneous OS renaming. If it detects a cross-device or cross-drive boundary (where OS renaming fails), it gracefully catches the error and executes a full byte-by-byte recursive `Copy + Delete` sequence.

---

## 📚 Documentation Index

Ready to dive in? Choose a section below for exhaustive command breakdowns:

1. [Setup & Installation](setup-and-installation.md)
2. [Project Management](project-management.md)
3. [Dependency Management](dependency-management.md)
4. [Execution & Scripts](execution.md)
5. [Filesystem Routing](filesystem.md)
6. [AI Features](ai-features.md)
7. [Configuration](configuration.md)
