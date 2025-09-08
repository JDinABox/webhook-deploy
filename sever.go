package githubwebhookdeploy

import "net/http"

type Config struct {
	listen string
}

type Option func(*Config) error

// WithListenAddr sets the server's listen address
func WithListenAddr(addr string) Option {
	return func(c *Config) error {
		c.listen = addr
		return nil
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
	router := newApp()

	// Start HTTP server
	server := &http.Server{
		Addr:    conf.listen,
		Handler: router,
	}

	return server.ListenAndServe()
}
