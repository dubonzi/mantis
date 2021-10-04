package modules

import (
	"context"

	"github.com/americanas-go/config"
	"github.com/dubonzi/wirego/pkg/app"
	"go.uber.org/fx"
)

func Main() fx.Option {
	config.Load()
	return fx.Options(
		fx.Provide(
			context.Background,
			app.NewHandler,
			app.NewMatcher,
		),
		serverModule(),
		loaderModule(),
	)
}

func loaderModule() fx.Option {
	mode := config.String("loader.mode")

	switch mode {
	case "db":
		return fx.Provide()
	default:
		return fx.Provide(
			app.NewFileLoader,
		)
	}

}
