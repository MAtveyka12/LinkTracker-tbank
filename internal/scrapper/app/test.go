package app

import (
	"log/slog"
	"os"

	"github.com/lmittmann/tint"
	"go.uber.org/fx"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	httpServer "github.com/es-debug/backend-academy-2024-go-template/pkg/http_server"
)

func A() {
	app := fx.New(
		fx.Provide(func() *slog.Logger {
			return slog.New(tint.NewHandler(os.Stdout, nil))
		}),
		fx.Provide(
			func() (*config.Config, error) {
				return config.LoadConfig("config/config.yaml", ".env")
			},
		),
		fx.Provide(httpServer.NewHTTPServer),
	)

	app.Run()
}
