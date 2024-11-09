package main_test

import (
	"reflect"
	"strings"
	"testing"

	yatta "github.com/AnthonyDickson/yatta"
	"golang.org/x/net/html"
)

func TestRenderer_TasksList(t *testing.T) {
	renderer := mustCreateRenderer(t)

	t.Run("renders tasks list", func(t *testing.T) {
		want := []yatta.Task{
			{ID: 0, Description: "eat"},
			{ID: 1, Description: "sleep"},
			{ID: 2, Description: "debug tests 🙃"},
		}

		htmlString, err := renderer.RenderTaskList(want)

		assertNoError(t, err)
		assertHTMLContainsTasks(t, string(htmlString), want)
	})
}

func mustCreateRenderer(t *testing.T) *yatta.HTMLRenderer {
	t.Helper()

	renderer, err := yatta.NewHTMLRenderer()

	if err != nil {
		t.Errorf("an error occurred creating the renderer: %v", err)
	}

	return renderer
}

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("an error occurred while rendering the task list: %v", err)
	}
}

func assertHTMLContainsTasks(t *testing.T, htmlString string, tasks []yatta.Task) {
	t.Helper()

	doc, err := html.Parse(strings.NewReader(htmlString))

	if err != nil {
		t.Fatalf("an error occurred while parsing the HTML string: %v", err)
	}

	got := extractTasksFromHTML(t, doc)

	var want []string

	for _, task := range tasks {
		want = append(want, task.Description)
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got task list %q want %q", got, want)
	}
}

func extractTasksFromHTML(t *testing.T, htmlFragment *html.Node) []string {
	t.Helper()

	tasks := []string{}

	var extractText func(*html.Node)
	extractText = func(node *html.Node) {
		if node.Type == html.TextNode {
			tasks = append(tasks, node.Data)
			return
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			extractText(child)
		}
	}

	var findTasks func(*html.Node)
	findTasks = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "li" {
			extractText(node)
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			findTasks(child)
		}
	}

	findTasks(htmlFragment)

	return tasks
}
