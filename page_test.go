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

func TestCont(t *testing.T) {
	t.Parallel()
	w := NewWikipedia()
	page := NewPage(w, "Argentina")
	for image := range page.Images() {
		if image.Err != nil {
			t.Error(fmt.Sprintf("error getting image: %s", image.Err))
			return
		}
	}
}

func TestImages(t *testing.T) {
	t.Parallel()
	w := NewWikipedia()
	w.SetImagesResults("5")
	page := NewPage(w, "Argentina")
	c := 0
	imageSet := make(map[string]bool)
	for imageRequest := range page.Images() {
		if imageRequest.Err != nil {
			t.Error(fmt.Sprintf("error getting page images %s", imageRequest.Err))
			return
		}
		image := imageRequest.Image
		if len(image.Title) == 0 {
			t.Error("got image with no title")
			return
		}
		if len(image.Url) == 0 {
			t.Error("got image with no url")
			return
		}
		if len(image.DescriptionUrl) == 0 {
			t.Error("got image with no description url")
			return
		}
		imageSet[image.Title] = true
		c++
		if c == 11 {
			break
		}
	}
	if c != 11 || len(imageSet) != 11 {
		t.Error("got less than 11 different images")
		return
	}
}

func TestExtlinks(t *testing.T) {
	t.Parallel()
	w := NewWikipedia()
	w.SetLinksResults("2")
	page := NewPage(w, "Argentina")
	c := 0
	referenceSet := make(map[string]bool)
	for referenceRequest := range page.Extlinks() {
		if referenceRequest.Err != nil {
			t.Error(fmt.Sprintf("error getting page external links %s", referenceRequest.Err))
			return
		}
		reference := referenceRequest.Reference
		if len(reference.Url) == 0 {
			t.Error("got reference with no url")
			return
		}
		referenceSet[reference.Url] = true
		c++
		if c == 5 {
			break
		}
	}
	if c != 5 || len(referenceSet) != 5 {
		t.Error("got less than 5 different external links")
		return
	}
}
