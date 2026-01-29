# pkt

`pkt` is a cross-platform project manager and dependency tracker for
JavaScript/Node.js projects.\
It organizes your projects, tracks dependencies, and abstracts package
manager differences while staying out of your way.

------------------------------------------------------------------------

## Why pkt?

Modern JS projects use different package managers (`pnpm`, `npm`,
`bun`), each with different commands and behaviors.\
`pkt` sits on top of them and gives you:

-   One place for **all projects**
-   One command set for **all package managers**
-   Accurate dependency tracking via `package.json`
-   Safe, predictable project management
-   A foundation for future GUI tools

------------------------------------------------------------------------

## Core Concepts

### Projects Root

All projects live inside **one root folder** chosen during setup.

### Project Identity

Each project has: - A **unique ID** - A name (duplicates allowed) - A
filesystem path - A project-specific package manager

Internally, `pkt` always uses the **project ID**.

------------------------------------------------------------------------

## First-Time Setup

You **must run this first**:

``` bash
pkt start
```

You will be prompted for: - Projects root folder (default:
`~/Documents/workspace`) - Default package manager (pnpm) - Code editor
command (e.g. `code`)

This creates:

    ~/.pkt/config.json

No other command works before this step.

------------------------------------------------------------------------

## Commands

### Setup

#### `pkt start`

Initializes pkt and stores configuration.

------------------------------------------------------------------------

### Project Management

#### `pkt create <project-name>`

Creates a new project folder inside the projects root and tracks it.

``` bash
pkt create music-app
```

-   No scaffolding
-   No `package.json` created yet
-   Project is assigned an ID

------------------------------------------------------------------------

#### `pkt open <project | id | .>`

Opens a project in your configured editor.

``` bash
pkt open music-app
pkt open pkt-01HJ9KJQ8D
pkt open .
```

If multiple projects share the same name, pkt will prompt you to select
one.

------------------------------------------------------------------------

#### `pkt list`

Lists all tracked projects.

Shows: - Name - ID - Package manager - Path

------------------------------------------------------------------------

#### `pkt delete <project | id>`

Deletes a project folder and removes it from the database.

``` bash
pkt delete music-app
```

------------------------------------------------------------------------

### Dependency Management

> Important: Dependency commands **must be run inside a project
> folder**.

------------------------------------------------------------------------

#### `pkt add <package>`

Adds a dependency using the project's package manager.

``` bash
pkt add axios
pkt add -D nodemon
```

Behavior: - Ensures you are inside a managed project - Creates
`package.json` lazily if missing - Uses the correct PM command - Syncs
dependencies into the database

------------------------------------------------------------------------

#### `pkt remove <package>`

Removes a dependency.

``` bash
pkt remove axios
```

------------------------------------------------------------------------

#### `pkt deps [project | id | .]`

Lists dependencies for a project.

``` bash
pkt deps
pkt deps music-app
pkt deps pkt-01HJ9KJQ8D
```

Always reads from `package.json` and syncs DB first.

------------------------------------------------------------------------

### Package Manager Control

#### `pkt pm set <pm> <project | id | .>`

Changes the package manager for a project.

``` bash
pkt pm set npm music-app
pkt pm set bun .
```

What happens: - Project is resolved by name, ID, or current folder - PM
availability is checked - Database is updated - `package.json` scripts
are rewritten using filesystem access - Dependencies are re-synced

------------------------------------------------------------------------

## Supported Package Managers

  PM     Add             Remove            Init
  ------ --------------- ----------------- ----------------
  pnpm   `pnpm add`      `pnpm remove`     `pnpm init -y`
  npm    `npm install`   `npm uninstall`   `npm init -y`
  bun    `bun add`       `bun remove`      `bun init`

Commands are resolved through an internal PM registry.

------------------------------------------------------------------------

## Safety Rules

-   `pkt` **only modifies `package.json`**
-   Never touches source files or configs
-   Database is always synced from `package.json`
-   `add` / `remove` only work inside project folders
-   Duplicate project names are always resolved interactively

------------------------------------------------------------------------

## Architecture Overview

-   Language: **Go**
-   Database: **PostgreSQL**
-   CLI Framework: **Cobra**
-   PM abstraction via command registry
-   File-system operations are isolated and safe

------------------------------------------------------------------------

## GUI Support

`pkt` is CLI-first but **GUI-ready**.

A GUI can: - Call `pkt` commands - Read project/dependency data from
DB - Avoid re-implementing logic

------------------------------------------------------------------------

## Future Ideas

-   PM plugins via JSON
-   Project health checks
-   Background sync watcher
-   GUI dashboard
-   Remote project metadata

------------------------------------------------------------------------

## Summary

`pkt` is opinionated where it matters and flexible where it counts.

If you manage multiple JS projects, switch package managers, or want
order without friction --- `pkt` is built for you.
