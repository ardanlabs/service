// Package statsviz allows visualizing Go runtime metrics data in real time in
// your browser.
//
// Register a Statsviz HTTP handlers with your server's [http.ServeMux]
// (preferred method):
//
//	mux := http.NewServeMux()
//	statsviz.Register(mux)
//
// Alternatively, you can register with [http.DefaultServeMux]:
//
//	ss := statsviz.Server{}
//	s.Register(http.DefaultServeMux)
//
// By default, Statsviz is served at http://host:port/debug/statsviz/. This, and
// other settings, can be changed by passing some [Option] to [NewServer].
//
// If your application is not already running an HTTP server, you need to start
// one. Add "net/http" and "log" to your imports, and use the following code in
// your main function:
//
//	go func() {
//	    log.Println(http.ListenAndServe("localhost:8080", nil))
//	}()
//
// Then open your browser and visit http://localhost:8080/debug/statsviz/.
//
// # Advanced usage:
//
// If you want more control over Statsviz HTTP handlers, examples are:
//   - you're using some HTTP framework
//   - you want to place Statsviz handler behind some middleware
//
// then use [NewServer] to obtain a [Server] instance. Both the [Server.Index] and
// [Server.Ws]() methods return [http.HandlerFunc].
//
//	srv, err := statsviz.NewServer(); // Create server or handle error
//	srv.Index()                       // UI (dashboard) http.HandlerFunc
//	srv.Ws()                          // Websocket http.HandlerFunc
package statsviz

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/arl/statsviz/internal/plot"
	"github.com/arl/statsviz/internal/static"
)

const (
	defaultRoot         = "/debug/statsviz"
	defaultSendInterval = time.Second
)

// RegisterDefault registers the Statsviz HTTP handlers on [http.DefaultServeMux].
//
// RegisterDefault should not be used in production.
func RegisterDefault(opts ...Option) error {
	return Register(http.DefaultServeMux, opts...)
}

// Register registers the Statsviz HTTP handlers on the provided mux.
func Register(mux *http.ServeMux, opts ...Option) error {
	srv, err := NewServer(opts...)
	if err != nil {
		return err
	}
	srv.Register(mux)
	return nil
}

// Server is the core component of Statsviz. It collects and periodically
// updates metrics data and provides two essential HTTP handlers:
//   - the Index handler serves Statsviz user interface, allowing you to
//     visualize runtime metrics on your browser.
//   - The Ws handler establishes a WebSocket connection allowing the connected
//     browser to receive metrics updates from the server.
//
// The zero value is not a valid Server, use NewServer to create a valid one.
type Server struct {
	interval  time.Duration // interval between consecutive metrics emission
	root      string        // HTTP path root
	plots     *plot.List    // plots shown on the user interface
	userPlots []plot.UserPlot
}

// NewServer constructs a new Statsviz Server with the provided options, or the
// default settings.
//
// Note that once the server is created, its HTTP handlers needs to be registered
// with some HTTP server. You can either use the Register method or register yourself
// the Index and Ws handlers.
func NewServer(opts ...Option) (*Server, error) {
	s := &Server{
		interval: defaultSendInterval,
		root:     defaultRoot,
	}

	for _, opt := range opts {
		if err := opt(s); err != nil {
			return nil, err
		}
	}

	pl, err := plot.NewList(s.userPlots)
	if err != nil {
		return nil, err
	}
	s.plots = pl
	return s, nil
}

// Option is a configuration option for the Server.
type Option func(*Server) error

// SendFrequency changes the interval between successive acquisitions of metrics
// and their sending to the user interface. The default interval is one second.
func SendFrequency(intv time.Duration) Option {
	return func(s *Server) error {
		if intv <= 0 {
			return fmt.Errorf("frequency must be a positive integer")
		}
		s.interval = intv
		return nil
	}
}

// Root changes the root path of the Statsviz user interface.
// The default is "/debug/statsviz".
func Root(path string) Option {
	return func(s *Server) error {
		s.root = path
		return nil
	}
}

// TimeseriesPlot adds a new time series plot to Statsviz. This options can
// be added multiple times.
func TimeseriesPlot(tsp TimeSeriesPlot) Option {
	return func(s *Server) error {
		s.userPlots = append(s.userPlots, plot.UserPlot{Scatter: tsp.timeseries})
		return nil
	}
}

// Register registers the Statsviz HTTP handlers on the provided mux.
func (s *Server) Register(mux *http.ServeMux) {
	mux.Handle(s.root+"/", s.Index())
	mux.HandleFunc(s.root+"/ws", s.Ws())
}

// Index returns the index handler, which responds with the Statsviz user
// interface HTML page. By default, the handler is served at the path specified
// by the root. Use [WithRoot] to change the path.
func (s *Server) Index() http.HandlerFunc {
	prefix := strings.TrimSuffix(s.root, "/") + "/"
	dist := http.FileServerFS(static.Assets())
	return http.StripPrefix(prefix, dist).ServeHTTP
}

func parseBoolEnv(name string) bool {
	env := os.Getenv(name)
	val, err := strconv.ParseBool(env)
	if err != nil {
		if env != "" {
			fmt.Fprintf(os.Stderr, "statsviz: malformed %s %v\n", name, err)
		}
	}
	return val
}

var debug = false

var wsUpgrader = sync.OnceValue(func() websocket.Upgrader {
	var checkOrigin func(r *http.Request) bool

	// Allow all origins for testing.
	if debug = parseBoolEnv("STATSVIZ_DEBUG"); debug {
		// passthrough
		checkOrigin = func(r *http.Request) bool { return true }
	}

	return websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 2048,
		CheckOrigin:     checkOrigin,
	}
})

// Ws returns the WebSocket handler used by Statsviz to send application
// metrics. The underlying net.Conn is used to upgrade the HTTP server
// connection to the WebSocket protocol.
func (s *Server) Ws() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		upgrader := wsUpgrader()
		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			if debug {
				fmt.Fprintf(os.Stderr, "statsviz: failed to upgrade connection: %v\n", err)
			}
			return
		}

		defer ws.Close()

		// Ignore websocket errors here. They mainly happen when the other end
		// of the connection closes. We can't handle them in any meaningful way
		// anyways, and since we're a library we don't want to spam the program
		// output streams. If really necessary, we could export a io.Writer the
		// user could connect to whatever logging facility they use, that seems
		// overkill for now.

		if err := s.sendConfig(ws); err != nil {
			if debug {
				fmt.Fprintf(os.Stderr, "statsviz: failed to send config: %v\n", err)
			}
		}
		if err := s.sendStats(ws, s.interval); err != nil {
			if debug {
				fmt.Fprintf(os.Stderr, "statsviz: failed to send stats: %v\n", err)
			}
		}
	}
}

type wsmsg struct {
	Event string `json:"event"`
	Data  any    `json:"data"`
}

func (s *Server) sendConfig(conn *websocket.Conn) error {
	return conn.WriteJSON(wsmsg{
		Event: "config",
		Data:  s.plots.Config(),
	})
}

// sendStats sends runtime statistics over the WebSocket connection.
func (s *Server) sendStats(conn *websocket.Conn, frequency time.Duration) error {
	tick := time.NewTicker(frequency)
	defer tick.Stop()

	for range tick.C {
		w, err := conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return err
		}
		if err := s.plots.WriteValues(w); err != nil {
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}
	}

	panic("unreachable")
}
