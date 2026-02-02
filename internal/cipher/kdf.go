package cipher

import (
	"github.com/SrPlugin/GhostEnv/internal/config"
	"golang.org/x/crypto/argon2"
)

func DeriveKey(password, salt []byte) []byte {
	cfg := config.Current()
	var time uint32 = config.Argon2Time
	var memory uint32 = config.Argon2Memory // already in KB (64*1024 = 64 MiB)
	var threads uint8 = config.Argon2Threads
	if cfg != nil {
		time = cfg.Argon2Iterations()
		memory = cfg.Argon2MemoryKB()
		threads = cfg.Argon2Parallelism()
	}
	return argon2.IDKey(
		password,
		salt,
		time,
		memory,
		threads,
		config.KeySize,
	)
}
