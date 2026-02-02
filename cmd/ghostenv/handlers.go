package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SrPlugin/GhostEnv/internal/audit"
	"github.com/SrPlugin/GhostEnv/internal/config"
	"github.com/SrPlugin/GhostEnv/internal/injector"
	"github.com/SrPlugin/GhostEnv/internal/shamir"
	"github.com/SrPlugin/GhostEnv/internal/storage"
	"github.com/SrPlugin/GhostEnv/internal/validator"
	"github.com/SrPlugin/GhostEnv/internal/vault"
)

type handlers struct {
	runner injector.Runner
}

func newHandlers(runner injector.Runner) *handlers {
	return &handlers{
		runner: runner,
	}
}

func (h *handlers) getVaultService(environment string) (vault.Service, error) {
	return getVaultService(environment)
}

func auditLog(action, vaultPath, env, key string, err error) {
	success := err == nil
	msg := ""
	if err != nil {
		msg = err.Error()
	}
	audit.Log(action, vaultPath, env, key, success, msg)
}

func (h *handlers) handleSet(key, value string, password []byte, environment string) (err error) {
	defer zeroBytes(password)
	vaultPath, _, _ := vault.GetVaultPath(environment)
	defer func() { auditLog(audit.ActionSet, vaultPath, environment, key, err) }()
	if err = validator.ValidateKey(key); err != nil {
		return fmt.Errorf("invalid key: %w", err)
	}

	vaultService, err := h.getVaultService(environment)
	if err != nil {
		return fmt.Errorf("failed to resolve vault: %w", err)
	}

	secrets := make(map[string]string)
	if vaultService.Exists() {
		existingSecrets, err := vaultService.Load(password)
		if err != nil && err != storage.ErrVaultNotFound {
			return fmt.Errorf("failed to load existing vault: %w", err)
		}
		if existingSecrets != nil {
			secrets = existingSecrets
		}
	}

	secrets[key] = value
	if err = vaultService.Save(secrets, password); err != nil {
		return fmt.Errorf("failed to save secret: %w", err)
	}

	fmt.Printf("Secret '%s' saved successfully\n", key)
	return nil
}

