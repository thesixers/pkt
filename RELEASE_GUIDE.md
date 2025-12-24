# Quick Start Guide for GitHub Release

This guide will help you publish PKT to GitHub in ~30 minutes.

## Step 1: Create GitHub Repository (5 min)

1. Go to https://github.com/new
2. Repository name: `pkt`
3. Description: "Universal package manager for Node.js, Python, Ruby, Java, and Go"
4. Public repository
5. **Don't** initialize with README (we have one)
6. Click "Create repository"

## Step 2: Initialize Git & Push (5 min)

```bash
cd /home/genesix/.gemini/antigravity/scratch/pkg-manager

# Initialize git
git init
git add .
git commit -m "Initial commit: PKT v1.0.0

- Multi-language package manager (Node, Python, Ruby, Java, Go)
- Global store with symlink-based dependencies
- Cross-platform support (Linux, macOS, Windows)
- 2,241 lines of C++ code"

# Add your GitHub repository (replace YOUR_USERNAME)
git remote add origin https://github.com/YOUR_USERNAME/pkt.git

# Push to GitHub
git branch -M main
git push -u origin main
```

## Step 3: Build Release Binaries (10 min)

```bash
# Build for your current platform
cd /home/genesix/.gemini/antigravity/scratch/pkg-manager
mkdir -p release
cd build

# Copy binary to release folder
cp pkt ../release/pkt-linux-x86_64

# Create checksum
cd ../release
sha256sum pkt-linux-x86_64 > checksums.txt
```

## Step 4: Create GitHub Release (10 min)

1. Go to your repository on GitHub
2. Click "Releases" → "Create a new release"
3. Tag version: `v1.0.0`
4. Release title: `PKT v1.0.0 - Initial Release`
5. Description:

````markdown
# PKT v1.0.0 - Universal Package Manager

First public release of PKT! 🎉

## Features

- Multi-language support (Node.js, Python, Ruby, Java, Go)
- Global package store with symlinks
- Cross-platform (Linux, macOS, Windows)
- Project and dependency management
- 2,241 lines of production C++ code

## Installation

### Linux

```bash
curl -L https://github.com/YOUR_USERNAME/pkt/releases/download/v1.0.0/pkt-linux-x86_64 -o pkt
chmod +x pkt
sudo mv pkt /usr/local/bin/pkt
```
````

### Quick Start

```bash
pkt create --language node
pkt add react@18.3.0
pkt deps list
```

## Documentation

- [README](https://github.com/YOUR_USERNAME/pkt#readme)
- [Architecture](https://github.com/YOUR_USERNAME/pkt/blob/main/ARCHITECTURE.md)
- [Examples](https://github.com/YOUR_USERNAME/pkt/blob/main/EXAMPLES.md)

Full changelog: [CHANGELOG.md](https://github.com/YOUR_USERNAME/pkt/blob/main/CHANGELOG.md)

````

6. Upload files:
   - `pkt-linux-x86_64`
   - `checksums.txt`

7. Click "Publish release"

## Step 5: Update install.sh (5 min)

Edit `install.sh` and replace `YOUR_USERNAME` with your actual GitHub username:

```bash
# Line 45
GITHUB_REPO="YOUR_ACTUAL_USERNAME/pkt"
````

Then commit and push:

```bash
git add install.sh
git commit -m "Update install script with actual GitHub username"
git push
```

## Step 6: Test Installation (5 min)

Test the one-line installer:

```bash
curl -sSL https://raw.githubusercontent.com/YOUR_USERNAME/pkt/main/install.sh | bash
```

## Done! 🎉

Your PKT is now publicly available!

## Next Steps

1. **Announce it**:

   - Post on Reddit (r/programming, r/commandline)
   - Share on Twitter/X
   - Submit to Hacker News

2. **Add more platforms**:

   - Build macOS binary
   - Build Windows binary
   - Set up GitHub Actions for automated releases

3. **Create packages**:
   - Homebrew formula
   - AUR package
   - Snap package

## Troubleshooting

**Git push fails**: Make sure you've set up SSH keys or use HTTPS with personal access token

**Release upload fails**: Check file size limits (100MB for free accounts)

**Install script fails**: Verify the download URL is correct

## Need Help?

Open an issue on GitHub or check the full deployment plan in `implementation_plan.md`.
