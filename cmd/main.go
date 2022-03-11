package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/americanas-go/log"
	"go.uber.org/fx"
)

func main() {

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	app := fx.New(
		mainModule(),
	)

	ctx := context.Background()

	err := app.Start(ctx)
	if err != nil {
		log.Fatal("error starting app: ", err)
	}

	<-stop
	app.Stop(ctx)
}
