// Package html converts the webapi records into html.
package html

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"time"

	"github.com/ardanlabs/service/app/tooling/docs/webapi"
)

//go:embed template.html
var document embed.FS

var uniqueGroups []string

// Transform converts the collection of webapi records into html.
func Transform(records []webapi.Record, browser bool) error {
	lastGroup := records[0].Group
	uniqueGroups = append(uniqueGroups, records[0].Group)
	for _, record := range records {
		if record.Group != lastGroup {
			lastGroup = record.Group
			uniqueGroups = append(uniqueGroups, record.Group)
		}
	}

	p := page{
		records: records,
	}

	http.HandleFunc("/", p.show)

	app := http.Server{
		Addr:    "localhost:8080",
		Handler: http.DefaultServeMux,
	}

	ch := make(chan error, 1)

	go func() {
		ch <- app.ListenAndServe()
	}()

	fmt.Printf("Listening on: %s\n", app.Addr)
	fmt.Println("Hit <ctrl> C to shutdown")
	defer fmt.Println("Shutdown complete")

	if browser {
		startBrowser("http://localhost:8080")
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)

	select {
	case err := <-ch:
		fmt.Println("ERROR:", err)

	case <-shutdown:
		fmt.Println("\nShutdown requested")
	}

	fmt.Println("Starting shutdown")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		app.Close()
	}

	return nil
}

type page struct {
	records []webapi.Record
}

func (p *page) show(w http.ResponseWriter, r *http.Request) {
	var funcMap = template.FuncMap{
		"minus":  minus,
		"status": status,
		"json":   toJSON,
		"groups": groups,
	}

	tmpl, err := template.New("").Funcs(funcMap).ParseFS(document, "template.html")
	if err != nil {
		http.Error(w, "Parse: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "template.html", p.records); err != nil {
		http.Error(w, "Exec:"+err.Error(), http.StatusInternalServerError)
	}
}

func groups() []string {
	return uniqueGroups
}

func minus(a, b int) int {
	return a - b
}

func status(status string) int {
	return webapi.Statuses[status]
}

func toJSON(v any) string {
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return ""
	}

	return string(data)
}

// startBrowser tries to open the URL in a browser, and returns
// whether it succeed.
func startBrowser(url string) bool {
	var args []string

	switch runtime.GOOS {
	case "darwin":
		args = []string{"open"}
	case "windows":
		args = []string{"cmd", "/c", "start"}
	default:
		args = []string{"xdg-open"}
	}

	cmd := exec.Command(args[0], append(args[1:], url)...)
	return cmd.Start() == nil
}
