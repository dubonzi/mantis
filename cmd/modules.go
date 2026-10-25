package main

import (
	"context"
	"fmt"
	"net"

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
		func(lc fx.Lifecycle, handler *app.Handler, shutdowner fx.Shutdowner) {
			srv := fiber.New(
				fiber.Config{
					AppName:               "Mantis Server",
					DisableStartupMessage: config.Bool("server.disableStartupMessage"),
				},
			)

			srv.Use("/*", fiberErrorLogger)
			srv.All("/*", handler.All)

			lc.Append(
				fx.Hook{
					OnStart: func(c context.Context) error {
						ln, err := net.Listen("tcp4", ":"+config.String("server.port"))
						if err != nil {
							return fmt.Errorf("error starting mantis server: %w", err)
						}
						go func() {
							if err := srv.Listener(ln); err != nil {
								log.Errorf("error serving mantis server: %v", err)
								if err := shutdowner.Shutdown(fx.ExitCode(1)); err != nil {
									log.Errorf("error requesting shutdown: %v", err)
								}
							}
						}()
						return nil
					},
					OnStop: func(c context.Context) error {
						log.Info("Shutting down server")
						return srv.ShutdownWithContext(c)
					},
				},
			)
		},
	)
}

func healthModule() fx.Option {
	return fx.Invoke(
		func(lc fx.Lifecycle, handler *app.Handler, shutdowner fx.Shutdowner) {
			srv := fiber.New(
				fiber.Config{
					AppName:               "Mantis Health Server",
					DisableStartupMessage: config.Bool("server.disableStartupMessage"),
				},
			)

			if config.Bool("server.pprof.enable") {
				srv.Use(pprof.New())
			}
			srv.Get("/health", handler.Health)

			lc.Append(
				fx.Hook{
					OnStart: func(c context.Context) error {
						ln, err := net.Listen("tcp4", ":"+config.String("health.port"))
						if err != nil {
							return fmt.Errorf("error starting health server: %w", err)
						}
						go func() {
							if err := srv.Listener(ln); err != nil {
								log.Errorf("error serving health server: %v", err)
								if err := shutdowner.Shutdown(fx.ExitCode(1)); err != nil {
									log.Errorf("error requesting shutdown: %v", err)
								}
							}
						}()
						return nil
					},
					OnStop: func(c context.Context) error {
						return srv.ShutdownWithContext(c)
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
