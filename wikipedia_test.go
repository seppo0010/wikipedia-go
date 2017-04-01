package wikipedia

import "testing"

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

func TestGetLanguages(t *testing.T) {
	w := NewWikipedia()
	languages, err := w.GetLanguages()
	if err != nil {
		t.Error("Failed to get languages")
		return
	}
	for _, lang := range languages {
		if lang.code == "en" {
			if lang.name == "English" {
				return
			}
			t.Error("en is not named English")
			return
		}
	}
	t.Error("Could not find English")
}

func TestBaseUrlLanguage(t *testing.T) {
	w := NewWikipedia()
	w.SetBaseUrl("http://wikipedia.com/{language}/test")
	url := w.GetBaseUrl()
	if url != "http://wikipedia.com/en/test" {
		t.Error("Got wrong url")
		return
	}
}

func TestBaseUrlNoLanguage(t *testing.T) {
	w := NewWikipedia()
	w.SetBaseUrl("http://wikipedia.com/test")
	url := w.GetBaseUrl()
	if url != "http://wikipedia.com/test" {
		t.Error("Got wrong url")
		return
	}
}

func TestSearch(t *testing.T) {
	w := NewWikipedia()
	results, err := w.Search("hello world")
	if err != nil {
		t.Error("Got error")
		return
	}
	if contains(results, "\"Hello, World!\" program") == false {
		t.Error("Expected results to contain hello world program")
		return
	}
}

func TestGeosearch(t *testing.T) {
	w := NewWikipedia()
	results, err := w.Geosearch(-34.603333, -58.381667, 10)
	if err != nil {
		t.Error("Got error")
		return
	}
	if contains(results, "Buenos Aires") == false {
		t.Error("Expected results to contain Buenos Aires")
		return
	}
}
