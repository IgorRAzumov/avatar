package worker_test

import (
	"context"
	"testing"

	"avatar/internal/config"
	"avatar/internal/logger"
	"avatar/internal/worker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunRequiresRabbitMQ(t *testing.T) {
	cfg := &config.Config{RabbitMQ: config.RabbitMQConfig{Enabled: false}}

	err := worker.Run(context.Background(), logger.Nop(), cfg)
	require.Error(t, err)
	assert.ErrorContains(t, err, "RABBITMQ_ENABLED=true")
}
