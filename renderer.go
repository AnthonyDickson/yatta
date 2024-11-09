package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
)

var (
	//go:embed "templates/*"
	taskListTemplate embed.FS
)

type Renderer interface {
	RenderTaskList(tasks []Task) ([]byte, error)
	RenderTask(task Task) ([]byte, error)
}

type HTMLRenderer struct {
	template *template.Template
}

func NewHTMLRenderer() (*HTMLRenderer, error) {
	renderer := new(HTMLRenderer)

	templatesPath := "templates/*.html"
	tmpl, err := template.ParseFS(taskListTemplate, templatesPath)

	if err != nil {
		return nil, fmt.Errorf("an error occurred while parsing the templates at %s: %v", templatesPath, err)
	}

	renderer.template = tmpl

	return renderer, nil
}

func (r *HTMLRenderer) RenderTaskList(tasks []Task) ([]byte, error) {
	body := new(bytes.Buffer)
	templatePath := "task_list.html"

	if err := r.template.ExecuteTemplate(body, templatePath, tasks); err != nil {
		return nil, fmt.Errorf("an error occurred while rendering the template for the template %q with data %q: %v", templatePath, tasks, err)
	}

	return body.Bytes(), nil
}

func (r *HTMLRenderer) RenderTask(task Task) ([]byte, error) {
	body := new(bytes.Buffer)
	templatePath := "task.html"

	if err := r.template.ExecuteTemplate(body, templatePath, task); err != nil {
		return nil, fmt.Errorf("an error occurred while rendering the template for the template %q with data %q: %v", templatePath, task, err)
	}

	return body.Bytes(), nil
}
