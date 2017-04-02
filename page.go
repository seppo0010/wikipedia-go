package wikipedia

import "errors"

type Page struct {
	wikipedia *Wikipedia
	title, id string
}

func NewPage(wikipedia *Wikipedia, title string) *Page {
	return &Page{
		title:     title,
		wikipedia: wikipedia,
	}
}
func NewPageFromId(wikipedia *Wikipedia, id string) *Page {
	return &Page{
		id:        id,
		wikipedia: wikipedia,
	}
}

func (page *Page) queryParam() (string, string) {
	if page.title != "" {
		return "titles", page.title
	}
	if page.id != "" {
		return "pageids", page.id
	}
	panic("Page must have a title or an id")
}

func (page *Page) redirect(r interface{}) (title string, redirect bool) {
	if v, ok := r.(map[string]interface{}); ok {
		if query, ok := v["query"].(map[string]interface{}); ok {
			if redirects, ok := query["redirects"].([]interface{}); ok {
				if len(redirects) > 0 {
					if to, ok := redirects[0].(string); ok {
						if to != "" {
							redirect = true
							title = to
						}
					}
				}
			}
		}
	}
	return
}

func (page *Page) Id() (pageId string, err *WikipediaError) {
	if page.id != "" {
		pageId = page.id
		return
	}
	k, v := page.queryParam()
	var f interface{}
	err = page.wikipedia.query(map[string][]string{
		"prop":      []string{"info|pageprops"},
		"inprop":    []string{"url"},
		"ppprop":    []string{"disambiguation"},
		"redirects": []string{""},
		"format":    []string{"json"},
		"action":    []string{"query"},
		k:           []string{v},
	}, &f)
	if err != nil {
		return
	}
	if title, redirect := page.redirect(f); redirect {
		pageId, err = NewPage(page.wikipedia, title).Id()
		return
	}
	if v, ok := f.(map[string]interface{}); ok {
		if query, ok := v["query"].(map[string]interface{}); ok {
			if pages, ok := query["pages"].(map[string]interface{}); ok {
				for pageString, _ := range pages {
					pageId = pageString
					break
				}
			}
		}
	}
	if pageId == "" {
		err = newError(ResponseError, errors.New("invalid json response"))
	}
	return
}

func (page *Page) Title() (pageTitle string, err *WikipediaError) {
	if page.title != "" {
		pageTitle = page.title
		return
	}
	k, v := page.queryParam()
	var f interface{}
	err = page.wikipedia.query(map[string][]string{
		"prop":      []string{"info|pageprops"},
		"inprop":    []string{"url"},
		"ppprop":    []string{"disambiguation"},
		"redirects": []string{""},
		"format":    []string{"json"},
		"action":    []string{"query"},
		k:           []string{v},
	}, &f)
	if err != nil {
		return
	}
	if title, redirect := page.redirect(f); redirect {
		pageTitle, err = NewPage(page.wikipedia, title).Title()
		return
	}
	if v, ok := f.(map[string]interface{}); ok {
		if query, ok := v["query"].(map[string]interface{}); ok {
			if pages, ok := query["pages"].(map[string]interface{}); ok {
				for _, page := range pages {
					if pageObject, ok := page.(map[string]interface{}); ok {
						if pageTitle, ok = pageObject["title"].(string); ok {
							break
						}
					}
				}
			}
		}
	}
	if pageTitle == "" {
		err = newError(ResponseError, errors.New("invalid json response"))
	}
	return
}

func getFirstPage(f interface{}) (value map[string]interface{}, found bool) {
	if v, ok := f.(map[string]interface{}); ok {
		if query, ok := v["query"].(map[string]interface{}); ok {
			if pages, ok := query["pages"].(map[string]interface{}); ok {
				for _, page := range pages {
					if val, ok := page.(map[string]interface{}); ok {
						value = val
						found = true
						break
					}
				}
			}
		}
	}
	return
}

func (page *Page) Content() (content string, err *WikipediaError) {
	k, v := page.queryParam()
	var f interface{}
	err = page.wikipedia.query(map[string][]string{
		"prop":        []string{"extracts|revisions"},
		"explaintext": []string{""},
		"rvprop":      []string{"ids"},
		"redirects":   []string{""},
		"format":      []string{"json"},
		"action":      []string{"query"},
		k:             []string{v},
	}, &f)
	if err != nil {
		return
	}
	if title, redirect := page.redirect(f); redirect {
		content, err = NewPage(page.wikipedia, title).Content()
		return
	}
	if v, ok := getFirstPage(f); ok {
		if extract, ok := v["extract"].(string); ok {
			content = extract
		}
	}
	if content == "" {
		err = newError(ResponseError, errors.New("invalid json response"))
	}
	return
}