func (h *handlers) handleRun(command string, args []string, password []byte, environment string) (err error) {
	defer zeroBytes(password)
	vaultPath, _, _ := vault.GetVaultPath(environment)
	defer func() { auditLog(audit.ActionRun, vaultPath, environment, command, err) }()
	vaultService, err := h.getVaultService(environment)
	if err != nil {
		return fmt.Errorf("failed to resolve vault: %w", err)
	}

	secrets, err := vaultService.Load(password)
	if err != nil {
		if err == storage.ErrVaultNotFound {
			return fmt.Errorf("vault not found. Run 'set' first")
		}
		return fmt.Errorf("failed to load vault: %w", err)
	}

	if err = h.runner.Run(command, args, secrets); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

func (h *handlers) handleList(password []byte, environment string) (err error) {
	defer zeroBytes(password)
	vaultPath, _, _ := vault.GetVaultPath(environment)
	defer func() { auditLog(audit.ActionList, vaultPath, environment, "", err) }()
	vaultService, err := h.getVaultService(environment)
	if err != nil {
		return fmt.Errorf("failed to resolve vault: %w", err)
	}

	secrets, err := vaultService.Load(password)
	if err != nil {
		if err == storage.ErrVaultNotFound {
			return fmt.Errorf("vault not found")
		}
		return fmt.Errorf("failed to load vault: %w", err)
	}

	fmt.Println("--- Stored Secret Keys ---")
	for key := range secrets {
		fmt.Printf("%s\n", key)
	}
	fmt.Printf("\nTotal: %d secrets\n", len(secrets))
	return nil
}

func (h *handlers) handleGet(key string, password []byte, environment string) (err error) {
	defer zeroBytes(password)
	vaultPath, _, _ := vault.GetVaultPath(environment)
	defer func() { auditLog(audit.ActionGet, vaultPath, environment, key, err) }()
	vaultService, err := h.getVaultService(environment)
	if err != nil {
		return fmt.Errorf("failed to resolve vault: %w", err)
	}

	secrets, err := vaultService.Load(password)
	if err != nil {
		if err == storage.ErrVaultNotFound {
			return fmt.Errorf("vault not found")
		}
		return fmt.Errorf("failed to load vault: %w", err)
	}

	if val, ok := secrets[key]; ok {
		fmt.Printf("%s = %s\n", key, val)
	} else {
		return fmt.Errorf("secret '%s' not found", key)
	}
	return nil
}

func (h *handlers) handleRemove(key string, password []byte, environment string) (err error) {
	defer zeroBytes(password)
	vaultPath, _, _ := vault.GetVaultPath(environment)
	defer func() { auditLog(audit.ActionRemove, vaultPath, environment, key, err) }()
	vaultService, err := h.getVaultService(environment)
	if err != nil {
		return fmt.Errorf("failed to resolve vault: %w", err)
	}

	secrets, err := vaultService.Load(password)
	if err != nil {
		if err == storage.ErrVaultNotFound {
			return fmt.Errorf("vault not found")
		}
		return fmt.Errorf("failed to load vault: %w", err)
	}

	if _, ok := secrets[key]; !ok {
		return fmt.Errorf("secret '%s' not found", key)
	}

	delete(secrets, key)
	if err := vaultService.Save(secrets, password); err != nil {
		return fmt.Errorf("failed to save vault: %w", err)
	}

	fmt.Printf("Secret '%s' removed successfully\n", key)
	return nil
}

func (h *handlers) handleImport(filePath string, password []byte, environment string) (err error) {
	defer zeroBytes(password)
	vaultPath, _, _ := vault.GetVaultPath(environment)
	defer func() { auditLog(audit.ActionImport, vaultPath, environment, filePath, err) }()
	vaultService, err := h.getVaultService(environment)
	if err != nil {
		return fmt.Errorf("failed to resolve vault: %w", err)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	secrets := make(map[string]string)
	if vaultService.Exists() {
		existingSecrets, err := vaultService.Load(password)
		if err != nil && err != storage.ErrVaultNotFound {
			return fmt.Errorf("failed to load existing vault: %w", err)
		}
		if existingSecrets != nil {
			secrets = existingSecrets
		}
	}

	lines := strings.Split(string(content), "\n")
	count := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			if key != "" {
				if err := validator.ValidateKey(key); err != nil {
					continue
				}
				secrets[key] = val
				count++
			}
		}
	}

	if err := vaultService.Save(secrets, password); err != nil {
		return fmt.Errorf("failed to save vault: %w", err)
	}

	fmt.Printf("Successfully imported %d secrets from %s\n", count, filePath)
	return nil
}

