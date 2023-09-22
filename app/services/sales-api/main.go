package main

import (
	"os"

	"github.com/ardanlabs/service/app/services/sales-api/v1/cmd"
	"github.com/ardanlabs/service/app/services/sales-api/v1/cmd/all"
)

/*
	Need to figure out timeouts for http service.
*/

var build = "develop"

func main() {
	if err := cmd.Main(build, all.Routes(), nil, nil, false); err != nil {
		os.Exit(1)
	}
}
