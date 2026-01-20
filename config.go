package main

import (
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// Config holds the application configuration
type Config struct {
	ListenAddr      string `yaml:"listen_addr"`
	AuthUser        string `yaml:"auth_user"`
	AuthPassword    string `yaml:"auth_password"`
	ReadOnly        bool   `yaml:"read_only"`
	RefreshInterval int    `yaml:"refresh_interval"`
	NFTBinary       string `yaml:"nft_binary"`
	TableFamily     string `yaml:"table_family"`
	TableName       string `yaml:"table_name"`
	ChainName       string `yaml:"chain_name"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		ListenAddr:      ":8080",
		AuthUser:        "",
		AuthPassword:    "",
		ReadOnly:        false,
		RefreshInterval: 5,
		NFTBinary:       "/usr/sbin/nft",
		TableFamily:     "inet",
		TableName:       "filter",
		ChainName:       "output",
	}
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig() (*Config, error) {
	cfg := DefaultConfig()

	// Try to load from config file
	configPaths := []string{
		"./config.yaml",
		"./config.yml",
		"/etc/nft-ui/config.yaml",
		"/etc/nft-ui/config.yml",
	}

	for _, path := range configPaths {
		if data, err := os.ReadFile(path); err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, err
			}
			break
		}
	}

	// Override with environment variables
	if v := os.Getenv("NFT_UI_LISTEN_ADDR"); v != "" {
		cfg.ListenAddr = v
	}
	if v := os.Getenv("NFT_UI_AUTH_USER"); v != "" {
		cfg.AuthUser = v
	}
	if v := os.Getenv("NFT_UI_AUTH_PASSWORD"); v != "" {
		cfg.AuthPassword = v
	}
	if v := os.Getenv("NFT_UI_READ_ONLY"); v != "" {
		cfg.ReadOnly = v == "true" || v == "1"
	}
	if v := os.Getenv("NFT_UI_REFRESH_INTERVAL"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			cfg.RefreshInterval = n
		}
	}
	if v := os.Getenv("NFT_UI_NFT_BINARY"); v != "" {
		cfg.NFTBinary = v
	}
	if v := os.Getenv("NFT_UI_TABLE_FAMILY"); v != "" {
		cfg.TableFamily = v
	}
	if v := os.Getenv("NFT_UI_TABLE_NAME"); v != "" {
		cfg.TableName = v
	}
	if v := os.Getenv("NFT_UI_CHAIN_NAME"); v != "" {
		cfg.ChainName = v
	}

	return cfg, nil
}

// AuthEnabled returns true if authentication is configured
func (c *Config) AuthEnabled() bool {
	return c.AuthUser != "" && c.AuthPassword != ""
}
