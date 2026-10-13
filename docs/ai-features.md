# AI Intelligence Features

`pkt` is uniquely positioned to offer world-class AI assistance because it already knows the absolute path, language, and dependency graph of your active project. 

The core AI engine (`internal/ai`) acts as a unified abstraction layer, allowing you to seamlessly swap between Groq, Gemini, OpenAI, or localized models like Ollama.

---

## 1. The Autonomous Agent (`pkt chat`)

**Command:** `pkt chat`

This launches a fully interactive REPL (Read-Eval-Print Loop) backed by `chzyer/readline`. 

**Capabilities:**
Unlike a standard ChatGPT window, the `pkt` agent has "Tools" natively injected into its prompt. 
- It can read your local files (`view_file`).
- It can execute bash commands on your host machine to compile code or run tests (`run_command`).
- It can actively write and replace code in your project.

**Usage:**
Run `pkt chat` and simply type: `"Find the authentication bug in src/auth.go, fix it, and run the tests to verify."` The agent will autonomously read the file, write the patch, execute `go test`, and report back to you.

---

## 2. Native Debugging (`pkt debug`)

**Command:** `pkt debug [filename]`

This command is engineered to instantly diagnose failing logs or stack traces.

**Piping Support:**
`pkt debug` is specifically designed to read from `os.Stdin`. This means you can pipe failing terminal commands directly into the AI for instant diagnosis without copy-pasting.

**Example 1 (File):**
`pkt debug crash.log`

**Example 2 (Pipe):**
`go run main.go 2>&1 | pkt debug`
*The AI will read the stdout/stderr stream from your crashed Go program and immediately tell you what went wrong.*

---

## 3. Semantic Package Searching (`pkt add --ai`)

**Command:** `pkt add --ai <description...>`

Instead of googling for package names, you can describe what you want the package to do. `pkt` sends your request (along with your project's language, e.g., "JavaScript") to the AI.

**The `survey/v2` Integration:**
1. The AI returns a structured JSON list of the top 3-5 packages that fit your description.
2. `pkt` parses this JSON and dynamically renders an interactive multi-select checkbox list in your terminal using the `survey/v2` UI library.
3. You use your Arrow Keys to navigate, `Spacebar` to select the packages you want, and `Enter` to confirm.
4. `pkt` immediately routes the selected packages to your native Package Manager for installation!

---

## 4. Agent Skills & Reusable Knowledge (`#skill`)

Just like Claude Code and Antigravity, `pkt chat` supports learning and applying **Skills** — modular, reusable guidelines, workflows, or rulesets that elevate the agent's performance.

### Calling Skills with the Dedicated `#` (or `$`) Sign
In the same way `/` is used for slash commands and `@` is used to attach files, `#` is the dedicated trigger sign for skills:

```bash
pkt [groq:llama-3.3-70b] ❯ refactor this handler according to #go-clean-arch
```

- You can mention `#<skill-name>` anywhere in your prompt.
- `pkt` automatically loads the skill instructions and injects them directly into the agent's context.
- Pressing **Tab** after `#` (or `$`) provides interactive auto-completion of all available skills.

### Skill Discovery & Storage
Skills are discovered automatically across project and global scopes:
- **Project skills**: `.skills/` (preferred hidden folder in project root) or `.pkt/skills/`.
- **Global skills**: `~/.pkt/skills/` (and reads/shares existing skills from `~/.gemini/config/skills/`).

**Storage Rules:**
- **Project skills** are saved to `.skills/<name>/SKILL.md` in the current project root.
- **Global skills** are **always** saved to `~/.pkt/skills/<name>/SKILL.md` (it will never write to `~/.gemini/config/skills/`, but can share and read them).

Skills can be defined either as Markdown files (`<name>.md` or `<name>/SKILL.md`) with standard YAML frontmatter:

```markdown
---
name: go-clean-arch
description: Clean architecture principles for Go services
---

Follow domain-driven directory layout. Handlers must never call DB directly.
```

### Managing & Teaching Skills in Chat
- `/skills` — List all registered skills across project and global scopes.
- `/skill <name>` — View full instructions and metadata of any skill.
- `/learn [--global|--project] <name> <instructions...>` — Teach the agent a new skill and save it to `.skills/` (or `~/.pkt/skills/`).
- **Autonomous Learning**: The agent is equipped with the `learn_skill` tool, allowing it to autonomously persist guidelines and techniques when instructed.
