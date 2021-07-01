package main

import (
	"os"
	"runtime"

	"github.com/ardanlabs/service/business/config"
	"github.com/ardanlabs/service/foundation/logger"
	"go.uber.org/automaxprocs/maxprocs"
	"go.uber.org/zap"
)

/*
Need to figure out timeouts for http service.
*/

// build is the git version of this program. It is set using build flags in the makefile.
var build = "develop"

func main() {

	// Construct the application logger.
	log := logger.New("SALES-API")
	defer log.Sync()

	// Make sure the program is using the correct
	// number of threads if a CPU quota is set.
	if _, err := maxprocs.Set(); err != nil {
		log.Errorw("startup", zap.Error(err))
		os.Exit(1)
	}
	log.Infow("startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// Perform the startup and shutdown sequence.
	if err := config.RunApi(log, build); err != nil {
		log.Errorw("startup", "ERROR", err)
		os.Exit(1)
	}
}
