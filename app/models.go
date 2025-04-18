package app

import (
	"database/sql"
	"time"
)

type Feed struct {
	ID   int64  `db:"id"`
	Name string `db:"name"`
	URL  string `db:"url"`
}

type Article struct {
	ID          int64        `db:"id"`
	FeedID      int64        `db:"feed_id"`
	Title       string       `db:"title"`
	Link        string       `db:"link"`
	Summary     string       `db:"summary"`
	PublishedAt time.Time    `db:"published_at"`
	BotPostedAt sql.NullTime `db:"bot_posted_at"`
}
