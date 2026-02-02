# GhostEnv

GhostEnv is a secure environment variable manager that encrypts sensitive configuration data and injects it directly into application processes at runtime. It eliminates the security risks of plain-text `.env` files by storing secrets in encrypted binary vaults.

## Description

GhostEnv provides a secure alternative to traditional environment variable management. Instead of storing secrets in plain-text files that can be accidentally committed to version control or accessed by unauthorized processes, GhostEnv encrypts all sensitive data using strong cryptographic algorithms and stores it in binary vault files.

The tool operates at the operating system level, injecting decrypted environment variables directly into the memory space of child processes before they start execution. This ensures that secrets never appear as plain text on disk during runtime and are only available to the specific process that needs them.

## Features

- **Encrypted Storage**: AES-256-GCM encryption with Argon2id key derivation
- **Project-Based Vaults**: Separate vaults for each project with environment support
- **Multiple Environments**: Support for dev, staging, production, and custom environments
- **Global Vault**: Fallback to global vault when not in a project directory
- **Runtime Injection**: Secrets injected directly into process environment
- **Version Control Safe**: Binary vault format prevents accidental exposure
- **Language Agnostic**: Works with any runtime that reads environment variables
- **Cross-Platform**: Supports Linux, macOS, and Windows
- **Input Validation**: Validates keys and prevents invalid characters
- **Secure Password Handling**: Passwords and secrets kept in `[]byte` and zeroed after use to avoid lingering in RAM
- **Vault Integrity (HMAC)**: Each vault file includes an HMAC; tampering or corruption is detected and the vault is refused
- **Atomic Writes**: Saves go to a temporary file then rename, so a crash during write does not corrupt the vault
- **Hide Password from Process List**: Prefer `GHOSTENV_PASS` environment variable over `-p` so the password does not appear in `ps aux`
- **Version Command**: Print version and build information
- **Change Password**: Re-encrypt vault with a new master password
- **Stats**: Vault statistics (path, type, environment, key count, last modified)
- **Export Formats**: Export as JSON or `.env` with `--format`; write to a file with `--output`
- **Shamir's Secret Sharing**: Split the master password into N shares; recover it with K shares (`create-shares`, `recover`)

## Installation

### Prerequisites

- Go 1.25.5 or higher
- Make (optional, for using Makefiles)

### Build from Source

#### Linux and macOS

```bash
git clone https://github.com/SrPlugin/GhostEnv.git
cd GhostEnv
make build
make install
```

#### Windows

```bash
git clone https://github.com/SrPlugin/GhostEnv.git
cd GhostEnv
make -f Makefile.windows build
```

Or build directly with Go:

```bash
go build -o bin/ghostenv.exe ./cmd/ghostenv
```

### Cross-Compilation

Build for different platforms from Linux/macOS:

```bash
make build-windows  # Windows (amd64)
make build-linux    # Linux (amd64)
make build-darwin   # macOS (amd64 and arm64)
```

## Usage

### Quick Start

```bash
# Set your first secret (will prompt for master password)
ghostenv set API_KEY "your-secret-key"

# Run your application with secrets injected
ghostenv run -- node app.js
```

### Commands Reference

#### Set Secret

Store or update a secret in the vault:

```bash
# Basic usage (prompts for password)
ghostenv set API_KEY "your-api-key"

# With password flag
ghostenv -p "master-password" set DATABASE_URL "postgres://localhost/db"

# In project with specific environment
ghostenv --env production set API_KEY "prod-key"

# Multiple secrets
ghostenv set API_KEY "key1"
ghostenv set DB_PASSWORD "pass123"
ghostenv set JWT_SECRET "secret-token"
```

#### Get Secret

Retrieve the value of a specific secret:

```bash
# Get secret from current environment
ghostenv get API_KEY

# Get from specific environment
ghostenv --env production get API_KEY

# With password flag
ghostenv -p "password" get DATABASE_URL
```

#### List Secrets

List all secret keys stored in the vault:

