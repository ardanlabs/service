package main

import (
	"context"
	"log"

	"github.com/ServiceWeaver/weaver"
	"github.com/ardanlabs/service/app/services/sales-api/v1/cmd"
	"github.com/ardanlabs/service/app/services/sales-api/v1/cmd/all"
)

/*
	TODOs:
	* Add secrets API to Service Weaver and use it.
	* More documentation in the dev.toml file or a link where to go.
	* Break things down by domain.
*/

//go:generate weaver generate

const build = "develop"

func main() {
	if err := weaver.Run(context.Background(), serve); err != nil {
		log.Fatal(err)
	}
}

type server struct {
	weaver.Implements[weaver.Main]
	api, debug weaver.Listener
}

func serve(ctx context.Context, s *server) error {
	if err := cmd.MainServiceWeaver(build, all.Routes(), s.debug, s.api); err != nil {
		return err
	}

	return nil
}
