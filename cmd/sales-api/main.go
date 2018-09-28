package main

import (
	"context"
	"encoding/json"
	_ "expvar"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ardanlabs/service/cmd/sales-api/handlers"
	"github.com/ardanlabs/service/internal/platform/db"
	"github.com/ardanlabs/service/internal/platform/flag"
	itrace "github.com/ardanlabs/service/internal/platform/trace"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/kelseyhightower/envconfig"
	"go.opencensus.io/trace"
)

/*
ZipKin: http://localhost:9411
AddLoad: hey -m GET -c 10 -n 10000 "http://localhost:3000/v1/users"
expvarmon -ports=":3001" -endpoint="/metrics" -vars="requests,goroutines,errors,mem:memstats.Alloc"
*/

/*
Need to figure out timeouts for http service.
You might want to reset your DB_HOST env var during test tear down.
Add pulling git version from build command line.
Service should start even without a DB running yet.
symbols in profiles: https://github.com/golang/go/issues/23376 / https://github.com/google/pprof/pull/366
*/

func main() {

	// =========================================================================
	// Logging

	log := log.New(os.Stdout, "SALES : ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)

	// =========================================================================
	// Configuration

	var cfg struct {
		Web struct {
			APIHost         string        `default:"0.0.0.0:3000" envconfig:"API_HOST"`
			DebugHost       string        `default:"0.0.0.0:4000" envconfig:"DEBUG_HOST"`
			ReadTimeout     time.Duration `default:"5s" envconfig:"READ_TIMEOUT"`
			WriteTimeout    time.Duration `default:"5s" envconfig:"WRITE_TIMEOUT"`
			ShutdownTimeout time.Duration `default:"5s" envconfig:"SHUTDOWN_TIMEOUT"`
		}
		DB struct {
			DialTimeout time.Duration `default:"5s" envconfig:"DIAL_TIMEOUT"`
			Host        string        `default:"mongo:27017/gotraining" envconfig:"HOST"`
		}
		Trace struct {
			Host         string        `default:"http://tracer:3002/v1/publish" envconfig:"HOST"`
			BatchSize    int           `default:"1000" envconfig:"BATCH_SIZE"`
			SendInterval time.Duration `default:"15s" envconfig:"SEND_INTERVAL"`
			SendTimeout  time.Duration `default:"500ms" envconfig:"SEND_TIMEOUT"`
		}
		Auth struct {
			KeyID          string `envconfig:"KEY_ID"`
			PrivateKeyFile string `envconfig:"PRIVATE_KEY_FILE"`
			Algorithm      string `default:"RS256" envconfig:"ALGORITHM"`
		}
	}

	if err := envconfig.Process("SALES", &cfg); err != nil {
		log.Fatalf("main : Parsing Config : %v", err)
	}

	if err := flag.Process(&cfg); err != nil {
		if err != flag.ErrHelp {
			log.Fatalf("main : Parsing Command Line : %v", err)
		}
		return // We displayed help.
	}

	// =========================================================================
	// App Starting

	log.Println("main : Started : Application Initializing")
	defer log.Println("main : Completed")

	cfgJSON, err := json.MarshalIndent(cfg, "", "    ")
	if err != nil {
		log.Fatalf("main : Marshalling Config to JSON : %v", err)
	}

	// TODO: Validate what is being written to the logs. We don't
	// want to leak credentials or anything that can be a security risk.
	log.Printf("main : Config : %v\n", string(cfgJSON))

	// =========================================================================
	// Find auth keys

	keyContents, err := ioutil.ReadFile(cfg.Auth.PrivateKeyFile)
	if err != nil {
		log.Fatalf("main : Reading auth private key : %v", err)
	}

	key, err := jwt.ParseRSAPrivateKeyFromPEM(keyContents)
	if err != nil {
		log.Fatalf("main : Parsing auth private key : %v", err)
	}

	userAuth := handlers.UserAuth{
		KeyID: cfg.Auth.KeyID,
		Alg:   cfg.Auth.Algorithm,
		Key:   key,
	}

	// =========================================================================
	// Start Mongo

	log.Println("main : Started : Initialize Mongo")
	masterDB, err := db.New(cfg.DB.Host, cfg.DB.DialTimeout)
	if err != nil {
		log.Fatalf("main : Register DB : %v", err)
	}
	defer masterDB.Close()

	// =========================================================================
	// Start Tracing Support

	logger := func(format string, v ...interface{}) {
		log.Printf(format, v...)
	}

	log.Printf("main : Tracing Started : %s", cfg.Trace.Host)
	exporter, err := itrace.NewExporter(logger, cfg.Trace.Host, cfg.Trace.BatchSize, cfg.Trace.SendInterval, cfg.Trace.SendTimeout)
	if err != nil {
		log.Fatalf("main : RegiTracingster : ERROR : %v", err)
	}
	defer func() {
		log.Printf("main : Tracing Stopping : %s", cfg.Trace.Host)
		batch, err := exporter.Close()
		if err != nil {
			log.Printf("main : Tracing Stopped : ERROR : Batch[%d] : %v", batch, err)
		} else {
			log.Printf("main : Tracing Stopped : Flushed Batch[%d]", batch)
		}
	}()

	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	// =========================================================================
	// Start Debug Service

	// /debug/vars - Added to the default mux by the expvars package.
	// /debug/pprof - Added to the default mux by the net/http/pprof package.

	debug := http.Server{
		Addr:           cfg.Web.DebugHost,
		Handler:        http.DefaultServeMux,
		ReadTimeout:    cfg.Web.ReadTimeout,
		WriteTimeout:   cfg.Web.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Not concerned with shutting this down when the
	// application is being shutdown.
	go func() {
		log.Printf("main : Debug Listening %s", cfg.Web.DebugHost)
		log.Printf("main : Debug Listener closed : %v", debug.ListenAndServe())
	}()

	// =========================================================================
	// Start API Service

	api := http.Server{
		Addr:           cfg.Web.APIHost,
		Handler:        handlers.API(log, masterDB, userAuth),
		ReadTimeout:    cfg.Web.ReadTimeout,
		WriteTimeout:   cfg.Web.WriteTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Make a channel to listen for errors coming from the listener. Use a
	// buffered channel so the goroutine can exit if we don't collect this error.
	serverErrors := make(chan error, 1)

	// Start the service listening for requests.
	go func() {
		log.Printf("main : API Listening %s", cfg.Web.APIHost)
		serverErrors <- api.ListenAndServe()
	}()

	// =========================================================================
	// Shutdown

	// Make a channel to listen for an interrupt or terminate signal from the OS.
	// Use a buffered channel because the signal package requires it.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)

	// =========================================================================
	// Stop API Service

	// Blocking main and waiting for shutdown.
	select {
	case err := <-serverErrors:
		log.Fatalf("main : Error starting server: %v", err)

	case <-osSignals:
		log.Println("main : Start shutdown...")

		// Create context for Shutdown call.
		ctx, cancel := context.WithTimeout(context.Background(), cfg.Web.ShutdownTimeout)
		defer cancel()

		// Asking listener to shutdown and load shed.
		if err := api.Shutdown(ctx); err != nil {
			log.Printf("main : Graceful shutdown did not complete in %v : %v", cfg.Web.ShutdownTimeout, err)
			if err := api.Close(); err != nil {
				log.Fatalf("main : Could not stop http server: %v", err)
			}
		}
	}
}