```bash
# List all secrets
ghostenv list

# List from specific environment
ghostenv --env staging list

# Output example:
# --- Stored Secret Keys ---
# API_KEY
# DATABASE_URL
# JWT_SECRET
# 
# Total: 3 secrets
```

#### Remove Secret

Delete a secret from the vault:

```bash
# Remove a secret
ghostenv remove API_KEY

# Remove from specific environment
ghostenv --env production remove OLD_KEY
```

#### Import from .env File

Import secrets from an existing `.env` file:

```bash
# Import from .env file
ghostenv import .env

# Import to specific environment
ghostenv --env production import .env.production

# Import with password flag
ghostenv -p "password" import secrets.env
```

The import command:
- Parses `KEY=VALUE` format
- Ignores empty lines and comments (lines starting with `#`)
- Merges with existing secrets in the vault
- Validates keys (skips invalid ones)

#### Export Secrets

Export secrets in JSON or `.env` format, to stdout or to a file (useful for CI/CD or sharing config):

```bash
# Export as JSON to stdout (default)
ghostenv export

# Export as .env format
ghostenv export --format env
ghostenv export -f env

# Write to file
ghostenv export --output secrets.json
ghostenv export -o .env.production -f env

# From specific environment
ghostenv --env production export -f env -o .env.prod

# Prefer GHOSTENV_PASS so password is not visible in process list
export GHOSTENV_PASS="your-password"
ghostenv export -o secrets.json
```

#### Version

Print version and build information:

```bash
ghostenv version
```

Output example:
```
ghostenv version 0.1.0
Go version: go1.25.5
```

#### Change Password

Change the master password for the vault. You will be prompted for the current password and then for the new password (twice to confirm):

```bash
# Change password for current environment
ghostenv change-password

# Change password for specific environment
ghostenv --env production change-password

# Provide current password via flag
ghostenv -p "current-password" change-password
```

#### Stats

Show vault statistics: path, type (project or global), environment, key count, and last modified time:

```bash
# Stats for current environment
ghostenv stats

# Stats for specific environment
ghostenv --env production stats

# With password flag
ghostenv -p "password" stats
```

Output example:
```
Vault Statistics
----------------
Path:        /home/user/my-project/.ghostenv/dev.gev
Type:        project
Environment: dev
Keys:        5
Modified:    2026-01-24T12:00:00Z
```

#### Create Shares (Shamir's Secret Sharing)

Split the master password into N secret shares so that K shares are required to recover it (K-of-N). Useful for backup or team recovery without storing the full password in one place.

```bash
# Create 3 shares, 2 required to recover (default: -n 3 -k 2)
ghostenv create-shares --output .ghostenv/shares

# Custom N and K: 5 shares, 3 required
ghostenv create-shares --parts 5 --threshold 3 --output ./backup-shares
ghostenv create-shares -n 5 -k 3 -o ./backup-shares

# You will be prompted for the master password (or use GHOSTENV_PASS / -p)
```

Share files are written as `share-1.txt`, `share-2.txt`, etc. (base64-encoded). Store each share in a separate, secure location. **Do not commit share files to version control.**

#### Recover Master Password from Shares

Recover the master password by combining at least K share files. The recovered password is printed to stdout; use it with `GHOSTENV_PASS` or `change-password` to apply it.

```bash
# Recover from two share files
ghostenv recover .ghostenv/shares/share-1.txt .ghostenv/shares/share-2.txt

# Recover and set as environment variable (Linux/macOS)
export GHOSTENV_PASS=$(ghostenv recover share-1.txt share-2.txt)
ghostenv list

# Recover then change vault password
ghostenv recover share-1.txt share-2.txt | xargs -I {} sh -c 'export GHOSTENV_PASS="{}"; ghostenv change-password'
```

**Security**: The recovered password is printed to stdout. Avoid piping to logs or untrusted commands. Prefer redirecting to a variable or using it only in memory (e.g. `GHOSTENV_PASS`).

#### Run Command with Secrets

Execute a command with secrets injected as environment variables:

