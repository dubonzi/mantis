package main

import (
	"context"
	"os"

	"github.com/americanas-go/log"
	"go.uber.org/fx"
)

func main() {
	os.Exit(run())
}

func run() int {
	app := fx.New(mainModule())

	startCtx, cancelStart := context.WithTimeout(context.Background(), app.StartTimeout())
	defer cancelStart()
	if err := app.Start(startCtx); err != nil {
		log.Errorf("error starting app: %v", err)
		return 1
	}

	signal := <-app.Wait()
	stopCtx, cancelStop := context.WithTimeout(context.Background(), app.StopTimeout())
	defer cancelStop()
	if err := app.Stop(stopCtx); err != nil {
		log.Errorf("error stopping app: %v", err)
		return 1
	}
	return signal.ExitCode
}
