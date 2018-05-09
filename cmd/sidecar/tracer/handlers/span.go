package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/ardanlabs/service/internal/platform/web"
	"go.opencensus.io/trace"
)

// Span represents the API to collect span data.
type Span struct {
	// ADD OTHER STATE LIKE THE LOGGER AND CONFIG HERE.
}

// Publish takes a batch and publishes that to a host system.
func (s *Span) Publish(ctx context.Context, w http.ResponseWriter, r *http.Request, params map[string]string) error {
	var sd []trace.SpanData
	if err := json.NewDecoder(r.Body).Decode(&sd); err != nil {
		log.Println("*********>", err)
	}
	log.Printf("*********> %+v\n", sd)

	// SEND THIS TO ZIPKIN

	web.Respond(ctx, w, nil, http.StatusNoContent)

	return nil
}
