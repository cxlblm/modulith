package config

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pelletier/go-toml/v2"

	"modular_monolith/internal/platform/httpserver"
	"modular_monolith/internal/platform/logging"
	"modular_monolith/internal/platform/mysql"
)

const (
	configFileEnv     = "CONFIG_FILE"
	defaultConfigFile = "config/local.toml"
)

type Config struct {
	Env   string
	HTTP  httpserver.Config
	Log   logging.Config
	MySQL mysql.Config
}

type fileConfig struct {
	Env   string          `toml:"env"`
	HTTP  fileHTTPConfig  `toml:"http"`
	Log   fileLogConfig   `toml:"log"`
	MySQL fileMySQLConfig `toml:"mysql"`
}

type fileHTTPConfig struct {
	Addr            string `toml:"addr"`
	ReadTimeout     string `toml:"read_timeout"`
	WriteTimeout    string `toml:"write_timeout"`
	ShutdownTimeout string `toml:"shutdown_timeout"`
}

type fileLogConfig struct {
	Level string `toml:"level"`
}

type fileMySQLConfig struct {
	DSN             string `toml:"dsn"`
	AutoMigrate     *bool  `toml:"auto_migrate"`
	MaxOpenConns    *int   `toml:"max_open_conns"`
	MaxIdleConns    *int   `toml:"max_idle_conns"`
	ConnMaxLifetime string `toml:"conn_max_lifetime"`
	ConnMaxIdleTime string `toml:"conn_max_idle_time"`
}

func Load() (Config, error) {
	path := os.Getenv(configFileEnv)
	if path == "" {
		path = defaultConfigFile
	}
	return LoadFile(path)
}

func LoadFile(path string) (Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return Config{}, fmt.Errorf("open config file %q: %w", path, err)
	}
	defer file.Close()

	cfg, err := LoadFromReader(file)
	if err != nil {
		return Config{}, fmt.Errorf("load config file %q: %w", path, err)
	}
	return cfg, nil
}

func LoadFromReader(reader io.Reader) (Config, error) {
	cfg := defaultConfig()
	if reader == nil {
		return cfg, nil
	}

	var fileCfg fileConfig
	if err := toml.NewDecoder(reader).DisallowUnknownFields().Decode(&fileCfg); err != nil {
		return Config{}, fmt.Errorf("decode toml config: %w", err)
	}

	if fileCfg.Env != "" {
		cfg.Env = fileCfg.Env
	}
	if fileCfg.HTTP.Addr != "" {
		cfg.HTTP.Addr = fileCfg.HTTP.Addr
	}
	if fileCfg.HTTP.ReadTimeout != "" {
		duration, err := parseDuration("http.read_timeout", fileCfg.HTTP.ReadTimeout)
		if err != nil {
			return Config{}, err
		}
		cfg.HTTP.ReadTimeout = duration
	}
	if fileCfg.HTTP.WriteTimeout != "" {
		duration, err := parseDuration("http.write_timeout", fileCfg.HTTP.WriteTimeout)
		if err != nil {
			return Config{}, err
		}
		cfg.HTTP.WriteTimeout = duration
	}
	if fileCfg.HTTP.ShutdownTimeout != "" {
		duration, err := parseDuration("http.shutdown_timeout", fileCfg.HTTP.ShutdownTimeout)
		if err != nil {
			return Config{}, err
		}
		cfg.HTTP.ShutdownTimeout = duration
	}

	if fileCfg.Log.Level != "" {
		cfg.Log.Level = fileCfg.Log.Level
	}

	if fileCfg.MySQL.DSN != "" {
		cfg.MySQL.DSN = fileCfg.MySQL.DSN
	}
	if fileCfg.MySQL.AutoMigrate != nil {
		cfg.MySQL.AutoMigrate = *fileCfg.MySQL.AutoMigrate
	}
	if fileCfg.MySQL.MaxOpenConns != nil {
		cfg.MySQL.MaxOpenConns = *fileCfg.MySQL.MaxOpenConns
	}
	if fileCfg.MySQL.MaxIdleConns != nil {
		cfg.MySQL.MaxIdleConns = *fileCfg.MySQL.MaxIdleConns
	}
	if fileCfg.MySQL.ConnMaxLifetime != "" {
		duration, err := parseDuration("mysql.conn_max_lifetime", fileCfg.MySQL.ConnMaxLifetime)
		if err != nil {
			return Config{}, err
		}
		cfg.MySQL.ConnMaxLifetime = duration
	}
	if fileCfg.MySQL.ConnMaxIdleTime != "" {
		duration, err := parseDuration("mysql.conn_max_idle_time", fileCfg.MySQL.ConnMaxIdleTime)
		if err != nil {
			return Config{}, err
		}
		cfg.MySQL.ConnMaxIdleTime = duration
	}

	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		Env: "development",
		HTTP: httpserver.Config{
			Addr:            ":8080",
			ReadTimeout:     5 * time.Second,
			WriteTimeout:    10 * time.Second,
			ShutdownTimeout: 10 * time.Second,
		},
		Log: logging.Config{
			Level: "info",
		},
		MySQL: mysql.Config{
			DSN:             "",
			AutoMigrate:     false,
			MaxOpenConns:    25,
			MaxIdleConns:    10,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: time.Minute,
		},
	}
}

func parseDuration(name string, value string) (time.Duration, error) {
	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", name, err)
	}
	return duration, nil
}
