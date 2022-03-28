package main

import (
	"context"

	"github.com/americanas-go/config"
	igfiber "github.com/americanas-go/ignite/gofiber/fiber.v2"
	"github.com/americanas-go/ignite/gofiber/fiber.v2/plugins/contrib/americanas-go/health.v1"
	"github.com/americanas-go/log"
	"github.com/americanas-go/log/contrib/go.uber.org/zap.v1"
	"github.com/dubonzi/wirego/pkg/app"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
)

func mainModule() fx.Option {
	loadDefaultConfig()
	config.Load()
	log.SetGlobalLogger(zap.NewLoggerWithOptions(zapOptions()))

	return fx.Options(
		fx.Provide(
			context.Background,
			app.NewHandler,
			app.NewRegexCache,
			app.NewFileLoader,
			func(loader *app.FileLoader) (app.Mappings, error) { return loader.GetMappings() },
			fx.Annotate(app.NewMatcher, fx.As(new(app.Matcher))),
		),
		serverModule(),
		healthModule(),
		fxLogger(),
	)
}

func serverModule() fx.Option {
	return fx.Invoke(
		func(lc fx.Lifecycle, ctx context.Context, handler *app.Handler) {
			srv := igfiber.NewServerWithOptions(
				ctx,
				serverFiberOptions(),
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

func healthModule() fx.Option {
	return fx.Invoke(
		func(lc fx.Lifecycle, handler *app.Handler) {
			ctx := context.Background()
			srv := igfiber.NewServerWithOptions(
				ctx,
				healthFiberOptions(),
				health.Register,
			)

			lc.Append(
				fx.Hook{
					OnStart: func(c context.Context) error {
						go srv.Serve(ctx)
						return nil
					},
					OnStop: func(c context.Context) error {
						return srv.App().Shutdown()
					},
				},
			)
		},
	)
}

func fxLogger() fx.Option {
	if config.Bool("fx.log.enable") {
		return fx.Provide()
	}
	return fx.NopLogger
}

func fiberErrorLogger(c *fiber.Ctx) error {
	err := c.Next()
	if err != nil {
		log.Error("captured error from handler: ", err)
	}
	return err
}
