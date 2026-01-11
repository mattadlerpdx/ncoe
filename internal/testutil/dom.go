package testutil

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

// DOM provides DOM traversal and assertion helpers for HTML content.
type DOM struct {
	Root *html.Node
	Raw  string
	t    *testing.T
}

// ParseDOM parses HTML content and returns a DOM helper.
func ParseDOM(t *testing.T, body string) *DOM {
	t.Helper()
	doc, err := html.Parse(strings.NewReader(body))
	if err != nil {
		t.Fatalf("ParseDOM: failed to parse HTML: %v", err)
	}
	return &DOM{Root: doc, Raw: body, t: t}
}

// --- Full Page vs Fragment Assertions ---

// AssertFullPage asserts this is a full HTML page with html, head, and body elements.
func (d *DOM) AssertFullPage() {
	d.t.Helper()
	if !d.hasElement("html") {
		d.t.Error("AssertFullPage: missing <html> element")
	}
	if !d.hasElement("head") {
		d.t.Error("AssertFullPage: missing <head> element")
	}
	if !d.hasElement("body") {
		d.t.Error("AssertFullPage: missing <body> element")
	}
}

// AssertFragment asserts this is an HTML fragment (not a full page).
// It should NOT contain html, head, or body elements in the raw source.
// Note: We check the raw string because html.Parse normalizes fragments to full docs.
func (d *DOM) AssertFragment() {
	d.t.Helper()
	rawLower := strings.ToLower(d.Raw)
	if strings.Contains(rawLower, "<html") {
		d.t.Error("AssertFragment: should not contain <html> element")
	}
	if strings.Contains(rawLower, "<head") {
		d.t.Error("AssertFragment: should not contain <head> element")
	}
	if strings.Contains(rawLower, "<body") {
		d.t.Error("AssertFragment: should not contain <body> element")
	}
	if strings.Contains(rawLower, "<!doctype") {
		d.t.Error("AssertFragment: should not contain <!DOCTYPE>")
	}
}

// --- Element Existence Assertions ---

// AssertHasElement asserts an element with the given tag exists.
func (d *DOM) AssertHasElement(tag string) {
	d.t.Helper()
	if !d.hasElement(tag) {
		d.t.Errorf("AssertHasElement: missing <%s> element", tag)
	}
}

// AssertHasElementByID asserts an element with the given ID exists.
func (d *DOM) AssertHasElementByID(id string) {
	d.t.Helper()
	if d.FindByID(id) == nil {
		d.t.Errorf("AssertHasElementByID: no element with id=%q", id)
	}
}

// AssertHasElementByClass asserts at least one element with the given class exists.
func (d *DOM) AssertHasElementByClass(class string) {
	d.t.Helper()
	if d.FindByClass(class) == nil {
		d.t.Errorf("AssertHasElementByClass: no element with class=%q", class)
	}
}

// --- Input/Form Assertions ---

// AssertHasInputName asserts an input/textarea/select with the given name exists.
func (d *DOM) AssertHasInputName(name string) {
	d.t.Helper()
	if d.FindInput(name) == nil {
		d.t.Errorf("AssertHasInputName: no input with name=%q", name)
	}
}

// AssertHasForm asserts a form with the given action exists.
func (d *DOM) AssertHasForm(action string) {
	d.t.Helper()
	if d.FindForm(action) == nil {
		d.t.Errorf("AssertHasForm: no form with action=%q", action)
	}
}

// AssertFormHasInputs asserts a form has all the specified input names.
func (d *DOM) AssertFormHasInputs(action string, names ...string) {
	d.t.Helper()
	form := d.FindForm(action)
	if form == nil {
		d.t.Errorf("AssertFormHasInputs: no form with action=%q", action)
		return
	}
	for _, name := range names {
		if findInputInNode(form, name) == nil {
			d.t.Errorf("AssertFormHasInputs: form %q missing input name=%q", action, name)
		}
	}
}

// --- Text Content Assertions ---

// AssertContainsText asserts the body contains the given text.
func (d *DOM) AssertContainsText(text string) {
	d.t.Helper()
	if !strings.Contains(d.Raw, text) {
		d.t.Errorf("AssertContainsText: body does not contain %q", text)
	}
}

// AssertNotContainsText asserts the body does not contain the given text.
func (d *DOM) AssertNotContainsText(text string) {
	d.t.Helper()
	if strings.Contains(d.Raw, text) {
		d.t.Errorf("AssertNotContainsText: body should not contain %q", text)
	}
}

// ContainsText returns true if the body contains the given text.
func (d *DOM) ContainsText(text string) bool {
	return strings.Contains(d.Raw, text)
}

// --- HTMX Attribute Assertions ---

// AssertHasHTMXAttr asserts an element with the given HTMX attribute exists.
func (d *DOM) AssertHasHTMXAttr(attr string) {
	d.t.Helper()
	found := d.findNode(func(n *html.Node) bool {
		return hasAttr(n, attr)
	})
	if found == nil {
		d.t.Errorf("AssertHasHTMXAttr: no element with %s attribute", attr)
	}
}

// --- Find Methods (return nodes, don't assert) ---

// FindByID finds an element by ID attribute.
func (d *DOM) FindByID(id string) *html.Node {
	return d.findNode(func(n *html.Node) bool {
		return getAttr(n, "id") == id
	})
}

// FindByClass finds the first element with a given class.
func (d *DOM) FindByClass(class string) *html.Node {
	return d.findNode(func(n *html.Node) bool {
		classes := getAttr(n, "class")
		for _, c := range strings.Fields(classes) {
			if c == class {
				return true
			}
		}
		return false
	})
}

// FindInput finds an input/textarea/select by name.
func (d *DOM) FindInput(name string) *html.Node {
	return d.findNode(func(n *html.Node) bool {
		if n.Type != html.ElementNode {
			return false
		}
		if n.Data != "input" && n.Data != "textarea" && n.Data != "select" {
			return false
		}
		return getAttr(n, "name") == name
	})
}

// FindForm finds a form by action.
func (d *DOM) FindForm(action string) *html.Node {
	return d.findNode(func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == "form" && getAttr(n, "action") == action
	})
}

// FindAllByTag finds all elements with the given tag.
func (d *DOM) FindAllByTag(tag string) []*html.Node {
	return d.findAllNodes(func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == tag
	})
}

// --- Internal Helpers ---

func (d *DOM) hasElement(tag string) bool {
	return d.findNode(func(n *html.Node) bool {
		return n.Type == html.ElementNode && n.Data == tag
	}) != nil
}

func (d *DOM) findNode(match func(*html.Node) bool) *html.Node {
	var result *html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if result != nil {
			return
		}
		if match(n) {
			result = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(d.Root)
	return result
}

func (d *DOM) findAllNodes(match func(*html.Node) bool) []*html.Node {
	var results []*html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if match(n) {
			results = append(results, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(d.Root)
	return results
}

func getAttr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}

func hasAttr(n *html.Node, key string) bool {
	for _, a := range n.Attr {
		if a.Key == key {
			return true
		}
	}
	return false
}

func findInputInNode(node *html.Node, name string) *html.Node {
	var result *html.Node
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if result != nil {
			return
		}
		if n.Type == html.ElementNode {
			if n.Data == "input" || n.Data == "textarea" || n.Data == "select" {
				if getAttr(n, "name") == name {
					result = n
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(node)
	return result
}
