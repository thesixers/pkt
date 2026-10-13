# Configuration & State

`pkt` is incredibly strict about preserving a clean state. It does not litter your workspace with `.pkt` folders in every project. Instead, it relies on a single, centralized Global Configuration file.

---

## 1. The Global Config File

**Location:** `~/.pkt/config.json`

This JSON file acts as the source of truth for your entire workspace environment.
```json
{
  "projects_root": "~/Documents/workspace",
  "default_pms": {
    "javascript": "pnpm",
    "python": "uv",
    "go": "go",
    "rust": "cargo"
  },
  "editor": "code",
  "initialized": true,
  "ai": {
    "provider": "groq",
    "providers": {
      "groq": {
        "api_key": "sk-xxxx",
        "model": "llama-3.3-70b-versatile",
        "url": ""
      }
    }
  }
}
```

---

## 2. Viewing and Modifying Configuration

**View:** `pkt config`

`pkt config` is read-only: it prints your active configuration and all registered AI providers.

**Change:** `pkt change <key> <value>`

### Core Settings
- `pkt change editor <cmd>`: Sets the binary to execute when you run `pkt open`. Valid inputs: `code`, `cursor`, `vim`, `nano`, `goland`, etc.
- `pkt change pm <pm>`: Changes the package manager for the current tracked project.
- `pkt change pm <pm> -l <lang> -g`: Changes the default package manager for future projects of a language. Languages accept `js`, `py`, `go`, and `rs`.
- `pkt change ai <provider>`: Selects the active AI provider.
- `pkt change root <path>`: Sets the projects root used for future `pkt create` and `pkt clone` commands. The path is normalized to an absolute path and created if it does not exist.

Changing the projects root does not move existing projects or rewrite their database records; tracked projects retain their current locations.

### Project Overrides (The `--global` Flag)
When you run a command like `pkt change pm npm`:
- **Inside a tracked project**: It updates that project's package manager.
- **Outside a tracked project**: pkt asks which language default should change unless you provide `-l`.
- **With the `--global` flag** (`pkt change pm npm -l js -g`): It updates the selected language default even when you are inside a project.

---

## 3. AI Provider Registry

The AI integration supports multiple providers, API keys, and models simultaneously.

### Adding Cloud Providers (API Keys Required)

```bash
pkt change ai add provider groq
pkt change ai add key sk-123456789 --provider groq
```

### Adding Local Providers (No Keys Required)

If you are running a local model engine (like `ollama`), register it without a key.
```bash
pkt change ai add provider ollama
```

### Custom Endpoints & Self-Hosted Models

```bash
pkt change ai add provider my-server --url http://192.168.1.100:1234/v1
```

### Adding and Pinning Models

```bash
pkt change ai add model llama-3.1-8b-instant --provider groq
pkt change ai add model llama-3.3-70b-versatile --provider groq
```

The first model added becomes the default. To change the default later:
```bash
pkt change ai add model llama-3.3-70b-versatile --provider groq
```
(adding an already-saved model is a no-op, so use `/model` in chat or remove and re-add)

### Switching Active Provider

```bash
pkt change ai groq    # Switch active provider
pkt change ai ollama
```

Or switch inside `pkt chat` without leaving the session:
```
/provider groq
/model llama-3.3-70b-versatile
```

### Removing Providers, Keys, and Models

```bash
pkt change ai remove key sk-... --provider groq
pkt change ai remove model llama-3.1-8b-instant --provider groq
pkt change ai remove provider groq
```
