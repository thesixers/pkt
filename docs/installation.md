# Installation

## Prerequisites

Before using `pkt`, you need the following tools installed based on the languages you work with.

### Required for All Users

| Tool | Why | Install |
| --- | --- | --- |
| **Git** | Required for `pkt clone` | [git-scm.com](https://git-scm.com/) |

### Per-Language Requirements

| Language | Required Tools | Install |
| --- | --- | --- |
| **JavaScript** | Node.js + npm (or pnpm/bun) | [nodejs.org](https://nodejs.org/) |
| **Python** | Python 3.8+ | [python.org](https://www.python.org/) |
| **Go** | Go 1.18+ | [go.dev](https://go.dev/) |
| **Rust** | Rust + Cargo | [rustup.rs](https://rustup.rs/) |

> **Note:** pkt will use whatever package manager is available. For JavaScript, it prefers pnpm > bun > npm. For Python, it prefers uv > poetry > pip.

## Downloading the Binary

The easiest way to install `pkt` is to download the latest pre-compiled binary for your platform from the [Releases](https://github.com/thesixers/pkt/releases) page.

Once downloaded, extract the binary and place it in your system's `PATH`.

### Supported Platforms

`pkt` compiles to a single binary with no external dependencies:

| Platform | Binary |
| --- | --- |
| Linux | `pkt-linux-amd64` or `pkt-linux-arm64` |
| macOS | `pkt-darwin-amd64` or `pkt-darwin-arm64` |
| Windows | `pkt-windows-amd64.exe` or `pkt-windows-arm64.exe` |

## Building from Source

If you have Go installed, you can easily build `pkt` from source:

```bash
git clone https://github.com/thesixers/pkt.git
cd pkt
make build        # creates bin/pkt
sudo make install # installs to /usr/local/bin
```

## Upgrading pkt

Once installed, you can upgrade `pkt` to the latest version automatically at any time by running:

```bash
pkt upgrade
```

This command safely connects to GitHub, downloads the latest binary for your OS and Architecture, and seamlessly replaces your current installation.
