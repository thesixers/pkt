# How to Test PKT Yourself - Quick Guide

## 🚀 Option 1: Automated Testing (Recommended)

Just run this one command:

```bash
cd /home/genesix/.gemini/antigravity/scratch/pkg-manager
./test-yourself.sh
```

This script will:

1. Check if libcurl is installed (install if needed)
2. Build the project
3. Test all commands
4. Create a sample project
5. Show you everything works

---

## 🔧 Option 2: Manual Step-by-Step

### Step 1: Install libcurl

```bash
sudo dnf install libcurl-devel
```

### Step 2: Build PKT

```bash
cd /home/genesix/.gemini/antigravity/scratch/pkg-manager

# Clean and build
rm -rf build
mkdir build && cd build
cmake ..
make -j$(nproc)
```

**Expected output:**

```
[ 12%] Building CXX object CMakeFiles/pkt.dir/src/Utils.cpp.o
[ 25%] Building CXX object CMakeFiles/pkt.dir/src/GlobalRegistry.cpp.o
...
[100%] Linking CXX executable pkt
[100%] Built target pkt
```

### Step 3: Test Basic Commands

```bash
# Still in build/ directory

# Test version
./pkt version
# Expected: PKT version 1.0.0

# Test help
./pkt help
# Expected: Full help text with all commands
```

### Step 4: Create a Test Project

```bash
# Create test directory
mkdir /tmp/my-test-app
cd /tmp/my-test-app

# Create project
/home/genesix/.gemini/antigravity/scratch/pkg-manager/build/pkt create --language node
# Expected: ✓ Created project 'my-test-app' (node)

# Check files created
ls -la
# Expected: .pkt.info, .pkt.deps, node_modules/

# View project info
cat .pkt.info
# Expected: JSON with project metadata

cat .pkt.deps
# Expected: {} (empty object)
```

### Step 5: Test Dependency Management

```bash
# Add a dependency
/home/genesix/.gemini/antigravity/scratch/pkg-manager/build/pkt add react@18.3.0
# Expected: ℹ Fetching latest version...
#           ℹ Downloading react@18.3.0...
#           ✓ Added react@18.3.0

# List dependencies
/home/genesix/.gemini/antigravity/scratch/pkg-manager/build/pkt deps list
# Expected: 📦 Project Dependencies:
#             • react@18.3.0

# Check .pkt.deps updated
cat .pkt.deps
# Expected: {"react": "18.3.0"}

# Check symlink created
ls -la node_modules/
# Expected: react -> /home/genesix/.pkt_global_store/node/node_modules/react/18.3.0
```

### Step 6: Verify Global Store

```bash
# Check global store created
ls -la ~/.pkt_global_store/
# Expected: node/, python/, ruby/, java/, go/

# Check node packages
ls -la ~/.pkt_global_store/node/node_modules/
# Expected: react/

# Check global .deps
cat ~/.pkt_global_store/node/.deps
# Expected: {"react": "18.3.0"}
```

### Step 7: Test Project Registry

```bash
# List all projects
/home/genesix/.gemini/antigravity/scratch/pkg-manager/build/pkt projects
# Expected: 📦 Registered Projects:
#             • my-test-app (node)
#               Path: /tmp/my-test-app
#               ID: proj-xxxxxxxx

# Check registry file
cat ~/.pkt_registry.json
# Expected: JSON with projects array
```

### Step 8: Test More Commands

```bash
# Set editor
/home/genesix/.gemini/antigravity/scratch/pkg-manager/build/pkt editor set code
# Expected: ✓ Set default editor to: code

# Search projects
/home/genesix/.gemini/antigravity/scratch/pkg-manager/build/pkt search test
# Expected: Shows matching projects

# Add another dependency
/home/genesix/.gemini/antigravity/scratch/pkg-manager/build/pkt add express
# Expected: ✓ Added express@x.x.x

# Update a dependency
/home/genesix/.gemini/antigravity/scratch/pkg-manager/build/pkt update react
# Expected: ✓ Updated react

# Remove a dependency
/home/genesix/.gemini/antigravity/scratch/pkg-manager/build/pkt remove express
# Expected: ✓ Removed express@x.x.x
```

---

## ✅ What Success Looks Like

After testing, you should see:

1. **Binary built**: `build/pkt` exists and runs
2. **Version works**: Shows "PKT version 1.0.0"
3. **Help works**: Shows all commands
4. **Project created**: `.pkt.info` and `.pkt.deps` files exist
5. **Dependencies work**: Packages added, symlinks created
6. **Global store**: `~/.pkt_global_store/` contains packages
7. **Registry works**: `~/.pkt_registry.json` tracks projects

---

## 🐛 Troubleshooting

### Build fails with "Could NOT find CURL"

```bash
# Install libcurl
sudo dnf install libcurl-devel
# Then rebuild
```

### "Not in a PKT project directory"

```bash
# Make sure you ran 'pkt create' first
# Or cd to a directory with .pkt.info
```

### Symlinks not created

```bash
# Check if package was downloaded
ls ~/.pkt_global_store/node/node_modules/

# Check permissions
ls -la node_modules/
```

---

## 📦 Install System-Wide (Optional)

After successful testing:

```bash
cd /home/genesix/.gemini/antigravity/scratch/pkg-manager/build
sudo make install
```

Now you can use `pkt` from anywhere:

```bash
pkt version
pkt help
```

---

## 🎯 Quick Verification Checklist

- [ ] libcurl-devel installed
- [ ] Project builds without errors
- [ ] `./pkt version` shows version
- [ ] `./pkt help` shows help
- [ ] Can create a project
- [ ] Can add dependencies
- [ ] Symlinks created in node_modules/
- [ ] Global store populated
- [ ] Registry tracks projects
- [ ] All commands work

---

## 💡 Tips

1. **Use absolute path** to pkt binary until installed:

   ```bash
   /home/genesix/.gemini/antigravity/scratch/pkg-manager/build/pkt
   ```

2. **Create alias** for easier testing:

   ```bash
   alias pkt='/home/genesix/.gemini/antigravity/scratch/pkg-manager/build/pkt'
   ```

3. **Check logs** if something fails - PKT shows colored error messages

4. **Clean test** by removing global files:
   ```bash
   rm -rf ~/.pkt_global_store ~/.pkt_registry.json
   ```
