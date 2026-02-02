package config

type Config struct {
	Project       ProjectConfig       `yaml:"project"`
	Storage       StorageConfig      `yaml:"storage"`
	Security      SecurityConfig     `yaml:"security"`
	Microservices MicroservicesConfig `yaml:"microservices"`
	Scripts       ScriptsConfig      `yaml:"scripts"`
	Audit         AuditConfig        `yaml:"audit"`
	Export        ExportConfig       `yaml:"export"`
}

type ProjectConfig struct {
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	DefaultEnv string `yaml:"default_env"`
}

type StorageConfig struct {
	VaultDir        string                    `yaml:"vault_dir"`
	RecursiveSearch bool                      `yaml:"recursive_search"`
	AutoBackup      AutoBackupConfig          `yaml:"auto_backup"`
	Environments    map[string]EnvEntry       `yaml:"environments"`
}

type EnvEntry struct {
	Dir   string `yaml:"dir,omitempty"`
	Vault string `yaml:"vault,omitempty"`
}

type AutoBackupConfig struct {
	Enabled       bool   `yaml:"enabled"`
	RetentionDays int    `yaml:"retention_days"`
	Path          string `yaml:"path"`
}

type SecurityConfig struct {
	Argon2 Argon2Config `yaml:"argon2"`
	Policy PolicyConfig `yaml:"policy"`
}

type Argon2Config struct {
	Memory      string `yaml:"memory"`
	Iterations  uint32 `yaml:"iterations"`
	Parallelism uint8  `yaml:"parallelism"`
}

type PolicyConfig struct {
	MaxAuthAttempts            int  `yaml:"max_auth_attempts"`
	ForceMemoryZeroing         bool `yaml:"force_memory_zeroing"`
	DisallowPasswordFlagInProd bool `yaml:"disallow_password_flag_in_prod"`
}

type MicroservicesConfig struct {
	Inheritance MicroInheritanceConfig `yaml:"inheritance"`
	Server      MicroServerConfig      `yaml:"server"`
	Postgres    PostgresConfig         `yaml:"postgres"`
}

type MicroInheritanceConfig struct {
	Enabled     bool   `yaml:"enabled"`
	SharedVault string `yaml:"shared_vault"`
}

type MicroServerConfig struct {
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
	UseTLS bool   `yaml:"use_tls"`
}

type PostgresConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	UserKey  string `yaml:"user_key"`
	PassKey  string `yaml:"pass_key"`
	SSLMode  string `yaml:"ssl_mode"`
}

type ScriptsConfig map[string]string

type AuditConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Output   string `yaml:"output"`
	FilePath string `yaml:"file_path"`
	LogLevel string `yaml:"log_level"`
	MaskKeys bool   `yaml:"mask_keys"`
}

type ExportConfig struct {
	DefaultFormat    string `yaml:"default_format"`
	IncludeTimestamp bool   `yaml:"include_timestamp"`
}
