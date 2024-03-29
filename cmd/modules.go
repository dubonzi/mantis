package main

import (
	"context"
	"fmt"

	"github.com/americanas-go/config"
	"github.com/americanas-go/log"
	"github.com/americanas-go/log/contrib/go.uber.org/zap.v1"
	"github.com/dubonzi/mantis/pkg/app"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
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
			app.NewLoader,
			app.NewJSONPathCache,
			app.NewMatcher,
			app.NewScenarioHandler,
			func(loader *app.Loader) (app.Mappings, error) { return loader.GetMappings() },
			fx.Annotate(app.NewResponseDelayer, fx.As(new(app.Delayer))),
			fx.Annotate(app.NewService, fx.As(new(app.ServiceMatcher))),
		),
		serverModule(),
		healthModule(),
		fxLogger(),
	)
}

func serverModule() fx.Option {
	return fx.Invoke(
		func(lc fx.Lifecycle, ctx context.Context, handler *app.Handler) {
			srv := fiber.New(
				fiber.Config{
					AppName:               "Mantis Server",
					DisableStartupMessage: config.Bool("server.disableStartupMessage"),
				},
			)

			srv.Use("/*", fiberErrorLogger)
			srv.Use(pprof.New())
			srv.All("/*", handler.All)

			lc.Append(
				fx.Hook{
					OnStart: func(c context.Context) error {
						go func() {
							if err := srv.Listen(":" + config.String("server.port")); err != nil {
								panic(fmt.Errorf("error starting mantis server: %s", err))
							}
						}()
						return nil
					},
					OnStop: func(c context.Context) error {
						log.Info("Shuting down server")
						return srv.Shutdown()
					},
				},
			)
		},
	)
}

func healthModule() fx.Option {
	return fx.Invoke(
		func(lc fx.Lifecycle, handler *app.Handler) {
			srv := fiber.New(
				fiber.Config{
					AppName:               "Mantis Health Server",
					DisableStartupMessage: config.Bool("server.disableStartupMessage"),
				},
			)

			srv.Get("/health", handler.Health)

			lc.Append(
				fx.Hook{
					OnStart: func(c context.Context) error {
						go func() {
							if err := srv.Listen(":" + config.String("health.port")); err != nil {
								panic(fmt.Errorf("error starting health server: %s", err))
							}
						}()
						return nil
					},
					OnStop: func(c context.Context) error {
						return srv.Shutdown()
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
