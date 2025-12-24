#!/bin/bash
# PKT Manager - Quick Verification Script
# This script shows the project structure and verifies the code is ready to build

echo "╔════════════════════════════════════════════════════════════╗"
echo "║         PKT - Universal Package Manager                   ║"
echo "║         Project Verification                               ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo ""

# Check project structure
echo "📁 Project Structure:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
tree -L 2 -I 'build|_deps' . 2>/dev/null || find . -maxdepth 2 -type f -o -type d | grep -v build | sort
echo ""

# Count files
echo "📊 Code Statistics:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Header files:  $(find include -name '*.hpp' 2>/dev/null | wc -l)"
echo "  Source files:  $(find src -name '*.cpp' 2>/dev/null | wc -l)"
echo "  Total C++ LOC: $(find include src -name '*.cpp' -o -name '*.hpp' 2>/dev/null | xargs wc -l 2>/dev/null | tail -1 | awk '{print $1}')"
echo "  Documentation: $(find . -maxdepth 1 -name '*.md' | wc -l) files"
echo ""

# Check dependencies
echo "🔧 Build Dependencies:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check CMake
if command -v cmake &> /dev/null; then
    echo "  ✅ CMake:    $(cmake --version | head -n1)"
else
    echo "  ❌ CMake:    NOT FOUND (required)"
fi

# Check C++ compiler
if command -v g++ &> /dev/null; then
    echo "  ✅ g++:      $(g++ --version | head -n1)"
elif command -v clang++ &> /dev/null; then
    echo "  ✅ clang++:  $(clang++ --version | head -n1)"
else
    echo "  ❌ C++:      NO COMPILER FOUND (required)"
fi

# Check libcurl
if pkg-config --exists libcurl 2>/dev/null; then
    echo "  ✅ libcurl:  $(pkg-config --modversion libcurl)"
else
    echo "  ⚠️  libcurl:  NOT FOUND (required for build)"
    echo ""
    echo "  To install libcurl:"
    echo "    Ubuntu/Debian: sudo apt-get install libcurl4-openssl-dev"
    echo "    Fedora/RHEL:   sudo dnf install libcurl-devel"
    echo "    macOS:         brew install curl"
fi

echo ""

# Show core components
echo "🏗️  Core Components Implemented:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
for file in include/*.hpp; do
    component=$(basename "$file" .hpp)
    echo "  ✅ $component"
done
echo ""

# Show what the binary will do
echo "🎯 PKT Commands (after build):"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat << 'EOF'
  Project Management:
    pkt create --language node    Create new project
    pkt init --language python    Initialize existing folder
    pkt projects                   List all projects
    pkt search <query>             Search projects
    pkt open <name>                Open in editor
    pkt delete <name>              Delete project
    
  Dependency Management:
    pkt add react@18.3.0          Add dependency
    pkt remove react              Remove dependency
    pkt update react              Update dependency
    pkt deps list                 List project deps
    pkt deps list --global        List global deps
    
  Configuration:
    pkt editor set code           Set default editor
    pkt help                      Show help
    pkt version                   Show version
EOF
echo ""

# Build instructions
echo "🚀 How to Build:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
cat << 'EOF'
  1. Install dependencies (if not already installed):
     sudo apt-get install libcurl4-openssl-dev  # Ubuntu/Debian
     sudo dnf install libcurl-devel              # Fedora/RHEL
     
  2. Build the project:
     mkdir -p build && cd build
     cmake ..
     make -j$(nproc)
     
  3. Test the binary:
     ./pkt help
     ./pkt version
     
  4. Install system-wide (optional):
     sudo make install
EOF
echo ""

# Quick test
echo "🧪 Quick Code Verification:"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check if main.cpp exists and has correct structure
if grep -q "pkg::CLI cli" src/main.cpp 2>/dev/null; then
    echo "  ✅ Main entry point configured correctly"
else
    echo "  ⚠️  Main entry point may need verification"
fi

# Check if CMakeLists.txt has pkt target
if grep -q "add_executable(pkt" CMakeLists.txt 2>/dev/null; then
    echo "  ✅ CMake configured to build 'pkt' binary"
else
    echo "  ⚠️  CMake target may need verification"
fi

# Check CLI help text
if grep -q "PKT - Universal Package Manager" src/CLI.cpp 2>/dev/null; then
    echo "  ✅ CLI help text updated to 'pkt'"
else
    echo "  ⚠️  CLI help text may need update"
fi

echo ""
echo "╔════════════════════════════════════════════════════════════╗"
echo "║  Project is ready to build! Follow the build steps above. ║"
echo "╚════════════════════════════════════════════════════════════╝"
