package wikipedia

import "net/http"
import "net/url"
import "errors"
import "fmt"
import "encoding/json"
import "strings"

const LANGUAGE_URL_MARKER = "{language}"

type Wikipedia interface {
	Page(title string) Page
	PageFromId(id string) Page
	GetBaseUrl() string
	SetBaseUrl(baseUrl string)
	SetImagesResults(imagesResults string)
	SetLinksResults(linksResults string)
	SetCategoriesResults(categoriesResults string)
	PreLanguageUrl() string
	PostLanguageUrl() string
	Language() string
	SearchResults() int
	GetLanguages() (languages []Language, err error)
	Search(query string) (results []string, err error)
	Geosearch(latitude float64, longitude float64, radius int) (results []string, err error)
	RandomCount(count uint) (results []string, err error)
	Random() (string, error)
	ImagesResults() string
	LinksResults() string
	CategoriesResults() string
}

type WikipediaClient struct {
	preLanguageUrl, postLanguageUrl, language      string
	imagesResults, linksResults, categoriesResults string
	searchResults                                  int
}

const (
	ParameterError = iota
	ResponseError  = iota
)

type WikipediaError struct {
	Type int
	Err  error
}

func newError(t int, e error) *WikipediaError {
	return &WikipediaError{Type: t, Err: e}
}

func (e *WikipediaError) Error() string {
	switch e.Type {
	case ParameterError:
		return fmt.Sprintf("parameter error: %s", e.Err.Error())
	case ResponseError:
		return fmt.Sprintf("response error: %s", e.Err.Error())
	default:
		return fmt.Sprintf("unknown error: %s", e.Err.Error())
	}
}

type Language struct {
	code, name string
}

func NewWikipedia() *WikipediaClient {
	return &WikipediaClient{
		preLanguageUrl:    "https://",
		postLanguageUrl:   ".wikipedia.org/w/api.php",
		language:          "en",
		searchResults:     10,
		imagesResults:     "max",
		linksResults:      "max",
		categoriesResults: "max",
	}
}

func (w *WikipediaClient) Page(title string) Page {
	return NewPage(w, title)
}

func (w *WikipediaClient) PageFromId(id string) Page {
	return NewPageFromId(w, id)
}

func (w *WikipediaClient) GetBaseUrl() string {
	return fmt.Sprintf("%s%s%s", w.preLanguageUrl, w.language, w.postLanguageUrl)
}

func (w *WikipediaClient) SetBaseUrl(baseUrl string) {
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

func (w *WikipediaClient) ImagesResults() string {
	return w.imagesResults
}

func (w *WikipediaClient) LinksResults() string {
	return w.linksResults
}

func (w *WikipediaClient) CategoriesResults() string {
	return w.categoriesResults
}

func (w *WikipediaClient) SetImagesResults(imagesResults string) {
	w.imagesResults = imagesResults
}

func (w *WikipediaClient) SetLinksResults(linksResults string) {
	w.linksResults = linksResults
}

func (w *WikipediaClient) SetCategoriesResults(categoriesResults string) {
	w.categoriesResults = categoriesResults
}

func query(w Wikipedia, q map[string][]string, v interface{}) error {
	resp, err := http.Get(fmt.Sprintf("%s?%s", w.GetBaseUrl(), url.Values(q).Encode()))
	if err != nil {
		return newError(ResponseError, err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&v)
	if err != nil {
		return newError(ResponseError, err)
	}
	return nil
}

func processResults(v interface{}, field string) ([]string, error) {
	results := make([]string, 0)
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
		return nil, newError(ResponseError, errors.New("invalid json response"))
	}
	return results, nil
}

func (w *WikipediaClient) PreLanguageUrl() string {
	return w.preLanguageUrl
}

func (w *WikipediaClient) PostLanguageUrl() string {
	return w.postLanguageUrl
}

func (w *WikipediaClient) Language() string {
	return w.language
}

func (w *WikipediaClient) SearchResults() int {
	return w.searchResults
}

func (w *WikipediaClient) GetLanguages() ([]Language, error) {
	var f interface{}
	err := query(w, map[string][]string{
		"meta":   {"siteinfo"},
		"siprop": {"languages"},
		"format": {"json"},
		"action": {"query"},
	}, &f)
	if err != nil {
		return nil, err
	}
	gotLangs := false
	languages := make([]Language, 0)
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
		return nil, newError(ResponseError, errors.New("invalid json response"))
	}
	return languages, nil
}

func (w *WikipediaClient) Search(q string) ([]string, error) {
	var f interface{}
	err := query(w, map[string][]string{
		"list":     {"search"},
		"srpop":    {""},
		"srlimit":  {fmt.Sprintf("%d", w.searchResults)},
		"srsearch": {q},
		"format":   {"json"},
		"action":   {"query"},
	}, &f)
	if err != nil {
		return nil, err
	}
	return processResults(f, "search")
}

func (w *WikipediaClient) Geosearch(latitude float64, longitude float64, radius int) ([]string, error) {
	if latitude < -90.0 || latitude > 90.0 {
		return nil, newError(ParameterError, errors.New("invalid latitude"))
	}
	if longitude < -180.0 || longitude > 180.0 {
		return nil, newError(ParameterError, errors.New("invalid longitude"))
	}
	if radius < -10 || radius > 10000 {
		return nil, newError(ParameterError, errors.New("invalid radius"))
	}
	var f interface{}
	err := query(w, map[string][]string{
		"list":     {"geosearch"},
		"gsradius": {fmt.Sprintf("%d", radius)},
		"gscoord":  {fmt.Sprintf("%f|%f", latitude, longitude)},
		"gslimit":  {fmt.Sprintf("%d", w.searchResults)},
		"format":   {"json"},
		"action":   {"query"},
	}, &f)
	if err != nil {
		return nil, err
	}
	return processResults(f, "geosearch")
}

func (w *WikipediaClient) RandomCount(count uint) ([]string, error) {
	var f interface{}
	err := query(w, map[string][]string{
		"list":        {"random"},
		"rnnamespace": {"0"},
		"rnlimit":     {fmt.Sprintf("%d", count)},
		"format":      {"json"},
		"action":      {"query"},
	}, &f)
	if err != nil {
		return nil, err
	}
	return processResults(f, "random")
}

func (w *WikipediaClient) Random() (string, error) {
	results, err := w.RandomCount(1)
	if err != nil {
		return "", err
	}
	if len(results) == 0 {
		return "", newError(ResponseError, errors.New("Got no results"))
	}
	return results[0], nil
}
