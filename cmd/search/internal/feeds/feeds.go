package feeds

import (
	"context"
	"encoding/xml"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ardanlabs/service/cmd/search/internal/search"
	"github.com/ardanlabs/service/internal/platform/web"
	gc "github.com/patrickmn/go-cache"
)

const (
	expiration = time.Minute * 15
	cleanup    = time.Hour
)

var cache = gc.New(expiration, cleanup)

var fetch = struct {
	sync.Mutex
	m map[string]*sync.Mutex
}{
	m: make(map[string]*sync.Mutex),
}

type (

	// Item defines the fields associated with the item tag in the RSS document.
	Item struct {
		XMLName     xml.Name `xml:"item"`
		PubDate     string   `xml:"pubDate"`
		Title       string   `xml:"title"`
		Description string   `xml:"description"`
		Link        string   `xml:"link"`
		Search      struct {
			Description string
		}
	}

	// Image defines the fields associated with the image tag in the RSS document.
	Image struct {
		XMLName xml.Name `xml:"image"`
		URL     string   `xml:"url"`
		Title   string   `xml:"title"`
		Link    string   `xml:"link"`
	}

	// Channel defines the fields associated with the channel tag in the RSS document.
	Channel struct {
		XMLName xml.Name `xml:"channel"`
		Image   Image    `xml:"image"`
		Items   []Item   `xml:"item"`
	}

	// Document defines the fields associated with the RSS document.
	Document struct {
		XMLName xml.Name `xml:"rss"`
		Channel Channel  `xml:"channel"`
	}
)

// rssSearch is used against any RSS feeds.
func rssSearch(ctx context.Context, log *log.Logger, term, engine, uri string) ([]search.Match, error) {
	var mu *sync.Mutex
	fetch.Lock()
	{
		var found bool
		if mu, found = fetch.m[uri]; !found {
			mu = &sync.Mutex{}
			fetch.m[uri] = mu
		}
	}
	fetch.Unlock()

	var d Document
	mu.Lock()
	{
		v, found := cache.Get(uri)
		switch {
		case found:
			d = v.(Document)

		default:
			resp, err := http.Get(uri)
			if err != nil {
				return []search.Match{}, err
			}
			defer resp.Body.Close()

			if err := xml.NewDecoder(resp.Body).Decode(&d); err != nil {
				return []search.Match{}, err
			}

			for i := range d.Channel.Items {
				d.Channel.Items[i].Search.Description = strings.ToLower(d.Channel.Items[i].Description)
			}

			cache.Set(uri, d, expiration)

			traceID := "MISSING"
			v, ok := ctx.Value(web.KeyValues).(*web.Values)
			if ok {
				traceID = v.TraceID
			}
			log.Printf("%s : feeds : %s : reloaded cache : %s", traceID, engine, uri)
		}
	}
	mu.Unlock()

	results := []search.Match{}
	term = strings.ToLower(term)

	for _, item := range d.Channel.Items {
		if strings.Contains(item.Search.Description, term) {
			results = append(results, search.Match{
				Engine:  engine,
				Title:   item.Title,
				Link:    item.Link,
				Content: item.Description,
			})
		}
	}

	return results, nil
}