func (h *handlers) handleExport(password []byte, environment, format, outputPath string) (err error) {
	defer zeroBytes(password)
	vaultPath, _, _ := vault.GetVaultPath(environment)
	defer func() { auditLog(audit.ActionExport, vaultPath, environment, "", err) }()
	if format == "" {
		if c := config.Current(); c != nil && c.Export.DefaultFormat != "" {
			format = c.Export.DefaultFormat
		}
		if format == "" {
			format = "json"
		}
	}
	vaultService, err := h.getVaultService(environment)
	if err != nil {
		return fmt.Errorf("failed to resolve vault: %w", err)
	}

	secrets, err := vaultService.Load(password)
	if err != nil {
		if err == storage.ErrVaultNotFound {
			return fmt.Errorf("vault not found")
		}
		return fmt.Errorf("failed to load vault: %w", err)
	}

	var out []byte
	switch format {
	case "env":
		var b strings.Builder
		for k, v := range secrets {
			b.WriteString(k)
			b.WriteString("=")
			b.WriteString(v)
			b.WriteString("\n")
		}
		out = []byte(b.String())
	default:
		var err error
		out, err = json.MarshalIndent(secrets, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal secrets: %w", err)
		}
	}

	if outputPath != "" {
		dir := filepath.Dir(outputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		if err := os.WriteFile(outputPath, out, 0600); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Exported to %s\n", outputPath)
	} else {
		fmt.Println(string(out))
	}
	return nil
}

func (h *handlers) handleChangePassword(currentPassword, newPassword []byte, environment string) (err error) {
	defer zeroBytes(currentPassword)
	defer zeroBytes(newPassword)
	vaultPath, _, _ := vault.GetVaultPath(environment)
	defer func() { auditLog(audit.ActionChangePassword, vaultPath, environment, "", err) }()
	vaultService, err := h.getVaultService(environment)
	if err != nil {
		return fmt.Errorf("failed to resolve vault: %w", err)
	}

	secrets, err := vaultService.Load(currentPassword)
	if err != nil {
		if err == storage.ErrVaultNotFound {
			return fmt.Errorf("vault not found")
		}
		return fmt.Errorf("failed to load vault (wrong password?): %w", err)
	}

	if err = vaultService.Save(secrets, newPassword); err != nil {
		return fmt.Errorf("failed to save vault with new password: %w", err)
	}

	fmt.Println("Password changed successfully")
	return nil
}

func (h *handlers) handleStats(password []byte, environment string) (err error) {
	defer zeroBytes(password)
	vaultPath, vaultType, err := vault.GetVaultPath(environment)
	defer func() { auditLog(audit.ActionStats, vaultPath, environment, "", err) }()
	if err != nil {
		return fmt.Errorf("failed to resolve vault: %w", err)
	}

	vaultService := vault.NewService(vaultPath)
	if !vaultService.Exists() {
		return fmt.Errorf("vault not found")
	}

	secrets, err := vaultService.Load(password)
	if err != nil {
		if err == storage.ErrVaultNotFound {
			return fmt.Errorf("vault not found")
		}
		return fmt.Errorf("failed to load vault: %w", err)
	}

	modTime, err := storage.VaultModTime(vaultPath)
	if err != nil {
		modTime = time.Time{}
	}

	envName := environment
	if envName == "" {
		envName = "dev"
	}

	fmt.Println("Vault Statistics")
	fmt.Println("----------------")
	fmt.Printf("Path:        %s\n", vaultPath)
	fmt.Printf("Type:        %s\n", vaultType)
	fmt.Printf("Environment: %s\n", envName)
	fmt.Printf("Keys:        %d\n", len(secrets))
	if !modTime.IsZero() {
		fmt.Printf("Modified:    %s\n", modTime.Format(time.RFC3339))
	}
	return nil
}

func (h *handlers) handleCreateShares(parts, threshold int, outputDir string, password []byte, environment string) (err error) {
	defer zeroBytes(password)
	vaultPath, _, _ := vault.GetVaultPath(environment)
	defer func() { auditLog(audit.ActionCreateShares, vaultPath, environment, outputDir, err) }()

	if parts < 2 || parts > 255 {
		return fmt.Errorf("parts must be between 2 and 255, got %d", parts)
	}
	if threshold < 2 || threshold > parts {
		return fmt.Errorf("threshold must be between 2 and parts (%d), got %d", parts, threshold)
	}

	shares, err := shamir.SplitSecret(password, parts, threshold)
	if err != nil {
		return fmt.Errorf("failed to split secret: %w", err)
	}
	defer shamir.ZeroShares(shares)

	for i, share := range shares {
		path := filepath.Join(outputDir, fmt.Sprintf("share-%d.txt", i+1))
		if err := shamir.WriteShareToFile(share, path); err != nil {
			return fmt.Errorf("failed to write share %d: %w", i+1, err)
		}
		fmt.Printf("Written %s\n", path)
	}
	fmt.Printf("Created %d shares; %d required to recover.\n", parts, threshold)
	return nil
}

func (h *handlers) handleRecover(sharePaths []string, environment string) (err error) {
	vaultPath, _, _ := vault.GetVaultPath(environment)
	defer func() { auditLog(audit.ActionRecover, vaultPath, environment, "", err) }()

	if len(sharePaths) < 2 {
		return fmt.Errorf("at least 2 share files required")
	}

	shares := make([][]byte, 0, len(sharePaths))
	for _, p := range sharePaths {
		data, err := shamir.ReadShareFromFile(p)
		if err != nil {
			return fmt.Errorf("failed to read share %s: %w", p, err)
		}
		shares = append(shares, data)
	}
	defer shamir.ZeroShares(shares)

	recovered, err := shamir.CombineShares(shares)
	if err != nil {
		return fmt.Errorf("failed to combine shares: %w", err)
	}
	defer zeroBytes(recovered)

	fmt.Println(string(recovered))
	return nil
}
