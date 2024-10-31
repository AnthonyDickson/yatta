package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

var (
	//go:embed "templates/*"
	todosListTemplate embed.FS
)

type Renderer interface {
	RenderTodosList(todos []string) ([]byte, error)
}

type HTMLRenderer struct {
	template *template.Template
}

func NewHTMLRenderer() (*HTMLRenderer, error) {
	renderer := new(HTMLRenderer)

	templatesPath := "templates/*.html"
	tmpl, err := template.ParseFS(todosListTemplate, templatesPath)

	if err != nil {
		return nil, fmt.Errorf("an error occurred while parsing the templates at %s: %v", templatesPath, err)
	}

	renderer.template = tmpl

	return renderer, nil
}

func (r *HTMLRenderer) RenderTodosList(todos []string) ([]byte, error) {
	body := new(bytes.Buffer)
	templatePath := "todo_list.html"

	if err := r.template.ExecuteTemplate(body, templatePath, todos); err != nil {
		return nil, fmt.Errorf("an error occurred while rendering the template for the template %q with data %q: %v", templatePath, todos, err)
	}

	return body.Bytes(), nil
}
