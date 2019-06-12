package feeds

import (
	"context"
	"log"

	"github.com/ardanlabs/service/cmd/search/internal/search"
)

var cnnFeeds = []string{
	"http://rss.cnn.com/rss/cnn_topstories.rss",
	"http://rss.cnn.com/rss/cnn_world.rss",
	"http://rss.cnn.com/rss/cnn_us.rss",
	"http://rss.cnn.com/rss/cnn_allpolitics.rss",
}

// CNN provides support for CNN searches.
type CNN struct{}

// NewCNN returns a CNN Searcher value.
func NewCNN() CNN {
	return CNN{}
}

// Search performs a search against the CNN RSS feeds.
func (cnn CNN) Search(ctx context.Context, log *log.Logger, term string) ([]search.Match, error) {
	matches := []search.Match{}

	for _, feed := range cnnFeeds {
		match, err := rssSearch(ctx, log, term, cnn.FeedName(), feed)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match...)
	}

	return matches, nil
}

// FeedName provides the name of this feed for logging.
func (CNN) FeedName() string {
	return "CNN"
}
