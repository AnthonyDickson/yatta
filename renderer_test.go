package main_test

import (
	"fmt"
	"log/slog"
	"reflect"
	"slices"
	"strings"
	"testing"

	yatta "github.com/AnthonyDickson/yatta"
	"github.com/AnthonyDickson/yatta/models"
	"github.com/AnthonyDickson/yatta/yattatest"
	"golang.org/x/net/html"
)

func TestRenderer_Index(t *testing.T) {
	t.Run("renders index page", func(t *testing.T) {
		renderer := mustCreateRenderer(t)
		users := []models.User{
			{ID: 1, Email: "test@example.com", Password: yattatest.MustCreatePasswordHash(t, "test")},
			{ID: 2, Email: "foo@bar.com", Password: yattatest.MustCreatePasswordHash(t, "baz")},
		}

		htmlString, err := renderer.RenderIndex(users)
		yattatest.AssertNoError(t, err)
		assertHTMLContainsUsers(t, string(htmlString), users, "a")
	})
}

func TestRenderer_TasksList(t *testing.T) {
	renderer := mustCreateRenderer(t)

	t.Run("renders tasks list", func(t *testing.T) {
		want := []models.Task{
			{ID: 0, Description: "eat"},
			{ID: 1, Description: "sleep"},
			{ID: 2, Description: "debug tests 🙃"},
		}

		htmlString, err := renderer.RenderTaskList(want)

		yattatest.AssertNoError(t, err)
		assertHTMLContainsTasks(t, string(htmlString), want, "li")
	})
}

func TestRenderer_Task(t *testing.T) {
	renderer := mustCreateRenderer(t)

	t.Run("renders task", func(t *testing.T) {
		want := []models.Task{
			{ID: 0, Description: "eat"},
		}

		htmlString, err := renderer.RenderTask(want[0])

		yattatest.AssertNoError(t, err)
		assertHTMLContainsTasks(t, string(htmlString), want, "p")
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

func assertHTMLContainsTasks(t *testing.T, htmlString string, tasks []models.Task, containerTag string) {
	t.Helper()

	doc, err := html.Parse(strings.NewReader(htmlString))

	if err != nil {
		t.Fatalf("an error occurred while parsing the HTML string: %v", err)
	}

	got := extractTextNodesFromHTML(t, doc, containerTag)

	var want []string

	for _, task := range tasks {
		want = append(want, task.Description)
	}

	if !reflect.DeepEqual(got, want) {
		slog.Error(fmt.Sprintf("Got HTML: %s", htmlString))
		t.Errorf("got task list %q want %q", got, want)
	}
}

func assertHTMLContainsUsers(t *testing.T, htmlString string, users []models.User, containerTag string) {
	t.Helper()

	doc, err := html.Parse(strings.NewReader(htmlString))

	if err != nil {
		t.Fatalf("an error occurred while parsing the HTML string: %v", err)
	}

	got := extractTextNodesFromHTML(t, doc, containerTag)

	for _, user := range users {
		if !slices.Contains(got, user.Email) {
			t.Errorf("could not find user %q in the HTML", user.Email)
		}
	}
}

func extractTextNodesFromHTML(t *testing.T, htmlFragment *html.Node, containerTag string) []string {
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
		if node.Type == html.ElementNode && node.Data == containerTag {
			extractText(node)
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			findTasks(child)
		}
	}

	findTasks(htmlFragment)

	return tasks
}
