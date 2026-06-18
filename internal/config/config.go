package config

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

// Config is the canonical configuration struct.
type Config struct {
	DatabaseURL         string        `env:"DATABASE_URL"           yaml:"database_url"`
	ServerHost          string        `env:"SERVER_HOST"            yaml:"server_host"`
	ServerPort          int           `env:"SERVER_PORT"            yaml:"server_port"            envDefault:"8080"`
	JWTPrivateKeyPath   string        `env:"JWT_PRIVATE_KEY_PATH"    yaml:"jwt_private_key_path"`
	JWTPublicKeyPath    string        `env:"JWT_PUBLIC_KEY_PATH"     yaml:"jwt_public_key_path"`
	CORSAllowedOrigins  []string      `env:"CORS_ALLOWED_ORIGINS"   yaml:"cors_allowed_origins"   envSeparator:","`
	RateLimitRequests   int           `env:"RATE_LIMIT_REQUESTS"    yaml:"rate_limit_requests"    envDefault:"100"`
	RateLimitWindow     time.Duration `env:"RATE_LIMIT_WINDOW"      yaml:"rate_limit_window"      envDefault:"1m"`
	DefaultPollInterval time.Duration `env:"DEFAULT_POLL_INTERVAL"  yaml:"default_poll_interval"  envDefault:"15m"`
	LogLevel            string        `env:"LOG_LEVEL"              yaml:"log_level"              envDefault:"info"`
	ConfigFile          string        `env:"CONFIG_FILE"            yaml:"-"`

	// Computed.
	Addr string `yaml:"-"`
}

// Load builds a Config by applying sources in priority order:
//
//	1. YAML config file (from -config flag or CONFIG_FILE env) — highest
//	2. .env file from CWD
//	3. Shell environment variables
//	4. Default values (envDefault struct tags) — lowest
//
// Each source fills fields not set by a higher-priority source.
func Load() (*Config, error) {
	// --- Bootstrap: minimal env parse to find CONFIG_FILE ---
	boot := &Config{}
	if err := env.Parse(boot); err != nil {
		return nil, fmt.Errorf("bootstrap: %w", err)
	}

	// --- CLI -config flag overrides CONFIG_FILE env ---
	if !flag.Parsed() {
		flag.StringVar(&boot.ConfigFile, "config", boot.ConfigFile, "path to YAML config file")
		flag.Parse()
	}

	// --- Parse each source independently ---

	// Source 1: YAML file (highest priority).
	yamlCfg := &Config{}
	if boot.ConfigFile != "" {
		data, err := os.ReadFile(boot.ConfigFile)
		if err != nil {
			return nil, fmt.Errorf("config file %s: %w", boot.ConfigFile, err)
		}
		if err := yaml.Unmarshal(data, yamlCfg); err != nil {
			return nil, fmt.Errorf("config file %s: %w", boot.ConfigFile, err)
		}
	}

	// Source 2: .env file. We load with godotenv's Overload so .env beats
	// existing shell env vars. Then parse env to pick them up.
	_ = godotenv.Load()
	_ = godotenv.Overload() // second call overwrites shell env with .env values

	// After Overload, the OS env now has .env values beating shell values.
	// Parse everything.
	envCfg := &Config{}
	if err := env.Parse(envCfg); err != nil {
		return nil, fmt.Errorf("env parse: %w", err)
	}

	// --- Merge with priority ---
	// envCfg already has: defaults + shell env + .env (in order of priority).
	// YAML beats everything, so overlay YAML on top.
	mergeNonZero(envCfg, yamlCfg)
	cfg := envCfg

	cfg.Addr = fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	cfg.ConfigFile = boot.ConfigFile
	return cfg, nil
}

// mergeNonZero copies non-zero exported fields from src into dst.
// src is higher priority — its non-zero values overwrite dst.
func mergeNonZero(dst, src *Config) {
	sv := reflect.ValueOf(src).Elem()
	dv := reflect.ValueOf(dst).Elem()
	st := sv.Type()

	for i := 0; i < st.NumField(); i++ {
		f := st.Field(i)
		if !f.IsExported() || f.Tag.Get("yaml") == "-" {
			continue
		}
		sf := sv.Field(i)
		if sf.IsZero() {
			continue
		}
		df := dv.Field(i)
		if df.CanSet() {
			df.Set(sf)
		}
	}
}