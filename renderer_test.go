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

func TestRenderer_RegistrationPage(t *testing.T) {
	t.Run("renders registration page", func(t *testing.T) {
		renderer := mustCreateRenderer(t)

		htmlString, err := renderer.RenderRegistrationPage()

		yattatest.AssertNoError(t, err)
		assertValidRegistrationPage(t, string(htmlString))
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

func findNodes(t *testing.T, htmlFragment *html.Node, containerTag string) []*html.Node {
	t.Helper()

	var foundNodes []*html.Node

	var findNodesHelper func(*html.Node)
	findNodesHelper = func(node *html.Node) {
		if node.Type == html.ElementNode && node.Data == containerTag {
			foundNodes = append(foundNodes, node)
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			findNodesHelper(child)
		}
	}

	findNodesHelper(htmlFragment)

	return foundNodes
}

func mustFindNode(t *testing.T, htmlFragment *html.Node, wantTag string) *html.Node {
	nodes := findNodes(t, htmlFragment, wantTag)

	return nodes[0]
}

func mustFindNodes(t *testing.T, htmlFragment *html.Node, wantTag string, wantCount int) []*html.Node {
	t.Helper()

	nodes := findNodes(t, htmlFragment, wantTag)

	if len(nodes) != wantCount {
		t.Fatalf("expected %d %s elements, got %d", wantCount, wantTag, len(nodes))
	}

	return nodes
}

func assertInputsHaveAttributes(t *testing.T, inputs []*html.Node, wantInputAttrs map[string]string) {
	t.Helper()

	for _, input := range inputs {
		for _, attribute := range input.Attr {
			want, ok := wantInputAttrs[attribute.Key]

			if !ok {
				continue
			}

			if attribute.Val != want {
				t.Errorf("got input attribute %q with value %q, want %s=%q", attribute.Key, attribute.Val, attribute.Key, want)
			}

			delete(wantInputAttrs, attribute.Key)
		}
	}

	for attrKey, attrVal := range wantInputAttrs {
		t.Errorf("got input without attribute %q, want input attribute %s=%q", attrKey, attrKey, attrVal)
	}
}

func assertHTMLNodeHasAttributes(t *testing.T, node *html.Node, wantAttrs map[string]string) {
	t.Helper()

	for _, attribute := range node.Attr {
		want, ok := wantAttrs[attribute.Key]

		if !ok {
			continue
		}

		if attribute.Val != want {
			t.Errorf("got attribute %q with value %q, want %q", attribute.Key, attribute.Val, want)
		}

		delete(wantAttrs, attribute.Key)
	}

	for attrKey, attrVal := range wantAttrs {
		t.Errorf("got node without attribute %q, want attribute %s=%q", attrKey, attrKey, attrVal)
	}
}

func assertValidRegistrationPage(t *testing.T, htmlString string) {
	t.Helper()

	htmlRootNode, err := html.Parse(strings.NewReader(htmlString))
	yattatest.AssertNoError(t, err)

	form := mustFindNode(t, htmlRootNode, "form")
	wantAttrs := map[string]string{"action": "/users", "method": "POST"}
	assertHTMLNodeHasAttributes(t, form, wantAttrs)

	inputs := mustFindNodes(t, form, "input", 1)
	wantInputAttrs := map[string]string{"type": "email"}
	assertInputsHaveAttributes(t, inputs, wantInputAttrs)

	button := mustFindNode(t, form, "button")
	wantButtonAttrs := map[string]string{"type": "submit"}
	assertHTMLNodeHasAttributes(t, button, wantButtonAttrs)
}
