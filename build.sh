#!/bin/bash
# PKG Manager - Build and Test Script

set -e

echo "🔧 PKG Manager - Build Script"
echo "=============================="
echo ""

# Check for dependencies
echo "📦 Checking dependencies..."

if ! command -v cmake &> /dev/null; then
    echo "❌ CMake not found. Please install CMake 3.15+"
    exit 1
fi

if ! command -v g++ &> /dev/null && ! command -v clang++ &> /dev/null; then
    echo "❌ C++ compiler not found. Please install g++ or clang++"
    exit 1
fi

echo "✅ CMake found: $(cmake --version | head -n1)"
echo "✅ Compiler found"

# Install libcurl if needed
if ! pkg-config --exists libcurl 2>/dev/null; then
    echo "⚠️  libcurl not found. Attempting to install..."
    
    if command -v apt-get &> /dev/null; then
        sudo apt-get update && sudo apt-get install -y libcurl4-openssl-dev
    elif command -v dnf &> /dev/null; then
        sudo dnf install -y libcurl-devel
    elif command -v brew &> /dev/null; then
        brew install curl
    else
        echo "❌ Could not install libcurl automatically."
        echo "   Please install libcurl development libraries manually."
        exit 1
    fi
fi

echo "✅ libcurl found"
echo ""

# Build
echo "🏗️  Building PKG..."
mkdir -p build
cd build

echo "  → Running CMake..."
cmake .. -DCMAKE_BUILD_TYPE=Release

echo "  → Compiling..."
make -j$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 2)

echo ""
echo "✅ Build complete!"
echo ""

# Test
echo "🧪 Running tests..."
ctest --output-on-failure

echo ""
echo "✅ All tests passed!"
echo ""

# Show binary
echo "📍 Binary location: $(pwd)/pkg"
echo ""
echo "To install system-wide, run:"
echo "  sudo make install"
echo ""
echo "To test the CLI, try:"
echo "  ./pkg help"
echo "  ./pkg version"
echo ""
