package config

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/docker/docker/api/types"
	"gopkg.in/yaml.v2"
)

type DBConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

type DockerConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Server   string `yaml:"server"`
}

type Config struct {
	MySQL  DBConfig     `yaml:"mysql"`
	Redis  DBConfig     `yaml:"redis"`
	Docker DockerConfig `yaml:"docker"`
}

var dockerAuthEncoded string

func LoadConfig(filePath string) (*Config, error) {
	config := &Config{}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	auth := types.AuthConfig{
		Username:      config.Docker.Username,
		Password:      config.Docker.Password,
		ServerAddress: config.Docker.Server,
	}

	jsonAuth, err := json.Marshal(auth)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal auth config: %w", err)
	}

	dockerAuthEncoded = base64.URLEncoding.EncodeToString(jsonAuth)

	return config, nil
}

func GetDockerAuth() string {
	return dockerAuthEncoded
}
