package wikipedia

import "fmt"
import "testing"

func testPageId(t *testing.T, page *Page, id string) {
	pageId, err := page.GetPageId()
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
