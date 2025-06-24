package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/es-debug/backend-academy-2024-go-template/internal/scrapper/infrastructure/repository"
)

type linkRepository struct {
	ctxManager repository.CtxManager
}

func (lr *linkRepository) GetLink(ctx context.Context, userID, linkID int) (link string, id int, err error) {
	slog.Debug("Getting link",
		slog.Int("user_id", userID),
		slog.Int("link_id", linkID),
		slog.String("operation", "get_link"),
	)

	tr := lr.ctxManager.ByKey(ctx, lr.ctxManager.CtxKey())
	if tr == nil {
		tr = lr.ctxManager.Default(ctx)
	}

	exec := tr.Begin().(*sql.Tx)
	query := `SELECT id, url FROM links WHERE chat_id = $1 AND id = $2`

	err = exec.QueryRowContext(ctx, query, userID, linkID).Scan(&id, &link)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Warn("Link not found",
				slog.Int("user_id", userID),
				slog.Int("link_id", linkID),
			)

			return "", 0, repository.ErrNoRows
		}

		slog.Error("Failed to get link",
			slog.Int("user_id", userID),
			slog.Int("link_id", linkID),
			slog.String("error", err.Error()),
			slog.String("query", query),
		)

		return "", 0, err
	}

	slog.Info("Link retrieved successfully",
		slog.Int("user_id", userID),
		slog.Int("link_id", id),
		slog.String("url", link),
	)

	return link, id, nil
}

func (lr *linkRepository) GetAllLinks(ctx context.Context, userID int) ([]string, error) {
	slog.Debug("Getting all links",
		slog.Int("user_id", userID),
		slog.String("operation", "get_all_links"),
	)

	tr := lr.ctxManager.ByKey(ctx, lr.ctxManager.CtxKey())
	if tr == nil {
		tr = lr.ctxManager.Default(ctx)
	}

	exec := tr.Begin().(*sql.Tx)
	query := `SELECT url FROM links WHERE chat_id = $1`
	rows, err := exec.QueryContext(ctx, query, userID)

	if err != nil {
		slog.Error("Failed to get all links",
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
			slog.String("query", query),
		)

		return nil, err
	}

	defer rows.Close()

	var links []string

	for rows.Next() {
		var link string
		if err := rows.Scan(&link); err != nil {
			slog.Error("Failed to scan link row",
				slog.Int("user_id", userID),
				slog.String("error", err.Error()),
			)

			return nil, err
		}

		links = append(links, link)
	}

	if err := rows.Err(); err != nil {
		slog.Error("Error after iterating links",
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
		)

		return nil, err
	}

	slog.Info("Retrieved all links successfully",
		slog.Int("user_id", userID),
		slog.Int("count", len(links)),
	)

	return links, nil
}

func (lr *linkRepository) AddLink(ctx context.Context, userID int, link string) (int, error) {
	slog.Debug("Adding link",
		slog.Int("user_id", userID),
		slog.String("url", link),
		slog.String("operation", "add_link"),
	)

	tr := lr.ctxManager.ByKey(ctx, lr.ctxManager.CtxKey())
	if tr == nil {
		tr = lr.ctxManager.Default(ctx)
	}

	exec := tr.Begin().(*sql.Tx)
	query := `INSERT INTO links (chat_id, url) VALUES ($1, $2) RETURNING id`

	var id int

	err := exec.QueryRowContext(ctx, query, userID, link).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Warn("No rows returned after adding link",
				slog.Int("user_id", userID),
				slog.String("url", link),
			)

			return 0, repository.ErrNoRows
		}

		slog.Error("Failed to add link",
			slog.Int("user_id", userID),
			slog.String("url", link),
			slog.String("error", err.Error()),
			slog.String("query", query),
		)

		return 0, err
	}

	slog.Info("Link added successfully",
		slog.Int("user_id", userID),
		slog.Int("link_id", id),
		slog.String("url", link),
	)

	return id, nil
}

