package commands

import (
	"bytes"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"
	"text/template"
)

//go:embed templates
var templates embed.FS

// Domain adds a new domain to the system.
func Domain(name string) error {
	if err := addAppLayer(name); err != nil {
		return fmt.Errorf("adding app layer files: %w", err)
	}

	if err := addBusinessLayer(name); err != nil {
		return fmt.Errorf("adding bus layer files: %w", err)
	}

	if err := addStorageLayer(name); err != nil {
		return fmt.Errorf("adding sto layer files: %w", err)
	}

	if err := addMiddlewareLayer(name); err != nil {
		return fmt.Errorf("adding middleware layer files: %w", err)
	}

	return nil
}

func addAppLayer(domain string) error {
	const basePath = "app/services/sales-api/v1/handlers"

	app, err := fs.Sub(templates, "templates/sales-api/app")
	if err != nil {
		return fmt.Errorf("switching to template/app folder: %w", err)
	}

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		return walkWork(domain, basePath, app, fileName, dirEntry, err)
	}

	fmt.Println("=======================================================")
	fmt.Println("APP LAYER CODE")

	if err := fs.WalkDir(app, ".", fn); err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	return nil
}

func addBusinessLayer(domain string) error {
	const basePath = "business/core"

	app, err := fs.Sub(templates, "templates/sales-api/business")
	if err != nil {
		return fmt.Errorf("switching to template/business folder: %w", err)
	}

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		return walkWork(domain, basePath, app, fileName, dirEntry, err)
	}

	fmt.Println("=======================================================")
	fmt.Println("BUSINESS LAYER CODE")

	if err := fs.WalkDir(app, ".", fn); err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	return nil
}

func addStorageLayer(domain string) error {
	basePath := fmt.Sprintf("business/core/%s/stores", domain)

	app, err := fs.Sub(templates, "templates/sales-api/storage")
	if err != nil {
		return fmt.Errorf("switching to template/storage folder: %w", err)
	}

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		return walkWork(domain, basePath, app, fileName, dirEntry, err)
	}

	fmt.Println("=======================================================")
	fmt.Println("STORAGE LAYER CODE")

	if err := fs.WalkDir(app, ".", fn); err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	return nil
}

func addMiddlewareLayer(domain string) error {
	basePath := "business/web/v1/mid"

	app, err := fs.Sub(templates, "templates/sales-api/mid")
	if err != nil {
		return fmt.Errorf("switching to template/mid folder: %w", err)
	}

	fn := func(fileName string, dirEntry fs.DirEntry, err error) error {
		return walkWork(domain, basePath, app, fileName, dirEntry, err)
	}

	fmt.Println("=======================================================")
	fmt.Println("MIDDLEWARE LAYER CODE")

	if err := fs.WalkDir(app, ".", fn); err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	return nil
}

func walkWork(domain string, basePath string, app fs.FS, fileName string, dirEntry fs.DirEntry, err error) error {
	if err != nil {
		return fmt.Errorf("walkdir failure: %w", err)
	}

	if dirEntry.IsDir() {
		return nil
	}

	f, err := app.Open(fileName)
	if err != nil {
		return fmt.Errorf("opening %s: %w", fileName, err)
	}

	data, err := io.ReadAll(f)
	if err != nil {
		return fmt.Errorf("reading %s: %w", fileName, err)
	}

	tmpl := template.Must(template.New("code").Parse(string(data)))

	domainVar := domain
	for _, c := range []string{"a", "e", "i", "o", "u"} {
		domainVar = strings.ReplaceAll(domainVar, c, "")
	}
	if len(domainVar) < 3 {
		domainVar = domainVar + domain[len(domain)-1:]
	}

	d := struct {
		DomainL      string
		DomainU      string
		DomainVar    string
		DomainVarU   string
		DomainNewVar string
		DomainUpdVar string
	}{
		DomainL:      domain,
		DomainU:      strings.ToUpper(domain[0:1]) + domain[1:],
		DomainVar:    domainVar,
		DomainVarU:   strings.ToUpper(domainVar[0:1]) + domainVar[1:],
		DomainNewVar: "n" + domain[0:1],
		DomainUpdVar: "u" + domain[0:1],
	}

	var b bytes.Buffer
	if err := tmpl.Execute(&b, d); err != nil {
		return err
	}

	if err := writeFile(basePath, domain, fileName, b); err != nil {
		return fmt.Errorf("writing %s: %w", fileName, err)
	}

	return nil
}

func writeFile(basePath string, domain string, fileName string, b bytes.Buffer) error {
	path := basePath
	switch {
	case basePath[:3] == "app":
		path = fmt.Sprintf("%s/%sgrp", basePath, domain)
	case strings.Contains(basePath, "stores"):
		path = fmt.Sprintf("%s/%sdb", basePath, domain)
	case strings.Contains(basePath, "core"):
		path = fmt.Sprintf("%s/%s", basePath, domain)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Println("Creating directory:", path)

		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return fmt.Errorf("write app directory: %w", err)
		}
	}

	path = fmt.Sprintf("%s/%s", path, fileName[:len(fileName)-1])
	path = strings.Replace(path, "new", domain, 1)

	fmt.Println("Add file:", path)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer f.Close()

	fmt.Println("Writing code:", path)
	if _, err := f.Write(b.Bytes()); err != nil {
		return fmt.Errorf("writing bytes: %w", err)
	}

	return nil
}
