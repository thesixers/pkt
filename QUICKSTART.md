# PKT - Quick Start Testing Guide

![Testing Guide](pkt_testing_guide_1766535060498.png)

## 🎯 Two Ways to Test

### ⚡ Option 1: Automated (Easiest)

```bash
cd /home/genesix/.gemini/antigravity/scratch/pkg-manager
./test-yourself.sh
```

This does everything automatically!

---

### 🔧 Option 2: Manual (Step-by-Step)

#### 1️⃣ Install libcurl

```bash
sudo dnf install libcurl-devel
```

#### 2️⃣ Build PKT

```bash
cd /home/genesix/.gemini/antigravity/scratch/pkg-manager
rm -rf build && mkdir build && cd build
cmake ..
make -j$(nproc)
```

#### 3️⃣ Test It Works

```bash
./pkt version    # Should show: PKT version 1.0.0
./pkt help       # Should show: Full help text
```

#### 4️⃣ Create a Test Project

```bash
cd /tmp
/home/genesix/.gemini/antigravity/scratch/pkg-manager/build/pkt create test-app --language node
cd test-app
/home/genesix/.gemini/antigravity/scratch/pkg-manager/build/pkt add react@18.3.0
echo 'const react = require("react"); console.log("React loaded:", !!react);' > index.js
/home/genesix/.gemini/antigravity/scratch/pkg-manager/build/pkt run index.js
/home/genesix/.gemini/antigravity/scratch/pkg-manager/build/pkt deps list
```

---

## ✅ Success Indicators

You'll know it works when you see:

- ✅ `./pkt version` → "PKT version 1.0.0"
- ✅ `./pkt help` → Full command list
- ✅ Project creates `.pkt.info` and `.pkt.deps`
- ✅ Dependencies show in `pkt deps list`
- ✅ Symlinks created in `node_modules/`
- ✅ Global store at `~/.pkt_global_store/`

---

## 📚 More Details

- **Full testing guide**: [TESTING.md](file:///home/genesix/.gemini/antigravity/scratch/pkg-manager/TESTING.md)
- **Complete walkthrough**: See artifacts
- **Architecture docs**: [ARCHITECTURE.md](file:///home/genesix/.gemini/antigravity/scratch/pkg-manager/ARCHITECTURE.md)

---

## 🚀 After Testing

Install system-wide:

```bash
cd /home/genesix/.gemini/antigravity/scratch/pkg-manager/build
sudo make install
```

Then use `pkt` from anywhere! 🎉
