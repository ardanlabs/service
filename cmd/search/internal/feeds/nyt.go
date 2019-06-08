package feeds

import (
	"context"
	"log"

	"github.com/ardanlabs/service/cmd/search/views"
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
func (nyt NYT) Search(ctx context.Context, log *log.Logger, term string) ([]views.Result, error) {
	results := []views.Result{}

	for _, feed := range nytFeeds {
		res, err := rssSearch(ctx, log, term, nyt.FeedName(), feed)
		if err != nil {
			log.Println("ERROR: ", err)
			return nil, err
		}
		results = append(results, res...)
	}

	return results, nil
}

// FeedName provides the name of this feed for logging.
func (NYT) FeedName() string {
	return "NYT"
}
