package wikipedia

import "fmt"
import "strings"
import "testing"

func testPageId(t *testing.T, page *Page, id string) {
	pageId, err := page.Id()
	if err != nil {
		t.Error(fmt.Sprintf("error getting page id %s", err))
		return
	}
	if pageId != id {
		t.Error("Invalid page id")
		return
	}
}

func TestPageId(t *testing.T) {
	t.Parallel()
	w := NewWikipedia()
	testPageId(t, NewPage(w, "Bikeshedding"), "4138548")
}

func TestPageIdRedirect(t *testing.T) {
	t.Parallel()
	w := NewWikipedia()
	testPageId(t, NewPage(w, "Law of triviality"), "4138548")
}

func TestPageIdPageId(t *testing.T) {
	t.Parallel()
	w := NewWikipedia()
	testPageId(t, NewPageFromId(w, "4138548"), "4138548")
}

func testPageTitle(t *testing.T, page *Page, title string) {
	pageTitle, err := page.Title()
	if err != nil {
		t.Error(fmt.Sprintf("error getting page title %s", err))
		return
	}
	if pageTitle != title {
		t.Error(fmt.Sprintf("Invalid page title (expected %s, got %s)", title, pageTitle))
		return
	}
}

func TestPageTitle(t *testing.T) {
	t.Parallel()
	w := NewWikipedia()
	testPageTitle(t, NewPageFromId(w, "4138548"), "Law of triviality")
}

func TestPageTitleFromTitle(t *testing.T) {
	t.Parallel()
	w := NewWikipedia()
	testPageTitle(t, NewPage(w, "Law of triviality"), "Law of triviality")
}

func TestContent(t *testing.T) {
	t.Parallel()
	w := NewWikipedia()
	page := NewPageFromId(w, "4138548")
	content, err := page.Content()
	if err != nil {
		t.Error(fmt.Sprintf("error getting page content %s", err))
		return
	}
	if strings.Contains(content, "bike-shedding") == false {
		t.Error("expected content to contain bike-shedding")
		return
	}
}

func TestHtmlContent(t *testing.T) {
	t.Parallel()
	w := NewWikipedia()
	page := NewPageFromId(w, "4138548")
	content, err := page.HtmlContent()
	if err != nil {
		t.Error(fmt.Sprintf("error getting page html content %s", err))
		return
	}
	if strings.Contains(content, "bike-shedding") == false {
		t.Error("expected content to contain bike-shedding")
		return
	}
}

func TestSummary(t *testing.T) {
	t.Parallel()
	w := NewWikipedia()
	page := NewPageFromId(w, "4138548")
	summary, err := page.Summary()
	if err != nil {
		t.Error(fmt.Sprintf("error getting page summary %s", err))
		return
	}
	if strings.Contains(summary, "bike-shedding") == false {
		t.Error("expected summary to contain bike-shedding")
		return
	}

	content, err := page.Content()
	if err != nil {
		t.Error(fmt.Sprintf("error getting page content %s", err))
		return
	}

	if len(summary) >= len(content) {
		t.Error("summary is at least as long as content, expected shorter")
		return
	}
}
