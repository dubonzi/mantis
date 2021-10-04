package main

import (
	"github.com/dubonzi/wirego/pkg/modules"
	"go.uber.org/fx"
)

func main() {

	app := fx.New(
		modules.Main(),
	)

	app.Run()
}
