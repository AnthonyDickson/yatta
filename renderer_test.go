package main_test

import (
	"bytes"
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

func assertInputsHaveAttributes(t *testing.T, inputs []*html.Node, wantInputAttrs attributeSets) {
	t.Helper()

	for _, input := range inputs {
		got := newAttributeSet(input)

		for _, want := range wantInputAttrs {
			intersection := got.intersection(want)

			if intersection == FULL_INTERSECTION {
				wantInputAttrs = wantInputAttrs.delete(want)
				break
			} else if intersection == PARTIAL_INTERSECTION {
				found := make(attributeSet)
				missing := make(attributeSet)

				for wantKey, wantValue := range want {
					gotValue, ok := got[wantKey]

					if ok && gotValue == wantValue {
						found[wantKey] = wantValue
					} else {
						missing[wantKey] = wantValue
					}
				}

				t.Errorf("got input with the attributes %v, was missing %s", found.format(), missing.format())
				wantInputAttrs = wantInputAttrs.delete(want)
				break
			}
		}
	}

	for _, attributes := range wantInputAttrs {
		t.Errorf("expected to find an input element with the attributes: %s", attributes.format())
	}
}

type attributeSet map[string]string

// newAttributeSet creates a attribute key-value pairs from an HTML node.
func newAttributeSet(node *html.Node) attributeSet {
	attributes := make(attributeSet)

	for _, attribute := range node.Attr {
		attributes[attribute.Key] = attribute.Val
	}

	return attributes
}

// Format each key-value pair in the attributeSet as `key="value"` separated with a space.
func (a attributeSet) format() string {
	b := new(bytes.Buffer)

	for k, v := range a {
		fmt.Fprintf(b, "%s=\"%s\" ", k, v)
	}

	// Remove trailing space
	b.Truncate(b.Len() - 1)

	return b.String()
}

const (
	NO_INTERSECTION      = 1
	PARTIAL_INTERSECTION = 2
	FULL_INTERSECTION    = 3
)

// Find the intersection between two attribute sets.
//
// A return value of FULL_INTERSECTION represents the case where all attributes in `want` are
// found in `got`, i.e. that `got` is a superset of `want`.
//
// A return value of PARTIAL_INTERSECTION represents the case where some attributes in `want` are
// found in `got`, but not all. This includes attributes with the same key but different values.
//
// A return value of NO_INTERSECTION represents the case where no attributes in `want` are found in `got`.
func (got attributeSet) intersection(want attributeSet) int {
	found := 0
	wantCount := len(want)

	for attribute, value := range want {
		gotValue, ok := got[attribute]

		if !ok || gotValue != value {
			continue
		}

		found++
	}

	if found == 0 {
		return NO_INTERSECTION
	} else if found == wantCount {
		return FULL_INTERSECTION
	} else {
		return PARTIAL_INTERSECTION
	}
}

type attributeSets []attributeSet

// Delete an attribute set from a slice of attribute sets.
//
// Note that you must assign the return value to the original slice, e.g. `a = a.delete(b)`.
func (a attributeSets) delete(attributes attributeSet) attributeSets {
	index := -1

	for i, attrs := range a {
		if reflect.DeepEqual(attributes, attrs) {
			index = i
		}
	}

	if index != -1 {
		a = append(a[:index], a[index+1:]...)
	} else {
		panic(fmt.Sprintf("could not find attribute set %v in %v", attributes, a))
	}

	return a
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
		t.Errorf("got %s element without attribute %q, want attribute %s=%q", node.Data, attrKey, attrKey, attrVal)
	}
}

func assertValidRegistrationPage(t *testing.T, htmlString string) {
	t.Helper()

	htmlRootNode, err := html.Parse(strings.NewReader(htmlString))
	yattatest.AssertNoError(t, err)

	form := mustFindNode(t, htmlRootNode, "form")
	wantAttrs := map[string]string{"action": "/users", "method": "POST"}
	assertHTMLNodeHasAttributes(t, form, wantAttrs)

	wantInputAttrs := attributeSets{
		attributeSet{"type": "email", "name": "email", "required": ""},
		attributeSet{"type": "password", "name": "password", "required": ""},
	}
	inputs := mustFindNodes(t, form, "input", len(wantInputAttrs))
	assertInputsHaveAttributes(t, inputs, wantInputAttrs)

	button := mustFindNode(t, form, "button")
	wantButtonAttrs := map[string]string{"type": "submit", "tabindex": "0"}
	assertHTMLNodeHasAttributes(t, button, wantButtonAttrs)
}
