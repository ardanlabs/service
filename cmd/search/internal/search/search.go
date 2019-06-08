package search

import (
	"context"
	"log"

	"github.com/ardanlabs/service/cmd/search/views"
	"github.com/ardanlabs/service/internal/platform/web"
)

// Searcher declares an interface used to leverage different
// search engines to find results.
type Searcher interface {
	Search(ctx context.Context, log *log.Logger, term string) ([]views.Result, error)
	FeedName() string
}

// Submit runs through the collection of searchers concurrently for better performance.
func Submit(ctx context.Context, log *log.Logger, term string, firstOnly bool, searchers []Searcher) []views.Result {
	g := len(searchers)
	results := make(chan []views.Result, g)

	// Perform the searches concurrently.
	for _, searcher := range searchers {
		go func(searcher Searcher) {
			result, err := searcher.Search(ctx, log, term)
			if err != nil {

				// Important to capture there errors.
				v, ok := ctx.Value(web.KeyValues).(*web.Values)
				if !ok {
					log.Printf("TRACEID MISSING : search : ERROR : %s : searching feed : %v", searcher.FeedName(), err)
				} else {
					log.Printf("%s : search : ERROR : %s : search feed : %v", v.TraceID, searcher.FeedName(), err)
				}
			}

			// Signal the result from this search.
			results <- result

			// Important to capture we are not leaking goroutines.
			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if !ok {
				log.Printf("TRACEID MISSING : search : %s : searching complete", searcher.FeedName())
			} else {
				log.Printf("%s : search : %s : searching complete", v.TraceID, searcher.FeedName())
			}
		}(searcher)
	}

	// Wait for the results to come back.
	var all []views.Result
	for i := 0; i < g; i++ {
		result := <-results

		// If we only want the first result, return it. This
		// code can exit because the buffered channel will prevent
		// the other goroutines from blocking.
		if firstOnly {
			return result
		}

		// Save the results since the user wants them all.
		all = append(all, result...)
	}

	return all
}
