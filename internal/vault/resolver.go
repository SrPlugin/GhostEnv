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
}

func NewResolver() Resolver {
	return &resolver{
		projectRoot: findProjectRoot(),
	}
}

func findProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}

	dir := wd
	for {
		vaultDir := filepath.Join(dir, config.ProjectVaultDir)
		if info, err := os.Stat(vaultDir); err == nil && info.IsDir() {
			return dir
		}

		vaultFile := filepath.Join(vaultDir, config.DefaultEnvironment+".gev")
		if _, err := os.Stat(vaultFile); err == nil {
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

	if r.projectRoot == "" || r.projectRoot == home || r.projectRoot == wd {
		absHome, _ := filepath.Abs(home)
		absProject, _ := filepath.Abs(r.projectRoot)
		if absProject == absHome {
			vaultPath := filepath.Join(home, config.VaultFileName)
			return vaultPath, VaultTypeGlobal, nil
		}
	}

	vaultDir := filepath.Join(r.projectRoot, config.ProjectVaultDir)
	if err := os.MkdirAll(vaultDir, 0755); err != nil {
		return "", "", err
	}
	vaultPath := filepath.Join(vaultDir, environment+".gev")
	return vaultPath, VaultTypeProject, nil
}

func GetVaultPath(environment string) (string, VaultType, error) {
	r := NewResolver()
	return r.ResolveVaultPath(environment)
}
