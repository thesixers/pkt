# pkt - Quick Start Guide

## Prerequisites

Before using pkt, ensure you have:

1. At least one language runtime installed (Node.js, Python, Go, or Rust)
2. The pkt binary built or downloaded from [Releases](https://github.com/thesixers/pkt/releases)

---

## Building pkt

```bash
git clone https://github.com/thesixers/pkt.git
cd pkt
make build        # creates bin/pkt
# Or install system-wide
make install      # installs to /usr/local/bin
```

---

## First-Time Setup

```bash
pkt start
```

You'll be prompted for:

- **Projects root folder** (default: `~/Documents/workspace`)
- **Automatically detected default package managers** for JavaScript, Python, Go, and Rust
- **Editor command** (e.g., `code`, `vim`, `cursor`)

This creates `~/.pkt/config.json` and an embedded SQLite database.

---

## Usage Examples

### Create and manage projects

```bash
pkt create my-app           # Create new JS project (prompts for language)
pkt create my-api -l python # Create with a specific language
pkt list                    # List all tracked projects
pkt list -a                 # Include size, ID, and package manager
pkt open my-app             # Open in configured editor
pkt delete my-app           # Delete project + database record
pkt rename my-app new-name  # Rename a tracked project
```

### Clone and track repos

```bash
pkt clone https://github.com/user/repo   # Clone and auto-track
```

### Dependencies

```bash
cd my-app
pkt add axios react          # Add production dependencies
pkt add -D typescript        # Add dev dependency
pkt remove axios             # Remove a dependency
pkt update                   # Update all dependencies
pkt outdated                 # Check for outdated packages
pkt deps                     # List project dependencies
```

### Run scripts

```bash
pkt run dev          # npm/pnpm run dev, or go run .
pkt run test         # run tests for any language
pkt run build        # build for any language
pkt exec my-app ls   # run a command in another project's context
```

---

## AI Commands

### Set up an AI provider first

```bash
# Cloud providers (require API key)
pkt change ai add provider groq
pkt change ai add key sk-xxxx --provider groq
pkt change ai add key AIxxxx --provider gemini
pkt change ai add key sk-xxxx --provider openai

# Local Ollama (no key needed)
pkt change ai add provider ollama
pkt change ai ollama

# Self-hosted at a custom URL
pkt change ai add provider myserver --url http://localhost:8080
pkt change ai add model phi3 --provider myserver
pkt change ai myserver
```

### Use AI features

```bash
pkt debug error.log           # or: cat error.log | pkt debug
pkt add --ai "I need a library for sending emails"
pkt chat                      # Launch autonomous agent REPL
```

---

## Configuration

```bash
pkt config                                                  # Show all config + provider registry
pkt change editor cursor                                    # Change editor
pkt change pm npm                                           # Change current project's manager
pkt change pm uv -l py -g                                  # Change Python's default manager
pkt change ai groq                                          # Switch active AI provider
pkt change root ~/Projects                                  # Set root for new projects and clones
pkt change ai add model llama-3.3-70b-versatile --provider groq   # Pin a model
```

Config file: `~/.pkt/config.json`

---

## All Commands

```
pkt start              Initialize pkt
pkt create <name>      Create new project
pkt init [path]        Track existing project
pkt list               List all tracked projects
pkt open <project>     Open in editor
pkt delete <project>   Delete project
pkt clone <url>        Clone and track repo
pkt rename <project>   Rename a project
pkt stats              Show workspace statistics
pkt status             Show git status for all projects
pkt clean              Prune node_modules / .venv caches

pkt add <pkg>          Add dependency
pkt remove <pkg>       Remove dependency
pkt update             Update dependencies
pkt outdated           Check for outdated packages
pkt deps               List dependencies

pkt run <script>       Run a script
pkt exec <proj> <cmd>  Run command in another project

pkt debug [file]       Debug errors with AI
pkt chat               Launch autonomous AI agent

pkt config             Show configuration
pkt change <key> <val> Change configuration
pkt change ai add ...  Add provider/key/model
pkt change ai remove ...  Remove provider/key/model
```

---

## Troubleshooting

**"pkt not initialized"** → Run `pkt start` first

**"not in a tracked project"** → Run `pkt init .` inside the folder, or `cd` into a tracked project

**"no AI provider configured"** → Run `pkt change ai add provider groq` then `pkt change ai add key sk-xxxx --provider groq`

**"rate limit exceeded"** → Add another key: `pkt change ai add key sk-yyy --provider groq` or use a different provider

---

**For full documentation, see [README.md](README.md)**
