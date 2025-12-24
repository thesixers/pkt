# PKG - Universal C++ Package Manager

A cross-platform CLI package manager written in C++ that provides unified dependency management across multiple programming languages with a global store architecture and symlink-based project dependencies.

## 🚀 Features

- **Multi-Language Support**: Manage dependencies for Node.js, Python, Ruby, Java, and Go from a single tool
- **Global Package Store**: Efficient storage with language-native folder structures
- **Symlink-Based Dependencies**: Minimal disk usage and fast updates
- **Project Management**: Create, initialize, search, and manage projects
- **Version Resolution**: Automatic latest version detection or specify exact versions
- **Cross-Platform**: Works on Linux, macOS, and Windows

## 📦 Installation

### Prerequisites

- C++17 or later compiler (g++, clang, or MSVC)
- CMake 3.15+
- libcurl development libraries

#### Ubuntu/Debian

```bash
sudo apt-get install build-essential cmake libcurl4-openssl-dev
```

#### macOS

```bash
brew install cmake curl
```

#### Windows

Install Visual Studio with C++ support and CMake.

### Build from Source

```bash
git clone https://github.com/yourusername/pkg-manager.git
cd pkg-manager
mkdir build && cd build
cmake ..
make
sudo make install
```

## 🎯 Quick Start

### Create a New Project

```bash
# Create a Node.js project
mkdir my-app && cd my-app
pkg create --language node

# Or initialize an existing directory
pkg init --language python
```

### Manage Dependencies

```bash
# Add a dependency (latest version)
pkg add react

# Add a specific version
pkg add fastify@4.27.0

# Remove a dependency
pkg remove react

# Update a dependency
pkg update fastify@5.0.0

# List project dependencies
pkg deps list

# List global dependencies
pkg deps list --global --lang node
pkg deps list --global --all
```

### Project Management

```bash
# List all projects
pkg projects

# Search for projects
pkg search my-app

# Set default editor
pkg editor set code

# Open project in editor
pkg open my-app

# Delete a project
pkg delete my-app
```

## 📁 Architecture

### Global Store Structure

```
~/.pkg_global_store/
├─ node/
│  ├─ node_modules/
│  │  ├─ react/18.3.0/
│  │  ├─ fastify/4.27.0/
│  ├─ .deps              # Global package tracking
├─ python/
│  ├─ site-packages/
│  │  ├─ requests/2.31.0/
│  ├─ .deps
├─ ruby/
│  ├─ gems/
│  ├─ .deps
```

### Project Structure

```
my-app/
├─ .pkg.info            # Project metadata
├─ .pkg.deps            # Project dependencies
├─ node_modules/        # Symlinks to global store
│  ├─ react -> ~/.pkg_global_store/node/node_modules/react/18.3.0
│  ├─ fastify -> ~/.pkg_global_store/node/node_modules/fastify/4.27.0
```

## 🌐 Supported Languages

| Language | Dependency Folder | Registry      |
| -------- | ----------------- | ------------- |
| Node.js  | `node_modules`    | npm           |
| Python   | `site-packages`   | PyPI          |
| Ruby     | `gems`            | RubyGems      |
| Java     | `maven`           | Maven Central |
| Go       | `pkg`             | Go Proxy      |

## 📖 Command Reference

### Project Commands

- `pkg create --language <lang>` - Create new project
- `pkg init --language <lang>` - Initialize existing directory
- `pkg open <name_or_id>` - Open project in editor
- `pkg editor set <command>` - Set default editor
- `pkg editor unset` - Unset default editor
- `pkg search <query>` - Search projects
- `pkg projects` - List all projects
- `pkg delete <name_or_id>` - Delete project

### Dependency Commands

- `pkg add <package>[@<version>]` - Add dependency
- `pkg remove <package>` - Remove dependency
- `pkg update <package>[@<version>]` - Update dependency
- `pkg deps list` - List project dependencies
- `pkg deps list --global [--lang <lang>] [--all]` - List global dependencies

## 🛠️ Development

### Project Structure

```
pkg-manager/
├─ include/          # Header files
├─ src/              # Implementation files
├─ tests/            # Test files
├─ CMakeLists.txt    # Build configuration
└─ README.md
```

### Building

```bash
mkdir build && cd build
cmake ..
make
```

### Running Tests

```bash
cd build
ctest --output-on-failure
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

MIT License - see LICENSE file for details

## 🙏 Acknowledgments

- Built with [nlohmann/json](https://github.com/nlohmann/json) for JSON parsing
- Uses libcurl for HTTP operations
- Inspired by modern package managers like npm, pip, and cargo

## 📞 Support

For issues and questions, please open an issue on GitHub.
