package audit

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/SrPlugin/GhostEnv/internal/config"
)

const (
	ActionSet             = "set"
	ActionGet             = "get"
	ActionList             = "list"
	ActionRemove           = "remove"
	ActionImport           = "import"
	ActionExport           = "export"
	ActionRun              = "run"
	ActionChangePassword   = "change-password"
	ActionStats            = "stats"
	ActionCreateShares     = "create-shares"
	ActionRecover          = "recover"
)

type Entry struct {
	Timestamp   string `json:"timestamp"`
	Action      string `json:"action"`
	Environment string `json:"environment,omitempty"`
	VaultPath   string `json:"vault_path,omitempty"`
	Key         string `json:"key,omitempty"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

var (
	mu       sync.Mutex
	disabled bool
)

func init() {
	if os.Getenv("GHOSTENV_AUDIT_DISABLE") == "1" || os.Getenv("GHOSTENV_AUDIT_DISABLE") == "true" {
		disabled = true
	}
}

func Log(action, vaultPath, environment, key string, success bool, errMsg string) {
	if disabled {
		return
	}
	cfg := config.Current()
	if cfg != nil && !cfg.Audit.Enabled {
		return
	}
	mu.Lock()
	defer mu.Unlock()

	entry := Entry{
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Action:      action,
		Environment: environment,
		VaultPath:   vaultPath,
		Key:         key,
		Success:     success,
		Error:       errMsg,
	}
	if cfg != nil && cfg.Audit.MaskKeys {
		entry.Key = "[redacted]"
	}

	logPath := resolveLogPath(vaultPath)
	if logPath == "" {
		return
	}

	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return
	}

	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintln(f, string(data))
}

func resolveLogPath(vaultPath string) string {
	if path := os.Getenv("GHOSTENV_AUDIT_LOG"); path != "" {
		return path
	}
	cfg := config.Current()
	if cfg != nil && cfg.Audit.FilePath != "" {
		p := cfg.Audit.FilePath
		if !filepath.IsAbs(p) {
			root := config.ProjectRoot()
			if root != "" {
				p = filepath.Join(root, p)
			}
		}
		return filepath.Clean(p)
	}
	if vaultPath == "" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".ghostenv", "audit.log")
	}
	dir := filepath.Dir(vaultPath)
	return filepath.Join(dir, "audit.log")
}
