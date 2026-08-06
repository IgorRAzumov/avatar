package main

import (
	"avatar/cmd"
	"avatar/internal/app"
	"avatar/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		cmd.ExitWithLog(nil, err)
	}

	log, runtime, err := app.InitApp(cfg, cfg.Observability.ServiceName)
	if err != nil {
		cmd.ExitWithLog(nil, err)
	}
	defer func() { cmd.ShutdownRuntime(runtime) }()

	if err := app.Run(log, cfg, runtime.Kit); err != nil {
		cmd.ShutdownRuntime(runtime)
		cmd.ExitWithLog(log, err)
	}
}
