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
- **Secure Password Handling**: Memory zeroing after password use

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

Export all secrets as JSON:

```bash
# Export to stdout
ghostenv export

# Export from specific environment
ghostenv --env production export

# Save to file
ghostenv export > secrets.json

# Export with password flag
ghostenv -p "password" export | jq
```

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

The master password protects all secrets in a vault. You can provide it via flag or be prompted:

```bash
# Prompt for password (recommended for security)
ghostenv set API_KEY "value"

# Provide password via flag (useful for scripts)
ghostenv -p "your-password" set API_KEY "value"

# Using environment variable (for CI/CD)
export GHOSTENV_PASS="your-password"
ghostenv -p "$GHOSTENV_PASS" run -- npm test
```

**Security Note**: Avoid passing passwords via command line in production. Use the prompt or environment variables.

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
│   │   └── kdf.go         # Argon2id key derivation
│   ├── storage/           # Vault file I/O
│   ├── vault/             # Vault service layer
│   │   ├── vault.go       # Vault operations
│   │   └── resolver.go    # Vault path resolution
│   ├── injector/          # Process execution
│   ├── validator/         # Input validation
│   └── config/            # Configuration constants
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
- **internal/config/**: Centralized configuration constants

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

### Best Practices

- Use strong, unique master passwords
- Never store master passwords in plain text
- Use a password manager for master password storage
- Back up vault files securely if needed
- Rotate secrets regularly
- Use different environments for different deployment stages

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
