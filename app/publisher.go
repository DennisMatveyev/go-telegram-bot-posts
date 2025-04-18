package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"
	"unicode/utf8"

	readability "github.com/go-shiori/go-readability"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Publisher struct {
	repo            IRepository
	summarizer      ISummarizer
	bot             *tgbotapi.BotAPI
	publishInterval time.Duration
	tgChanID        int64
	log             *slog.Logger
}

func NewPublisher(
	repo IRepository,
	summarizer ISummarizer,
	bot *tgbotapi.BotAPI,
	publishInterval time.Duration,
	tgChanID int64,
	log *slog.Logger,
) *Publisher {
	return &Publisher{
		repo:            repo,
		summarizer:      summarizer,
		bot:             bot,
		publishInterval: publishInterval,
		tgChanID:        tgChanID,
		log:             log,
	}
}

func (p *Publisher) Start(ctx context.Context) error {
	ticker := time.NewTicker(p.publishInterval)
	defer ticker.Stop()

	p.publish(ctx)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			p.publish(ctx)
		}
	}
}

func (p *Publisher) publish(ctx context.Context) {
	notPosted, err := p.repo.GetNotPostedArticles(ctx)
	if err != nil || len(notPosted) == 0 {
		return
	}
	for _, article := range notPosted {
		// goroutine might be applicable here, but both TG and OpenAI have rate limits
		summary, err := p.getSummary(ctx, article)
		if err == nil {
			if err := p.makePost(article, summary); err == nil {
				p.repo.MarkArticleAsPosted(ctx, article.ID)
			}
		}
	}
}

func (p *Publisher) getSummary(ctx context.Context, article *Article) (string, error) {
	if utf8.RuneCountInString(article.Summary) > 400 {
		return article.Summary, nil
	}
	content, err := p.getContent(article.Link)
	if err != nil {
		return "", err
	}
	summary, err := p.summarizer.Summarize(ctx, content)
	if err != nil {
		return "", err
	}

	return "\n\n" + summary, nil
}

func (p *Publisher) getContent(link string) (string, error) {
	resp, err := http.Get(link)
	if err != nil {
		p.log.Error("failed to get article content", "url", link, "error", err.Error())
		return "", ErrGetContentFailed
	}
	defer resp.Body.Close()

	articleURL, err := url.Parse(link)
	if err != nil {
		p.log.Error("failed to parse article URL", "url", link, "error", err.Error())
		return "", ErrGetContentFailed
	}
	doc, err := readability.FromReader(resp.Body, articleURL)
	if err != nil {
		p.log.Error("failed to read article content", "url", link, "error", err.Error())
		return "", ErrGetContentFailed
	}
	return cleanText(doc.TextContent), nil
}

func (p *Publisher) makePost(article *Article, summary string) error {
	msg := tgbotapi.NewMessage(
		p.tgChanID,
		fmt.Sprintf(
			"*%s*%s\n\n%s",
			article.Title,
			summary,
			article.Link,
		),
	)
	msg.ParseMode = tgbotapi.ModeMarkdownV2

	_, err := p.bot.Send(msg)
	if err != nil {
		p.log.Error("failed to send message", "error", err.Error())
		return ErrCouldNotMakePost
	}
	return nil
}
