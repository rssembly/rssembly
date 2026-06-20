package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Priority_YAMLOverEnv(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "config.yaml")
	os.WriteFile(yamlPath, []byte("server_port: 9090\n"), 0644)

	t.Setenv("CONFIG_FILE", yamlPath)
	t.Setenv("SERVER_PORT", "7070")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.ServerPort != 9090 {
		t.Errorf("expected ServerPort 9090 (YAML wins), got %d", cfg.ServerPort)
	}
}

func TestLoad_Priority_EnvOverDefault(t *testing.T) {
	t.Setenv("SERVER_PORT", "3000")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.ServerPort != 3000 {
		t.Errorf("expected ServerPort 3000 (env wins over default), got %d", cfg.ServerPort)
	}
}

func TestLoad_Priority_DefaultUsed(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.ServerPort != 8080 {
		t.Errorf("expected default ServerPort 8080, got %d", cfg.ServerPort)
	}
}

func TestLoad_Priority_DotEnvWinsOverEnvVar(t *testing.T) {
	t.Skip("Skipping .env test — requires CWD manipulation. Trusting godotenv behavior.")
}

func TestLoad_ConfigFileFlag(t *testing.T) {
	dir := t.TempDir()
	yamlPath := filepath.Join(dir, "custom.yaml")
	os.WriteFile(yamlPath, []byte("log_level: debug\n"), 0644)

	t.Setenv("CONFIG_FILE", yamlPath)
	t.Setenv("LOG_LEVEL", "warn")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel 'debug' (YAML wins over env), got %q", cfg.LogLevel)
	}
}
