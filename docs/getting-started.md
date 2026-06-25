# Getting Started

## First-Time Setup

Before you can manage projects, you must initialize your workspace.

```bash
pkt start
```

You'll be prompted for:
- **Projects root folder** (default: `~/Documents/workspace`)
- **Default package manager** (e.g., pnpm, npm, or bun)
- **Editor command** (e.g., `code`, `vim`, `cursor`)

This creates `~/.pkt/config.json` and initializes an embedded SQLite database to track your projects.

## Creating Your First Project

Let's create a new project. You can let `pkt` prompt you for the language, or specify it directly.

```bash
# Interactive creation
pkt create my-app

# Direct creation specifying the language
pkt create my-api -l python
pkt create my-cli -l go
```

## Initializing an Existing Project

If you already have a repository, you can bring it under `pkt` management:

```bash
cd /path/to/existing-project
pkt init .
```
`pkt` will automatically detect the language (based on `package.json`, `go.mod`, `Cargo.toml`, etc.) and start tracking it.

## Adding Dependencies

Now let's add some packages to your project:

```bash
cd my-app
pkt add react react-dom        # JavaScript
pkt add requests flask         # Python
pkt add -D typescript eslint   # Dev dependencies
```

`pkt` will automatically stream the native progress bar from your package manager directly to your terminal!

## Running Scripts

You can execute your project's scripts directly through `pkt` without needing to remember the underlying package manager syntax:

```bash
pkt run dev                    # Runs npm/pnpm run dev
pkt run test                   # Runs tests for any language
```

**`pkt run` examples by language:**

| Language | Commands |
| --- | --- |
| **JavaScript** | `pkt run dev`, `pkt run build`, `pkt run <any-script>` |
| **Python** | `pkt run test` (pytest), `pkt run main.py` |
| **Go** | `pkt run run`, `pkt run test`, `pkt run build` |
| **Rust** | `pkt run run`, `pkt run test`, `pkt run build` |
