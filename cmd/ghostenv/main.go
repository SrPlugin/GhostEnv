package main

import (
	"fmt"
	"os"

	"github.com/SrPlugin/GhostEnv/internal/injector"
	"github.com/spf13/cobra"
)

var (
	masterPassword string
	environment    string
)

func main() {
	runner := injector.NewRunner()
	h := newHandlers(runner)

	var rootCmd = &cobra.Command{
		Use:   "ghostenv",
		Short: "GhostEnv: Secure Environment Variable Manager",
		Long:  "GhostEnv - Securely encrypt and inject environment variables.\nDeveloped by Sebastian Cheikh",
	}

	rootCmd.PersistentFlags().StringVarP(&masterPassword, "pass", "p", "", "Master password for the vault (optional, will prompt if missing)")
	rootCmd.PersistentFlags().StringVarP(&environment, "env", "e", "", "Environment name (default: dev, uses global vault if not in project)")

	var setCmd = &cobra.Command{
		Use:  "set [KEY] [VALUE]",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword(masterPassword)
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}
			return h.handleSet(args[0], args[1], pw, environment)
		},
	}

	var runCmd = &cobra.Command{
		Use:  "run -- [command]",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword(masterPassword)
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}
			return h.handleRun(args[0], args[1:], pw, environment)
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all stored keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword(masterPassword)
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}
			return h.handleList(pw, environment)
		},
	}

	var getCmd = &cobra.Command{
		Use:   "get [KEY]",
		Short: "Show the value of a specific secret",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword(masterPassword)
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}
			return h.handleGet(args[0], pw, environment)
		},
	}

	var removeCmd = &cobra.Command{
		Use:   "remove [KEY]",
		Short: "Delete a secret from the vault",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword(masterPassword)
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}
			return h.handleRemove(args[0], pw, environment)
		},
	}

	var importCmd = &cobra.Command{
		Use:   "import [FILE_PATH]",
		Short: "Import secrets from a .env file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword(masterPassword)
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}
			return h.handleImport(args[0], pw, environment)
		},
	}

	var exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export all secrets in JSON format",
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword(masterPassword)
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}
			return h.handleExport(pw, environment)
		},
	}

	rootCmd.AddCommand(setCmd, runCmd, listCmd, getCmd, removeCmd, importCmd, exportCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
