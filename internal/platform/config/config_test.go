package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLoadFromReader_Defaults(t *testing.T) {
	cfg, err := LoadFromReader(nil)
	if err != nil {
		t.Fatalf("LoadFromReader() error = %v", err)
	}

	if cfg.Env != "development" {
		t.Fatalf("Env = %q, want %q", cfg.Env, "development")
	}
	if cfg.HTTP.Addr != ":8080" {
		t.Fatalf("HTTP.Addr = %q, want %q", cfg.HTTP.Addr, ":8080")
	}
	if cfg.HTTP.ReadTimeout != 5*time.Second {
		t.Fatalf("HTTP.ReadTimeout = %s, want %s", cfg.HTTP.ReadTimeout, 5*time.Second)
	}
	if cfg.HTTP.WriteTimeout != 10*time.Second {
		t.Fatalf("HTTP.WriteTimeout = %s, want %s", cfg.HTTP.WriteTimeout, 10*time.Second)
	}
	if cfg.HTTP.ShutdownTimeout != 10*time.Second {
		t.Fatalf("HTTP.ShutdownTimeout = %s, want %s", cfg.HTTP.ShutdownTimeout, 10*time.Second)
	}
	if cfg.Log.Level != "info" {
		t.Fatalf("Log.Level = %q, want %q", cfg.Log.Level, "info")
	}
	if cfg.MySQL.MaxOpenConns != 25 {
		t.Fatalf("MySQL.MaxOpenConns = %d, want %d", cfg.MySQL.MaxOpenConns, 25)
	}
	if cfg.MySQL.MaxIdleConns != 10 {
		t.Fatalf("MySQL.MaxIdleConns = %d, want %d", cfg.MySQL.MaxIdleConns, 10)
	}
}

func TestLoadFromReader_Overrides(t *testing.T) {
	cfg, err := LoadFromReader(strings.NewReader(`
env = "test"

[http]
addr = "127.0.0.1:9000"
read_timeout = "2s"
write_timeout = "3s"
shutdown_timeout = "4s"

[log]
level = "debug"

[mysql]
dsn = "user:pass@tcp(127.0.0.1:3306)/app?parseTime=true"
auto_migrate = true
max_open_conns = 7
max_idle_conns = 3
conn_max_lifetime = "30s"
conn_max_idle_time = "15s"
`))
	if err != nil {
		t.Fatalf("LoadFromReader() error = %v", err)
	}

	if cfg.Env != "test" {
		t.Fatalf("Env = %q, want %q", cfg.Env, "test")
	}
	if cfg.HTTP.Addr != "127.0.0.1:9000" {
		t.Fatalf("HTTP.Addr = %q, want %q", cfg.HTTP.Addr, "127.0.0.1:9000")
	}
	if cfg.HTTP.ReadTimeout != 2*time.Second {
		t.Fatalf("HTTP.ReadTimeout = %s, want %s", cfg.HTTP.ReadTimeout, 2*time.Second)
	}
	if cfg.HTTP.WriteTimeout != 3*time.Second {
		t.Fatalf("HTTP.WriteTimeout = %s, want %s", cfg.HTTP.WriteTimeout, 3*time.Second)
	}
	if cfg.HTTP.ShutdownTimeout != 4*time.Second {
		t.Fatalf("HTTP.ShutdownTimeout = %s, want %s", cfg.HTTP.ShutdownTimeout, 4*time.Second)
	}
	if cfg.Log.Level != "debug" {
		t.Fatalf("Log.Level = %q, want %q", cfg.Log.Level, "debug")
	}
	if cfg.MySQL.DSN == "" {
		t.Fatal("MySQL.DSN is empty, want configured DSN")
	}
	if !cfg.MySQL.AutoMigrate {
		t.Fatal("MySQL.AutoMigrate = false, want true")
	}
	if cfg.MySQL.MaxOpenConns != 7 {
		t.Fatalf("MySQL.MaxOpenConns = %d, want %d", cfg.MySQL.MaxOpenConns, 7)
	}
	if cfg.MySQL.MaxIdleConns != 3 {
		t.Fatalf("MySQL.MaxIdleConns = %d, want %d", cfg.MySQL.MaxIdleConns, 3)
	}
	if cfg.MySQL.ConnMaxLifetime != 30*time.Second {
		t.Fatalf("MySQL.ConnMaxLifetime = %s, want %s", cfg.MySQL.ConnMaxLifetime, 30*time.Second)
	}
	if cfg.MySQL.ConnMaxIdleTime != 15*time.Second {
		t.Fatalf("MySQL.ConnMaxIdleTime = %s, want %s", cfg.MySQL.ConnMaxIdleTime, 15*time.Second)
	}
}

func TestLoadFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.toml")
	if err := os.WriteFile(path, []byte(`
env = "file"

[http]
addr = ":9090"
`), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	cfg, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile() error = %v", err)
	}

	if cfg.Env != "file" {
		t.Fatalf("Env = %q, want %q", cfg.Env, "file")
	}
	if cfg.HTTP.Addr != ":9090" {
		t.Fatalf("HTTP.Addr = %q, want %q", cfg.HTTP.Addr, ":9090")
	}
}

func TestLoad_UsesConfigFileEnv(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.toml")
	if err := os.WriteFile(path, []byte(`
env = "from-env-path"

[log]
level = "warn"
`), 0o600); err != nil {
		t.Fatalf("write config file: %v", err)
	}
	t.Setenv(configFileEnv, path)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Env != "from-env-path" {
		t.Fatalf("Env = %q, want %q", cfg.Env, "from-env-path")
	}
	if cfg.Log.Level != "warn" {
		t.Fatalf("Log.Level = %q, want %q", cfg.Log.Level, "warn")
	}
}

func TestLoadFromReader_InvalidDuration(t *testing.T) {
	_, err := LoadFromReader(strings.NewReader(`
[http]
read_timeout = "not-a-duration"
`))
	if err == nil {
		t.Fatal("LoadFromReader() error = nil, want error")
	}
}

func TestLoadFromReader_UnknownField(t *testing.T) {
	_, err := LoadFromReader(strings.NewReader(`
[http]
read_timout = "5s"
`))
	if err == nil {
		t.Fatal("LoadFromReader() error = nil, want error")
	}
}
