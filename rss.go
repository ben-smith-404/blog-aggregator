package main

import (
	"context"
	"encoding/xml"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/ben-smith-404/blog-aggregator/internal/database"
	"github.com/google/uuid"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// fetches a feed from a given URL
func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	client := &http.Client{}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	checkError(err)
	request.Header.Set("User-Agent", "gator")
	response, err := client.Do(request)
	checkError(err)
	body, err := io.ReadAll(response.Body)
	checkError(err)
	var feed RSSFeed
	err = xml.Unmarshal(body, &feed)
	checkError(err)
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for x, rssItem := range feed.Channel.Item {
		rssItem.Title = html.UnescapeString(rssItem.Title)
		rssItem.Description = html.UnescapeString(rssItem.Description)
		feed.Channel.Item[x] = rssItem
	}
	return &feed, nil
}

// the almighty feed scraper. Gets the oldest feed in the db, retrieves the feed info using fetchFeed()
// and saves them to the database in the posts table
func scrapeFeeds(s *state) {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	checkError(err)
	err = s.db.MarkFeedFetched(context.Background(), feed.ID)
	checkError(err)
	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	checkError(err)
	for _, rssItem := range rssFeed.Channel.Item {
		pubDate, _ := time.Parse(time.RFC1123Z, rssItem.PubDate)
		err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       rssItem.Title,
			Url:         rssItem.Link,
			PublishedAt: pubDate,
			FeedID:      feed.ID,
		})
	}
}
