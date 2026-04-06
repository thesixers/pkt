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
- **Default package manager** (pnpm, npm, or bun)
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
pkt search react            # Search through project names
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
pkt install                  # Install all dependencies
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
pkt config set-ai groq   sk-xxxx
pkt config set-ai gemini AIxxxx
pkt config set-ai openai sk-xxxx

# Local Ollama (no key needed)
pkt config set-ai ollama
pkt config ai ollama

# Self-hosted at a custom URL
pkt config set-ai myserver --url http://localhost:8080
pkt config set-model myserver phi3
pkt config ai myserver
```

### Use AI features

```bash
pkt ask "how do I add authentication to this project?"
pkt generate "REST API endpoint for user login"
pkt debug error.log           # or: cat error.log | pkt debug
pkt add --ai "I need a library for sending emails"
pkt chat                      # Launch autonomous agent REPL
```

---

## Configuration

```bash
pkt config                          # Show all config + provider registry
pkt config editor cursor            # Change editor
pkt config pm npm                   # Change default package manager
pkt config ai groq                  # Switch active AI provider
pkt config set-model groq llama-3.3-70b-versatile   # Pin a model
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
pkt search <query>     Search projects
pkt stats              Show workspace statistics
pkt status             Show git status for all projects
pkt clean              Prune node_modules / .venv caches

pkt add <pkg>          Add dependency
pkt remove <pkg>       Remove dependency
pkt install            Install all dependencies
pkt update             Update dependencies
pkt outdated           Check for outdated packages
pkt deps               List dependencies

pkt run <script>       Run a script
pkt exec <proj> <cmd>  Run command in another project

pkt ask <question>     Ask AI about your project
pkt generate <desc>    Generate code with AI
pkt debug [file]       Debug errors with AI
pkt chat               Launch autonomous AI agent

pkt config             View/update configuration
pkt pm set <pm> <proj> Change package manager
```

---

## Troubleshooting

**"pkt not initialized"** → Run `pkt start` first

**"not in a tracked project"** → Run `pkt init .` inside the folder, or `cd` into a tracked project

**"no AI provider configured"** → Run `pkt config set-ai groq sk-xxxx` then `pkt config ai groq`

**"rate limit exceeded"** → Switch model: `pkt config set-model groq llama-3.1-8b-instant` or use a different provider

---

**For full documentation, see [README.md](README.md)**
