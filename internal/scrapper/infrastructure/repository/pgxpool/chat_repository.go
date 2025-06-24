package pgxpool

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository"
)

type chatRepository struct {
	ctxManager repository.CtxManager
}

func (cr *chatRepository) RegisterChat(ctx context.Context, userID int) error {
	slog.Debug("Registering chat",
		slog.Int("user_id", userID),
		slog.String("operation", "insert"),
	)

	tr := cr.ctxManager.ByKey(ctx, cr.ctxManager.CtxKey())
	if tr == nil {
		slog.Debug("Using default transaction")

		tr = cr.ctxManager.Default(ctx)
	}

	exec := tr.Begin().(pgx.Tx)

	sql, args, err := sq.Insert("chats").
		Columns("id").
		Values(userID).
		Suffix("ON CONFLICT (id) DO NOTHING").
		PlaceholderFormat(sq.Dollar).
		ToSql()
	if err != nil {
		slog.Error("Failed to build SQL query",
			slog.String("error", err.Error()),
			slog.String("operation", "insert"),
		)

		return fmt.Errorf("failed to build query: %w", err)
	}

	slog.Debug("Executing SQL query",
		slog.String("query", sql),
		slog.Any("args", args),
	)

	_, err = exec.Exec(ctx, sql, args...)
	if err != nil {
		slog.Error("Failed to execute query",
			slog.String("error", err.Error()),
			slog.String("query", sql),
			slog.Int("user_id", userID),
		)

		slog.Error("Failed to exec", slog.String("query", sql))

		return err
	}

	if err := exec.Commit(ctx); err != nil {
		slog.Error("Failed to commit transaction",
			slog.String("error", err.Error()),
		)

		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	slog.Info("Chat registered successfully",
		slog.Int("user_id", userID),
	)

	return nil
}

func (cr *chatRepository) DeleteChat(ctx context.Context, userID int) error {
	slog.Debug("Deleting chat",
		slog.Int("user_id", userID),
		slog.String("operation", "delete"),
	)

	tr := cr.ctxManager.ByKey(ctx, cr.ctxManager.CtxKey())
	if tr == nil {
		slog.Debug("Using default transaction")

		tr = cr.ctxManager.Default(ctx)
	}

	exec, ok := tr.Begin().(pgx.Tx)
	if !ok {
		return errors.New("failed to assert transaction type")
	}

	var err error
	defer func() {
		if err != nil {
			_ = exec.Rollback(ctx)
		} else {
			_ = exec.Commit(ctx)
		}
	}()

	sql, args, err := sq.Delete("chats").
		Where(sq.Eq{"id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		slog.Error("Failed to perform", slog.String("query", sql))
		return err
	}

	slog.Debug("Executing SQL query",
		slog.String("query", sql),
		slog.Any("args", args),
	)

	if _, err = exec.Exec(ctx, sql, args...); err != nil {
		slog.Error("Failed to execute query",
			slog.String("error", err.Error()),
			slog.String("query", sql),
			slog.Int("user_id", userID),
		)

		return fmt.Errorf("failed to exec query: %w", err)
	}

	slog.Info("Chat deleted successfully",
		slog.Int("user_id", userID),
	)

	return nil
}

func (cr *chatRepository) ChatExists(ctx context.Context, userID int) (bool, error) {
	slog.Debug("Checking chat existence",
		slog.Int("user_id", userID),
		slog.String("operation", "check_exists"),
	)

	tr := cr.ctxManager.ByKey(ctx, cr.ctxManager.CtxKey())
	if tr == nil {
		slog.Debug("Using default transaction")

		tr = cr.ctxManager.Default(ctx)
	}

	exec := tr.Begin().(pgx.Tx)
	sql, args, err := sq.Select("1").
		From("chats").
		Where(sq.Eq{"id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		slog.Error("Failed to perform", slog.String("query", sql))
		return false, err
	}

	slog.Debug("Executing SQL query",
		slog.String("query", sql),
		slog.Any("args", args),
	)

	var exists bool
	if err = exec.QueryRow(ctx, sql, args...).Scan(&exists); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Debug("Chat does not exist",
				slog.Int("user_id", userID),
			)

			return false, nil
		}

		slog.Error("Failed to check chat existence",
			slog.String("error", err.Error()),
			slog.String("query", sql),
			slog.Int("user_id", userID),
		)

		return false, fmt.Errorf("failed to check existence: %w", err)
	}

	slog.Debug("Chat existence check result",
		slog.Int("user_id", userID),
		slog.Bool("exists", exists),
	)

	return exists, nil
}

func NewChatRepository(ctxManager repository.CtxManager) repository.ChatRepository {
	slog.Info("Initializing chat repository")

	return &chatRepository{
		ctxManager: ctxManager,
	}
}
