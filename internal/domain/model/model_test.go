package model

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProcessingStatuses(t *testing.T) {
	assert.Equal(t, "pending", ProcessingStatusPending)
	assert.Equal(t, "processing", ProcessingStatusProcessing)
	assert.Equal(t, "completed", ProcessingStatusCompleted)
}
