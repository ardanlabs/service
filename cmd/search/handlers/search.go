package handlers

import (
	"context"
	"log"
	"net/http"

	"github.com/ardanlabs/service/cmd/search/internal/feeds"

	"github.com/ardanlabs/service/cmd/search/internal/search"
	"github.com/ardanlabs/service/cmd/search/views"
	"github.com/ardanlabs/service/internal/platform/web"
)

// Search provides support for orchestration searches.
type Search struct {
	log *log.Logger
}

// NewSearch constructs a Search for a given set of feeds.
func NewSearch(log *log.Logger) *Search {
	return &Search{
		log: log,
	}
}

// Query performs a search against the datastore.
func (s *Search) Query(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	fv, options := formValues(r)

	var results []views.Result
	if r.Method == "POST" && options.Term != "" {
		var searchers []search.Searcher
		if options.CNN {
			searchers = append(searchers, feeds.NewCNN())
		}
		if options.BBC {
			searchers = append(searchers, feeds.NewBBC())
		}
		if options.NYT {
			searchers = append(searchers, feeds.NewNYT())
		}
		results = search.Submit(ctx, s.log, options.Term, options.First, searchers)
	}

	markup, err := views.Render(fv, results)
	if err != nil {
		return err
	}

	web.RespondHTML(ctx, w, markup, http.StatusOK)
	return nil
}

// formValues extracts the form data.
func formValues(r *http.Request) (map[string]interface{}, views.Options) {
	fv := make(map[string]interface{})
	var options views.Options

	fv["term"] = r.FormValue("term")
	options.Term = r.FormValue("term")

	if r.FormValue("cnn") == "on" {
		fv["cnn"] = "checked"
		options.CNN = true
	} else {
		fv["cnn"] = ""
	}

	if r.FormValue("nyt") == "on" {
		fv["nyt"] = "checked"
		options.NYT = true
	} else {
		fv["nyt"] = ""
	}

	if r.FormValue("bbc") == "on" {
		fv["bbc"] = "checked"
		options.BBC = true
	} else {
		fv["bbc"] = ""
	}

	if r.FormValue("first") == "on" {
		fv["first"] = "checked"
		options.First = true
	} else {
		fv["first"] = ""
	}

	return fv, options
}
