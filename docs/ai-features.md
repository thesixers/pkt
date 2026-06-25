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

## 2. Contextual Querying (`pkt ask`)

**Command:** `pkt ask <query...>`

A quick-access command for one-off questions. Because `pkt` knows your current directory, it passes your project's context directly to the LLM.

**Example:**
`pkt ask "how do I configure a database connection pool in this Rust project?"`

---

## 3. Code Scaffolding (`pkt generate`)

**Command:** `pkt generate <description...>`

Instructs the AI to generate boilerplate code. The engine forces the LLM to output structured markdown blocks, which `pkt` can then parse and potentially pipe directly into files.

**Example:**
`pkt generate "A python FastAPI router for user registration"`

---

## 4. Native Debugging (`pkt debug`)

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

## 5. Semantic Package Searching (`pkt add --ai`)

**Command:** `pkt add --ai <description...>`

Instead of googling for package names, you can describe what you want the package to do. `pkt` sends your request (along with your project's language, e.g., "JavaScript") to the AI.

**The `survey/v2` Integration:**
1. The AI returns a structured JSON list of the top 3-5 packages that fit your description.
2. `pkt` parses this JSON and dynamically renders an interactive multi-select checkbox list in your terminal using the `survey/v2` UI library.
3. You use your Arrow Keys to navigate, `Spacebar` to select the packages you want, and `Enter` to confirm.
4. `pkt` immediately routes the selected packages to your native Package Manager for installation!
