package sql

import (
	"context"
	"database/sql"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository"
)

type ctxManager struct {
	db *sql.DB
}

func (ctxm *ctxManager) ByKey(ctx context.Context, key repository.CtxKey) repository.Transaction {
	tx, ok := ctx.Value(key).(*sql.Tx)
	if !ok {
		return nil
	}

	return NewTransaction(tx)
}

func (ctxm *ctxManager) Default(_ context.Context) repository.Transaction {
	tx, err := ctxm.db.Begin()
	if err != nil {
		return nil
	}

	return NewTransaction(tx)
}

func (ctxm *ctxManager) CtxKey() repository.CtxKey {
	return repository.CtxKey{}
}

func NewCtxManager(db *sql.DB) repository.CtxManager {
	return &ctxManager{
		db: db,
	}
}
