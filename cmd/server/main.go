package main

import (
	"os"

	"avatar/internal/app"
	"avatar/internal/logger"
)

func main() {
	log := logger.New(os.Stdout)

	if err := app.Run(log); err != nil {
		log.Error("application error", "error", err)
		os.Exit(1)
	}
}
