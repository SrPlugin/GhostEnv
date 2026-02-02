package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
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
	ErrVaultIntegrity     = errors.New("vault integrity check failed: file may be corrupted or tampered")
)

func zeroBytes(b []byte) {
	for i := range b {
		b[i] = 0
	}
}

func Encrypt(plaintext, password []byte) ([]byte, error) {
	salt := make([]byte, config.SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEncryptionFailed, err)
	}

	key := DeriveKey(password, salt)
	defer zeroBytes(key)

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
	payload := make([]byte, len(salt)+len(ciphertext))
	copy(payload[:len(salt)], salt)
	copy(payload[len(salt):], ciphertext)

	mac := hmac.New(sha256.New, key)
	mac.Write(payload)
	sum := mac.Sum(nil)

	result := make([]byte, len(payload)+len(sum))
	copy(result, payload)
	copy(result[len(payload):], sum)
	return result, nil
}

func Decrypt(data, password []byte) ([]byte, error) {
	if len(data) < config.SaltSize {
		return nil, ErrInvalidVaultData
	}

	salt := data[:config.SaltSize]
	key := DeriveKey(password, salt)
	defer zeroBytes(key)

	if len(data) > config.SaltSize+config.HMACSize {
		payload := data[:len(data)-config.HMACSize]
		expectedMAC := data[len(data)-config.HMACSize:]
		mac := hmac.New(sha256.New, key)
		mac.Write(payload)
		sum := mac.Sum(nil)
		if hmac.Equal(sum, expectedMAC) {
			data = payload
		}
	}

	ciphertext := data[config.SaltSize:]
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
