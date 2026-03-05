package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const configFileName = "profiles.json"

type Profile struct {
	Name        string `json:"name"`
	Account     string `json:"account"`
	Project     string `json:"project"`
	Region      string `json:"region"`
	Zone        string `json:"zone"`
	Domain      string `json:"domain"`
	Description string `json:"description"`
}

type Config struct {
	ActiveProfile string    `json:"active_profile"`
	Profiles      []Profile `json:"profiles"`
}

func configDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".gcp-switcher")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

func configFilePath() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

func loadConfig() (*Config, error) {
	path, err := configFilePath()
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return &Config{Profiles: []Profile{}}, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	return &cfg, nil
}

func saveConfig(cfg *Config) error {
	path, err := configFilePath()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	return os.WriteFile(path, data, 0600)
}

func (c *Config) FindProfile(name string) (*Profile, int) {
	for i, p := range c.Profiles {
		if p.Name == name {
			return &c.Profiles[i], i
		}
	}
	return nil, -1
}
