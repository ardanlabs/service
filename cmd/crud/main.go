package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ardanlabs/service/cmd/crud/handlers"
	"github.com/ardanlabs/service/internal/platform/db"
)

/*
hey -m GET -c 10 -n 10000 "http://localhost:3000/v1/users"
expvarmon -ports=":4000" -vars="requests,goroutines,errors,mem:memstats.Alloc"
*/

/*
Need to figure out timeouts for http service.
You might want to reset your DB_HOST env var during test tear down
*/

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {

	// ============================================================
	// Configuration

	readTimeout := 5 * time.Second
	writeTimeout := 10 * time.Second
	shutdownTimeout := 5 * time.Second
	dbDialTimeout := 5 * time.Second
	apiHost := os.Getenv("API_HOST")
	if apiHost == "" {
		apiHost = ":3000"
	}
	debugHost := os.Getenv("DEBUG_HOST")
	if debugHost == "" {
		debugHost = ":4000"
	}
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "got:got2015@ds039441.mongolab.com:39441/gotraining"
	}

	// ============================================================
	// Start Mongo

	log.Println("main : Started : Initialize Mongo")
	masterDB, err := db.New(dbHost, dbDialTimeout)
	if err != nil {
		log.Fatalf("startup : Register DB : %v", err)
	}
	defer masterDB.Close()

	// ============================================================
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
		log.Printf("startup : Debug Listening %s", debugHost)
		log.Printf("shutdown : Debug Listener closed : %v", debug.ListenAndServe())
	}()

	// ============================================================
	// Start API Service

	api := http.Server{
		Addr:           apiHost,
		Handler:        handlers.API(masterDB),
		ReadTimeout:    readTimeout,
		WriteTimeout:   writeTimeout,
		MaxHeaderBytes: 1 << 20,
	}

	// Starting the service, listening for requests.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		log.Printf("startup : API Listening %s", apiHost)
		log.Printf("shutdown : API Listener closed : %v", api.ListenAndServe())
		wg.Done()
	}()

	// ============================================================
	// Shutdown

	// Blocking main and waiting for shutdown.
	osSignals := make(chan os.Signal, 1)
	signal.Notify(osSignals, os.Interrupt, syscall.SIGTERM)
	<-osSignals

	// Create context for Shutdown call.
	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	// Asking listener to shutdown and load shed.
	if err := api.Shutdown(ctx); err != nil {
		log.Printf("shutdown : Graceful shutdown did not complete in %v : %v", shutdownTimeout, err)

		if err := api.Close(); err != nil {
			log.Printf("shutdown : Error killing server : %v", err)
		}
	}

	// Waiting for service to complete that load shedding.
	wg.Wait()
	log.Println("main : Completed")
}
