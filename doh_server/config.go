package doh

import (
	"flag"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Host        string `yaml:"host"`
		Port        string `yaml:"port"`
		TLSCertPath string `yaml:"tls_cert_filepath`
		TLSKeyPath  string `yaml:"tls_key_filepath`
		Timeout     struct {
			Shutdown time.Duration `yaml:"shutdown"`
			Write    time.Duration `yaml:"write"`
			Read     time.Duration `yaml:"read"`
			Idle     time.Duration `yaml:"idle"`
		} `yaml:"timeout_sec"`
	} `yaml:"server"`
}

func parseConfigFile(filepath string) (*Config, error) {

	config := &Config{}

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
