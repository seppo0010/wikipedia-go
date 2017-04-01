package wikipedia

import "net/http"
import "net/url"
import "errors"
import "fmt"
import "encoding/json"
import "strings"

const LANGUAGE_URL_MARKER = "{language}"

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

func (w *Wikipedia) GetBaseUrl() string {
	return fmt.Sprintf("%s%s%s", w.preLanguageUrl, w.language, w.postLanguageUrl)
}

func (w *Wikipedia) SetBaseUrl(baseUrl string) {
	index := strings.Index(baseUrl, LANGUAGE_URL_MARKER)
	if index == -1 {
		w.preLanguageUrl = baseUrl
		w.language = ""
		w.postLanguageUrl = ""
	} else {
		w.preLanguageUrl = baseUrl[0:index]
		w.postLanguageUrl = baseUrl[index+len(LANGUAGE_URL_MARKER):]
	}
}

func (w *Wikipedia) query(q map[string][]string, v interface{}) error {
	resp, err := http.Get(fmt.Sprintf("%s?%s", w.GetBaseUrl(), url.Values(q).Encode()))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return json.NewDecoder(resp.Body).Decode(&v)
}

func (w *Wikipedia) results(v interface{}, field string) (results []string, err error) {
	gotResults := false
	if r, ok := v.(map[string]interface{}); ok {
		if query, ok := r["query"].(map[string]interface{}); ok {
			if values, ok := query[field].([]interface{}); ok {
				gotResults = true
				for _, l := range values {
					if lang, ok := l.(map[string]interface{}); ok {
						if title, ok := lang["title"].(string); ok {
							results = append(results, title)
						}
					}
				}
			}
		}
	}
	if gotResults == false {
		err = errors.New("Invalid JSON response")
	}
	return
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
	if err != nil {
		return
	}
	gotLangs := false
	if r, ok := f.(map[string]interface{}); ok {
		if query, ok := r["query"].(map[string]interface{}); ok {
			if langs, ok := query["languages"].([]interface{}); ok {
				gotLangs = true
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
	if gotLangs == false {
		err = errors.New("Invalid JSON response")
	}
	return
}

func (w *Wikipedia) Search(query string) (results []string, err error) {
	var f interface{}
	err = w.query(map[string][]string{
		"list":     []string{"search"},
		"srpop":    []string{""},
		"srlimit":  []string{fmt.Sprintf("%d", w.searchResults)},
		"srsearch": []string{query},
		"format":   []string{"json"},
		"action":   []string{"query"},
	}, &f)
	if err != nil {
		return
	}
	results, err = w.results(f, "search")
	return
}

func (w *Wikipedia) Geosearch(latitude float64, longitude float64, radius int) (results []string, err error) {
	if latitude < -90.0 || latitude > 90.0 {
		err = errors.New("Invalid latitude")
		return
	}
	if longitude < -180.0 || longitude > 180.0 {
		err = errors.New("Invalid longitude")
		return
	}
	if radius < -10 || radius > 10000 {
		err = errors.New("Invalid radius")
		return
	}
	var f interface{}
	err = w.query(map[string][]string{
		"list":     []string{"geosearch"},
		"gsradius": []string{fmt.Sprintf("%d", radius)},
		"gscoord":  []string{fmt.Sprintf("%f|%f", latitude, longitude)},
		"gslimit":  []string{fmt.Sprintf("%d", w.searchResults)},
		"format":   []string{"json"},
		"action":   []string{"query"},
	}, &f)
	if err != nil {
		return
	}
	results, err = w.results(f, "geosearch")
	return
}
