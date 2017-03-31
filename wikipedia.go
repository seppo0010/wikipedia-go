package main

import "net/http"
import "net/url"
import "io/ioutil"
import "fmt"
import "encoding/json"

type Wikipedia struct {
	preLanguageUrl, postLanguageUrl, language      string
	imagesResults, linksResults, categoriesResults string
	searchResults                                  int
}

type Language struct {
	code, name string
}

func NewWikipedia() *Wikipedia {
	return &Wikipedia{
		preLanguageUrl:    "https://",
		postLanguageUrl:   ".wikipedia.org/w/api.php",
		language:          "en",
		searchResults:     10,
		imagesResults:     "max",
		linksResults:      "max",
		categoriesResults: "max",
	}
}

func (w *Wikipedia) baseUrl() string {
	return fmt.Sprintf("%s%s%s", w.preLanguageUrl, w.language, w.postLanguageUrl)
}

func (w *Wikipedia) query(q map[string][]string, v interface{}) error {
	resp, err := http.Get(fmt.Sprintf("%s?%s", w.baseUrl(), url.Values(q).Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	json.Unmarshal(data, &v)
	return nil
}

func (w *Wikipedia) PreLanguageUrl() string {
	return w.preLanguageUrl
}

func (w *Wikipedia) PostLanguageUrl() string {
	return w.postLanguageUrl
}

func (w *Wikipedia) Language() string {
	return w.language
}

func (w *Wikipedia) SearchResults() int {
	return w.searchResults
}

func (w *Wikipedia) GetLanguages() (languages []Language, err error) {
	var f interface{}
	err = w.query(map[string][]string{
		"meta":   []string{"siteinfo"},
		"siprop": []string{"languages"},
		"format": []string{"json"},
		"action": []string{"query"},
	}, &f)
	if r, ok := f.(map[string]interface{}); ok {
		if query, ok := r["query"].(map[string]interface{}); ok {
			if langs, ok := query["languages"].([]interface{}); ok {
				for _, l := range langs {
					if lang, ok := l.(map[string]interface{}); ok {
						if code, ok := lang["code"].(string); ok {
							if name, ok := lang["*"].(string); ok {
								languages = append(languages, Language{code, name})
							}
						}
					}
				}
			}
		}
	}
	return
}

func main() {
	w := NewWikipedia()
	languages, _ := w.GetLanguages()
	fmt.Printf("%+v\n", languages)
}
