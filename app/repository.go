package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/mmcdole/gofeed"
)

type repository struct {
	db  *sqlx.DB
	log *slog.Logger
}

func NewRepository(db *sqlx.DB, log *slog.Logger) IRepository {
	return &repository{
		db:  db,
		log: log,
	}
}

func (r *repository) GetFeeds(ctx context.Context) ([]*Feed, error) {
	var feeds []*Feed
	if err := r.db.SelectContext(ctx, &feeds, "SELECT * FROM feeds"); err != nil {
		r.log.Error("failed to select from feeds table", "error", err.Error())
		return nil, ErrDatabaseOperation
	}
	return feeds, nil
}

func (r *repository) AddArticle(ctx context.Context, item *gofeed.Item, feedID int64) {
	stmt := `
		INSERT INTO articles (feed_id, title, link, summary, published_at) 
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (link) DO NOTHING
	`
	if _, err := r.db.ExecContext(
		ctx, stmt, feedID, item.Title, item.Link, item.Description, item.PublishedParsed.UTC(),
	); err != nil {
		r.log.Error("failed to insert record into articles table", "error", err.Error())
	}
}

func (r *repository) GetNotPostedArticles(ctx context.Context) ([]*Article, error) {
	var articles []*Article
	query := `SELECT * FROM articles WHERE bot_posted_at IS NULL ORDER BY published_at DESC`
	if err := r.db.SelectContext(ctx, &articles, query); err != nil {
		r.log.Error("failed to select not posted from articles table", "error", err.Error())
		return nil, ErrDatabaseOperation
	}
	return articles, nil
}

func (r *repository) MarkArticleAsPosted(ctx context.Context, id int64) {
	stmt := `UPDATE articles SET bot_posted_at = $1::timestamp WHERE id = $2`
	if _, err := r.db.ExecContext(ctx, stmt, time.Now().UTC(), id); err != nil {
		r.log.Error("failed to mark article as posted", "id", id, "error", err.Error())
	}
}

func (r *repository) DeleteOldArticles(ctx context.Context, days int) {
	stmt := `DELETE FROM articles WHERE bot_posted_at < NOW() - make_interval(days => $1)`
	if _, err := r.db.ExecContext(ctx, stmt, days); err != nil {
		r.log.Error("failed to delete old articles", "error", err.Error())
	}
}
