package wikipedia

import "errors"
import "fmt"

type Page struct {
	wikipedia *Wikipedia
	title, id string
}

type Image struct {
	Url, Title, DescriptionUrl string
}

type ImagesRequest struct {
	images []Image
	cont   map[string][]string
}

type ImageRequest struct {
	Image Image
	Err   error
}

type Reference struct {
	Url string
}

type ReferencesRequest struct {
	references []Reference
	cont       map[string][]string
}

type ReferenceRequest struct {
	Reference Reference
	Err       error
}

type Link struct {
	Title string
}

type LinksRequest struct {
	links []Link
	cont  map[string][]string
}

type LinkRequest struct {
	Link Link
	Err  error
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

func (page *Page) Id() (pageId string, err error) {
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

func (page *Page) Title() (pageTitle string, err error) {
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

func (page *Page) Content() (content string, err error) {
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

func (page *Page) HtmlContent() (content string, err error) {
	k, v := page.queryParam()
	var f interface{}
	err = page.wikipedia.query(map[string][]string{
		"prop":        []string{"revisions"},
		"explaintext": []string{""},
		"rvprop":      []string{"content"},
		"rvlimit":     []string{"1"},
		"rvparse":     []string{""},
		"redirects":   []string{""},
		"format":      []string{"json"},
		"action":      []string{"query"},
		k:             []string{v},
	}, &f)
	if err != nil {
		return
	}
	if title, redirect := page.redirect(f); redirect {
		content, err = NewPage(page.wikipedia, title).HtmlContent()
		return
	}
	if v, ok := getFirstPage(f); ok {
		if revisions, ok := v["revisions"].([]interface{}); ok {
			for _, revisionInterface := range revisions {
				if revision, ok := revisionInterface.(map[string]interface{}); ok {
					if html, ok := revision["*"].(string); ok {
						content = html
						break
					}
				}
			}
		}
	}
	if content == "" {
		err = newError(ResponseError, errors.New("invalid json response"))
	}
	return
}

func (page *Page) Summary() (summary string, err error) {
	k, v := page.queryParam()
	var f interface{}
	err = page.wikipedia.query(map[string][]string{
		"prop":        []string{"extracts"},
		"explaintext": []string{""},
		"exintro":     []string{""},
		"redirects":   []string{""},
		"format":      []string{"json"},
		"action":      []string{"query"},
		k:             []string{v},
	}, &f)
	if err != nil {
		return
	}
	if title, redirect := page.redirect(f); redirect {
		summary, err = NewPage(page.wikipedia, title).Summary()
		return
	}
	if v, ok := getFirstPage(f); ok {
		if extract, ok := v["extract"].(string); ok {
			summary = extract
		}
	}
	if summary == "" {
		err = newError(ResponseError, errors.New("invalid json response"))
	}
	return
}

func parseCont(q interface{}) (params map[string][]string, err error) {
	params = make(map[string][]string)
	if q2, ok := q.(map[string]interface{}); ok {
		if cont, ok := q2["continue"].(map[string]interface{}); ok {
			for k, vUntyped := range cont {
				switch v := vUntyped.(type) {
				case int:
					params[k] = []string{fmt.Sprintf("%d", v)}
				case nil:
					params[k] = []string{""}
				case bool:
					if v {
						params[k] = []string{"1"}
					} else {
						params[k] = []string{"0"}
					}
				case float64:
					params[k] = []string{fmt.Sprintf("%f", v)}
				case string:
					params[k] = []string{v}
				default:
					err = errors.New("invalid continue parameter")
					return
				}
			}
		}
	}
	return
}

func (page *Page) requestImages(params map[string][]string) (imagesRequest ImagesRequest, err error) {
	k, v := page.queryParam()
	var f interface{}
	if len(params) == 0 {
		params["continue"] = []string{""}
	}
	for k, v := range map[string][]string{
		"generator": []string{"images"},
		"gimlimit":  []string{page.wikipedia.imagesResults},
		"prop":      []string{"imageinfo"},
		"iiprop":    []string{"url"},
		"format":    []string{"json"},
		"action":    []string{"query"},
		k:           []string{v},
	} {
		params[k] = v
	}
	err = page.wikipedia.query(params, &f)
	if err != nil {
		return
	}
	imagesRequest.cont, err = parseCont(f)
	if err != nil {
		return
	}
	if v, ok := f.(map[string]interface{}); ok {
		if query, ok := v["query"].(map[string]interface{}); ok {
			if pages, ok := query["pages"].(map[string]interface{}); ok {
				for _, page := range pages {
					if v, ok := page.(map[string]interface{}); ok {
						title, _ := v["title"].(string)
						url := ""
						descriptionUrl := ""
						if imageinfo, ok := v["imageinfo"].([]interface{}); ok {
							if len(imageinfo) > 0 {
								if info, ok := imageinfo[0].(map[string]interface{}); ok {
									url, _ = info["url"].(string)
									descriptionUrl, _ = info["descriptionurl"].(string)
								}
							}
						}
						imagesRequest.images = append(imagesRequest.images, Image{Title: title, Url: url, DescriptionUrl: descriptionUrl})
					}
				}
			}
		}
	}
	if len(imagesRequest.images) == 0 {
		err = newError(ResponseError, errors.New("invalid json response"))
	}
	return

}

func (page *Page) Images() <-chan ImageRequest {
	ch := make(chan ImageRequest)
	go func() {
		defer close(ch)
		cont := make(map[string][]string)
		for {
			imagesRequest, err := page.requestImages(cont)
			if err != nil {
				ch <- ImageRequest{Err: err}
				return
			}
			for _, image := range imagesRequest.images {
				ch <- ImageRequest{Image: image}
			}
			cont = imagesRequest.cont
			if len(cont) == 0 {
				break
			}
		}
	}()
	return ch
}

func (page *Page) requestExtlinks(params map[string][]string) (referencesRequest ReferencesRequest, err error) {
	k, v := page.queryParam()
	var f interface{}
	if len(params) == 0 {
		params["continue"] = []string{""}
	}
	for k, v := range map[string][]string{
		"prop":    []string{"extlinks"},
		"ellimit": []string{page.wikipedia.linksResults},
		"format":  []string{"json"},
		"action":  []string{"query"},
		k:         []string{v},
	} {
		params[k] = v
	}
	err = page.wikipedia.query(params, &f)
	if err != nil {
		return
	}
	referencesRequest.cont, err = parseCont(f)
	if err != nil {
		return
	}

	if v, ok := f.(map[string]interface{}); ok {
		if query, ok := v["query"].(map[string]interface{}); ok {
			if pages, ok := query["pages"].(map[string]interface{}); ok {
				for _, page := range pages {
					if v, ok := page.(map[string]interface{}); ok {
						if extlinks, ok := v["extlinks"].([]interface{}); ok {
							for _, elI := range extlinks {
								if el, ok := elI.(map[string]interface{}); ok {
									if url, ok := el["*"].(string); ok {
										referencesRequest.references = append(referencesRequest.references, Reference{Url: url})
									}
								}
							}
						}
					}
				}
			}
		}
	}
	if len(referencesRequest.references) == 0 {
		err = newError(ResponseError, errors.New("invalid json response"))
	}
	return

}

func (page *Page) Extlinks() <-chan ReferenceRequest {
	ch := make(chan ReferenceRequest)
	go func() {
		defer close(ch)
		cont := make(map[string][]string)
		for {
			referencesRequest, err := page.requestExtlinks(cont)
			if err != nil {
				ch <- ReferenceRequest{Err: err}
				return
			}
			for _, reference := range referencesRequest.references {
				ch <- ReferenceRequest{Reference: reference}
			}
			cont = referencesRequest.cont
			if len(cont) == 0 {
				break
			}
		}
	}()
	return ch
}

func (page *Page) requestLinks(params map[string][]string) (linksRequest LinksRequest, err error) {
	k, v := page.queryParam()
	var f interface{}
	if len(params) == 0 {
		params["continue"] = []string{""}
	}
	for k, v := range map[string][]string{
		"prop":        []string{"links"},
		"pllimit":     []string{page.wikipedia.linksResults},
		"plnamespace": []string{"0"},
		"format":      []string{"json"},
		"action":      []string{"query"},
		k:             []string{v},
	} {
		params[k] = v
	}
	err = page.wikipedia.query(params, &f)
	if err != nil {
		return
	}
	linksRequest.cont, err = parseCont(f)
	if err != nil {
		return
	}

	if v, ok := f.(map[string]interface{}); ok {
		if query, ok := v["query"].(map[string]interface{}); ok {
			if pages, ok := query["pages"].(map[string]interface{}); ok {
				for _, page := range pages {
					if v, ok := page.(map[string]interface{}); ok {
						if links, ok := v["links"].([]interface{}); ok {
							for _, elI := range links {
								if el, ok := elI.(map[string]interface{}); ok {
									if title, ok := el["title"].(string); ok {
										linksRequest.links = append(linksRequest.links, Link{Title: title})
									}
								}
							}
						}
					}
				}
			}
		}
	}
	if len(linksRequest.links) == 0 {
		err = newError(ResponseError, errors.New("invalid json response"))
	}
	return

}

func (page *Page) Links() <-chan LinkRequest {
	ch := make(chan LinkRequest)
	go func() {
		defer close(ch)
		cont := make(map[string][]string)
		for {
			linksRequest, err := page.requestLinks(cont)
			if err != nil {
				ch <- LinkRequest{Err: err}
				return
			}
			for _, link := range linksRequest.links {
				ch <- LinkRequest{Link: link}
			}
			cont = linksRequest.cont
			if len(cont) == 0 {
				break
			}
		}
	}()
	return ch
}
