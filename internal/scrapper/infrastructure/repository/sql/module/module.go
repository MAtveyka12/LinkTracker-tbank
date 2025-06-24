package module

import (
	sqll "database/sql"

	// Register PostgreSQL driver.
	_ "github.com/lib/pq"

	"go.uber.org/fx"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository/sql"
)

var SQLModule = fx.Provide(
	func(cfg *config.Config) (*sqll.DB, error) {
		db, err := sqll.Open("postgres", cfg.Database.GetDSN())
		if err != nil {
			return nil, err
		}
		return db, nil
	},
	sql.NewCtxManager,
	sql.NewLinkRepository,
	sql.NewChatRepository,
)
