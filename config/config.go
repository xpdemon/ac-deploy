package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type AppConfig struct {
	DockerContexts   []DockerContext `json:"docker_contexts"`
	DockerRegistries []string        `json:"docker_registries"`
}

type DockerContext struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Host        string `json:"host"`
	// ...
}

// In-memory configuration instance
var Cfg AppConfig

// getConfigPath returns the path ~/.xpdemon-deploy/config.json
func getConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	cfgDir := filepath.Join(home, ".xpdemon-deploy")
	err = os.MkdirAll(cfgDir, 0755)
	if err != nil {
		return "", err
	}
	return filepath.Join(cfgDir, "config.json"), nil
}

// LoadConfig loads the configuration from the file ~/.xpdemon-deploy/config.json
func LoadConfig() error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}
	// If the file does not exist, initialize an empty config
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		Cfg = AppConfig{
			DockerContexts:   []DockerContext{},
			DockerRegistries: []string{},
		}
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, &Cfg)
	if err != nil {
		return err
	}
	return nil
}

// SaveConfig saves the configuration to the file ~/.xpdemon-deploy/config.json
func SaveConfig() error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(Cfg, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	fmt.Println("Configuration saved to", path)
	return nil
}
