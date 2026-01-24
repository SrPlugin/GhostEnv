package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/SrPlugin/GhostEnv/internal/config"
	"github.com/SrPlugin/GhostEnv/internal/injector"
	"github.com/SrPlugin/GhostEnv/internal/storage"
	"github.com/SrPlugin/GhostEnv/internal/vault"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var masterPassword string

func getVaultPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, config.VaultFileName)
}

func getPassword() (string, error) {
	if masterPassword != "" {
		return masterPassword, nil
	}
	fmt.Print("Enter Master Password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	if len(bytePassword) == 0 {
		return "", fmt.Errorf("password cannot be empty")
	}
	return string(bytePassword), nil
}

func main() {
	vaultPath := getVaultPath()
	vaultService := vault.NewService(vaultPath)

	var rootCmd = &cobra.Command{
		Use:   "ghostenv",
		Short: "GhostEnv: Secure Environment Variable Manager",
		Long:  "GhostEnv - Securely encrypt and inject environment variables.\nDeveloped by Sebastian Cheikh",
	}

	rootCmd.PersistentFlags().StringVarP(&masterPassword, "pass", "p", "", "Master password for the vault (optional, will prompt if missing)")

	var setCmd = &cobra.Command{
		Use:  "set [KEY] [VALUE]",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword()
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}

			secrets := make(map[string]string)
			if vaultService.Exists() {
				existingSecrets, err := vaultService.Load(pw)
				if err != nil && err != storage.ErrVaultNotFound {
					return fmt.Errorf("failed to load existing vault: %w", err)
				}
				if existingSecrets != nil {
					secrets = existingSecrets
				}
			}

			secrets[args[0]] = args[1]
			if err := vaultService.Save(secrets, pw); err != nil {
				return fmt.Errorf("failed to save secret: %w", err)
			}

			fmt.Printf("Secret '%s' saved successfully\n", args[0])
			return nil
		},
	}

	var runCmd = &cobra.Command{
		Use:  "run -- [command]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword()
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}

			secrets, err := vaultService.Load(pw)
			if err != nil {
				if err == storage.ErrVaultNotFound {
					return fmt.Errorf("vault not found. Run 'set' first")
				}
				return fmt.Errorf("failed to load vault: %w", err)
			}

			if err := injector.Run(args[0], args[1:], secrets); err != nil {
				return fmt.Errorf("command execution failed: %w", err)
			}

			return nil
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all stored keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword()
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}

			secrets, err := vaultService.Load(pw)
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
		},
	}

	var getCmd = &cobra.Command{
		Use:   "get [KEY]",
		Short: "Show the value of a specific secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword()
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}

			secrets, err := vaultService.Load(pw)
			if err != nil {
				if err == storage.ErrVaultNotFound {
					return fmt.Errorf("vault not found")
				}
				return fmt.Errorf("failed to load vault: %w", err)
			}

			if val, ok := secrets[args[0]]; ok {
				fmt.Printf("%s = %s\n", args[0], val)
			} else {
				return fmt.Errorf("secret '%s' not found", args[0])
			}
			return nil
		},
	}

	var removeCmd = &cobra.Command{
		Use:   "remove [KEY]",
		Short: "Delete a secret from the vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword()
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}

			secrets, err := vaultService.Load(pw)
			if err != nil {
				if err == storage.ErrVaultNotFound {
					return fmt.Errorf("vault not found")
				}
				return fmt.Errorf("failed to load vault: %w", err)
			}

			if _, ok := secrets[args[0]]; !ok {
				return fmt.Errorf("secret '%s' not found", args[0])
			}

			delete(secrets, args[0])
			if err := vaultService.Save(secrets, pw); err != nil {
				return fmt.Errorf("failed to save vault: %w", err)
			}

			fmt.Printf("Secret '%s' removed successfully\n", args[0])
			return nil
		},
	}

	var importCmd = &cobra.Command{
		Use:   "import [FILE_PATH]",
		Short: "Import secrets from a .env file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword()
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}

			content, err := os.ReadFile(args[0])
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			secrets := make(map[string]string)
			if vaultService.Exists() {
				existingSecrets, err := vaultService.Load(pw)
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
						secrets[key] = val
						count++
					}
				}
			}

			if err := vaultService.Save(secrets, pw); err != nil {
				return fmt.Errorf("failed to save vault: %w", err)
			}

			fmt.Printf("Successfully imported %d secrets from %s\n", count, args[0])
			return nil
		},
	}

	var exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export all secrets in JSON format",
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword()
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}

			secrets, err := vaultService.Load(pw)
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
		},
	}

	rootCmd.AddCommand(setCmd, runCmd, listCmd, getCmd, removeCmd, importCmd, exportCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
