# GhostEnv Roadmap

Ideas for functionality and security improvements. Prioritize by impact and effort.

---

## Security

### High Priority

| Item | Description | Effort |
|------|-------------|--------|
| **Rate limiting** | Limit failed password attempts (e.g. 5 attempts, then lockout or delay) to reduce brute-force risk | Medium |
| **Password strength** | Optional check on master password (length, complexity) with clear messages | Low |
| **Vault integrity** | Command to verify vault (decrypt + HMAC/checksum) and detect corruption or tampering | Low |
| **No password in process list** | Avoid `-p "password"` showing in `ps`; prefer env var `GHOSTENV_PASS` and document it | Low (docs + warning) |
| **Secure memory for secrets** | Where possible, avoid copying decrypted secrets; use and clear in place (Go limits how much we can do) | Medium |

### Medium Priority

| Item | Description | Effort |
|------|-------------|--------|
| **Argon2 parameters** | Make time/memory configurable or increase defaults for stronger KDF | Low |
| **Vault version/metadata** | Add version field in vault format for future migrations and compatibility checks | Low |
| **Audit / access log** | Optional log of operations (e.g. set/remove/export) without logging secret values | Medium |
| **Timeouts** | Context-based timeouts for decrypt and I/O to avoid hanging on bad input or slow storage | Low |

### Lower Priority

| Item | Description | Effort |
|------|-------------|--------|
| **Key derivation per env** | Optional per-environment salt or KDF tweak so one master password doesn’t protect all envs identically | Medium |
| **Pluggable backends** | Interface for vault storage (e.g. encrypted file vs future KMS) without changing core logic | High |

---

## Functionality

### High Priority

| Item | Description | Effort |
|------|-------------|--------|
| **Change master password** | `change-password` (or `re-encrypt`) to re-encrypt vault with a new password without changing secrets | Medium |
| **Export formats** | Export to `.env` and optionally YAML, not only JSON; flag e.g. `--format env \| json \| yaml` | Low |
| **Export to file** | `export -o file` or `export --output file` to write to file instead of stdout | Low |
| **Import formats** | Import from JSON (and optionally YAML), not only `.env`; detect by content or `--format` | Low |
| **Import/export feedback** | Import: report count of imported/updated/skipped/invalid keys; optional `--dry-run` | Low |
| **Unit tests** | Tests for cipher, vault, storage, resolver, handlers (with mocks where needed) | Medium |
| **Version command** | `ghostenv version` with version and build info (e.g. from ldflags) | Low |

### Medium Priority

| Item | Description | Effort |
|------|-------------|--------|
| **Backup / restore** | `backup` (e.g. copy vault to timestamped file) and `restore` from backup | Low |
| **Vault stats** | `stats` or `info`: key count, last modified, vault path, environment | Low |
| **List with filter** | `list --filter "API_*"` or `list --prefix PREFIX` to show only matching keys | Low |
| **Get multiple keys** | `get KEY1 KEY2` or `get --keys key1,key2` to print several values (e.g. for scripts) | Low |
| **Bulk remove** | `remove KEY1 KEY2` or `remove --all` (with confirmation) for cleanup | Low |
| **Run with key subset** | `run --only "API_*,DB_*" -- command` to inject only matching keys | Medium |
| **CI-friendly output** | `export --format env` or `run --env-file` to produce env file for CI without interactive password | Low |
| **Config file** | Optional config (e.g. `~/.ghostenv.yaml`) for default env, vault path, Argon2 params | Medium |

### Lower Priority

| Item | Description | Effort |
|------|-------------|--------|
| **Copy between envs** | `copy --from dev --to staging` to clone or merge one environment into another | Medium |
| **Templates** | `run --template "script.sh.tpl"` to render template with secrets and then run | High |
| **Shell completion** | Improve Cobra completion for commands and flags (e.g. list envs, list keys) | Low |
| **Vault migration** | One-off migration from old vault format if we ever change it | Low when needed |

---

## Summary by Area

- **Security**: Rate limiting, password strength, vault integrity, no password in argv, secure memory usage, optional stronger Argon2 and audit.
- **Functionality**: Change password, export/import formats and to file, better import feedback, tests, version command, backup/restore, stats, list filter, multi-get, bulk remove, optional config file.

Suggested order to implement first:

1. **Version command** (quick win).
2. **Export to file + export as .env** (very useful for backups and CI).
3. **Change master password** (important for security operations).
4. **Unit tests** (stability and refactors).
5. **Rate limiting** (security).
6. **Vault verify / integrity** (reliability and security).
