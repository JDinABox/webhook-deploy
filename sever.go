package githubwebhookdeploy

import (
	"fmt"
	"net/http"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	listen string
	app    *AppConfig
}
type AppConfig struct {
	Secret      string              `yaml:"secret"`
	Deployments map[string][]string `yaml:"deployments"`
}

func loadConfig(configPath string) (*AppConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AppConfig
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config data: %w", err)
	}

	return &config, nil
}

type Option func(*Config) error

// WithListenAddr sets the server's listen address
func WithListenAddr(addr string) Option {
	return func(c *Config) error {
		c.listen = addr
		return nil
	}
}
func WithConfigFile(path string) Option {
	return func(c *Config) error {
		appConf, err := loadConfig(path)
		c.app = appConf
		return err
	}
}

// NewConfig creates a Config instance with optional settings
func NewConfig(opts ...Option) (*Config, error) {
	// Set defaults
	cfg := &Config{
		listen: "127.0.0.1:8080",
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
	router := newApp(conf.app)

	// Start HTTP server
	server := &http.Server{
		Addr:    conf.listen,
		Handler: router,
	}

	return server.ListenAndServe()
}
