// This program can take the structured log output and make it readable.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		s := scanner.Text()
		m := make(map[string]interface{})
		err := json.Unmarshal([]byte(s), &m)
		if err != nil {
			fmt.Println(s)
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
