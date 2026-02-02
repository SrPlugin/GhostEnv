package vault

import (
	"encoding/json"
	"fmt"

	"github.com/SrPlugin/GhostEnv/internal/cipher"
	"github.com/SrPlugin/GhostEnv/internal/storage"
)

type Service interface {
	Load(password []byte) (map[string]string, error)
	Save(secrets map[string]string, password []byte) error
	Exists() bool
}

type service struct {
	vaultPath string
}

func NewService(vaultPath string) Service {
	return &service{
		vaultPath: vaultPath,
	}
}

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func (s *service) Load(password []byte) (map[string]string, error) {
	data, err := storage.LoadVault(s.vaultPath)
	if err != nil {
		return nil, err
	}

	decrypted, err := cipher.Decrypt(data, password)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	defer zeroBytes(decrypted)

	var secrets map[string]string
	if err := json.Unmarshal(decrypted, &secrets); err != nil {
		return nil, fmt.Errorf("failed to unmarshal vault data: %w", err)
	}

	return secrets, nil
}

func (s *service) Save(secrets map[string]string, password []byte) error {
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

func (s *service) Exists() bool {
	return storage.VaultExists(s.vaultPath)
}
