package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/ardanlabs/service/app/tooling/docs/output/html"
	"github.com/ardanlabs/service/app/tooling/docs/output/json"
	"github.com/ardanlabs/service/app/tooling/docs/output/text"
	"github.com/ardanlabs/service/app/tooling/docs/webapi"
)

var output = flag.String("out", "text", "json, text, html")

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

	switch *output {
	case "text":
		err = text.Transform(records)

	case "json":
		err = json.Transform(records)

	case "html":
		err = html.Transform(records)
	}

	if err != nil {
		return fmt.Errorf("transform, %w", err)
	}

	return nil
}