```bash
# Run Node.js application
ghostenv run -- node app.js

# Run with arguments
ghostenv run -- npm start --port 3000

# Run Python application
ghostenv run -- python app.py

# Run with specific environment
ghostenv --env production run -- node app.js

# Run Docker container
ghostenv run -- docker run -it myimage

# Run shell script
ghostenv run -- ./deploy.sh
```

### Password Management

The master password protects all secrets in a vault. GhostEnv checks, in order: environment variable `GHOSTENV_PASS`, then flag `-p`, then an interactive prompt.

**Prefer `GHOSTENV_PASS` over `-p`**: Using `-p "password"` makes the password visible in the process list (e.g. `ps aux` on Linux). Use the environment variable so the password is not exposed:

```bash
# Interactive prompt (recommended when typing manually)
ghostenv set API_KEY "value"

# Environment variable (recommended for scripts and CI/CD; not visible in ps)
export GHOSTENV_PASS="your-password"
ghostenv run -- npm test

# Flag -p (avoid in production; password may appear in process list)
ghostenv -p "your-password" set API_KEY "value"
```

**Security Note**: Prefer `GHOSTENV_PASS` or an interactive prompt. Avoid `-p` on shared systems or in production.

### Project Vaults and Environments

GhostEnv automatically detects project directories and creates environment-specific vaults. This allows you to manage different secrets for development, staging, and production.

#### How It Works

1. **Project Detection**: GhostEnv looks for a `.ghostenv/` directory in the current directory or parent directories
2. **Environment Selection**: Each environment has its own encrypted vault file
3. **Default Behavior**: If no environment is specified, uses `dev`
4. **Global Fallback**: Outside project directories, uses the global vault at `~/.ghostenv.gev`

#### Project Structure

When you use GhostEnv in a project, it creates:

```
your-project/
  .ghostenv/
    dev.gev          # Development environment (default)
    production.gev   # Production environment
    staging.gev      # Staging environment
    test.gev         # Test environment
    custom.gev       # Any custom environment name
```

#### Environment Resolution Logic

```
Current Directory → Has .ghostenv/ or is project?
  ├─ YES → Use project vault with specified environment (default: dev)
  └─ NO  → Use global vault at ~/.ghostenv.gev
```

#### Usage Examples

**Working in a Project:**

```bash
# Navigate to your project
cd ~/my-project

# Set secret in dev environment (default)
ghostenv set API_KEY "dev-key"
# → Saves to .ghostenv/dev.gev

# Set secret in production environment
ghostenv --env production set API_KEY "prod-key"
# → Saves to .ghostenv/production.gev

# Set secret in staging
ghostenv --env staging set DATABASE_URL "staging-db"
# → Saves to .ghostenv/staging.gev

# List secrets from dev (default)
ghostenv list
# → Shows secrets from .ghostenv/dev.gev

# List secrets from production
ghostenv --env production list
# → Shows secrets from .ghostenv/production.gev

# Run with dev environment (default)
ghostenv run -- npm start
# → Injects secrets from .ghostenv/dev.gev

# Run with production environment
ghostenv --env production run -- npm start
# → Injects secrets from .ghostenv/production.gev
```

**Working Outside a Project:**

```bash
# Navigate to home or any non-project directory
cd ~

# Set secret in global vault
ghostenv set GLOBAL_KEY "value"
# → Saves to ~/.ghostenv.gev

# List global secrets
ghostenv list
# → Shows secrets from ~/.ghostenv.gev

# Run with global vault
ghostenv run -- some-command
# → Injects secrets from ~/.ghostenv.gev
```

#### Common Workflows

**Development Workflow:**

```bash
cd my-project

# Set up development secrets
ghostenv set DATABASE_URL "postgres://localhost/devdb"
ghostenv set API_KEY "dev-api-key"

# Run development server
ghostenv run -- npm run dev
```

**Production Deployment:**

```bash
cd my-project

# Set production secrets
ghostenv --env production set DATABASE_URL "postgres://prod-server/proddb"
ghostenv --env production set API_KEY "prod-api-key"

# Deploy with production secrets
ghostenv --env production run -- npm run deploy
```

