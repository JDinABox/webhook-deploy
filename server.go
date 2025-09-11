package githubwebhookdeploy

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Listen      string `yaml:"listen,omitempty"`
	Deployments map[string]struct {
		Secret   string   `yaml:"secret"`
		Commands []string `yaml:"commands"`
	} `yaml:"deployments"`
}

func loadConfig(config *Config, configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to read config file: %w", err)
		}
		log.Println("Warning: config file not found")
	}

	err = yaml.Unmarshal(data, config)
	return err
}

type Option func(*Config) error

func WithConfigFile(path string) Option {
	return func(c *Config) error {
		return loadConfig(c, path)
	}
}

// NewConfig creates a Config instance with optional settings
func NewConfig(opts ...Option) (*Config, error) {
	// Set defaults
	cfg := &Config{
		Listen: "127.0.0.1:8080",
	}

	// Apply options
	for _, opt := range opts {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func Start(opts ...Option) error {
	// Create config from options
	conf, err := NewConfig(opts...)
	if err != nil {
		return err
	}

	// Initialize HTTP router
	router := newApp(conf)

	// Start HTTP server
	server := &http.Server{
		Addr:    conf.Listen,
		Handler: router,
	}

	return server.ListenAndServe()
}
