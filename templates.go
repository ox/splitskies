package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"text/template"
)

const (
	layoutsDir   = "templates/layouts"
	templatesDir = "templates"
	extension    = "/*.html"
)

var (
	//go:embed templates/* templates/layouts/*
	templateFiles embed.FS
)

type TemplateEngine struct {
	templates map[string]*template.Template
}

func NewTemplateEngine() (*TemplateEngine, error) {
	e := &TemplateEngine{
		templates: make(map[string]*template.Template),
	}
	if err := e.ParseFiles(); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *TemplateEngine) ParseFiles() error {
	tmplFiles, err := fs.ReadDir(templateFiles, templatesDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, tmpl := range tmplFiles {
		if tmpl.IsDir() {
			continue
		}

		pt, err := template.ParseFS(templateFiles, templatesDir+"/"+tmpl.Name(), layoutsDir+extension)
		if err != nil {
			return err
		}

		e.templates[tmpl.Name()] = pt
		log.Printf("Added template %s", tmpl.Name())
	}

	return nil
}

func (e *TemplateEngine) Execute(w http.ResponseWriter, name string, data any) error {
	t, ok := e.templates[name]
	if !ok {
		return fmt.Errorf("could not find template named %s", name)
	}

	return t.Execute(w, data)
}

func (e *TemplateEngine) MustExecute(w http.ResponseWriter, name string, data any) {
	if err := e.Execute(w, name, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error executing template %s: %s", name, err.Error())))
		return
	}
}
