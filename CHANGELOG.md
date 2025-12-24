# Changelog

All notable changes to PKT will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-12-24

### Added

- Initial release of PKT Universal Package Manager
- Multi-language support (Node.js, Python, Ruby, Java, Go)
- Global package store with symlink-based dependencies
- Project management (create, init, open, delete, search)
- Dependency management (add, remove, update, list)
- Cross-platform support (Linux, macOS, Windows)
- Editor integration
- Fuzzy project search
- Colored CLI output
- Comprehensive documentation

### Features

- **Global Store**: Efficient package storage at `~/.pkg_global_store/`
- **Symlinks**: Minimal disk usage with symlink-based dependencies
- **Registry**: Track all projects in `~/.pkg_registry.json`
- **Version Resolution**: Automatic latest version detection or specify exact versions
- **Multi-Language**: Single tool for multiple package ecosystems

### Technical

- Written in C++17
- 2,241 lines of production code
- Uses nlohmann/json for JSON parsing
- Uses libcurl for HTTP operations
- CMake build system
- GitHub Actions CI/CD

[1.0.0]: https://github.com/YOUR_USERNAME/pkt/releases/tag/v1.0.0
