package main

import (
	"context"
	_ "expvar"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/service/cmd/crud/handlers"
	"github.com/ardanlabs/service/internal/platform/cfg"
	"github.com/ardanlabs/service/internal/platform/db"
	itrace "github.com/ardanlabs/service/internal/platform/trace"
	"go.opencensus.io/trace"
)

/*
hey -m GET -c 10 -n 10000 "http://localhost:3000/v1/users"
expvarmon -ports=":4000" -vars="requests,goroutines,errors,mem:memstats.Alloc"
*/

/*
Need to figure out timeouts for http service.
You might want to reset your DB_HOST env var during test tear down.
Add pulling git version from build command line.
Service should start even without a DB running yet.
*/

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {

	// =========================================================================
	// Configuration

	c, err := cfg.New(cfg.EnvProvider{Namespace: "CRUD"})
	if err != nil {
		log.Printf("config : %s. All config defaults in use.", err)
	}
	readTimeout, err := c.Duration("READ_TIMEOUT")
	if err != nil {
		readTimeout = 5 * time.Second
	}
	writeTimeout, err := c.Duration("WRITE_TIMEOUT")
	if err != nil {
		writeTimeout = 5 * time.Second
	}
	shutdownTimeout, err := c.Duration("SHUTDOWN_TIMEOUT")
	if err != nil {
		shutdownTimeout = 5 * time.Second
	}
	dbDialTimeout, err := c.Duration("DB_DIAL_TIMEOUT")
	if err != nil {
		dbDialTimeout = 5 * time.Second
	}
	apiHost, err := c.String("API_HOST")
	if err != nil {
		apiHost = "0.0.0.0:3000"
	}
	debugHost, err := c.String("DEBUG_HOST")
	if err != nil {
		debugHost = "0.0.0.0:4000"
	}
	dbHost, err := c.String("DB_HOST")
	if err != nil {
		dbHost = "got:got2015@ds039441.mongolab.com:39441/gotraining"
	}
	traceHost, err := c.String("TRACE_HOST")
	if err != nil {
		traceHost = "http://0.0.0.0:5000"
	}
	traceBatchSize, err := c.Int("TRACE_BATCH_SIZE")
	if err != nil {
		traceBatchSize = 1000
	}
	traceInterval, err := c.Duration("TRACE_INTERVAL")
	if err != nil {
		traceInterval = time.Minute
	}

	log.Printf("config : %s=%v", "READ_TIMEOUT", readTimeout)
	log.Printf("config : %s=%v", "WRITE_TIMEOUT", writeTimeout)
	log.Printf("config : %s=%v", "SHUTDOWN_TIMEOUT", shutdownTimeout)
	log.Printf("config : %s=%v", "DB_DIAL_TIMEOUT", dbDialTimeout)
	log.Printf("config : %s=%v", "API_HOST", apiHost)
	log.Printf("config : %s=%v", "DEBUG_HOST", debugHost)
	log.Printf("config : %s=%v", "DB_HOST", dbHost)
	log.Printf("config : %s=%v", "TRACE_HOST", traceHost)
	log.Printf("config : %s=%v", "TRACE_BATCH_SIZE", traceBatchSize)
	log.Printf("config : %s=%v", "TRACE_INTERVAL", traceInterval)

	// =========================================================================
	// Start Mongo

	log.Println("main : Started : Initialize Mongo")
	masterDB, err := db.New(dbHost, dbDialTimeout)
	if err != nil {
		log.Fatalf("main : Register DB : %v", err)
	}
	defer masterDB.Close()

	// =========================================================================
	// Start Tracing Support

	log.Printf("main : Tracing Started : %s", traceHost)
	exporter := itrace.NewExporter(traceHost, traceBatchSize, traceInterval)
	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	// =========================================================================
	// Start Debug Service

	// /debug/vars - Added to the default mux by the expvars package.
	// /debug/pprof - Added to the default mux by the net/http/pprof package.

	debug := http.Server{
		Addr:           debugHost,
		Handler:        http.DefaultServeMux,
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Not concerned with shutting this down when the
	// application is being shutdown.
	go func() {
		log.Printf("main : Debug Listening %s", debugHost)
		log.Printf("main : Debug Listener closed : %v", debug.ListenAndServe())
	}()

	// =========================================================================
	// Start API Service

	api := http.Server{
		Addr:           apiHost,
		Handler:        handlers.API(masterDB),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Printf("main : API Listening %s", apiHost)
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	log.Println("main : Started")
	defer log.Println("main : Completed")

	// =========================================================================
	// Stop API Service

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("main : Error starting server: %v", err)

	case <-osSignals:
		log.Println("main : Start shutdown...")

		// Create context for Shutdown call.
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		if err := api.Shutdown(ctx); err != nil {
			log.Printf("main : Graceful shutdown did not complete in %v : %v", shutdownTimeout, err)
			if err := api.Close(); err != nil {
				log.Fatalf("main : Could not stop http server: %v", err)
			}
		}
	}
}
