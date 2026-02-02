package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

var (
	current     *Config
	projectRoot string
	currentMu   sync.RWMutex
)

func SetCurrent(c *Config) {
	currentMu.Lock()
	defer currentMu.Unlock()
	current = c
}

func SetProjectRoot(root string) {
	currentMu.Lock()
	defer currentMu.Unlock()
	projectRoot = root
}

func Current() *Config {
	currentMu.RLock()
	defer currentMu.RUnlock()
	return current
}

func ProjectRoot() string {
	currentMu.RLock()
	defer currentMu.RUnlock()
	return projectRoot
}

const (
	ProjectConfigName = ".ghostenv.yml"
	GlobalConfigName  = "config.yml"
	GlobalConfigDir   = ".config/ghostenv"
	LegacyGlobalName  = ".ghostenv.yml"
)

func Default() *Config {
	c := &Config{}
	applyDefaults(c)
	return c
}

func Load(projectRoot string) (*Config, error) {
	global := loadFile(globalConfigPath())
	project := loadFile(projectConfigPath(projectRoot))

	merged := merge(global, project)
	applyDefaults(merged)
	if err := validateAndParse(merged); err != nil {
		return nil, err
	}
	return merged, nil
}

func globalConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	p := filepath.Join(home, GlobalConfigDir, GlobalConfigName)
	if _, err := os.Stat(p); err == nil {
		return p
	}
	legacy := filepath.Join(home, LegacyGlobalName)
	if _, err := os.Stat(legacy); err == nil {
		return legacy
	}
	return p
}

func projectConfigPath(projectRoot string) string {
	if projectRoot == "" {
		wd, _ := os.Getwd()
		projectRoot = wd
	}
	return filepath.Join(projectRoot, ProjectConfigName)
}

func loadFile(path string) *Config {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil
	}
	return &c
}

func merge(global, project *Config) *Config {
	out := &Config{}
	if global != nil {
		*out = *global
	}
	if project == nil {
		return out
	}
	if project.Project.Name != "" {
		out.Project.Name = project.Project.Name
	}
	if project.Project.Version != "" {
		out.Project.Version = project.Project.Version
	}
	if project.Project.DefaultEnv != "" {
		out.Project.DefaultEnv = project.Project.DefaultEnv
	}
	if project.Storage.VaultDir != "" {
		out.Storage.VaultDir = project.Storage.VaultDir
	}
	if project.Storage.RecursiveSearch {
		out.Storage.RecursiveSearch = true
	}
	if project.Storage.AutoBackup.Path != "" {
		out.Storage.AutoBackup = project.Storage.AutoBackup
	} else if project.Storage.AutoBackup.Enabled {
		out.Storage.AutoBackup.Enabled = true
		if project.Storage.AutoBackup.RetentionDays > 0 {
			out.Storage.AutoBackup.RetentionDays = project.Storage.AutoBackup.RetentionDays
		}
	}
	if len(project.Storage.Environments) > 0 {
		if out.Storage.Environments == nil {
			out.Storage.Environments = make(map[string]EnvEntry)
		}
		for k, v := range project.Storage.Environments {
			out.Storage.Environments[k] = v
		}
	}
	if project.Security.Argon2.Memory != "" {
		out.Security.Argon2 = project.Security.Argon2
	}
	if project.Security.Policy.MaxAuthAttempts > 0 {
		out.Security.Policy.MaxAuthAttempts = project.Security.Policy.MaxAuthAttempts
	}
	if project.Security.Policy.ForceMemoryZeroing {
		out.Security.Policy.ForceMemoryZeroing = true
	}
	if project.Security.Policy.DisallowPasswordFlagInProd {
		out.Security.Policy.DisallowPasswordFlagInProd = true
	}
	if project.Microservices.Inheritance.SharedVault != "" {
		out.Microservices.Inheritance = project.Microservices.Inheritance
	}
	if project.Microservices.Server.Host != "" || project.Microservices.Server.Port != 0 {
		out.Microservices.Server = project.Microservices.Server
	}
	if project.Microservices.Postgres.Enabled {
		out.Microservices.Postgres = project.Microservices.Postgres
	}
	if len(project.Scripts) > 0 {
		if out.Scripts == nil {
			out.Scripts = make(ScriptsConfig)
		}
		for k, v := range project.Scripts {
			out.Scripts[k] = v
		}
	}
	if project.Audit.FilePath != "" {
		out.Audit = project.Audit
	} else if project.Audit.Enabled {
		out.Audit.Enabled = true
		if project.Audit.Output != "" {
			out.Audit.Output = project.Audit.Output
		}
		if project.Audit.LogLevel != "" {
			out.Audit.LogLevel = project.Audit.LogLevel
		}
		out.Audit.MaskKeys = project.Audit.MaskKeys
	}
	if project.Export.DefaultFormat != "" {
		out.Export.DefaultFormat = project.Export.DefaultFormat
	}
	if project.Export.IncludeTimestamp {
		out.Export.IncludeTimestamp = true
	}
	return out
}

