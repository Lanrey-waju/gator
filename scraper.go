package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/Lanrey-waju/gator.git/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

func scrapeFeeds(s *state) error {
	var err error
	feeds, err := s.db.GetNextFeedsToFetch(context.Background(), 3)
	if err != nil {
		return fmt.Errorf("error fetching feeds from database: %v", err)
	}
	for _, feed := range feeds {
		s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
			LastFetchedAt: sql.NullTime{Time: time.Now().UTC(), Valid: true},
			UpdatedAt:     time.Now().UTC(),
			ID:            feed.ID,
		})
		rssFeed, err := fetchFeed(context.Background(), feed.Url)
		if err != nil {
			return fmt.Errorf("error fetching feed %s over the network: %v", feed.Url, err)
		}
		for _, item := range rssFeed.Channel.Item {
			publishedAt, err := parsePubDate(item)
			if err != nil {
				return fmt.Errorf("error parsing publish date: %v", err)
			}
			post, err := s.db.CreatePost(context.Background(), database.CreatePostParams{
				ID:          uuid.New(),
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
				Title:       item.Title,
				Url:         item.Link,
				Description: sql.NullString{String: item.Description, Valid: true},
				PublishedAt: sql.NullTime{Time: publishedAt, Valid: true},
				FeedID:      feed.ID,
			})
			if err != nil {
				if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
					// 23505 is the error code for unique violation
					fmt.Printf("duplicate post, URL already exists: %v\n", item.Title)
					continue
				} else {
					return fmt.Errorf("error creating post %v: %v", post.Title, err)
				}
			}

		}

	}
	return nil
}

func parsePubDate(item RSSItem) (time.Time, error) {
	timeFormats := []string{time.RFC1123, time.RFC1123Z, time.RFC3339, time.RFC3339Nano, time.RFC822, time.RFC822Z, time.RFC850}
	var parsedTime time.Time
	var err error
	for _, timeFormat := range timeFormats {
		parsedTime, err = time.Parse(timeFormat, item.PubDate)
		if err == nil {
			return parsedTime, err
		}
	}
	return time.Time{}, fmt.Errorf("error parsing date for feed: %v", err)
}
