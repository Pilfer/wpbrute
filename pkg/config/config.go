package config

import (
	"io/ioutil"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	DBFile        string         `yaml:"db" json:"db"`
	DBConfig      DatabaseConfig `yaml:"database_config" json:"database_config"`
	Debug         bool           `yaml:"debug" json:"debug"`
	WPBruteConfig WPBruteConfig  `yaml:"wp_brute_config" json:"wp_brute_config"`
}

type DatabaseConfig struct {
	Database string `yaml:"database" json:"database"`
	Username string `yaml:"username" json:"username"`
	Host     string `yaml:"host" json:"host"`
	Password string `yaml:"password" json:"password"`
	Port     int32  `yaml:"port" json:"port"`
}

type WPBruteConfig struct {
	HTTPTimeout    int64 `yaml:"http_timeout" json:"http_timeout"`
	MaxConcurrency int   `yaml:"max_concurrency" json:"max_concurrency"`
	WorkerCount    int   `yaml:"worker_count" json:"worker_count"`
}

func NewFromYaml(path string) *Config {
	// Check if path provided exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Fatalf("Configuration file `%s` not found", path)
	}

	// Read configuration file
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}

	// Unmarshal configuration from file
	c := &Config{}
	err = yaml.Unmarshal([]byte(data), c)
	if err != nil {
		log.Fatalf("Could not load configuration file: %s", err)
	}

	if c.WPBruteConfig.WorkerCount == 0 {
		c.WPBruteConfig.WorkerCount = 1000
	}

	return c
}
