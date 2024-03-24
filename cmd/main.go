package main

import (
	"context"
	"os"
	"os/signal"

	amercfg "github.com/americanas-go/config"
	"github.com/americanas-go/log"
	"github.com/dubonzi/mantis/pkg/config"
	"go.uber.org/fx"
)

func main() {
	config.SetDefaultConfig()
	amercfg.Load()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	setupOtel()

	app := fx.New(
		mainModule(),
	)

	ctx := context.Background()

	err := app.Start(ctx)
	if err != nil {
		log.Fatal("error starting app: ", err)
	}

	<-stop
	err = app.Stop(ctx)
	if err != nil {
		log.Errorf("error stopping app: ", err)
		os.Exit(1)
	}
}
