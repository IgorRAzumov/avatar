package config

import (
	"time"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	Server   ServerConfig
	Postgres PostgresConfig
	S3       S3Config
	RabbitMQ RabbitMQConfig
}

type ServerConfig struct {
	Addr        string `env:"SERVER_ADDR"`
	BaseURL     string `env:"BASE_URL"`
	MaxUploadMB int64  `env:"MAX_UPLOAD_MB"`
	// ReadTimeout/WriteTimeout are not env-configurable; set from defaults.
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type PostgresConfig struct {
	DSN string `env:"DATABASE_DSN"`
}

type S3Config struct {
	Endpoint  string `env:"S3_ENDPOINT"`
	AccessKey string `env:"S3_ACCESS_KEY"`
	SecretKey string `env:"S3_SECRET_KEY"`
	Bucket    string `env:"S3_BUCKET"`
	UseSSL    bool   `env:"S3_USE_SSL"`
	Region    string `env:"S3_REGION"`
}

type RabbitMQConfig struct {
	URL      string `env:"RABBITMQ_URL"`
	Enabled  bool   `env:"RABBITMQ_ENABLED"`
	Exchange string `env:"RABBITMQ_EXCHANGE"`
}

// Load builds the configuration from defaults and overrides any field whose
// corresponding environment variable is set. Malformed values are ignored and
// the default is kept.
func Load() *Config {
	cfg := defaults()
	// env.Parse only overrides fields whose env var is present, leaving the
	// prefilled defaults intact otherwise. On a malformed value we keep defaults.
	_ = env.Parse(cfg)
	return cfg
}

func defaults() *Config {
	return &Config{
		Server: ServerConfig{
			Addr:         DefaultServerAddr,
			BaseURL:      DefaultBaseURL,
			MaxUploadMB:  DefaultMaxUploadMB,
			ReadTimeout:  DefaultServerReadTimeout,
			WriteTimeout: DefaultServerWriteTimeout,
		},
		Postgres: PostgresConfig{
			DSN: DefaultPostgresDSN,
		},
		S3: S3Config{
			Endpoint:  DefaultS3Endpoint,
			AccessKey: DefaultS3AccessKey,
			SecretKey: DefaultS3SecretKey,
			Bucket:    DefaultS3Bucket,
			UseSSL:    false,
			Region:    DefaultS3Region,
		},
		RabbitMQ: RabbitMQConfig{
			URL:      DefaultRabbitMQURL,
			Enabled:  false,
			Exchange: DefaultRabbitMQExchange,
		},
	}
}

func (config *Config) MaxUploadBytes() int64 {
	return MaxUploadBytesFromMB(config.Server.MaxUploadMB)
}

func MaxUploadBytesFromMB(mb int64) int64 {
	return mb * bytesPerMB
}

func DefaultMaxUploadBytes() int64 {
	return MaxUploadBytesFromMB(DefaultMaxUploadMB)
}
