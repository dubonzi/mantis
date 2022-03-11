package main

import (
	"context"

	"github.com/americanas-go/config"
	"github.com/americanas-go/ignite/gofiber/fiber.v2"
	status "github.com/americanas-go/ignite/gofiber/fiber.v2/plugins/contrib/americanas-go/rest-response.v1"
	igzerolog "github.com/americanas-go/ignite/rs/zerolog.v1"
	"github.com/americanas-go/log"
	"github.com/dubonzi/wirego/pkg/app"
	"go.uber.org/fx"
)

func mainModule() fx.Option {
	config.Load()
	igzerolog.NewLogger()

	return fx.Options(
		fx.Provide(
			context.Background,
			app.NewHandler,
			app.NewMatcher,
		),
		serverModule(),
		loaderModule(),
		fxLogger(),
	)
}

func loaderModule() fx.Option {
	mode := config.String("loader.mode")

	switch mode {
	case "db":
		return fx.Provide()
	default:
		return fx.Provide(
			func() app.Loader { return app.NewFileLoader() },
		)
	}

}

func serverModule() fx.Option {
	return fx.Invoke(
		func(lc fx.Lifecycle, ctx context.Context, handler *app.Handler) {
			srv := fiber.NewServer(
				ctx,
				status.Register,
			)

			srv.All("/*", handler.All)

			lc.Append(
				fx.Hook{
					OnStart: func(c context.Context) error {
						go srv.Serve(ctx)
						return nil
					},
					OnStop: func(c context.Context) error {
						log.Info("Shuting down server")
						return srv.App().Shutdown()
					},
				},
			)

		},
	)
}

func fxLogger() fx.Option {
	if config.Bool("fx.disableLogging") {
		return fx.NopLogger
	}
	return fx.Provide()
}
