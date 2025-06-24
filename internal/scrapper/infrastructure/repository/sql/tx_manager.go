package sql

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository"
)

type CtxKey struct{}

type txManager struct {
	db         *sql.DB
	logger     *slog.Logger
	ctxManager repository.CtxManager
}

func NewTxManager(db *sql.DB, log *slog.Logger, ctxManager repository.CtxManager) repository.TxManager {
	return &txManager{
		db:         db,
		logger:     log,
		ctxManager: ctxManager,
	}
}

func (txm *txManager) Do(ctx context.Context, fn func(context.Context) error) error {
	txm.logger.Info("Start of a transaction (Do)", slog.String("isolation", "serializable"))

	tx, err := txm.db.Begin()

	if err != nil {
		txm.logger.Error("Couldn't start a transaction", slog.Any("error", err))
		return err
	}

	newTx := tx
	newCtx := context.WithValue(ctx, txm.ctxManager.CtxKey(), newTx)

	if err := fn(newCtx); err != nil {
		rollBackErr := tx.Rollback()

		for rollBackErr != nil {
			rollBackErr = tx.Rollback()
		}

		return err
	}

	txm.logger.Debug("Commit of a transaction")

	if err := tx.Commit(); err != nil {
		txm.logger.Error("Error when committing a transaction", slog.Any("error", err))
		return err
	}

	txm.logger.Info("The transaction has been completed successfully")

	return nil
}
