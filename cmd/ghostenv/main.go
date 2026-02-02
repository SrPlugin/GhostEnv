package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/SrPlugin/GhostEnv/internal/injector"
	"github.com/SrPlugin/GhostEnv/internal/version"
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

	rootCmd.PersistentFlags().StringVarP(&masterPassword, "pass", "p", "", "Master password (prefer GHOSTENV_PASS env to avoid visibility in process list)")
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

	var exportFormat string
	var exportOutput string
	var exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export secrets in JSON or .env format",
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword(masterPassword)
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}
			return h.handleExport(pw, environment, exportFormat, exportOutput)
		},
	}
	exportCmd.Flags().StringVarP(&exportFormat, "format", "f", "", "Output format: json or env (default from config or json)")
	exportCmd.Flags().StringVarP(&exportOutput, "output", "o", "", "Write to file instead of stdout")

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version and build information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("ghostenv version %s\n", version.Version)
			if version.BuildDate != "" {
				fmt.Printf("Build date: %s\n", version.BuildDate)
			}
			if version.Commit != "" {
				fmt.Printf("Commit: %s\n", version.Commit)
			}
			fmt.Printf("Go version: %s\n", runtime.Version())
			return nil
		},
	}

	var changePasswordCmd = &cobra.Command{
		Use:   "change-password",
		Short: "Change the master password for the vault",
		RunE: func(cmd *cobra.Command, args []string) error {
			currentPw, err := getPassword(masterPassword)
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}
			newPw, err := getNewPassword()
			if err != nil {
				zeroBytes(currentPw)
				return fmt.Errorf("new password error: %w", err)
			}
			err = h.handleChangePassword(currentPw, newPw, environment)
			if err != nil {
				return err
			}
			return nil
		},
	}

	var statsCmd = &cobra.Command{
		Use:   "stats",
		Short: "Show vault statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword(masterPassword)
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}
			return h.handleStats(pw, environment)
		},
	}

	var createSharesParts int
	var createSharesThreshold int
	var createSharesOutput string
	var createSharesCmd = &cobra.Command{
		Use:   "create-shares",
		Short: "Split master password into Shamir secret shares",
		Long:  "Prompts for the master password and splits it into N shares; K shares are required to recover it (K-of-N).",
		RunE: func(cmd *cobra.Command, args []string) error {
			pw, err := getPassword(masterPassword)
			if err != nil {
				return fmt.Errorf("password error: %w", err)
			}
			return h.handleCreateShares(createSharesParts, createSharesThreshold, createSharesOutput, pw, environment)
		},
	}
	createSharesCmd.Flags().IntVarP(&createSharesParts, "parts", "n", 3, "Total number of shares to create")
	createSharesCmd.Flags().IntVarP(&createSharesThreshold, "threshold", "k", 2, "Minimum shares required to recover")
	createSharesCmd.Flags().StringVarP(&createSharesOutput, "output", "o", "", "Directory to write share files (required)")
	createSharesCmd.MarkFlagRequired("output")

	var recoverCmd = &cobra.Command{
		Use:   "recover [SHARE_FILE...]",
		Short: "Recover master password from Shamir shares",
		Long:  "Reads share files and prints the recovered master password. Use with GHOSTENV_PASS or change-password to apply it.",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			return h.handleRecover(args, environment)
		},
	}

	rootCmd.AddCommand(setCmd, runCmd, listCmd, getCmd, removeCmd, importCmd, exportCmd, versionCmd, changePasswordCmd, statsCmd, createSharesCmd, recoverCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
