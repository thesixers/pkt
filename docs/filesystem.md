# Filesystem Routing

`pkt` overrides standard OS filesystem commands (`cp` and `mv`) to introduce "Project Routing"—the ability to seamlessly mix absolute paths, relative paths, and abstract project names in a single command.

---

## 1. The `@` Project Prefix

To differentiate between a folder named `api` and the tracked project `api`, `pkt` uses the `@` prefix. 

When the parser encounters an argument starting with `@` (e.g., `@backend-api/config.json`), it activates the **Resolution Engine**:
1. It strips the `@` and splits the string at the first `/`.
2. The first chunk (`backend-api`) is passed to the SQLite database.
3. The database returns the absolute path of the project (e.g., `/home/user/workspace/backend-api`).
4. The engine stitches the rest of the path (`config.json`) onto the absolute path.

This means you never have to type `../../backend-api/config.json` again.

---

## 2. Copying Files (`pkt cp`)

**Command:** `pkt cp <source> <destination>`

The standard Go standard library does not contain a recursive directory copy function, so `pkt cp` is backed by a custom, highly robust `CopyItem` engine (`internal/utils/copy.go`).

### Behaviors
- **Files to Files**: Exact byte-for-byte stream via `io.Copy`, preserving the `os.FileMode` permissions.
- **Files to Directories**: If you copy a file to an existing directory (`pkt cp file.txt @my-proj/assets/`), `pkt` automatically appends the base filename (`file.txt`) so it correctly lands *inside* the folder.
- **Directories to Directories**: Uses `filepath.WalkDir` to recursively recreate the entire nested folder structure and copy every individual file.

**Examples:**
```bash
# Project to Project
pkt cp @api-server/config.json @frontend/config.json

# Local to Project
pkt cp ./local.txt @my-project/assets/

# Recursive Project to Local
pkt cp -r @legacy-app/data ./data
```

---

## 3. Moving Files (`pkt mv`)

**Command:** `pkt mv <source> <destination>`

Operates identically to `cp` regarding the `@` resolution engine, but executes a filesystem move.

### Cross-Device Fallbacks
Moving a file is usually instantaneous because `os.Rename` simply updates the hard drive's filesystem pointer. However, if you attempt to move a file from an external USB drive into a project on your main SSD, the OS block-level rename will fail with an `EXDEV` (Cross-device link) error.

To ensure `pkt mv` never fails silently, it catches `Rename` errors and automatically falls back to a full memory-buffered `CopyItem` execution, followed by an `os.RemoveAll` on the source!
