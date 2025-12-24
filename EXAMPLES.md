# Example Usage Guide

This guide demonstrates how to use PKG after building it.

## Setup

```bash
# Build the project
./build.sh

# Install system-wide (optional)
cd build
sudo make install
```

## Example 1: Node.js Project

```bash
# Create a new directory for your project
mkdir my-node-app
cd my-node-app

# Initialize as a PKG project
pkg create --language node

# Add dependencies
pkg add react@18.3.0
pkg add express
pkg add lodash@4.17.21

# List dependencies
pkg deps list

# Check what's in node_modules (should see symlinks)
ls -la node_modules/

# View global store
cat ~/.pkg_global_store/node/.deps
ls -la ~/.pkg_global_store/node/node_modules/

# Set your editor
pkg editor set code  # or 'vim', 'nano', 'subl', etc.

# Register the project
pkg projects
```

## Example 2: Python Project

```bash
mkdir my-python-app
cd my-python-app

pkg create --language python

pkg add requests
pkg add flask@2.3.0
pkg add numpy

pkg deps list
ls -la site-packages/
```

## Example 3: Multi-Project Management

```bash
# Create multiple projects
mkdir -p ~/projects/web-app && cd ~/projects/web-app
pkg create --language node
pkg editor set code

mkdir -p ~/projects/api-server && cd ~/projects/api-server
pkg create --language python
pkg editor set vim

mkdir -p ~/projects/cli-tool && cd ~/projects/cli-tool
pkg create --language go

# List all projects
pkg projects

# Search for a project
pkg search web

# Open a project
pkg open web-app  # Opens in VS Code

# Delete a project
pkg delete cli-tool
```

## Example 4: Dependency Management

```bash
cd my-node-app

# Add latest version
pkg add axios

# Add specific version
pkg add typescript@5.0.0

# Update a dependency
pkg update axios@1.6.0

# Remove a dependency
pkg remove lodash

# List project dependencies
pkg deps list

# List global dependencies for Node.js
pkg deps list --global --lang node

# List all global dependencies (all languages)
pkg deps list --global --all
```

## Example 5: Working with Existing Projects

```bash
# If you have an existing project folder
cd ~/existing-project

# Initialize it as a PKG project
pkg init --language node

# Now you can manage dependencies
pkg add express
pkg add mongoose
```

## Verifying Symlinks

```bash
# In your project directory
ls -la node_modules/react
# Should show: react -> /home/user/.pkg_global_store/node/node_modules/react/18.3.0

# Check the actual package
ls ~/.pkg_global_store/node/node_modules/react/18.3.0/
# Should contain package.json and other files
```

## Global Store Structure

After adding some packages, your global store will look like:

```
~/.pkg_global_store/
├── node/
│   ├── node_modules/
│   │   ├── react/
│   │   │   └── 18.3.0/
│   │   │       └── package.json
│   │   ├── express/
│   │   │   └── 4.18.2/
│   │   └── axios/
│   │       └── 1.6.0/
│   └── .deps
├── python/
│   ├── site-packages/
│   │   ├── requests/
│   │   │   └── 2.31.0/
│   │   └── flask/
│   │       └── 2.3.0/
│   └── .deps
```

## Tips

1. **Use specific versions** for production projects:

   ```bash
   pkg add react@18.3.0
   ```

2. **Check global dependencies** before adding:

   ```bash
   pkg deps list --global --lang node
   ```

3. **Set editor per project** for convenience:

   ```bash
   pkg editor set code
   pkg open my-app  # Opens in VS Code
   ```

4. **Search projects** with fuzzy matching:

   ```bash
   pkg search api  # Finds "api-server", "my-api", etc.
   ```

5. **Clean up** unused projects:
   ```bash
   pkg delete old-project
   ```
