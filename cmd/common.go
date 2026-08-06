package cmd

import (
	"context"
	"os"
	"time"

	"avatar/internal/logger"
	"avatar/internal/observability"
)

func ShutdownRuntime(runtime *observability.Runtime) {
	if runtime == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := runtime.Shutdown(ctx); err != nil {
		_, _ = os.Stderr.WriteString("shutdown observability: " + err.Error() + "\n")
	}
}

func ExitWithLog(log *logger.Logger, err error) {
	if err != nil {
		if log != nil {
			log.Error(context.Background(), "application error", "error", err)
		} else {
			_, _ = os.Stderr.WriteString("application error: " + err.Error() + "\n")
		}
	}
	os.Exit(1)
}
