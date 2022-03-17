package main

import (
	"context"

	"github.com/americanas-go/config"
	igzerolog "github.com/americanas-go/ignite/go.uber.org/zap.v1"
	igfiber "github.com/americanas-go/ignite/gofiber/fiber.v2"
	status "github.com/americanas-go/ignite/gofiber/fiber.v2/plugins/contrib/americanas-go/rest-response.v1"
	"github.com/americanas-go/log"
	"github.com/dubonzi/wirego/pkg/app"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

func mainModule() fx.Option {
	config.Load()
	log.SetGlobalLogger(igzerolog.NewLogger())

	return fx.Options(
		fx.Provide(
			context.Background,
			app.NewHandler,
			fx.Annotate(app.NewMatcher, fx.As(new(app.Matcher))),
			func(loader app.Loader) (app.Mappings, error) { return loader.GetMappings() },
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
			fx.Annotate(app.NewFileLoader, fx.As(new(app.Loader))),
		)
	}

}

func serverModule() fx.Option {
	return fx.Invoke(
		func(lc fx.Lifecycle, ctx context.Context, handler *app.Handler) {
			srv := igfiber.NewServer(
				ctx,
				status.Register,
			)

			srv.App().Use("/*", fiberErrorLogger)
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

func fiberErrorLogger(c *fiber.Ctx) error {
	err := c.Next()
	if err != nil {
		log.Error("captured error from handler: ", err)
	}
	return err
}
