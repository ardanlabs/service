package mid

import (
	"context"
	"net/http"
	"time"

	"github.com/ardanlabs/service/business/sys/database"
	"github.com/ardanlabs/service/business/sys/validate"
	"github.com/ardanlabs/service/foundation/web"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// response represents an individual response record.
type response struct {
	ID          string    `db:"id"`           // Unique identifier.
	TraceID     string    `db:"trace_id"`     // Trace ID for this response.
	Response    string    `db:"response"`     // Response body sent.
	DateCreated time.Time `db:"date_created"` // When the Response was sent.
}

// etchWriter writes a response into db for debugging purpose, because writing in logs is insecure
type etchWriter struct {
	http.ResponseWriter
	Logger  *zap.SugaredLogger
	DB      sqlx.ExtContext
	TraceID string
}

// Write extends the default behavior, inserting the response into DB before responding
func (ew etchWriter) Write(resp []byte) (int, error) {

	// Construct the response struct that would be etched into DB.
	r := response{
		ID:          validate.GenerateID(),
		TraceID:     ew.TraceID,
		Response:    string(resp),
		DateCreated: time.Now(),
	}

	// Construct the query to insert the response.
	const q = `
	INSERT INTO responses
		(id, trace_id, response, date_created)
	VALUES
		(:id, :trace_id, :response, :date_created)`

	// Insert the response into DB in best-effort manner.
	_ = database.NamedExecContext(context.Background(), ew.Logger, ew.DB, q, r)

	// Respond in the default way
	return ew.ResponseWriter.Write(resp)
}

// Logger writes some information about the request to the logs in the
// format: TraceID : (200) GET /foo -> IP ADDR (latency)
func Logger(log *zap.SugaredLogger, db sqlx.ExtContext) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// If the context is missing this value, request the service
			// to be shutdown gracefully.
			v, err := web.GetValues(ctx)
			if err != nil {
				return web.NewShutdownError("web value missing from context")
			}

			log.Infow("request started", "traceid", v.TraceID, "method", r.Method, "path", r.URL.Path,
				"remoteaddr", r.RemoteAddr)

			// Construct an enhanced writer, passing in logger and DB handle so that it can etch a response into DB.
			ew := etchWriter{
				ResponseWriter: w,
				Logger:         log,
				DB:             db,
				TraceID:        v.TraceID,
			}

			// Call the next handler with the enhanced writer.
			err = handler(ctx, ew, r)

			log.Infow("request completed", "traceid", v.TraceID, "method", r.Method, "path", r.URL.Path,
				"remoteaddr", r.RemoteAddr, "statuscode", v.StatusCode, "since", time.Since(v.Now))

			// Return the error so it can be handled further up the chain.
			return err
		}

		return h
	}

	return m
}
