package pgxpool

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository"
)

type transaction struct {
	Tx pgx.Tx
}

func NewTransaction(tx pgx.Tx) repository.Transaction {
	return &transaction{
		Tx: tx,
	}
}

func (t *transaction) Begin() interface{} {
	return t.Tx
}

func (t *transaction) Commit(ctx context.Context) error {
	return t.Tx.Commit(ctx)
}

func (t *transaction) Rollback(ctx context.Context) error {
	return t.Tx.Rollback(ctx)
}
