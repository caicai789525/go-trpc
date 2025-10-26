package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	SelfIP string `json:"self_ip"`
	PeerIP string `json:"peer_ip"`
	Port   string `json:"port"`
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile("config.json")
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
