# Configuration & State

`pkt` is incredibly strict about preserving a clean state. It does not litter your workspace with `.pkt` folders in every project. Instead, it relies on a single, centralized Global Configuration file.

---

## 1. The Global Config File

**Location:** `~/.pkt/config.json`

This JSON file acts as the source of truth for your entire workspace environment.
```json
{
  "projects_root": "~/Documents/workspace",
  "default_pm": "pnpm",
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

## 2. Modifying Configuration (`pkt config`)

**Command:** `pkt config [subcommand]`

If you run `pkt config` with no arguments, it prints a stylized readout of your entire active configuration, including all registered AI providers.

### Core Settings
- `pkt config editor <cmd>`: Sets the binary to execute when you run `pkt open`. Valid inputs: `code`, `cursor`, `vim`, `nano`, `goland`, etc.
- `pkt config pm <pm>`: Changes the global default package manager used when you run `pkt create` for a JavaScript project.

### Project Overrides (The `--global` Flag)
When you run a command like `pkt config pm npm`:
- **Inside a tracked project**: It strictly updates the SQLite database record for *that specific project* to use `npm`.
- **With the `--global` flag** (`pkt config pm npm -g`): It updates your `~/.pkt/config.json` so all future projects default to `npm`.

---

## 3. AI Provider Registry

The AI integration supports multiple providers simultaneously, allowing you to seamlessly hot-swap between models depending on your needs.

### Adding Cloud Providers (API Keys Required)
**Command:** `pkt config set-ai <provider> <api-key>`

To use OpenAI, Groq, or Gemini, you must provide an API key.
```bash
pkt config set-ai groq sk-123456789
```

### Adding Local Providers (No Keys Required)
**Command:** `pkt config set-ai <provider>`

If you are running a local model engine (like `ollama`), you can register it without a key.
```bash
pkt config set-ai ollama
```

### Custom Endpoints & Self-Hosted Models
**Command:** `pkt config set-ai <provider> --url <endpoint>`

If you are hosting your own inference server (e.g., vLLM or LM Studio), you can point `pkt` to your custom URL.
```bash
pkt config set-ai my-server --url http://192.168.1.100:1234/v1
```

### Pinning Models and Hot-Swapping
**Command:** `pkt config set-model <provider> <model>`

Most providers (like Groq) have multiple models available. You can pin a specific model to the provider.
```bash
pkt config set-model groq llama-3.1-8b-instant
```

Once you have multiple providers registered, you can hot-swap which one the `pkt chat`, `ask`, and `debug` commands use in an instant:
```bash
pkt config ai ollama  # Switches all AI commands to use local ollama
pkt config ai groq    # Switches back to Groq
```
