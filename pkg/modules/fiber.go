package modules

import (
	"context"

	"github.com/americanas-go/ignite/gofiber/fiber.v2"
	"github.com/americanas-go/ignite/gofiber/fiber.v2/plugins/contrib/americanas-go/health.v1"
	status "github.com/americanas-go/ignite/gofiber/fiber.v2/plugins/contrib/americanas-go/rest-response.v1"
	"github.com/dubonzi/wirego/pkg/app"
	"go.uber.org/fx"
)

func serverModule() fx.Option {
	return fx.Invoke(
		func(lc fx.Lifecycle, ctx context.Context, handler *app.Handler) {
			srv := fiber.NewServer(
				ctx,
				health.Register,
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
						return srv.App().Shutdown()
					},
				},
			)

		},
	)
}