**Multiple Environments:**

```bash
cd my-project

# Set up all environments
ghostenv set DEV_VAR "dev-value"
ghostenv --env staging set STAGING_VAR "staging-value"
ghostenv --env production set PROD_VAR "prod-value"

# Test each environment
ghostenv run -- npm test
ghostenv --env staging run -- npm test
ghostenv --env production run -- npm test
```

**Import Existing .env Files:**

```bash
cd my-project

# Import to dev environment
ghostenv import .env
# → Imports to .ghostenv/dev.gev

# Import to production
ghostenv --env production import .env.production
# → Imports to .ghostenv/production.gev
```

## Configuration

GhostEnv supports a YAML config file for **global** and **per-project** settings. Project config overrides global; missing values use built-in defaults.

### Config file locations

- **Global**: `~/.config/ghostenv/config.yml` or `~/.ghostenv.yml`
- **Project**: `./.ghostenv.yml` in the project root (same directory as `.ghostenv/` or where the config file lives)

### Schema (see `example.yml`)

| Section | Description |
|--------|-------------|
| **project** | `name`, `version`, `default_env` (default environment when `--env` is not set) |
| **storage** | `vault_dir` (path to vaults), `recursive_search`, `auto_backup` (enabled, retention_days, path), optional `environments` (per-env dir overrides) |
| **security** | **argon2**: `memory` (e.g. `64MB`), `iterations`, `parallelism`. **policy**: `max_auth_attempts`, `force_memory_zeroing`, `disallow_password_flag_in_prod` |
| **microservices** | **inheritance**: `enabled`, `shared_vault`. **server**: `host`, `port`, `use_tls`. **postgres**: `enabled`, `host`, `port`, `database`, `user_key` / `pass_key` (vault keys for credentials), `ssl_mode` |
| **scripts** | Alias commands (e.g. `dev: "run --env dev -- node dist/main.js"`) — for future `ghostenv run <alias>` |
| **audit** | `enabled`, `output` (file/stdout/syslog), `file_path`, `log_level`, `mask_keys` (redact key names in log) |
| **export** | `default_format` (json/env), `include_timestamp` |

Project root is detected by the presence of `.ghostenv/` or `.ghostenv.yml`. Relative paths in config (e.g. `./.ghostenv/vaults`) are resolved from the project root.

## Architecture

### Project Structure

```
GhostEnv/
├── cmd/
│   └── ghostenv/          # CLI application
│       ├── main.go        # Command definitions
│       ├── handlers.go    # Command handlers
│       ├── password.go    # Secure password handling
│       └── vault_resolver.go
├── internal/
│   ├── cipher/            # Encryption engine
│   │   ├── cipher.go      # AES-256-GCM encryption
│   │   └── kdf.go         # Argon2id key derivation (uses config for Argon2 params)
│   ├── storage/           # Vault file I/O
│   ├── vault/             # Vault service layer
│   │   ├── vault.go       # Vault operations
│   │   └── resolver.go    # Vault path resolution (uses config for vault_dir, default_env)
│   ├── injector/          # Process execution
│   ├── validator/         # Input validation
│   ├── shamir/            # Shamir's Secret Sharing (split/combine)
│   ├── audit/             # Audit logging (uses config for path, enabled, mask_keys)
│   └── config/            # Configuration: constants, schema, loader (global + project YAML)
├── example.yml            # Example config (project, storage, security, policy, microservices, audit, export)
├── Makefile               # Build system (Linux/macOS)
├── Makefile.windows       # Build system (Windows)
└── README.md
```

### Components

