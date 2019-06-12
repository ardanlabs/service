package feeds

import (
	"context"
	"log"

	"github.com/ardanlabs/service/cmd/search/internal/search"
)

var nytFeeds = []string{
	"http://rss.nytimes.com/services/xml/rss/nyt/HomePage.xml",
	"http://rss.nytimes.com/services/xml/rss/nyt/US.xml",
	"http://rss.nytimes.com/services/xml/rss/nyt/Politics.xml",
	"http://rss.nytimes.com/services/xml/rss/nyt/Business.xml",
}

// NYT provides support for NYT searches.
type NYT struct{}

// NewNYT returns a NYT Searcher value.
func NewNYT() NYT {
	return NYT{}
}

// Search performs a search against the NYT RSS feeds.
func (nyt NYT) Search(ctx context.Context, log *log.Logger, term string) ([]search.Match, error) {
	matches := []search.Match{}

	for _, feed := range nytFeeds {
		match, err := rssSearch(ctx, log, term, nyt.FeedName(), feed)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match...)
	}

	return matches, nil
}

// FeedName provides the name of this feed for logging.
func (NYT) FeedName() string {
	return "NYT"
}
