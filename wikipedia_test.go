package wikipedia

import "testing"

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
