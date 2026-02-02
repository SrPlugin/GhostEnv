package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SrPlugin/GhostEnv/internal/config"
)

var (
	ErrVaultNotFound    = fmt.Errorf("vault not found")
	ErrVaultReadFailed  = fmt.Errorf("failed to read vault")
	ErrVaultWriteFailed = fmt.Errorf("failed to write vault")
)

func SaveVault(path string, data []byte) error {
	dir := filepath.Dir(path)
	tmpPath := filepath.Join(dir, filepath.Base(path)+".tmp")

	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, config.VaultFilePerm)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrVaultWriteFailed, err)
	}

	_, err = f.Write(data)
	if err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("%w: %v", ErrVaultWriteFailed, err)
	}

	if err := f.Sync(); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("%w: %v", ErrVaultWriteFailed, err)
	}

	if err := f.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("%w: %v", ErrVaultWriteFailed, err)
	}

	_ = os.Remove(path)
	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
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

func VaultModTime(path string) (time.Time, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return time.Time{}, ErrVaultNotFound
		}
		return time.Time{}, fmt.Errorf("%w: %v", ErrVaultReadFailed, err)
	}
	return info.ModTime(), nil
}
