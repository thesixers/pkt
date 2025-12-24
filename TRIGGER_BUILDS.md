# How to Trigger Automated Builds

You've created the tag `v1.0.0` successfully! Now you need to create a release to trigger the automated builds.

## Option 1: Create Release via GitHub UI (Easiest)

1. **Go to your repository**: https://github.com/theskers/pkt
2. **Click "Releases"** (on the right sidebar or in the Code tab)
3. **Click "Draft a new release"**
4. **Choose tag**: Select `v1.0.0` from the dropdown
5. **Release title**: `PKT v1.0.0 - Initial Release`
6. **Description**: Add release notes (see template below)
7. **Click "Publish release"**

**This will trigger the GitHub Actions workflow** to build binaries for all platforms!

### Release Notes Template

````markdown
# PKT v1.0.0 - Universal Package Manager

First public release! 🎉

## Features

- Multi-language support (Node.js, Python, Ruby, Java, Go)
- Global package store with symlinks for efficient storage
- Cross-platform (Linux, macOS, Windows)
- Project and dependency management
- 2,241 lines of production C++ code

## Installation

### Linux (x86_64)

```bash
curl -L https://github.com/theskers/pkt/releases/download/v1.0.0/pkt-linux-x86_64 -o pkt
chmod +x pkt
sudo mv pkt /usr/local/bin/pkt
```
````

### macOS

```bash
curl -L https://github.com/theskers/pkt/releases/download/v1.0.0/pkt-macos-x86_64 -o pkt
chmod +x pkt
sudo mv pkt /usr/local/bin/pkt
```

### Windows

Download `pkt-windows-x64.exe` from the assets below.

## Quick Start

```bash
pkt create --language node
pkt add react@18.3.0
pkt deps list
```

## Documentation

- [README](https://github.com/theskers/pkt#readme)
- [Architecture](https://github.com/theskers/pkt/blob/main/ARCHITECTURE.md)
- [Examples](https://github.com/theskers/pkt/blob/main/EXAMPLES.md)

Full changelog: [CHANGELOG.md](https://github.com/theskers/pkt/blob/main/CHANGELOG.md)

````

---

## Option 2: Create Release via Command Line

```bash
# Install GitHub CLI if you don't have it
# Ubuntu/Debian: sudo apt install gh
# macOS: brew install gh
# Then authenticate: gh auth login

# Create release
gh release create v1.0.0 \
  --title "PKT v1.0.0 - Initial Release" \
  --notes "First public release of PKT Universal Package Manager"
````

---

## What Happens Next

After you create the release:

1. **GitHub Actions starts** (check the "Actions" tab)
2. **3 runners spin up** in parallel:
   - Ubuntu runner → Builds Linux binaries
   - macOS runner → Builds macOS binaries
   - Windows runner → Builds Windows binary
3. **~5-10 minutes later**: All binaries appear in your release
4. **Users can download** platform-specific binaries

---

## Check Progress

1. Go to **Actions** tab: https://github.com/theskers/pkt/actions
2. You'll see "Release" workflow running
3. Click on it to see build progress for each platform
4. Green checkmarks = success!

---

## Expected Binaries

After the workflow completes, your release will have:

- `pkt-linux-x86_64` (Linux Intel/AMD)
- `pkt-linux-arm64` (Linux ARM)
- `pkt-macos-x86_64` (macOS Intel)
- `pkt-macos-arm64` (macOS Apple Silicon)
- `pkt-windows-x64.exe` (Windows)

---

## If It Doesn't Work

The workflow might need a small update. Let me know if you see any errors in the Actions tab and I can fix it!
