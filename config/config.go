// ABOUTME: Loads and exposes typed application configuration from config.json.
// ABOUTME: config.json is the single source of truth — no env var fallback.

package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	DatabaseURL  string `json:"database_url"`
	JWTSecret    string `json:"jwt_secret"`
	DeviceSecret string `json:"device_secret"`
	Port         string `json:"port"`
	TokenTTL     string `json:"token_ttl"`
}

func Load(path string) (*Config, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg Config
	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
