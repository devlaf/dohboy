package doh

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Host        string `yaml:"host" default:"127.0.0.1"`
		Port        string `yaml:"port" default:"8080"`
		TLSCertPath string `yaml:"tls_cert_filepath`
		TLSKeyPath  string `yaml:"tls_key_filepath`
		Timeout     struct {
			Shutdown time.Duration `yaml:"shutdown" default:"30"`
			Write    time.Duration `yaml:"write" default:"10"`
			Read     time.Duration `yaml:"read" default:"15"`
			Idle     time.Duration `yaml:"idle" default:"5"`
		} `yaml:"timeout_sec"`
		IPRateLimit struct {
			Enabled              bool   `yaml:"enabled" default:"true"`
			KeyWhitelist         string `yaml:"key_whitelist"`
			RecoverXTokensPerSec int    `yaml:"recover_x_tokens_per_sec" default:"5"`
			MaxTokens            int    `yaml:"max_tokens" default:"25"`
		} `yaml:"ip_rate_limit"`
	} `yaml:"server"`
}

func parseConfigFile(filepath string) (*Config, error) {

	config := &Config{}
	defaults.Set(config)

	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	d := yaml.NewDecoder(file)
	if err := d.Decode(&config); err != nil {
		return nil, err
	}

	return config, nil
}

func ensureFileExists(filepath string) error {
	if fileInfo, err := os.Stat(filepath); err != nil {
		return err
	} else if fileInfo.IsDir() {
		return fmt.Errorf("Provided path [%v] is not a file.", filepath)
	}
	return nil
}

func validateConfig(config *Config) error {
	if (config.Server.TLSCertPath == "") != (config.Server.TLSKeyPath == "") {
		return fmt.Errorf("Both a cert path and a key path must be provided to configure TLS.")
	}

	if config.Server.TLSCertPath != "" {
		if err := ensureFileExists(config.Server.TLSCertPath); err != nil {
			return err
		}
	}

	if config.Server.TLSKeyPath != "" {
		if err := ensureFileExists(config.Server.TLSKeyPath); err != nil {
			return err
		}
	}

	return nil
}

func FetchConfig() (*Config, error) {
	var configPath string

	flag.StringVar(&configPath, "config", "./config.yml", "path to config file")
	flag.Parse()

	cfg, err := parseConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	if invalid := validateConfig(cfg); invalid != nil {
		return nil, err
	}

	return cfg, nil
}
