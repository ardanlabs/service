package search

import (
	"context"
	"html/template"
	"log"

	"github.com/ardanlabs/service/internal/platform/web"
)

// Options ...
type Options struct {
	Term  string
	CNN   bool
	NYT   bool
	BBC   bool
	First bool
}

// Match respresents data that matched the search criteria.
type Match struct {
	Engine  string
	Title   string
	Link    string
	Content string
}

// TitleHTML fixes encoding issues during rendering.
func (m *Match) TitleHTML() template.HTML {
	return template.HTML(m.Title)
}

// ContentHTML fixes encoding issues during rendering.
func (m *Match) ContentHTML() template.HTML {
	return template.HTML(m.Content)
}

// Result is the data that respresents the results of a search.
type Result struct {
	Matches []Match
	Err     error
}

// Searcher declares an interface used to leverage different
// search engines to find results.
type Searcher interface {
	Search(ctx context.Context, log *log.Logger, term string) ([]Match, error)
	FeedName() string
}

// Submit runs through the collection of searchers concurrently for better performance.
func Submit(ctx context.Context, log *log.Logger, term string, firstOnly bool, searchers []Searcher) []Result {
	traceID := "MISSING"
	v, ok := ctx.Value(web.KeyValues).(*web.Values)
	if ok {
		traceID = v.TraceID
	}

	g := len(searchers)
	results := make(chan Result, g)

	for _, searcher := range searchers {
		go func(searcher Searcher) {
			matches, err := searcher.Search(ctx, log, term)
			if err != nil {
				results <- Result{Err: err}
				log.Printf("%s : search : %s : searching complete : %v", traceID, searcher.FeedName(), err)
				return
			}

			results <- Result{Matches: matches}
			log.Printf("%s : search : %s : searching complete", traceID, searcher.FeedName())
		}(searcher)
	}

	all := []Result{}
	for i := 0; i < g; i++ {
		result := <-results
		if result.Err != nil {
			continue
		}

		all = append(all, result)

		// If we only want the first result, don't wait for any
		// other. This code can exit because the buffered
		// channel will prevent the other goroutines from blocking.
		if firstOnly {
			break
		}
	}

	return all
}
