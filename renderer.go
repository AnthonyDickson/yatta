package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"

	"github.com/AnthonyDickson/yatta/models"
)

var (
	//go:embed "templates/*"
	templatesFS embed.FS
)

// The paths to HTML templates relative to the project root dir.
const (
	baseTemplatePath     = "templates/base.html"
	taskTemplatePath     = "templates/task.html"
	taskListTemplatePath = "templates/task_list.html"
)

// Renderer converts tasks into a string representation.
type Renderer interface {
	// Render a single task.
	RenderTask(task models.Task) ([]byte, error)

	// Render tasks as list.
	RenderTaskList(tasks []models.Task) ([]byte, error)
}

// Renders responses as HTML pages.
type HTMLRenderer struct {
	// A mapping between a template path and the parsed template.
	templates map[string]*template.Template
}

// Create a new HTMLRenderer.
//
// Renturns an error if any of the templates cannot be parsed.
func NewHTMLRenderer() (*HTMLRenderer, error) {
	renderer := new(HTMLRenderer)
	renderer.templates = make(map[string]*template.Template)

	templates := []string{taskTemplatePath, taskListTemplatePath}

	for _, templatePath := range templates {
		tmpl, err := template.ParseFS(templatesFS, templatePath, baseTemplatePath)

		if err != nil {
			return nil, fmt.Errorf("could not parse the templates at %q and %q: %v", templatePath, baseTemplatePath, err)
		}

		renderer.templates[templatePath] = tmpl
	}

	return renderer, nil
}

// Render the HTML page for a single task.
//
// Returns an error if the template could not be found or rendered.
func (r *HTMLRenderer) RenderTask(task models.Task) ([]byte, error) {
	return r.renderHTMLTemplate(taskTemplatePath, task)
}

// Render the HTML page for a list of tasks.
//
// Returns an error if the template could not be found or rendered.
func (r *HTMLRenderer) RenderTaskList(tasks []models.Task) ([]byte, error) {
	return r.renderHTMLTemplate(taskListTemplatePath, tasks)
}

// Render data with the template at templatePath.
//
// This function assumes that templatePath points to a template that extends the base template [baseTemplatePath].
//
// Returns an error if the template could not be found or rendered.
func (r *HTMLRenderer) renderHTMLTemplate(templatePath string, data any) ([]byte, error) {
	tmpl := r.templates[templatePath]

	if tmpl == nil {
		return nil, fmt.Errorf("could not find template for %q, did you parse it in the constructor?", templatePath)
	}

	body := new(bytes.Buffer)

	if err := tmpl.Execute(body, data); err != nil {
		return nil, fmt.Errorf("could not render the template %q with data %q: %v", templatePath, data, err)
	}

	return body.Bytes(), nil
}
