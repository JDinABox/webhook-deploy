package webhookdeploy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gopkg.in/yaml.v3"
)

type Deployments struct {
	Secret   string   `yaml:"secret"`
	Remote   Remote   `yaml:"remote"`
	Commands []string `yaml:"commands"`
}

type Remote struct {
	User       string `yaml:"user"`
	ServerIP   string `yaml:"server_ip"`
	PrivateKey string `yaml:"private_key"`
}

type WebInterface struct {
	Enabled  bool   `yaml:"enabled"`
	Listen   string `yaml:"listen,omitempty"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}
type Config struct {
	Listen       string                 `yaml:"listen,omitempty"`
	KnowHosts    string                 `yaml:"ssh-known-hosts"`
	WebInterface WebInterface           `yaml:"web-interface,omitempty"`
	Deployments  map[string]Deployments `yaml:"deployments"`
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
		Listen:    "127.0.0.1:8080",
		KnowHosts: "/etc/webhook-deploy/known_hosts",
		WebInterface: WebInterface{
			Enabled:  false,
			Listen:   "127.0.0.1:9080",
			Username: "",
			Password: "",
		},
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
	router, webInterface := newApp(conf)

	// Channel to listen for errors from either server
	serverErrChan := make(chan error, 1)

	// Start webhook server
	server := &http.Server{
		Addr:    conf.Listen,
		Handler: router,
	}
	go func() {
		log.Printf("Webhook server listening on %s", conf.Listen)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrChan <- fmt.Errorf("Webhook error: %w", err)
		}
	}()

	// Start web interface server if enabled
	var webServer *http.Server
	if conf.WebInterface.Enabled {
		webServer = &http.Server{
			Addr:    conf.WebInterface.Listen,
			Handler: webInterface,
		}
		go func() {
			log.Printf("Web interface server listening on %s", conf.WebInterface.Listen)
			if err := webServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				serverErrChan <- fmt.Errorf("Web interface error: %w", err)
			}
		}()
	}

	// Wait for a shutdown signal or a server error
	quitChan := make(chan os.Signal, 1)
	signal.Notify(quitChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrChan:
		log.Printf("A server has failed: %v", err)
	case sig := <-quitChan:
		log.Printf("Received signal %v. Shutting down...", sig)
	}

	// Create a context with a timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown the servers
	log.Println("Shutting down webhook server...")
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Webhook server shutdown failed: %v", err)
	}

	if webServer != nil {
		log.Println("Shutting down web interface server...")
		if err := webServer.Shutdown(ctx); err != nil {
			log.Printf("Web interface server shutdown failed: %v", err)
		}
	}

	log.Println("Shutdown complete.")
	return nil
}
