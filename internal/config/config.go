package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v2"
)

// Config represents the labman configuration file structure
type Config struct {
	Defaults Defaults            `yaml:"defaults"`
	Hosts    map[string]Host     `yaml:"hosts"`
	Groups   map[string][]string `yaml:"groups,omitempty"`
}

// Defaults holds default values for connections
type Defaults struct {
	Username          string        `yaml:"username,omitempty"`
	Port              int           `yaml:"port,omitempty"`
	ConnectionTimeout time.Duration `yaml:"connection_timeout,omitempty"`
}

// Host represents a single host configuration
type Host struct {
	Host     string `yaml:"host"`
	Username string `yaml:"username,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	KeyFile  string `yaml:"key_file,omitempty"`
}

var configCache *Config

// Load reads the configuration file from the standard locations
func Load() (*Config, error) {
	if configCache != nil {
		return configCache, nil
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return nil, fmt.Errorf("get config path: %w", err)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return empty config if file doesn't exist
		return &Config{
			Defaults: Defaults{
				Port:              22,
				ConnectionTimeout: 30 * time.Second,
			},
			Hosts:  make(map[string]Host),
			Groups: make(map[string][]string),
		}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	// Set defaults if not specified
	if cfg.Defaults.Port == 0 {
		cfg.Defaults.Port = 22
	}
	if cfg.Defaults.ConnectionTimeout == 0 {
		cfg.Defaults.ConnectionTimeout = 30 * time.Second
	}
	if cfg.Hosts == nil {
		cfg.Hosts = make(map[string]Host)
	}
	if cfg.Groups == nil {
		cfg.Groups = make(map[string][]string)
	}

	configCache = &cfg
	return &cfg, nil
}

// LoadWithPath loads configuration from a specific file path
func LoadWithPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	// Set defaults
	if cfg.Defaults.Port == 0 {
		cfg.Defaults.Port = 22
	}
	if cfg.Defaults.ConnectionTimeout == 0 {
		cfg.Defaults.ConnectionTimeout = 30 * time.Second
	}
	if cfg.Hosts == nil {
		cfg.Hosts = make(map[string]Host)
	}
	if cfg.Groups == nil {
		cfg.Groups = make(map[string][]string)
	}

	return &cfg, nil
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() (string, error) {
	// Check XDG_CONFIG_HOME first
	if xdgConfig := os.Getenv("XDG_CONFIG_HOME"); xdgConfig != "" {
		return filepath.Join(xdgConfig, "labman", "config.yaml"), nil
	}

	// Fall back to ~/.labman/config.yaml
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}
	return filepath.Join(homeDir, ".labman", "config.yaml"), nil
}

// ResolveHost resolves a host identifier (alias or IP) to connection details
func (c *Config) ResolveHost(identifier string) (host, username string, port int, keyFile string, err error) {
	// Check if identifier is a configured host alias
	if hostConfig, exists := c.Hosts[identifier]; exists {
		username = hostConfig.Username
		if username == "" {
			username = c.Defaults.Username
		}

		port = hostConfig.Port
		if port == 0 {
			port = c.Defaults.Port
		}

		return hostConfig.Host, username, port, hostConfig.KeyFile, nil
	}

	// Treat as direct IP/hostname
	username = c.Defaults.Username
	port = c.Defaults.Port
	if port == 0 {
		port = 22
	}

	return identifier, username, port, "", nil
}

// Validate checks the configuration for errors
func (c *Config) Validate() error {
	// Validate hosts
	for alias, host := range c.Hosts {
		if host.Host == "" {
			return fmt.Errorf("host '%s' missing required 'host' field", alias)
		}
		if host.Port != 0 && (host.Port < 1 || host.Port > 65535) {
			return fmt.Errorf("host '%s' has invalid port: %d", alias, host.Port)
		}
		if host.KeyFile != "" {
			expandedPath := os.ExpandEnv(host.KeyFile)
			if _, err := os.Stat(expandedPath); err != nil {
				return fmt.Errorf("host '%s' key file not found: %s", alias, expandedPath)
			}
		}
	}

	// Validate groups
	for groupName, members := range c.Groups {
		for _, member := range members {
			if _, exists := c.Hosts[member]; !exists {
				return fmt.Errorf("group '%s' references undefined host '%s'", groupName, member)
			}
		}
	}

	return nil
}

// CreateDefaultConfig creates a sample configuration file
func CreateDefaultConfig() error {
	configPath, err := GetConfigPath()
	if err != nil {
		return fmt.Errorf("get config path: %w", err)
	}

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		return fmt.Errorf("config file already exists at %s", configPath)
	}

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	// Sample configuration
	sampleConfig := `# LabMan Configuration File
# See: https://github.com/tinotenda-alfaneti/labman-cli

defaults:
  username: ubuntu
  port: 22
  connection_timeout: 30s

hosts:
  homelab-prod:
    host: 192.168.1.10
    username: admin
    # port: 22  # optional, defaults to 22
    # key_file: ~/.ssh/id_homelab  # optional, for SSH key auth

  k8s-master:
    host: 192.168.1.20
    username: ubuntu

  # Add more hosts as needed
  # my-server:
  #   host: example.com
  #   port: 2222

groups:
  production:
    - homelab-prod
    - k8s-master
  
  # staging:
  #   - staging-01
  #   - staging-02
`

	file, err := os.OpenFile(configPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("create config file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(sampleConfig); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}

	return nil
}

// ClearCache clears the cached configuration (useful for testing)
func ClearCache() {
	configCache = nil
}
