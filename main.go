package main

import (
	"context"
	"errors"
	"log/slog"
	"os/signal"
	"sync"
	"syscall"
	"tg-bot-news/app"
	"tg-bot-news/config"
	"tg-bot-news/db"
	"tg-bot-news/logger"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("failed to load .env file: " + err.Error())
	}
	cfg := config.MustLoad()
	log := logger.Setup(cfg)

	dbConn := db.MustInitDB(cfg, log)
	defer dbConn.Close()

	botAPI, err := tgbotapi.NewBotAPI(cfg.TGBotToken)
	if err != nil {
		panic("failed to create bot API: " + err.Error())
	}
	repo := app.NewRepository(dbConn, log)
	fetcher := app.NewFetcher(repo, cfg.FetchInterval, log)
	publisher := app.NewPublisher(
		repo,
		app.NewOpenAIsummarizer(cfg.OpenAIKey, log),
		botAPI,
		cfg.PublishInterval,
		cfg.TGChannelID,
		log,
	)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	notifyPublisherCh := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(1)
	go func(ctx context.Context, fetcher *app.Fetcher, log *slog.Logger, stop context.CancelFunc, ch chan struct{}) {
		defer wg.Done()
		log.Info("starting fetcher")
		if err := fetcher.Start(ctx, notifyPublisherCh); err != nil {
			if errors.Is(err, context.Canceled) {
				log.Info("fetcher got canceled")
				return
			} else {
				log.Error("fetcher error", "error", err.Error())
				stop()
			}
		}
	}(ctx, fetcher, log, stop, notifyPublisherCh)

	wg.Add(1)
	go func(ctx context.Context, publisher *app.Publisher, log *slog.Logger) {
		defer wg.Done()
		<-notifyPublisherCh
		close(notifyPublisherCh)
		log.Info("starting publisher")
		if err := publisher.Start(ctx); err != nil {
			log.Info("publisher got canceled")
			return
		}
	}(ctx, publisher, log)

	wg.Add(1)
	go func(ctx context.Context, days int, repo app.IRepository, log *slog.Logger) {
		defer wg.Done()
		log.Info("starting periodic task - deleting old articles")
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Info("deleting old articles task got canceled")
				return
			case <-ticker.C:
				repo.DeleteOldArticles(ctx, days)
			}
		}
	}(ctx, cfg.DaysClearOld, repo, log)

	<-ctx.Done()
	log.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	quit := make(chan struct{})
	go func() {
		wg.Wait()
		close(quit)
	}()

	select {
	case <-quit:
		log.Info("graceful shutdown completed")
	case <-shutdownCtx.Done():
		log.Warn("forced shutdown due to timeout")
	}
}
