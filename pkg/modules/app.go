package modules

import (
	"context"

	"github.com/americanas-go/config"
	"go.uber.org/fx"
)

func Main() fx.Option {
	return fx.Options(
		fx.Provide(
			context.Background,
		),
		fx.Invoke(
			config.Load,
		),
		serverModule(),
	)
}
