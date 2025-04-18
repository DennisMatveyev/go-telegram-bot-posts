package app

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"
)

type Fetcher struct {
	repo           IRepository
	fetchInterval  time.Duration
	log            *slog.Logger
	parser         *gofeed.Parser
	dbFeedsCache   []*Feed
	cacheTimestamp time.Time
	cacheTTL       time.Duration
}

func NewFetcher(
	repo IRepository,
	fetchInterval time.Duration,
	log *slog.Logger,
) *Fetcher {
	return &Fetcher{
		repo:          repo,
		fetchInterval: fetchInterval,
		log:           log,
		parser:        gofeed.NewParser(),
		cacheTTL:      1 * time.Hour,
	}
}

func (f *Fetcher) Start(ctx context.Context, notifyPublisherCh chan struct{}) error {
	ticker := time.NewTicker(f.fetchInterval)
	defer ticker.Stop()

	if err := f.fetch(ctx); err != nil {
		return ErrFirstFetchFailed
	}
	notifyPublisherCh <- struct{}{}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			f.fetch(ctx)
		}
	}
}

func (f *Fetcher) fetch(ctx context.Context) error {
	var feeds []*Feed
	if f.dbFeedsCache != nil && time.Since(f.cacheTimestamp) < f.cacheTTL {
		feeds = f.dbFeedsCache
	} else {
		feedsDB, err := f.repo.GetFeeds(ctx)
		if err != nil && f.dbFeedsCache != nil {
			f.log.Error("fetcher failed to get feeds from database, using cache", "error", err.Error())
			feeds = f.dbFeedsCache
		} else if err != nil {
			f.log.Error("fetcher failed to get feeds from database", "error", err.Error())
			return err
		} else {
			feeds = feedsDB
			f.dbFeedsCache = feedsDB
			f.cacheTimestamp = time.Now()
		}
	}
	var wg sync.WaitGroup
	for _, feed := range feeds {
		wg.Add(1)

		go func(feed *Feed) {
			defer wg.Done()
			parsed, err := f.parser.ParseURL(feed.URL)
			if err != nil {
				f.log.Error("failed to load feed", "url", feed.URL, "error", err.Error())
				return
			}
			for _, item := range parsed.Items {
				item.Description = cleanText(item.Description)
				f.repo.AddArticle(ctx, item, feed.ID)
			}
		}(feed)
	}
	wg.Wait()

	return nil
}
