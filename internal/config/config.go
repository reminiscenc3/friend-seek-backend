package config

import (
	"fmt"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
)

const (
	EnvProduction = "production"
	EnvDebug      = "debug"
)

type Config struct {
	Env        string     `env:"ENV" env-default:"debug"`
	HTTPServer HTTPServer `env-prefix:"HTTP_"`
	Database   Database   `env-prefix:"DB_"`
	Game       Game       `env-prefix:"GAME_"`
}

type HTTPServer struct {
	Address         string        `env:"ADDRESS" env-default:":8080"`
	ReadTimeout     time.Duration `env:"READ_TIMEOUT" env-default:"5s"`
	WriteTimeout    time.Duration `env:"WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
	DefaultTimeout  time.Duration `env:"DEFAULT_TIMEOUT" env-default:"5s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"10s"`
}

type Database struct {
	Path        string        `env:"PATH" env-default:"friend_seek.db"`
	BusyTimeout time.Duration `env:"BUSY_TIMEOUT" env-default:"5s"`
	MaxConns    int           `env:"MAX_CONNS" env-default:"4"`
}

type Game struct {
	LocationTTL time.Duration `env:"LOCATION_TTL" env-default:"3m"`
}

func New() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("read env: %w", err)
	}

	if cfg.Env != EnvProduction && cfg.Env != EnvDebug {
		return nil, fmt.Errorf("ENV must be either %s or %s, got %q", EnvProduction, EnvDebug, cfg.Env)
	}

	return &cfg, nil
}

func (d Database) GetSQLiteDSN() string {
	return fmt.Sprintf(
		"file:%s?_pragma=journal_mode(WAL)&_pragma=busy_timeout(%d)&_pragma=foreign_keys(on)",
		d.Path, d.BusyTimeout.Milliseconds(),
	)
}
