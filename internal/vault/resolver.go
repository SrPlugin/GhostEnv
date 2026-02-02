package vault

import (
	"os"
	"path/filepath"

	"github.com/SrPlugin/GhostEnv/internal/config"
)

type Resolver interface {
	ResolveVaultPath(environment string) (string, VaultType, error)
}

type VaultType string

const (
	VaultTypeProject VaultType = "project"
	VaultTypeGlobal  VaultType = "global"
)

type resolver struct {
	projectRoot string
	cfg         *config.Config
}

func NewResolver() Resolver {
	root := findProjectRoot()
	cfg, _ := config.Load(root)
	if cfg == nil {
		cfg = config.Default()
	}
	config.SetCurrent(cfg)
	config.SetProjectRoot(root)
	return &resolver{projectRoot: root, cfg: cfg}
}

func findProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}

	dir := wd
	for {
		// Project: .ghostenv/ directory
		vaultDir := filepath.Join(dir, config.ProjectVaultDir)
		if info, err := os.Stat(vaultDir); err == nil && info.IsDir() {
			return dir
		}
		vaultFile := filepath.Join(vaultDir, config.DefaultEnvironment+".gev")
		if _, err := os.Stat(vaultFile); err == nil {
			return dir
		}
		// Project: .ghostenv.yml config file
		if _, err := os.Stat(filepath.Join(dir, config.ProjectConfigName)); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return wd
}

func (r *resolver) ResolveVaultPath(environment string) (string, VaultType, error) {
	if environment == "" && r.cfg != nil {
		environment = r.cfg.Project.DefaultEnv
	}
	if environment == "" {
		environment = config.DefaultEnvironment
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", err
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}

	// Global vault when project root is home (no project detected in a meaningful way)
	absHome, _ := filepath.Abs(home)
	absProject, _ := filepath.Abs(r.projectRoot)
	if r.projectRoot == "" || absProject == absHome || r.projectRoot == wd {
		if absProject == absHome {
			vaultPath := filepath.Join(home, config.VaultFileName)
			return vaultPath, VaultTypeGlobal, nil
		}
	}

	// Project vault: use config vault_dir (relative to project root) or default
	vaultDir := r.cfg.Storage.VaultDir
	if vaultDir == "" {
		vaultDir = filepath.Join(".", config.ProjectVaultDir)
	}
	if !filepath.IsAbs(vaultDir) {
		vaultDir = filepath.Join(r.projectRoot, vaultDir)
	}
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		return "", "", err
	}
	// Per-environment override from config.Storage.Environments
	if r.cfg.Storage.Environments != nil {
		if e, ok := r.cfg.Storage.Environments[environment]; ok && e.Dir != "" {
			vaultDir = filepath.Join(vaultDir, e.Dir)
			_ = os.MkdirAll(vaultDir, 0755)
		}
	}
	vaultPath := filepath.Join(vaultDir, environment+".gev")
	return vaultPath, VaultTypeProject, nil
}

func GetVaultPath(environment string) (string, VaultType, error) {
	r := NewResolver()
	return r.ResolveVaultPath(environment)
}
