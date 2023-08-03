package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/ardanlabs/service/app/tooling/docs/webapi"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	records, err := webapi.Records("productgrp")
	if err != nil {
		return fmt.Errorf("findWebAPIRecords, %w", err)
	}

	for _, record := range records {
		fmt.Print("\n============================\n\n")

		fmt.Printf("Route  : (%s) %s\n", record.Method, record.Route)
		fmt.Printf("Status : %s (%d)\n", record.Status, webapi.Statuses[record.Status])

		for _, comment := range record.Comments {
			fmt.Println(comment)
		}

		fmt.Print("\n")
		fmt.Println("Input Model :", webapi.ToJSON(record.InputDoc))
		fmt.Print("\n")
		fmt.Println("Output Model :", webapi.ToJSON(record.OutputDoc))
		fmt.Print("\n")
		fmt.Printf("Paging   : %v\n", strings.Join(record.QueryVars.Paging, ", "))
		fmt.Printf("FilterBy : %v\n", strings.Join(record.QueryVars.FilterBy, ", "))
		fmt.Printf("OrderBy  : %v\n", strings.Join(record.QueryVars.OrderBy, ", "))
	}

	return nil
}
