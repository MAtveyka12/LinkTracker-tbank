package pgxpool

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository"
)

type ctxManager struct {
	pool *pgxpool.Pool
}

func NewCtxManager(pool *pgxpool.Pool) repository.CtxManager {
	return &ctxManager{
		pool: pool,
	}
}

func (p *ctxManager) ByKey(ctx context.Context, key repository.CtxKey) repository.Transaction {
	tx, ok := ctx.Value(key).(pgx.Tx)
	if !ok {
		return nil
	}

	return NewTransaction(tx)
}

func (p *ctxManager) Default(ctx context.Context) repository.Transaction {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return nil
	}

	return NewTransaction(tx)
}

func (p *ctxManager) CtxKey() repository.CtxKey {
	return repository.CtxKey{}
}
