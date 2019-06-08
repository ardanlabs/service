package views

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
)

// These constants represent the different layouts in use.
const (
	RESULTS = "Results"
	SEARCH  = "Search"
	LAYOUT  = "Layout"
)

// Options ...
type Options struct {
	Term  string
	CNN   bool
	NYT   bool
	BBC   bool
	First bool
}

// Result ...
type Result struct {
	Engine  string
	Title   string
	Link    string
	Content string
}

// TitleHTML fixes encoding issues. The templates expect this
// method for rendering.
func (r *Result) TitleHTML() template.HTML {
	return template.HTML(r.Title)
}

// ContentHTML fixes encoding issues. The templates expect this
// method for rendering.
func (r *Result) ContentHTML() template.HTML {
	return template.HTML(r.Content)
}

// views contains a map of static templates for rendering views.
var views = make(map[string]*template.Template)

// Render generates the HTML response for this route.
func Render(fv map[string]interface{}, results interface{}) ([]byte, error) {
	var markup bytes.Buffer

	// Generate the markup for the results template.
	if results != nil {
		vars := map[string]interface{}{"Items": results}
		if err := views[RESULTS].Execute(&markup, vars); err != nil {
			return nil, fmt.Errorf("error processing results template : %v", err)
		}
		fv[RESULTS] = template.HTML(markup.String())
	}

	// Generate the markup for the search template.
	markup.Reset()
	if err := views[SEARCH].Execute(&markup, fv); err != nil {
		return nil, fmt.Errorf("error processing search template : %v", err)
	}

	// Generate the final markup with the layout template.
	vars := map[string]interface{}{"LayoutContent": template.HTML(markup.String())}
	markup.Reset()
	if err := views[LAYOUT].Execute(&markup, vars); err != nil {
		return nil, fmt.Errorf("error processing layout template : %v", err)
	}

	return markup.Bytes(), nil
}

// Init loads the existing templates for use to generate views.
func Init() error {

	// In order for the endpoint tests to run this needs to be
	// physically located. Trying to avoid configuration for now.
	pwd, _ := os.Getwd()
	if err := loadTemplate(LAYOUT, pwd+"/views/basic_layout.html"); err != nil {
		return err
	}

	if err := loadTemplate(SEARCH, pwd+"/views/search.html"); err != nil {
		return err
	}

	if err := loadTemplate(RESULTS, pwd+"/views/results.html"); err != nil {
		return err
	}
	return nil
}

// loadTemplate reads the specified template file for use.
func loadTemplate(name string, path string) error {

	// Read the html template file.
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	// Create a template value for this code.
	tmpl, err := template.New(name).Parse(string(data))
	if err != nil {
		return err
	}

	// Have we processed this file already?
	if _, exists := views[name]; exists {
		return fmt.Errorf("template %s already in use", name)
	}

	// Store the template for use.
	views[name] = tmpl
	return nil
}
