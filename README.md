# feat

Feature-centric context management for agentic coding.

## What's New: The feat.yaml Schema

The manifest file defines your project's feature hierarchy:

```yaml
config:
  max_files: 3                 # Max files per feature + ancestors
  workflow: [scaffold, fix, build, test, done]
tree:
  name: my-project
  files: [go.mod, README.md]   # Root-level files
  children:
    auth:                      # Boundary (has children)
      files: [auth/interface.go]
      children:
        login:                 # Feature (no children, has files/tests)
          files: [auth/login.go]
          tests: [auth/login_test.go]
```

**Key fields:**
- `config.max_files` — Maximum files to load (feature + ancestor files)
- `config.workflow` — Custom workflow steps for features
- `tree.name` — Project name
- `tree.files` — Shared files at root level
- `tree.children` — Feature hierarchy (boundaries nest, features are leaves)
- Node fields: `files` (implementation), `tests` (test files), `children` (sub-features)

## Overview

`feat` organizes code by feature, not by layer. Instead of loading entire packages, you work on specific features with their relevant files and ancestor context.



## Installation

### Requirements

- Go 1.21 or later

### From Source

Clone the repository and build:

```bash
git clone https://github.com/plor/feat.git
cd feat
go build -o feat ./cmd/feat
```

Or install directly with `go install`:

```bash
go install github.com/plor/feat/cmd/feat@latest
```

### Verify Installation

```bash
feat --version
```

## Quick Start Tutorial

Create a new project and initialize it with feat:

```bash
mkdir myproject && cd myproject
```

Initialize feat in your project:

```bash
feat init
```

This creates:
- `.feat.yml` — The manifest file defining your feature hierarchy
- `.feat/` — Directory containing state and metadata

### Generated .feat.yml

After running `feat init`, you'll have a basic manifest:

```yaml
config:
  max_files: 3
  workflow: [scaffold, fix, build, test, done]
tree:
  name: myproject
  files: []
  children: {}
```

### Initial .feat/state.json

The state file tracks your current context:

```json
{
  "current_feature": "",
  "workflow_state": ""
}
```

### View Available Features

List all features in your project:

```bash
feat list
```

### Start Working on a Feature

Begin working on a specific feature:

```bash
feat work <feature-id>
```

This loads the feature's context, including its files and ancestor context.

## Configuration Reference

The `feat.yaml` file (or `.feat.yml`) is the single source of truth for your project's feature hierarchy. This section provides a complete reference for all configuration options.

### File Structure

```yaml
config:
  max_files: 3
  workflow: [scaffold, fix, build, test, done]
tree:
  name: my-project
  files: [go.mod, README.md]
  children:
    # Features and boundaries defined here
```

### Config Section

The `config` section controls global behavior:

#### `max_files` (integer)
Maximum number of files to load when working on a feature. This includes the feature's own files plus ancestor files from parent boundaries.

- **Default:** 3
- **Purpose:** Prevents context overflow when agentic tools load feature context
- **Behavior:** Files are collected from the current feature up through each parent boundary

Example:
```yaml
config:
  max_files: 5  # Allow more files for larger features
```

#### `workflow` (array of strings)
Defines the steps a feature progresses through. Used by `feat transition` to track work state.

- **Default:** `[scaffold, fix, build, test, done]`
- **Purpose:** Standardizes feature lifecycle across your project
- **Customization:** Use any steps that match your process

Example:
```yaml
config:
  workflow: [draft, review, implement, verify, complete]
```

### Tree Section

The `tree` section defines your project's feature hierarchy:

#### `name` (string)
Project identifier. Used in output and logging.

```yaml
tree:
  name: my-api-service
```

#### `files` (array of strings)
Root-level files always included when loading any feature. Use for project-wide configuration and documentation.

```yaml
tree:
  files: [go.mod, go.sum, README.md, Makefile]
```

#### `children` (map)
Feature hierarchy. Keys are feature IDs, values are node definitions.

```yaml
tree:
  children:
    auth:
      files: [auth/interface.go]
      children:
        login:
          files: [auth/login.go]
          tests: [auth/login_test.go]
```

### Node Types

`feat` distinguishes between two node types:

#### Boundary Nodes
Group related features. Have children, no direct files/tests.

Characteristics:
- Has `children` map
- No `files` or `tests` arrays
- Serves as namespace/interface boundary
- Files listed in a boundary define the interface for its children

Example:
```yaml
auth:                      # Boundary
  files: [auth/interface.go]
  children:
    login: {}              # Features inside
```

#### Feature Nodes
Leaf nodes representing actual work. Have files/tests, no children.

