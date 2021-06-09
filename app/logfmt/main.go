// This program can take the structured log output and make it readable.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

var service string

func init() {
	flag.StringVar(&service, "service", "", "filter which service to see")
}

func main() {
	flag.Parse()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		s := scanner.Text()
		m := make(map[string]interface{})
		err := json.Unmarshal([]byte(s), &m)
		if err != nil {
			if service == "" {
				fmt.Println(s)
			}
			continue
		}
		if service != "" && m["service"] != service {
			continue
		}
		b, err := json.MarshalIndent(m, "", "    ")
		if err != nil {
			continue
		}
		fmt.Println(string(b))
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}
