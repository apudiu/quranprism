// Package config loads typed runtime configuration from the process
// environment. Every other platform package depends on Config; nothing in
// internal/modules touches env vars directly. New configurable values
// should be added here with a sensible default for local dev and a
// `,required` tag where there is no safe fallback.
package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	App    AppConfig
	HTTP   HTTPConfig
	DB     DBConfig
	Redis  RedisConfig
	NATS   NATSConfig
	JWT    JWTConfig
	Mailer MailerConfig
}

// AppConfig — environment selector and the user-facing base URL used when
// constructing links inside emails (verify, password reset, etc).
type AppConfig struct {
	Env     string `env:"APP_ENV"      envDefault:"development"`
	BaseURL string `env:"APP_BASE_URL" envDefault:"http://localhost:3002"`
}

// IsProduction returns true when the app is running in the production
// environment. Drives logger format, cookie Secure flag, etc.
func (a AppConfig) IsProduction() bool { return a.Env == "production" }

type HTTPConfig struct {
	Port int `env:"PORT" envDefault:"3000"`
}

type DBConfig struct {
	URL string `env:"DATABASE_URL,required"`
}

type RedisConfig struct {
	URL string `env:"REDIS_URL,required"`
}

type NATSConfig struct {
	URL string `env:"NATS_URL,required"`
}

// JWTConfig holds the access-token signing secret(s) and TTLs.
//
// Secrets is a rotating list: position 0 is the current signing secret;
// later entries are accepted on verify but never produce new tokens. To
// rotate, prepend a new value and keep the old one around long enough for
// every issued token to expire (AccessTTL).
type JWTConfig struct {
	Secrets    []string      `env:"JWT_SECRETS,required" envSeparator:","`
	AccessTTL  time.Duration `env:"JWT_ACCESS_TTL"       envDefault:"15m"`
	RefreshTTL time.Duration `env:"JWT_REFRESH_TTL"      envDefault:"720h"`
}

// MailerConfig — SMTP connection details. Same envelope works for the
// dev Mailpit container (plain, no auth) and for AWS SES SMTP in prod
// (TLS + STARTTLS auth) by toggling User / Password / TLSPolicy.
type MailerConfig struct {
	SMTPHost  string `env:"SMTP_HOST"     envDefault:"localhost"`
	SMTPPort  int    `env:"SMTP_PORT"     envDefault:"1025"`
	User      string `env:"SMTP_USER"     envDefault:""`
	Password  string `env:"SMTP_PASSWORD" envDefault:""`
	TLSPolicy string `env:"SMTP_TLS"      envDefault:"none"` // none | opportunistic | mandatory
	From      string `env:"MAILER_FROM,required"`
}

// Load reads and validates the Config from the environment. Returned errors
// describe the first missing/invalid field; fail fast at startup.
func Load() (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("config: %w", err)
	}
	if len(cfg.JWT.Secrets) == 0 || cfg.JWT.Secrets[0] == "" {
		return nil, fmt.Errorf("config: JWT_SECRETS must contain at least one non-empty secret")
	}
	return &cfg, nil
}
