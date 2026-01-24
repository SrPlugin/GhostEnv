package vault

import (
	"encoding/json"
	"fmt"

	"github.com/SrPlugin/GhostEnv/internal/cipher"
	"github.com/SrPlugin/GhostEnv/internal/storage"
)

type Service struct {
	vaultPath string
}

func NewService(vaultPath string) *Service {
	return &Service{
		vaultPath: vaultPath,
	}
}

func (s *Service) Load(password string) (map[string]string, error) {
	data, err := storage.LoadVault(s.vaultPath)
	if err != nil {
		return nil, err
	}

	decrypted, err := cipher.Decrypt(data, password)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	var secrets map[string]string
	if err := json.Unmarshal(decrypted, &secrets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vault data: %w", err)
	}

	return secrets, nil
}

func (s *Service) Save(secrets map[string]string, password string) error {
	payload, err := json.Marshal(secrets)
	if err != nil {
		return fmt.Errorf("failed to marshal secrets: %w", err)
	}

	encrypted, err := cipher.Encrypt(payload, password)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	if err := storage.SaveVault(s.vaultPath, encrypted); err != nil {
		return fmt.Errorf("failed to save vault: %w", err)
	}

	return nil
}

func (s *Service) Exists() bool {
	return storage.VaultExists(s.vaultPath)
}