func applyDefaults(c *Config) {
	if c.Project.DefaultEnv == "" {
		c.Project.DefaultEnv = DefaultEnvironment
	}
	if c.Storage.VaultDir == "" {
		c.Storage.VaultDir = "." + string(filepath.Separator) + ProjectVaultDir
	}
	if c.Security.Argon2.Memory == "" {
		c.Security.Argon2.Memory = "64MB"
	}
	if c.Security.Argon2.Iterations == 0 {
		c.Security.Argon2.Iterations = Argon2Time
	}
	if c.Security.Argon2.Parallelism == 0 {
		c.Security.Argon2.Parallelism = Argon2Threads
	}
	if c.Security.Policy.MaxAuthAttempts == 0 {
		c.Security.Policy.MaxAuthAttempts = 5
	}
	if c.Audit.Output == "" {
		c.Audit.Output = "file"
	}
	if c.Audit.LogLevel == "" {
		c.Audit.LogLevel = "info"
	}
	if c.Audit.FilePath == "" && c.Audit.Output == "" {
		c.Audit.Enabled = true
	}
	if c.Export.DefaultFormat == "" {
		c.Export.DefaultFormat = "json"
	}
	if c.Microservices.Server.Port == 0 {
		c.Microservices.Server.Port = 8080
	}
	if c.Microservices.Postgres.Port == 0 {
		c.Microservices.Postgres.Port = 5432
	}
	if c.Microservices.Postgres.SSLMode == "" {
		c.Microservices.Postgres.SSLMode = "prefer"
	}
}

func validateAndParse(c *Config) error {
	if c.Security.Argon2.Memory != "" {
		if _, err := parseMemoryToKB(c.Security.Argon2.Memory); err != nil {
			return fmt.Errorf("config security.argon2.memory: %w", err)
		}
	}
	return nil
}

func parseMemoryToKB(s string) (uint32, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	var mult uint64 = 1
	if strings.HasSuffix(s, "MB") {
		mult = 1024
		s = strings.TrimSuffix(s, "MB")
	} else if strings.HasSuffix(s, "KB") {
		s = strings.TrimSuffix(s, "KB")
	} else if strings.HasSuffix(s, "GB") {
		mult = 1024 * 1024
		s = strings.TrimSuffix(s, "GB")
	}
	n, err := strconv.ParseUint(strings.TrimSpace(s), 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid memory %q: %w", s, err)
	}
	return uint32(n * mult), nil
}

func (c *Config) Argon2MemoryKB() uint32 {
	if c == nil || c.Security.Argon2.Memory == "" {
		return Argon2Memory / 1024
	}
	kb, err := parseMemoryToKB(c.Security.Argon2.Memory)
	if err != nil {
		return Argon2Memory / 1024
	}
	return kb
}

func (c *Config) Argon2Iterations() uint32 {
	if c == nil || c.Security.Argon2.Iterations == 0 {
		return Argon2Time
	}
	return c.Security.Argon2.Iterations
}

func (c *Config) Argon2Parallelism() uint8 {
	if c == nil || c.Security.Argon2.Parallelism == 0 {
		return Argon2Threads
	}
	return c.Security.Argon2.Parallelism
}
