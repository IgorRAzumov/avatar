package logger

import (
	"io"
	"log/slog"
)

type Logger struct {
	log *slog.Logger
}

func New(out io.Writer) *Logger {
	return &Logger{
		log: slog.New(slog.NewJSONHandler(out, nil)),
	}
}

func Nop() *Logger {
	return &Logger{log: slog.New(slog.DiscardHandler)}
}

func (logger *Logger) Info(msg string, args ...any) {
	logger.log.Info(msg, args...)
}

func (logger *Logger) Warn(msg string, args ...any) {
	logger.log.Warn(msg, args...)
}

func (logger *Logger) Error(msg string, args ...any) {
	logger.log.Error(msg, args...)
}
