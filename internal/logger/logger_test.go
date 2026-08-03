package logger_test

import (
	"bytes"
	"strings"
	"testing"

	"avatar/internal/logger"

	"github.com/stretchr/testify/assert"
)

func TestLoggerWritesJSON(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf)

	log.Info("hello", "key", "value")

	out := buf.String()
	assert.True(t, strings.Contains(out, "hello"))
	assert.True(t, strings.Contains(out, "key"))
	assert.True(t, strings.Contains(out, "value"))
}

func TestLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	log := logger.New(&buf)

	log.Info("i")
	log.Warn("w")
	log.Error("e")

	out := buf.String()
	assert.Contains(t, out, "\"i\"")
	assert.Contains(t, out, "\"w\"")
	assert.Contains(t, out, "\"e\"")
}

func TestNopLoggerDoesNotPanic(t *testing.T) {
	log := logger.Nop()
	assert.NotPanics(t, func() {
		log.Info("i")
		log.Warn("w")
		log.Error("e")
	})
}
