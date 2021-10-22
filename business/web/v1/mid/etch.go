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

// Response represents an individual response record.
type Response struct {
	ID          string    `db:"id"`           // Unique identifier.
	TraceID     string    `db:"trace_id"`     // Trace ID for this response.
	Response    string    `db:"response"`     // Response body sent.
	DateCreated time.Time `db:"date_created"` // When the Response was sent.
}

// Etch writes a response into db for debugging purpose, because writing in logs is insecure
func Etch(log *zap.SugaredLogger, db sqlx.ExtContext) web.Middleware {

	// This is the actual middleware function to be executed.
	m := func(handler web.Handler) web.Handler {

		// Create the handler that will be attached in the middleware chain.
		h := func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {

			// Call the next handler first, which, upon responding,
			// will set the response body in context.
			if err := handler(ctx, w, r); err != nil {
				return err
			}

			// Construct the response struct that would be etched into DB.
			response := Response{
				ID:          validate.GenerateID(),
				TraceID:     web.GetTraceID(ctx),
				Response:    string(web.GetResponse(ctx)),
				DateCreated: time.Now(),
			}

			const q = `
	INSERT INTO responses
		(id, trace_id, response, date_created)
	VALUES
		(:id, :trace_id, :response, :date_created)`

			// This should be done in a best-effort manner.
			_ = database.NamedExecContext(context.Background(), log, db, q, response)

			return nil
		}

		return h
	}

	return m
}
