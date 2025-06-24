package sql

import (
	"context"
	"database/sql"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository"
)

type transaction struct {
	Tx *sql.Tx
}

func NewTransaction(tx *sql.Tx) repository.Transaction {
	return &transaction{
		Tx: tx,
	}
}

func (t *transaction) Begin() interface{} {
	return t.Tx
}

func (t *transaction) Commit(_ context.Context) error {
	return t.Tx.Commit()
}

func (t *transaction) Rollback(_ context.Context) error {
	return t.Tx.Rollback()
}
