package main_test

import (
	"reflect"
	"strings"
	"testing"

	yatta "github.com/AnthonyDickson/yatta"
	"golang.org/x/net/html"
)

func TestRenderer_TodosList(t *testing.T) {
	renderer := mustCreateRenderer(t)

	t.Run("renders todos list", func(t *testing.T) {
		want := []string{
			"eat",
			"sleep",
			"debug tests 🙃",
		}

		htmlString, err := renderer.RenderTodosList(want)

		assertNoError(t, err)
		assertHTMLContainsTodos(t, string(htmlString), want)
	})
}

func mustCreateRenderer(t *testing.T) *yatta.Renderer {
	t.Helper()

	renderer, err := yatta.NewRenderer()

	if err != nil {
		t.Errorf("an error occurred creating the renderer: %v", err)
	}

	return renderer
}

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("an error occurred while rendering the todo list: %v", err)
	}
}

func assertHTMLContainsTodos(t *testing.T, htmlString string, want []string) {
	t.Helper()

	doc, err := html.Parse(strings.NewReader(htmlString))

	if err != nil {
		t.Fatalf("an error occurred while parsing the HTML string: %v", err)
	}

	got := extractTodosFromHTML(t, doc)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got todo list %q want %q", got, want)
	}
}

func extractTodosFromHTML(t *testing.T, htmlFragment *html.Node) []string {
	t.Helper()

	todos := []string{}

	var extractText func(*html.Node)
	extractText = func(node *html.Node) {
		if node.Type == html.TextNode {
			todos = append(todos, node.Data)
			return
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			extractText(child)
		}
	}

	var findTodos func(*html.Node)
	findTodos = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == "li" {
			extractText(node)
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			findTodos(child)
		}
	}

	findTodos(htmlFragment)

	return todos
}
