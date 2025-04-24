package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Connection  struct {
		Host     string `json:"host"`
		Port     int    `json:"port"`
		Protocol string `json:"protocol"`
	} `json:"connection"`
	Database struct {
		Type        string `json:"type"`
		URL         string `json:"url"`
		Name        string `json:"name"`
		Collections struct {
			Models     string `json:"models"`
			Contexts   string `json:"contexts"`
			Executions string `json:"executions"`
			Data       string `json:"data"`
		} `json:"collections"`
	} `json:"database"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(file, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
