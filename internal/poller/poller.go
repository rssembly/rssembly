// Package poller implements background feed polling that fetches RSS/Atom feeds,
// creates articles, and schedules the next poll.

package poller

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"

	"github.com/rssembly/rssembly/internal/models"
	"github.com/rssembly/rssembly/internal/repo"
)

// Poller fetches due feeds and creates articles.
type Poller struct {
	feedRepo    *repo.FeedRepo
	articleRepo *repo.ArticleRepo
	httpClient  *http.Client
	fp          *gofeed.Parser
}

// New creates a new Poller.
func New(feedRepo *repo.FeedRepo, articleRepo *repo.ArticleRepo) *Poller {
	return &Poller{
		feedRepo:    feedRepo,
		articleRepo: articleRepo,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		fp: gofeed.NewParser(),
	}
}

// PollOnce fetches all due feeds and creates articles.
// Returns counts for logging/metrics.
func (p *Poller) PollOnce(ctx context.Context) (feedsPolled, articlesCreated int, _ error) {
	feeds, err := p.feedRepo.ListFeedsDueForPoll(ctx, 50)
	if err != nil {
		return 0, 0, fmt.Errorf("list due feeds: %w", err)
	}

	for _, feed := range feeds {
		created, err := p.pollFeed(ctx, feed)
		if err != nil {
			slog.Error("poll feed failed",
				"feed_id", feed.ID,
				"feed_url", feed.FeedURL,
				"error", err,
			)
			feed.Status = models.FeedStatusError
			if _, updateErr := p.feedRepo.UpdateFeed(ctx, feed); updateErr != nil {
				slog.Error("mark feed error failed", "feed_id", feed.ID, "error", updateErr)
			}
			continue
		}

		feedsPolled++
		articlesCreated += created

		slog.Debug("polled feed",
			"feed_id", feed.ID,
			"feed_url", feed.FeedURL,
			"articles", created,
		)
	}

	return feedsPolled, articlesCreated, nil
}

// pollFeed fetches a single feed, parses it, creates new articles, and
// updates the feed's polling metadata.
func (p *Poller) pollFeed(ctx context.Context, feed *models.Feed) (articlesCreated int, _ error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feed.FeedURL, nil)
	if err != nil {
		return 0, fmt.Errorf("create request: %w", err)
	}

	if feed.ETag != "" {
		req.Header.Set("If-None-Match", feed.ETag)
	}
	if feed.LastModified != "" {
		req.Header.Set("If-Modified-Since", feed.LastModified)
	}
	req.Header.Set("User-Agent", "rssembly/0.1.0")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	// 304 Not Modified — no new content.
	if resp.StatusCode == http.StatusNotModified {
		feed.NextPollAt = time.Now().Add(feed.PollInterval)
		feed.LastFetchedAt = timePtr(time.Now())
		if _, err := p.feedRepo.UpdateFeed(ctx, feed); err != nil {
			return 0, fmt.Errorf("update feed: %w", err)
		}
		return 0, nil
	}

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("read body: %w", err)
	}

	parsed, err := p.fp.ParseString(string(body))
	if err != nil {
		return 0, fmt.Errorf("parse feed: %w", err)
	}

	// Update feed metadata from parsed feed.
	if parsed.Title != "" {
		feed.Title = parsed.Title
	}
	if parsed.Description != "" {
		feed.Description = parsed.Description
	}
	if parsed.Link != "" {
		feed.SiteURL = parsed.Link
	}
	if parsed.Image != nil && parsed.Image.URL != "" {
		feed.IconURL = parsed.Image.URL
	}

	if etag := resp.Header.Get("ETag"); etag != "" {
		feed.ETag = etag
	}
	if lm := resp.Header.Get("Last-Modified"); lm != "" {
		feed.LastModified = lm
	}

	feed.Status = models.FeedStatusOK
	feed.LastFetchedAt = timePtr(time.Now())
	feed.NextPollAt = time.Now().Add(feed.PollInterval)
	feed.IsPaused = false

	created := 0
	for _, item := range parsed.Items {
		guid := item.GUID
		if guid == "" && item.Link != "" {
			guid = item.Link
		}
		if guid == "" {
			continue
		}

		article := &models.Article{
			FeedID: feed.ID,
			GUID:   guid,
			URL:    item.Link,
			Title:  item.Title,
		}

		if item.Content != "" {
			article.Content = item.Content
		} else if item.Description != "" {
			article.Content = item.Description
		}

		if item.PublishedParsed != nil {
			article.PublishedAt = item.PublishedParsed
		}

		if _, err := p.articleRepo.CreateArticle(ctx, article); err != nil {
			slog.Warn("create article failed",
				"feed_id", feed.ID,
				"guid", guid,
				"error", err,
			)
			continue
		}
		created++
	}

	if _, err := p.feedRepo.UpdateFeed(ctx, feed); err != nil {
		return created, fmt.Errorf("update feed metadata: %w", err)
	}

	return created, nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}
