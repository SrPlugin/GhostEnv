package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/SrPlugin/GhostEnv/internal/injector"
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

func (h *handlers) handleSet(key, value, password, environment string) error {
	if err := validator.ValidateKey(key); err != nil {
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
	if err := vaultService.Save(secrets, password); err != nil {
		return fmt.Errorf("failed to save secret: %w", err)
	}

	fmt.Printf("Secret '%s' saved successfully\n", key)
	return nil
}

func (h *handlers) handleRun(command string, args []string, password, environment string) error {
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

	if err := h.runner.Run(command, args, secrets); err != nil {
		return fmt.Errorf("command execution failed: %w", err)
	}

	return nil
}

func (h *handlers) handleList(password, environment string) error {
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

func (h *handlers) handleGet(key, password, environment string) error {
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

func (h *handlers) handleRemove(key, password, environment string) error {
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

func (h *handlers) handleImport(filePath, password, environment string) error {
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

func (h *handlers) handleExport(password, environment string) error {
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

	output, err := json.MarshalIndent(secrets, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal secrets: %w", err)
	}

	fmt.Println(string(output))
	return nil
}
