package module

import (
	"go.uber.org/fx"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
)

var ConfigModule = fx.Provide(
	func() (*config.Config, error) {
		configPath := "../../../../../../config/config.yaml"
		envPath := "../../../../../../.env"
		cfg, err := config.LoadConfig(configPath, envPath)

		if err != nil {
			return nil, err
		}

		return cfg, nil
	},
)
