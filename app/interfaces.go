package app

import (
	"context"

	"github.com/mmcdole/gofeed"
)

type ISummarizer interface {
	Summarize(ctx context.Context, text string) (string, error)
}

type IRepository interface {
	GetFeeds(ctx context.Context) ([]*Feed, error)
	AddArticle(ctx context.Context, item *gofeed.Item, feedID int64)
	GetNotPostedArticles(ctx context.Context) ([]*Article, error)
	MarkArticleAsPosted(ctx context.Context, id int64)
	DeleteOldArticles(ctx context.Context, days int)
}
