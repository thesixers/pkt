# Contributing to PKT

Thank you for your interest in contributing to PKT! We welcome contributions from the community.

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue on GitHub with:

- A clear description of the problem
- Steps to reproduce
- Expected vs actual behavior
- Your environment (OS, PKT version)

### Suggesting Features

We love new ideas! Open an issue with:

- Clear description of the feature
- Use cases and benefits
- Possible implementation approach (optional)

### Pull Requests

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature/amazing-feature`
3. **Make your changes**
4. **Test thoroughly**: Ensure all tests pass and add new tests if needed
5. **Commit**: Use clear, descriptive commit messages
6. **Push**: `git push origin feature/amazing-feature`
7. **Open a Pull Request**

### Development Setup

```bash
# Clone your fork
git clone https://github.com/thesixers/pkt.git
cd pkt

# Build
mkdir build && cd build
cmake ..
make

# Test
./pkt version
./pkt help
```

### Code Style

- Follow existing code style
- Use meaningful variable names
- Comment complex logic
- Keep functions focused and small

### Testing

- Test your changes on multiple platforms if possible
- Add unit tests for new features
- Ensure existing tests still pass

### Commit Messages

Use clear, descriptive commit messages:

- `feat: add Python 3.12 support`
- `fix: resolve symlink issue on Windows`
- `docs: update installation instructions`
- `refactor: simplify dependency resolution`

## Code of Conduct

Be respectful, inclusive, and constructive. We're all here to make PKT better!

## Questions?

Open an issue or start a discussion on GitHub. We're happy to help!

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
