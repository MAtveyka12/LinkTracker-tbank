package module_test

import (
	"context"
	"database/sql"
	"testing"

	"go.uber.org/fx"

	"github.com/es-debug/backend-academy-2024-go-template/internal/config"
	mdl "github.com/es-debug/backend-academy-2024-go-template/internal/config/module"
	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository/sql/module"
)

func TestSQLModule(t *testing.T) {
	app := fx.New(
		mdl.ConfigModule,
		module.SQLModule,
		fx.Invoke(func(db *sql.DB, cfg *config.Config) {
			if db == nil || cfg == nil {
				t.Error("expected *sql.DB || *config.Config to be provided")
			}
		}),
	)

	ctx := context.Background()

	if err := app.Start(ctx); err != nil {
		t.Fatalf("failed to start fx app: %v", err)
	}

	if err := app.Stop(ctx); err != nil {
		t.Fatalf("failed to stop fx app: %v", err)
	}
}
