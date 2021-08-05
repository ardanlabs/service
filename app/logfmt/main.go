// This program can take the structured log output and make it readable.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var service string

func init() {
	flag.StringVar(&service, "service", "", "filter which service to see")
}

func main() {
	flag.Parse()
	scanner := bufio.NewScanner(os.Stdin)
	var b strings.Builder
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

		traceID := "00000000000000000000000000000000"
		if v, ok := m["traceid"]; ok {
			traceID = fmt.Sprintf("%v", v)
		}

		var level string
		if v, ok := m["level"]; ok {
			level = fmt.Sprintf("%v", v)
		}

		var ts string
		if v, ok := m["ts"]; ok {
			ts = fmt.Sprintf("%v", v)
		}

		var caller string
		if v, ok := m["caller"]; ok {
			caller = fmt.Sprintf("%v", v)
		}

		var msg string
		if v, ok := m["msg"]; ok {
			msg = fmt.Sprintf("%v", v)
		}

		b.Reset()
		b.WriteString(fmt.Sprintf("%s: %s: %s: %s: %s: %s: ", service, level, ts, traceID, caller, msg))
		for k, v := range m {
			switch k {
			case "traceid", "service", "level", "caller", "msg", "ts":
				continue
			}
			b.WriteString(fmt.Sprintf("%v: ", v))
		}
		out := b.String()
		fmt.Println(out[:len(out)-2])
	}

	if err := scanner.Err(); err != nil {
		log.Println(err)
	}
}
