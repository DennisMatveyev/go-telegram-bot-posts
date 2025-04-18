-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS articles (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    link VARCHAR(255) NOT NULL UNIQUE,
    summary TEXT NOT NULL,
    published_at TIMESTAMPTZ NOT NULL,
    bot_posted_at TIMESTAMPTZ,
    feed_id INT NOT NULL,
    CONSTRAINT fk_article_feed FOREIGN KEY (feed_id) REFERENCES feeds(id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS articles;
-- +goose StatementEnd
