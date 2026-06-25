# Setup & Installation

This document covers the exhaustive details of installing, updating, and initializing the `pkt` ecosystem on your machine.

---

## 1. Installation

### Option A: Binary Download
`pkt` compiles down to a single binary with zero external dependencies. The fastest way to install is to grab the latest release from GitHub.

| Platform | Binary |
| --- | --- |
| Linux | `pkt-linux-amd64` |
| macOS | `pkt-darwin-amd64` (Intel) / `pkt-darwin-arm64` (Apple Silicon) |
| Windows | `pkt-windows-amd64.exe` |

*Download the binary, mark it as executable (`chmod +x pkt`), and move it to a folder in your `$PATH` (e.g., `/usr/local/bin`).*

### Option B: Building from Source
If you prefer compiling locally, ensure you have Go 1.24+ installed.

```bash
git clone https://github.com/thesixers/pkt.git
cd pkt
make build        # Creates the binary in ./bin/pkt
sudo make install # Moves the binary to /usr/local/bin
```
*Note: The `Makefile` dynamically embeds the exact git tag/commit hash into the binary using `-ldflags` during compilation.*

---

## 2. Initialization (`pkt start`)

**Command:** `pkt start`

You must run `pkt start` before using any other commands. This initializes the global configuration and provisions the embedded SQLite database.

**What it does:**
1. **Interactive Prompt**: It asks for your primary workspace folder (default: `~/Documents/workspace`), your preferred default JavaScript package manager (`npm`, `pnpm`, `bun`), and your default terminal editor (`code`, `vim`, `cursor`).
2. **Database Provisioning**: It creates the `~/.pkt/` directory and initializes `pkt2.db`.
3. **Runtime Auditing**: It actively scans your system `$PATH` for required and optional runtimes (Node.js, Go, Python, Rust, Cargo, Bun, pnpm, uv, poetry).
4. **Onboarding**: If you are missing a critical runtime, it halts the setup and provides a precise, OS-specific copy-paste snippet (e.g., `curl -fsSL https://bun.sh/install | bash`) to install the missing tool immediately.

---

## 3. Shell Auto-Completion (`pkt completion`)

**Command:** `pkt completion <shell>`

`pkt` features incredibly robust shell completion. When configured, pressing `Tab` will auto-complete not only command names, but also **tracked project names**, language flags, and package manager arguments.

### Bash Installation
```bash
echo 'source <(pkt completion bash)' >> ~/.bashrc
source ~/.bashrc
```

### Zsh Installation (macOS default)
```bash
echo 'source <(pkt completion zsh)' >> ~/.zshrc
source ~/.zshrc
```

### Fish Installation
```fish
pkt completion fish | source
```

---

## 4. Maintenance Commands

### Checking the Version (`pkt version`)
**Command:** `pkt version` (or `pkt -v`, `pkt --version`)
Outputs the exact version of the binary. This is generated natively at compile time via Go's `-ldflags`.

### Automatic Self-Updating (`pkt upgrade`)
**Command:** `pkt upgrade`

Instead of manually checking GitHub for new releases, `pkt` features a native, atomic self-updater.

**How it works:**
1. Makes an API call to `https://api.github.com/repos/thesixers/pkt/releases/latest`.
2. Compares the OS and Architecture (`runtime.GOOS`, `runtime.GOARCH`) to find the exact matching pre-compiled asset.
3. Downloads the binary to a temporary file in your system's temp directory.
4. Executes an atomic `os.Rename` to seamlessly overwrite the currently running `pkt` executable in your `$PATH`.
5. Requires zero external dependencies (no Go compiler required on the user's machine to upgrade).
