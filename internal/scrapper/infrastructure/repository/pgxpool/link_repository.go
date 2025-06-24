package pgxpool

import (
	"context"
	"errors"
	"log/slog"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

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

	exec := tr.Begin().(pgx.Tx)
	sql, args, err := sq.Select("id", "url").
		From("links").
		Where(sq.Eq{"chat_id": userID, "id": linkID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		slog.Error("Failed to build SQL query",
			slog.Int("user_id", userID),
			slog.Int("link_id", linkID),
			slog.String("error", err.Error()),
		)

		return "", 0, err
	}

	if err = exec.QueryRow(ctx, sql, args...).Scan(&id, &link); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Warn("Link not found",
				slog.Int("user_id", userID),
				slog.Int("link_id", linkID),
			)

			return "", 0, errors.New("not found")
		}

		slog.Error("Failed to get link",
			slog.Int("user_id", userID),
			slog.Int("link_id", linkID),
			slog.String("sql", sql),
			slog.Any("args", args),
			slog.String("error", err.Error()),
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

	exec := tr.Begin().(pgx.Tx)
	sql, args, err := sq.Select("url").
		From("links").
		Where(sq.Eq{"chat_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		slog.Error("Failed to build SQL query",
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
		)

		return nil, err
	}

	rows, err := exec.Query(ctx, sql, args...)
	if err != nil {
		slog.Error("Failed to execute query",
			slog.Int("user_id", userID),
			slog.String("sql", sql),
			slog.Any("args", args),
			slog.String("error", err.Error()),
		)

		return nil, err
	}

	defer rows.Close()

	var links []string

	for rows.Next() {
		var link string

		if err := rows.Scan(&link); err != nil {
			slog.Error("Failed to scan row",
				slog.Int("user_id", userID),
				slog.String("error", err.Error()),
			)

			return nil, err
		}

		links = append(links, link)
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

	exec, ok := tr.Begin().(pgx.Tx)
	if !ok {
		return -1, errors.New("failed to assert transaction type")
	}

	var err error
	defer func() {
		if err != nil {
			_ = exec.Rollback(ctx)
		} else {
			_ = exec.Commit(ctx)
		}
	}()

	sql, args, err := sq.Insert("links").
		Columns("chat_id", "url").
		Values(userID, link).
		Suffix("RETURNING id").
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		slog.Error("Failed to build SQL query",
			slog.Int("user_id", userID),
			slog.String("url", link),
			slog.String("error", err.Error()),
		)

		return 0, err
	}

	var id int

	err = exec.QueryRow(ctx, sql, args...).Scan(&id)
	if err != nil {
		slog.Error("Failed to add link",
			slog.Int("user_id", userID),
			slog.String("url", link),
			slog.String("sql", sql),
			slog.Any("args", args),
			slog.String("error", err.Error()),
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

	exec := tr.Begin().(pgx.Tx)
	sql, args, err := sq.Delete("links").
		Where(sq.Eq{"chat_id": userID, "id": linkID, "url": link}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		slog.Error("Failed to build SQL query",
			slog.Int("user_id", userID),
			slog.Int("link_id", linkID),
			slog.String("url", link),
			slog.String("error", err.Error()),
		)

		return err
	}

	cmdTag, err := exec.Exec(ctx, sql, args...)
	if err != nil {
		slog.Error("Failed to delete link",
			slog.Int("user_id", userID),
			slog.Int("link_id", linkID),
			slog.String("url", link),
			slog.String("sql", sql),
			slog.Any("args", args),
			slog.String("error", err.Error()),
		)

		return err
	}

	if cmdTag.RowsAffected() == 0 {
		slog.Warn("Link not found for deletion",
			slog.Int("user_id", userID),
			slog.Int("link_id", linkID),
			slog.String("url", link),
		)

		return errors.New("not found")
	}

	slog.Info("Link deleted successfully",
		slog.Int("user_id", userID),
		slog.Int("link_id", linkID),
		slog.String("url", link),
		slog.Int64("rows_affected", cmdTag.RowsAffected()),
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

	exec := tr.Begin().(pgx.Tx)
	sql, args, err := sq.Delete("links").
		Where(sq.Eq{"chat_id": userID}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		slog.Error("Failed to build SQL query",
			slog.Int("user_id", userID),
			slog.String("error", err.Error()),
		)

		return err
	}

	cmdTag, err := exec.Exec(ctx, sql, args...)
	if err != nil {
		slog.Error("Failed to delete all links",
			slog.Int("user_id", userID),
			slog.String("sql", sql),
			slog.Any("args", args),
			slog.String("error", err.Error()),
		)

		return err
	}

	slog.Info("All links deleted successfully",
		slog.Int("user_id", userID),
		slog.Int64("rows_affected", cmdTag.RowsAffected()),
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

	exec := tr.Begin().(pgx.Tx)
	sql, args, err := sq.Select("id").
		From("links").
		Where(sq.Eq{"chat_id": userID, "url": url}).
		PlaceholderFormat(sq.Dollar).
		ToSql()

	if err != nil {
		slog.Error("Failed to build SQL query",
			slog.Int("user_id", userID),
			slog.String("url", url),
			slog.String("error", err.Error()),
		)

		return 0, err
	}

	var id int
	if err = exec.QueryRow(ctx, sql, args...).Scan(&id); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			slog.Warn("Link not found by URL",
				slog.Int("user_id", userID),
				slog.String("url", url),
			)

			return 0, errors.New("not found")
		}

		slog.Error("Failed to get link ID by URL",
			slog.Int("user_id", userID),
			slog.String("url", url),
			slog.String("sql", sql),
			slog.Any("args", args),
			slog.String("error", err.Error()),
		)

		return 0, err
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
