package cmd

import (
	"avatar/internal/logger"
	"avatar/internal/observability"
	"os"

	"golang.org/x/net/context"
)

func ShutdownRuntime(runtime *observability.Runtime) {
	if runtime == nil {
		return
	}
	if err := runtime.Shutdown(context.Background()); err != nil {
		_, _ = os.Stderr.WriteString("shutdown observability: " + err.Error() + "\n")
	}
}

func ExitWithLog(log *logger.Logger, err error) {
	if err != nil {
		if log != nil {
			log.Error("application error", "error", err)
		} else {
			_, _ = os.Stderr.WriteString("application error: " + err.Error() + "\n")
		}
	}
	os.Exit(1)
}
