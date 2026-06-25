# Execution & Scripts

`pkt` features a "Zero-Flag Execution Architecture" that allows you to run scripts, compile binaries, and execute tests without needing to memorize the unique syntax of the underlying runtime.

---

## 1. Running Scripts (`pkt run`)

**Command:** `pkt run <target>`

This command intelligently executes code based on two contextual layers: the active project environment, and the file extension of the target.

### Layer 1: Contextual Project Execution
If you are currently inside a tracked project, `pkt` delegates execution to that project's package manager.

- **JavaScript**: `pkt run dev` translates exactly to `npm run dev` (or `pnpm run dev`, `bun run dev`).
- **Python**: `pkt run test` executes your configured test suite (e.g. `pytest`) inside your `.venv`.
- **Go / Rust**: `pkt run build` triggers `go build` or `cargo build`.

### Layer 2: Extension Inference (Direct Execution)
If you run `pkt run main.go` from a directory that is *not* tracked by `pkt` (or even from your desktop), the execution engine skips the package manager entirely and inspects the file extension (`filepath.Ext`).

It then automatically invokes the correct interpreter:
- `.js`, `.ts`, `.tsx`, `.mjs`: Executes via `bun` (if available) or `node`/`tsx`.
- `.py`: Executes via `python3`.
- `.go`: Executes via `go run`.
- `.rs`: Compiles and executes via `rustc`.

If the file has no recognized extension (e.g., `./binary`), it drops down to the OS and executes it natively.

---

## 2. Remote Context Execution (`pkt exec`)

**Command:** `pkt exec <project-name> <command...>`

This is one of `pkt`'s most powerful features. It allows you to execute arbitrary bash commands *in the directory context of another project* without ever leaving your current terminal working directory.

### How it works
1. It queries the SQLite database for the ULID or Name.
2. It retrieves the absolute path of the target project.
3. It spawns a new OS process (`exec.Command`) with the `Dir` field set to the target's absolute path.
4. It streams the `Stdout` and `Stderr` directly back to your active shell.

### Use Cases
Imagine you are working in your `frontend` project, but you need to restart the backend database or build the API server.
```bash
# You are currently inside ~/workspace/frontend
pkt exec backend-api go build -o server .
pkt exec database docker-compose up -d
```
You have full access to control your entire workspace from anywhere.
