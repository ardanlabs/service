package feeds

import (
	"context"
	"log"

	"github.com/ardanlabs/service/cmd/search/views"
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
func (cnn CNN) Search(ctx context.Context, log *log.Logger, term string) ([]views.Result, error) {
	results := []views.Result{}

	for _, feed := range cnnFeeds {
		res, err := rssSearch(ctx, log, term, cnn.FeedName(), feed)
		if err != nil {
			return nil, err
		}
		results = append(results, res...)
	}

	return results, nil
}

// FeedName provides the name of this feed for logging.
func (CNN) FeedName() string {
	return "CNN"
}