Characteristics:
- Has `files` and/or `tests` arrays
- No `children` map
- Represents implementable unit of work
- Full path forms the feature ID (e.g., `auth/login`)

Example:
```yaml
login:                     # Feature
  files: [auth/login.go]
  tests: [auth/login_test.go]
```

### Node Fields

All nodes (boundaries and features) support these fields:

#### `files` (array of strings)
Implementation files for this node. For boundaries, these define the interface contract. For features, these are the files to edit.

```yaml
files: [auth/login.go, auth/session.go]
```

#### `tests` (array of strings)
Test files associated with this feature. Only valid for feature nodes (leaves).

```yaml
tests: [auth/login_test.go, auth/session_test.go]
```

#### `children` (map)
Nested features. Only valid for boundary nodes. Keys are feature IDs, values are node definitions.

```yaml
children:
  login:
    files: [auth/login.go]
  logout:
    files: [auth/logout.go]
```

### Common Patterns

#### Flat Structure (No Boundaries)
Simple projects with just features, no nesting:

```yaml
tree:
  name: simple-project
  files: [go.mod]
  children:
    config:
      files: [config.go]
    server:
      files: [server.go]
    handlers:
      files: [handlers.go]
```

#### Layered Architecture
Group by architectural layer:

```yaml
tree:
  children:
    models:                    # Boundary
      files: [models/interface.go]
      children:
        user:
          files: [models/user.go]
        account:
          files: [models/account.go]
    handlers:                  # Boundary
      files: [handlers/interface.go]
      children:
        auth:
          files: [handlers/auth.go]
        api:
          files: [handlers/api.go]
```

#### Domain-Driven Design
Group by business domain:

```yaml
tree:
  children:
    billing:                   # Domain boundary
      files: [billing/interface.go]
      children:
        invoices:
          files: [billing/invoices.go]
        payments:
          files: [billing/payments.go]
    users:                     # Domain boundary
      files: [users/interface.go]
      children:
        profiles:
          files: [users/profiles.go]
        settings:
          files: [users/settings.go]
```

#### Deep Nesting
For complex projects with multiple levels:

```yaml
tree:
  children:
    api:
      files: [api/interface.go]
      children:
        v1:
          files: [api/v1/interface.go]
          children:
            users:
              files: [api/v1/users.go]
            posts:
              files: [api/v1/posts.go]
```

**When to use deep nesting:**
- Large projects with clear subdomain divisions
- API versioning
- Multi-tenant architectures

**When to avoid:**
- Creates long feature IDs (`api/v1/admin/users/list`)
- Makes `feat work` typing harder
- Consider flatter structure if >3 levels deep

### Validation Rules

`feat validate` checks your manifest for these issues:

#### Required Fields
- `tree.name` — Must be non-empty string
- `config.max_files` — Must be positive integer
- `config.workflow` — Must be non-empty array

#### Node Validation
- A node cannot have both `children` and `files`/`tests`
- Feature IDs must be unique within siblings
- File paths must be non-empty strings

#### Common Errors

**Error:** `node cannot be both boundary and feature`
```yaml
# Invalid: has children AND files
auth:
  files: [auth.go]           # ❌ Features can't have children
  children:
    login: {}
```
**Fix:** Move files to a boundary or remove children:
```yaml
auth:                        # ✅ Boundary with interface
  files: [auth/interface.go]
  children:
    login:
      files: [auth/login.go] # ✅ Feature with files only
```

**Error:** `duplicate feature ID`
```yaml
children:
  auth:
    files: [auth.go]
  auth:                      # ❌ Duplicate key
    files: [other.go]
```
**Fix:** Use unique IDs:
```yaml
children:
  auth:
    files: [auth.go]
  auth_v2:                   # ✅ Unique ID
    files: [auth_v2.go]
```

**Error:** `max_files must be at least 1`
```yaml
config:
  max_files: 0               # ❌ Must be >= 1
```
**Fix:** Set to valid positive integer:
```yaml
config:
  max_files: 3               # ✅ Valid
```

## Commands

- `feat init` — Create a new feat.yaml manifest
- `feat list` — Show feature tree
- `feat work <feature>` — Load a feature's context
- `feat split <parent> <name>` — Create a new feature
- `feat add <feature> <file>` — Add a file to an existing feature
- `feat status` — Show current feature context
- `feat transition <step>` — Update feature workflow state
- `feat validate` — Check manifest for issues

## Example

```bash
# Initialize a project
feat init --name my-app

# Create a feature
feat split "" auth
feat split auth login

# Work on a feature

# Add files to a feature
feat add auth/login auth/login_test.go
feat work auth/login

# Check status
feat status

# Move to next workflow step
feat transition build
```

## License

MIT