func (lr *linkRepository) DeleteLink(ctx context.Context, userID, linkID int, link string) error {
	slog.Debug("Deleting link",
		slog.Int("user_id", userID),
		slog.Int("link_id", linkID),
		slog.String("url", link),
		slog.String("operation", "delete_link"),
	)

	tr := lr.ctxManager.ByKey(ctx, lr.ctxManager.CtxKey())
	if tr == nil {
		tr = lr.ctxManager.Default(ctx)
	}

	exec := tr.Begin().(*sql.Tx)
	query := `DELETE FROM links WHERE chat_id = $1 AND id = $2 AND url = $3`
	result, err := exec.ExecContext(ctx, query, userID, linkID, link)

	if err != nil {
		slog.Error("Failed to delete link",
			slog.Int("user_id", userID),
			slog.Int("link_id", linkID),
			slog.String("url", link),
			slog.String("error", err.Error()),
			slog.String("query", query),
		)

		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("Failed to get rows affected when deleting link",
			slog.Int("user_id", userID),
			slog.Int("link_id", linkID),
			slog.String("error", err.Error()),
		)

		return err
	}

	if rowsAffected == 0 {
		slog.Warn("No links deleted - possibly not found",
			slog.Int("user_id", userID),
			slog.Int("link_id", linkID),
		)

		return repository.ErrNoRows
	}

	slog.Info("Link deleted successfully",
		slog.Int("user_id", userID),
		slog.Int("link_id", linkID),
		slog.Int64("rows_affected", rowsAffected),
	)

	return nil
}

func (lr *linkRepository) DeleteAllLinks(ctx context.Context, userID int) error {
	slog.Debug("Deleting all links",
		slog.Int("user_id", userID),
		slog.String("operation", "delete_all_links"),
	)

	tr := lr.ctxManager.ByKey(ctx, lr.ctxManager.CtxKey())
	if tr == nil {
		tr = lr.ctxManager.Default(ctx)
	}

	exec := tr.Begin().(*sql.Tx)
	query := `DELETE FROM links WHERE chat_id = $1`
	result, err := exec.ExecContext(ctx, query, userID)

	if err != nil {
		slog.Error("Failed to delete all links",
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
			slog.String("query", query),
		)

		return fmt.Errorf("failed to exec: %s", err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("Failed to get rows affected when deleting all links",
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
		)

		return fmt.Errorf("failed to get rows affected: %s", err.Error())
	}

	if rowsAffected == 0 {
		slog.Warn("No links deleted - possibly none found",
			slog.Int("user_id", userID),
		)

		return repository.ErrNoRows
	}

	slog.Info("All links deleted successfully",
		slog.Int("user_id", userID),
		slog.Int64("rows_affected", rowsAffected),
	)

	return nil
}

func (lr *linkRepository) GetLinkIDByURL(ctx context.Context, userID int, url string) (int, error) {
	slog.Debug("Getting link ID by URL",
		slog.Int("user_id", userID),
		slog.String("url", url),
		slog.String("operation", "get_link_id_by_url"),
	)

	tr := lr.ctxManager.ByKey(ctx, lr.ctxManager.CtxKey())

	if tr == nil {
		tr = lr.ctxManager.Default(ctx)
	}

	exec := tr.Begin().(*sql.Tx)
	query := `SELECT id FROM links WHERE chat_id = $1 AND url = $2`

	var id int

	err := exec.QueryRowContext(ctx, query, userID, url).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Warn("Link not found by URL",
				slog.Int("user_id", userID),
				slog.String("url", url),
			)

			return 0, repository.ErrNoRows
		}

		slog.Error("Failed to get link ID by URL",
			slog.Int("user_id", userID),
			slog.String("url", url),
			slog.String("error", err.Error()),
			slog.String("query", query),
		)

		return 0, fmt.Errorf("failed to get link ID: %s", err.Error())
	}

	slog.Info("Link ID retrieved by URL successfully",
		slog.Int("user_id", userID),
		slog.String("url", url),
		slog.Int("link_id", id),
	)

	return id, nil
}

func NewLinkRepository(ctxManager repository.CtxManager) repository.LinkRepository {
	slog.Info("Initializing link repository")

	return &linkRepository{
		ctxManager: ctxManager,
	}
}
