package config

const (
	SaltSize      = 16
	KeySize       = 32
	NonceSize     = 12
	VaultFileName = ".ghostenv.gev"
	VaultFilePerm = 0600
)

const (
	Argon2Time    = 1
	Argon2Memory  = 64 * 1024
	Argon2Threads = 4
)

const (
	DefaultEnvironment = "dev"
	ProjectVaultDir    = ".ghostenv"
)
