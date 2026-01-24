package cipher

import (
	"github.com/SrPlugin/GhostEnv/internal/config"
	"golang.org/x/crypto/argon2"
)

func DeriveKey(password string, salt []byte) []byte {
	return argon2.IDKey(
		[]byte(password),
		salt,
		config.Argon2Time,
		config.Argon2Memory,
		config.Argon2Threads,
		config.KeySize,
	)
}
