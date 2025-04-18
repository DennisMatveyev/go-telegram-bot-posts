package db

import (
	"database/sql"
	"log/slog"
	"tg-bot-news/config"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

func MustInitDB(cfg config.Config, log *slog.Logger) *sqlx.DB {
	db, err := sqlx.Connect("postgres", cfg.DBUrl)
	if err != nil {
		panicError("failed to open database connection", err, log)
	}
	applyMigrations(db.DB, log)

	var exists int
	err = db.Get(&exists, "SELECT 1 FROM feeds LIMIT 1")
	if err != nil {
		if err == sql.ErrNoRows {
			seedFeeds(db, log)
			log.Info("initial data seeded into feeds table")
		} else {
			panicError("failed to check feeds table", err, log)
		}
	}
	log.Info("database connection established, migrations applied")

	return db
}

func applyMigrations(db *sql.DB, log *slog.Logger) {
	err := goose.SetDialect("postgres")
	if err != nil {
		panicError("failed to set goose dialect", err, log)
	}
	err = goose.Up(db, "db/migrations")
	if err != nil {
		panicError("failed to apply migrations", err, log)
	}
}

type FeedsSeed struct {
	Feeds []struct {
		Name string `yaml:"name"`
		URL  string `yaml:"url"`
	} `yaml:"feeds"`
}

func seedFeeds(db *sqlx.DB, log *slog.Logger) {
	var source FeedsSeed
	if err := cleanenv.ReadConfig("db/feeds_seed.yaml", &source); err != nil {
		panicError("failed to read feeds seed file", err, log)
	}
	query := `
		INSERT INTO feeds (name, url)
		VALUES ($1, $2)
		ON CONFLICT (url) DO NOTHING;
	`
	for _, feed := range source.Feeds {
		_, err := db.Exec(query, feed.Name, feed.URL)
		if err != nil {
			panicError("failed to insert initial data into feeds table", err, log)
		}
	}
}

func panicError(msg string, err error, log *slog.Logger) {
	log.Error(msg, "error", err.Error())
	panic(msg + ": " + err.Error())
}
