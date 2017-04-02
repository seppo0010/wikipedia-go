package wikipedia

import "fmt"
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
