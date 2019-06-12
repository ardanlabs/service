package feeds

import (
	"context"
	"log"

	"github.com/ardanlabs/service/cmd/search/internal/search"
)

var bbcFeeds = []string{
	"http://feeds.bbci.co.uk/news/rss.xml",
	"http://feeds.bbci.co.uk/news/world/rss.xml",
	"http://feeds.bbci.co.uk/news/politics/rss.xml",
	"http://feeds.bbci.co.uk/news/world/us_and_canada/rss.xml",
}

// BBC provides support for BBC searches.
type BBC struct{}

// NewBBC returns a BBC Searcher value.
func NewBBC() BBC {
	return BBC{}
}

// Search performs a search against the BBC RSS feeds.
func (bbc BBC) Search(ctx context.Context, log *log.Logger, term string) ([]search.Match, error) {
	matches := []search.Match{}

	for _, feed := range bbcFeeds {
		match, err := rssSearch(ctx, log, term, bbc.FeedName(), feed)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match...)
	}

	return matches, nil
}

// FeedName provides the name of this feed for logging.
func (BBC) FeedName() string {
	return "BBC"
}
