# Example Usage Guide

This guide demonstrates how to use PKT after building it.

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
# Create a new Node.js project
pkt create my-node-app --language node
cd my-node-app

# Add dependencies
pkt add react@18.3.0
pkt add express
pkt add lodash@4.17.21

# Run a script
echo 'const express = require("express"); console.log("Express loaded:", !!express);' > server.js
pkt run server.js

# List dependencies
pkt deps list

# Check what's in node_modules (should see symlinks)
ls -la node_modules/

# View global store
cat ~/.pkt_global_store/node/.deps
ls -la ~/.pkt_global_store/node/node_modules/

# Set your editor
pkt editor set code  # or 'vim', 'nano', 'subl', etc.

# Register the project
pkt projects
```

## Example 2: Python Project

```bash
pkt create my-python-app --language python
cd my-python-app

pkt add requests
pkt add flask@2.3.0
pkt add numpy

# Run a script
echo 'import requests; print("Requests version:", requests.__version__)' > main.py
pkt run main.py

pkt deps list
ls -la site-packages/
```

## Example 3: Multi-Project Management

```bash
# Create multiple projects
mkdir -p ~/projects/web-app && cd ~/projects/web-app
pkt create --language node
pkt editor set code

mkdir -p ~/projects/api-server && cd ~/projects/api-server
pkt create --language python
pkt editor set vim

mkdir -p ~/projects/cli-tool && cd ~/projects/cli-tool
pkt create --language go

# List all projects
pkt projects

# Search for a project
pkt search web

# Open a project
pkt open web-app  # Opens in VS Code

# Delete a project
pkt delete cli-tool
```

## Example 4: Dependency Management

```bash
cd my-node-app

# Add latest version
pkt add axios

# Add specific version
pkt add typescript@5.0.0

# Update a dependency
pkt update axios@1.6.0

# Remove a dependency
pkt remove lodash

# List project dependencies
pkt deps list

# List global dependencies for Node.js
pkt deps list --global --lang node

# List all global dependencies (all languages)
pkt deps list --global --all
```

## Example 5: Working with Existing Projects

```bash
# If you have an existing project folder
cd ~/existing-project

# Initialize it as a PKG project
pkt init --language node

# Now you can manage dependencies
pkt add express
pkt add mongoose
```

## Verifying Symlinks

```bash
# In your project directory
ls -la node_modules/react
# Should show: react -> /home/user/.pkt_global_store/node/node_modules/react/18.3.0

# Check the actual package
ls ~/.pkt_global_store/node/node_modules/react/18.3.0/
# Should contain package.json and other files
```

## Global Store Structure

After adding some packages, your global store will look like:

```
~/.pkt_global_store/
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
   pkt add react@18.3.0
   ```

2. **Check global dependencies** before adding:

   ```bash
   pkt deps list --global --lang node
   ```

3. **Set editor per project** for convenience:

   ```bash
   pkt editor set code
   pkt open my-app  # Opens in VS Code
   ```

4. **Search projects** with fuzzy matching:

   ```bash
   pkt search api  # Finds "api-server", "my-api", etc.
   ```

5. **Clean up** unused projects:
   ```bash
   pkt delete old-project
   ```
