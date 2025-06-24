package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository"
)

type chatRepository struct {
	ctxManager repository.CtxManager
}

func (cr *chatRepository) execChatOperation(
	ctx context.Context,
	userID int,
	operation string,
	query string,
	successLog string,
	notFoundLog string,
) error {
	slog.Debug(operation,
		slog.Int("user_id", userID),
		slog.String("operation", operation),
	)

	tr := cr.ctxManager.ByKey(ctx, cr.ctxManager.CtxKey())
	if tr == nil {
		tr = cr.ctxManager.Default(ctx)
	}

	exec := tr.Begin().(*sql.Tx)
	result, err := exec.ExecContext(ctx, query, userID)

	if err != nil {
		return fmt.Errorf("failed to %s: %w", operation, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error(fmt.Sprintf("Failed to %s", operation),
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
			slog.String("query", query),
		)

		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		slog.Warn(notFoundLog, slog.Int("user_id", userID))
		return repository.ErrNoRows
	}

	slog.Info(successLog, slog.Int("user_id", userID))

	return nil
}

func (cr *chatRepository) RegisterChat(ctx context.Context, userID int) error {
	return cr.execChatOperation(
		ctx,
		userID,
		"insert_chat",
		`INSERT INTO chats (id) VALUES ($1) ON CONFLICT (id) DO NOTHING`,
		"Chat registered successfully",
		"Chat already exists",
	)
}

func (cr *chatRepository) DeleteChat(ctx context.Context, userID int) error {
	return cr.execChatOperation(
		ctx,
		userID,
		"delete_chat",
		`DELETE FROM chats WHERE id = $1`,
		"Chat deleted successfully",
		"Chat not found for deletion",
	)
}

func (cr *chatRepository) ChatExists(ctx context.Context, userID int) (bool, error) {
	slog.Debug("Checking chat existence",
		slog.Int("user_id", userID),
		slog.String("operation", "check_chat"),
	)

	tr := cr.ctxManager.ByKey(ctx, cr.ctxManager.CtxKey())
	if tr == nil {
		tr = cr.ctxManager.Default(ctx)
	}

	exec := tr.Begin().(*sql.Tx)
	query := `
		SELECT EXISTS(
			SELECT 1 FROM chats 
			WHERE id = $1
		)
	`

	var exists bool

	err := exec.QueryRowContext(ctx, query, userID).Scan(&exists)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Debug("Chat does not exist", slog.Int("user_id", userID))
			return false, repository.ErrNoRows
		}

		slog.Error("Failed to check chat existence",
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
			slog.String("query", query),
		)

		return false, err
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
