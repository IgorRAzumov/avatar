package config

import "time"

const (
	DefaultServiceName      = "avatar-service"
	DefaultServiceVersion   = "1.0.0"
	DefaultEnvironment      = "development"
	DefaultLogLevel         = "info"
	DefaultTracingEnabled   = true
	DefaultLogsEnabled      = true
	DefaultMetricsEnabled   = true
	DefaultOTLPEndpoint     = "localhost:4317"
	DefaultTraceSampleRatio = 1.0

	DefaultMaxUploadMB         int64 = 10
	DefaultServerAddr                = ":8080"
	DefaultMetricsAddr               = ":9090"
	DefaultBaseURL                   = "http://localhost:8080"
	DefaultPostgresDSN               = "postgres://avatar:avatar@localhost:5432/avatar?sslmode=disable"
	DefaultPostgresAutoMigrate       = true
	DefaultServerReadTimeout         = 15 * time.Second
	DefaultServerWriteTimeout        = 15 * time.Second
	DefaultShutdownTimeout           = 10 * time.Second
	DefaultMigrateTimeout            = 2 * time.Minute

	DefaultS3Endpoint               = "localhost:9000"
	DefaultS3AccessKey              = "minioadmin"
	DefaultS3SecretKey              = "minioadmin"
	DefaultS3Bucket                 = "avatars"
	DefaultS3Region                 = "us-east-1"
	DefaultRabbitMQURL              = "amqp://guest:guest@localhost:5672/"
	DefaultRabbitMQExchange         = "avatars.exchange"
	DefaultRabbitMQUploadQueue      = "avatar.upload"
	DefaultRabbitMQDeleteQueue      = "avatar.delete"
	DefaultRabbitMQUploadRoutingKey = "avatar.uploaded"
	DefaultRabbitMQDeleteRoutingKey = "avatar.deleted"

	DefaultRateLimitEnabled = true
	DefaultRateLimitRPS     = 50
)

const bytesPerMB int64 = 1024 * 1024
