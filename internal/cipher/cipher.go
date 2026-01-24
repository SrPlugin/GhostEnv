package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"

	"github.com/SrPlugin/GhostEnv/internal/config"
)

var (
	ErrInvalidVaultData   = errors.New("invalid vault data")
	ErrCiphertextTooShort = errors.New("ciphertext too short")
	ErrEncryptionFailed   = errors.New("encryption failed")
	ErrDecryptionFailed   = errors.New("decryption failed")
)

func Encrypt(plaintext []byte, password string) ([]byte, error) {
	salt := make([]byte, config.SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	key := DeriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	result := make([]byte, len(salt)+len(ciphertext))
	copy(result[:len(salt)], salt)
	copy(result[len(salt):], ciphertext)

	return result, nil
}

func Decrypt(data []byte, password string) ([]byte, error) {
	if len(data) < config.SaltSize {
		return nil, ErrInvalidVaultData
	}

	salt := data[:config.SaltSize]
	ciphertext := data[config.SaltSize:]
	key := DeriveKey(password, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, ErrCiphertextTooShort
	}

	nonce := ciphertext[:nonceSize]
	actualCiphertext := ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDecryptionFailed, err)
	}

	return plaintext, nil
}
