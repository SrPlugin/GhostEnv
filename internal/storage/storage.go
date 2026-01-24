package storage

import (
	"fmt"
	"os"

	"github.com/SrPlugin/GhostEnv/internal/config"
)

var (
	ErrVaultNotFound    = fmt.Errorf("vault not found")
	ErrVaultReadFailed  = fmt.Errorf("failed to read vault")
	ErrVaultWriteFailed = fmt.Errorf("failed to write vault")
)

func SaveVault(path string, data []byte) error {
	if err := os.WriteFile(path, data, config.VaultFilePerm); err != nil {
		return fmt.Errorf("%w: %v", ErrVaultWriteFailed, err)
	}
	return nil
}

func LoadVault(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrVaultNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrVaultReadFailed, err)
	}
	return data, nil
}

func VaultExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
