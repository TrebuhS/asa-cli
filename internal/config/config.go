package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	ClientID       string `mapstructure:"client_id"`
	TeamID         string `mapstructure:"team_id"`
	KeyID          string `mapstructure:"key_id"`
	OrgID          string `mapstructure:"org_id"`
	PrivateKeyPath string `mapstructure:"private_key_path"`
}

var (
	configDir  string
	cfgProfile string
)

func SetProfile(profile string) {
	cfgProfile = profile
}

func ConfigDir() string {
	if configDir != "" {
		return configDir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot determine home directory: %v\n", err)
		os.Exit(3)
	}
	configDir = filepath.Join(home, ".asa-cli")
	return configDir
}

func Load() (*Config, error) {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("cannot create config directory: %w", err)
	}

	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(dir)

	// Environment variable overrides
	v.SetEnvPrefix("ASA")
	v.AutomaticEnv()
	v.BindEnv("client_id")
	v.BindEnv("team_id")
	v.BindEnv("key_id")
	v.BindEnv("org_id")
	v.BindEnv("private_key_path")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}

	cfg := &Config{}

	if cfgProfile != "" && cfgProfile != "default" {
		sub := v.Sub("profiles." + cfgProfile)
		if sub == nil {
			return nil, fmt.Errorf("profile %q not found in config", cfgProfile)
		}
		if err := sub.Unmarshal(cfg); err != nil {
			return nil, fmt.Errorf("error parsing profile %q: %w", cfgProfile, err)
		}
	} else {
		if err := v.Unmarshal(cfg); err != nil {
			return nil, fmt.Errorf("error parsing config: %w", err)
		}
	}

	// Env vars always override
	if val := os.Getenv("ASA_CLIENT_ID"); val != "" {
		cfg.ClientID = val
	}
	if val := os.Getenv("ASA_TEAM_ID"); val != "" {
		cfg.TeamID = val
	}
	if val := os.Getenv("ASA_KEY_ID"); val != "" {
		cfg.KeyID = val
	}
	if val := os.Getenv("ASA_ORG_ID"); val != "" {
		cfg.OrgID = val
	}
	if val := os.Getenv("ASA_PRIVATE_KEY_PATH"); val != "" {
		cfg.PrivateKeyPath = val
	}

	return cfg, nil
}

func Save(cfg *Config, profile string) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	configPath := filepath.Join(dir, "config.yaml")

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	// Try to read existing config
	_ = v.ReadInConfig()

	if profile != "" && profile != "default" {
		v.Set("profiles."+profile+".client_id", cfg.ClientID)
		v.Set("profiles."+profile+".team_id", cfg.TeamID)
		v.Set("profiles."+profile+".key_id", cfg.KeyID)
		v.Set("profiles."+profile+".org_id", cfg.OrgID)
		v.Set("profiles."+profile+".private_key_path", cfg.PrivateKeyPath)
	} else {
		v.Set("client_id", cfg.ClientID)
		v.Set("team_id", cfg.TeamID)
		v.Set("key_id", cfg.KeyID)
		v.Set("org_id", cfg.OrgID)
		v.Set("private_key_path", cfg.PrivateKeyPath)
	}

	if err := v.WriteConfigAs(configPath); err != nil {
		return fmt.Errorf("error writing config: %w", err)
	}

	// Ensure restrictive permissions
	return os.Chmod(configPath, 0600)
}
