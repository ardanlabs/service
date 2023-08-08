package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ardanlabs/service/app/tooling/docs/output/json"
	"github.com/ardanlabs/service/app/tooling/docs/output/text"
	"github.com/ardanlabs/service/app/tooling/docs/webapi"
)

var output = flag.String("out", "text", "json, text")

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	records, err := webapi.Records("productgrp")
	if err != nil {
		return fmt.Errorf("findWebAPIRecords, %w", err)
	}

	var out string

	switch *output {
	case "text":
		out, err = text.Transform(records)
	case "json":
		out, err = json.Transform(records)
	}

	if err != nil {
		return fmt.Errorf("transform, %w", err)
	}

	fmt.Print(out)

	return nil
}
