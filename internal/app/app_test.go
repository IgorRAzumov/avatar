package app

import (
	"context"
	"testing"

	"avatar/internal/config"
	"avatar/internal/logger"
	"avatar/internal/testutil"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileStoreS3(t *testing.T) {
	cfg := &config.Config{S3: config.S3Config{
		Endpoint:  "127.0.0.1:9000",
		AccessKey: "minioadmin",
		SecretKey: "minioadmin",
		Bucket:    "avatars",
		UseSSL:    false,
		Region:    "us-east-1",
	}}

	store, err := NewFileStore(cfg)
	if err != nil {
		t.Skipf("minio not available: %v", err)
	}
	require.NotNil(t, store)

	type pinger interface {
		Ping(context.Context) error
	}
	if p, ok := store.(pinger); ok {
		if err := p.Ping(context.Background()); err != nil {
			t.Skipf("minio not available: %v", err)
		}
	}
}

func TestNewPublisherInProcess(t *testing.T) {
	cfg := &config.Config{RabbitMQ: config.RabbitMQConfig{Enabled: false}}
	repo := testutil.NewMemoryAvatarStore()
	store := testutil.NewMemoryStorage()

	publisher, broker, err := NewPublisher(cfg, repo, repo, store, logger.Nop())
	require.NoError(t, err)
	assert.NotNil(t, publisher)
	assert.Nil(t, broker)
}
