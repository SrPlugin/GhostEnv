# GhostEnv

GhostEnv is a secure environment variable manager that encrypts sensitive configuration data and injects it directly into application processes at runtime.

## Description

GhostEnv eliminates the security risks of plain-text `.env` files by storing secrets in an encrypted binary vault. Secrets are decrypted only when needed and injected directly into the target process environment, ensuring they never appear as plain text on disk during execution.

## Usage

### Installation

```bash
make build
make install
```

### Basic Commands

```bash
# Store a secret
ghostenv set API_KEY "your-api-key"

# List all secrets
ghostenv list

# Get a specific secret
ghostenv get API_KEY

# Run a command with secrets injected
ghostenv run -- node app.js

# Import from .env file
ghostenv import .env

# Export all secrets as JSON
ghostenv export

# Remove a secret
ghostenv remove API_KEY
```

### Password

Provide the master password via flag or you will be prompted:

```bash
ghostenv -p "your-password" set KEY "value"
```

## Why Use GhostEnv

- **Encrypted Storage**: All secrets are encrypted using AES-256-GCM with Argon2id key derivation
- **No Plain-Text Files**: Eliminates `.env` files that can be accidentally committed to version control
- **Runtime Injection**: Secrets are decrypted only when needed and injected directly into process memory
- **Version Control Safe**: Binary vault format prevents accidental exposure of readable secrets
- **Centralized Management**: Single encrypted vault for all secrets across multiple applications
- **Language Agnostic**: Works with any runtime that reads environment variables

## Disadvantages of Not Using GhostEnv

- **At-Rest Exposure**: Plain-text files can be read by any process with filesystem access
- **Version Control Leaks**: `.env` files frequently get committed to Git, exposing secrets permanently
- **Backup Security Issues**: Backups may include unencrypted secrets in less secure locations
- **Process Memory Dumps**: Plain-text environment variables can be extracted from crash dumps
- **Compliance Failures**: Organizations may fail security audits if sensitive data is stored in plain text

## How It Works

1. **Encryption**: Secrets are encrypted using AES-256-GCM with a master password. The password is derived into a key using Argon2id.
2. **Storage**: Encrypted data is stored in a binary vault file (`~/.ghostenv.gev`) with restricted permissions (0600).
3. **Decryption**: When running a command, the vault is decrypted using the master password.
4. **Injection**: Decrypted secrets are injected into the target process environment before execution.
5. **Memory Safety**: Secrets exist only in memory during process execution and are never written to disk in plain text.

## Architecture

- **cmd/ghostenv/**: CLI interface using Cobra framework
- **internal/cipher/**: AES-256-GCM encryption and Argon2id key derivation
- **internal/storage/**: Binary vault file I/O operations
- **internal/vault/**: Vault service layer for secret management
- **internal/injector/**: Process execution with environment injection
- **internal/config/**: Centralized configuration constants

## Security

- Master password protection with secure key derivation
- Authenticated encryption preventing tampering
- Random salt and nonce generation for each encryption
- File permissions restricted to owner-only access
- Secrets only exist in memory during execution

## License

Copyright 2026 Sebastian Cheikh

This project is licensed under a permissive open-source license. See the LICENSE file for details.