- **cmd/ghostenv/**: CLI interface using Cobra framework with command handlers
- **internal/cipher/**: AES-256-GCM encryption and Argon2id key derivation
- **internal/storage/**: Binary vault file I/O with restricted permissions (0600)
- **internal/vault/**: Vault service layer and path resolution (project/env/global)
- **internal/injector/**: Process execution with environment variable injection
- **internal/validator/**: Input validation for keys and values
- **internal/config/**: Configuration constants, YAML schema, and loader (global + project merge); drives vault paths, Argon2, audit, export defaults

### Technology Stack

- **Language**: Go 1.25.5+
- **CLI Framework**: Cobra
- **Encryption**: AES-256-GCM (Galois/Counter Mode)
- **Key Derivation**: Argon2id (1 iteration, 64MB memory, 4 threads, 32-byte key)
- **Storage**: Binary vault files with restricted permissions

## Security

### Encryption

- **Algorithm**: AES-256-GCM for authenticated encryption
- **Key Derivation**: Argon2id with secure parameters
- **Salt**: 16-byte random salt per encryption operation
- **Nonce**: Random nonce per encryption operation
- **Authentication**: GCM mode prevents tampering

### Storage

- **File Permissions**: Restricted to owner-only access (0600)
- **Binary Format**: Encrypted binary vault prevents accidental reading
- **Memory Safety**: Secrets only exist in memory during execution
- **Password Handling**: Memory zeroing after password use

### Shamir's Secret Sharing

GhostEnv can split the master password into N shares so that K shares (K ≤ N) are required to recover it. This allows:

- **Backup**: Distribute shares to trusted people or locations; no single share reveals the password
- **Recovery**: Regain access to the vault if the password is lost, by combining enough shares
- **Team recovery**: Require multiple people to combine their shares (e.g. 2-of-3) to recover access

Shares are stored as base64-encoded files. Keep each share in a separate, secure location and never commit them to version control.

### Best Practices

- Use strong, unique master passwords
- Never store master passwords in plain text
- Use a password manager for master password storage
- Back up vault files securely if needed
- Rotate secrets regularly
- Use different environments for different deployment stages
- Consider Shamir shares for backup: create shares and store them in separate secure locations

## Why Use GhostEnv

- **Encrypted Storage**: All secrets encrypted with industry-standard algorithms
- **No Plain-Text Files**: Eliminates `.env` files that can be committed to Git
- **Runtime Injection**: Secrets decrypted only when needed
- **Version Control Safe**: Binary format prevents accidental exposure
- **Centralized Management**: Single encrypted vault per environment
- **Language Agnostic**: Works with any runtime that reads environment variables
- **Project Isolation**: Separate vaults for different projects and environments

## Disadvantages of Not Using GhostEnv

- **At-Rest Exposure**: Plain-text files readable by any process with filesystem access
- **Version Control Leaks**: `.env` files frequently committed to Git repositories
- **Backup Security Issues**: Backups may include unencrypted secrets
- **Process Memory Dumps**: Plain-text environment variables extractable from crash dumps
- **Compliance Failures**: Organizations may fail security audits

## Cross-Platform Support

GhostEnv is fully cross-platform and works on:

- **Linux**: Tested on various distributions
- **macOS**: Supports both Intel and Apple Silicon
- **Windows**: Native support with Windows-specific build files

### Platform-Specific Notes

- **Linux/macOS**: Use standard Makefile
- **Windows**: Use `Makefile.windows` or build directly with Go
- **Cross-Compilation**: Build for any platform from any platform using Go's cross-compilation

## Language Compatibility

GhostEnv works with any programming language or runtime that reads environment variables:

- Node.js / JavaScript
- Python
- Java / Spring Boot
- Go
- PHP
- Ruby
- Rust
- C/C++
- Shell Scripts
- Docker Containers
- And any other language with environment variable support

## Development

### Building

```bash
make build          # Build for current platform
make build-windows  # Cross-compile for Windows
make build-linux    # Cross-compile for Linux
make build-darwin   # Cross-compile for macOS
```

### Testing

```bash
make test
```

### Code Quality

```bash
make fmt    # Format code
make vet    # Run go vet
make lint   # Run linter
```

## License

Copyright 2026 Sebastian Cheikh

This project is licensed under a permissive open-source license. See the LICENSE file for details.

## Contributing

Contributions are welcome. Please ensure that any security-related changes are thoroughly reviewed and tested.

## Author

Developed by Sebastian Cheikh
