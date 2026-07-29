package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, ":8080", cfg.Server.Addr)
	assert.Equal(t, "http://localhost:8080", cfg.Server.BaseURL)
	assert.Equal(t, int64(10), cfg.Server.MaxUploadMB)
	assert.Equal(t, DefaultMaxUploadBytes(), cfg.MaxUploadBytes())
	assert.Equal(t, DefaultPostgresDSN, cfg.Postgres.DSN)
	assert.Equal(t, DefaultS3Endpoint, cfg.S3.Endpoint)
	assert.Equal(t, DefaultS3Bucket, cfg.S3.Bucket)
	assert.False(t, cfg.S3.UseSSL)
	assert.False(t, cfg.RabbitMQ.Enabled)
	assert.Equal(t, DefaultRabbitMQExchange, cfg.RabbitMQ.Exchange)
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("SERVER_ADDR", ":9090")
	t.Setenv("BASE_URL", "https://avatars.example.com")
	t.Setenv("MAX_UPLOAD_MB", "5")
	t.Setenv("DATABASE_DSN", "postgres://custom/db")
	t.Setenv("S3_ENDPOINT", "s3.example.com")
	t.Setenv("S3_ACCESS_KEY", "key")
	t.Setenv("S3_SECRET_KEY", "secret")
	t.Setenv("S3_BUCKET", "my-bucket")
	t.Setenv("S3_USE_SSL", "true")
	t.Setenv("RABBITMQ_URL", "amqp://custom:5672/")
	t.Setenv("RABBITMQ_ENABLED", "true")
	t.Setenv("RABBITMQ_EXCHANGE", "custom.exchange")

	cfg, err := Load()
	require.NoError(t, err)
	assert.Equal(t, ":9090", cfg.Server.Addr)
	assert.Equal(t, "https://avatars.example.com", cfg.Server.BaseURL)
	assert.Equal(t, int64(5), cfg.Server.MaxUploadMB)
	assert.Equal(t, "postgres://custom/db", cfg.Postgres.DSN)
	assert.Equal(t, "s3.example.com", cfg.S3.Endpoint)
	assert.Equal(t, "key", cfg.S3.AccessKey)
	assert.Equal(t, "secret", cfg.S3.SecretKey)
	assert.Equal(t, "my-bucket", cfg.S3.Bucket)
	assert.True(t, cfg.S3.UseSSL)
	assert.Equal(t, "amqp://custom:5672/", cfg.RabbitMQ.URL)
	assert.True(t, cfg.RabbitMQ.Enabled)
	assert.Equal(t, "custom.exchange", cfg.RabbitMQ.Exchange)
}

func TestLoadMalformedValuesReturnError(t *testing.T) {
	t.Setenv("MAX_UPLOAD_MB", "not-a-number")

	_, err := Load()
	require.Error(t, err)
}
