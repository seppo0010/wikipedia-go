package wikipedia

import "errors"
import "fmt"
import "strings"

type Page interface {
	Id() (pageId string, err error)
	Title() (pageTitle string, err error)
	Content() (content string, err error)
	HtmlContent() (content string, err error)
	Summary() (summary string, err error)
	Images() <-chan ImageRequest
	Extlinks() <-chan ReferenceRequest
	Links() <-chan LinkRequest
	Categories() <-chan CategoryRequest
	Sections() (titles []string, err error)
	SectionContent(title string) (sectionContent string, err error)
}

type PageClient struct {
	wikipedia Wikipedia
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

type Category struct {
	Name string
}

type CategoriesRequest struct {
	categories []Category
	cont       map[string][]string
}

type CategoryRequest struct {
	Category Category
	Err      error
}

func NewPage(wikipedia Wikipedia, title string) *PageClient {
	return &PageClient{
		title:     title,
		wikipedia: wikipedia,
	}
}
func NewPageFromId(wikipedia Wikipedia, id string) *PageClient {
	return &PageClient{
		id:        id,
		wikipedia: wikipedia,
	}
}

func (page *PageClient) queryParam() (string, string) {
	if page.title != "" {
		return "titles", page.title
	}
	if page.id != "" {
		return "pageids", page.id
	}
	panic("Page must have a title or an id")
}

func (page *PageClient) redirect(r interface{}) (title string, redirect bool) {
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

func (page *PageClient) Id() (string, error) {
	if page.id != "" {
		return page.id, nil
	}
	k, v := page.queryParam()
	var f interface{}
	err := query(page.wikipedia, map[string][]string{
		"prop":      {"info|pageprops"},
		"inprop":    {"url"},
		"ppprop":    {"disambiguation"},
		"redirects": {""},
		"format":    {"json"},
		"action":    {"query"},
		k:           {v},
	}, &f)
	if err != nil {
		return "", err
	}
	if title, redirect := page.redirect(f); redirect {
		return NewPage(page.wikipedia, title).Id()
	}
	if v, ok := f.(map[string]interface{}); ok {
		if query, ok := v["query"].(map[string]interface{}); ok {
			if pages, ok := query["pages"].(map[string]interface{}); ok {
				for pageString := range pages {
					return pageString, nil
				}
			}
		}
	}
	return "", newError(ResponseError, errors.New("invalid json response"))
}

func (page *PageClient) Title() (string, error) {
	if page.title != "" {
		return page.title, nil
	}
	k, v := page.queryParam()
	var f interface{}
	err := query(page.wikipedia, map[string][]string{
		"prop":      {"info|pageprops"},
		"inprop":    {"url"},
		"ppprop":    {"disambiguation"},
		"redirects": {""},
		"format":    {"json"},
		"action":    {"query"},
		k:           {v},
	}, &f)
	if err != nil {
		return "", err
	}
	if title, redirect := page.redirect(f); redirect {
		return NewPage(page.wikipedia, title).Title()
	}
	if v, ok := f.(map[string]interface{}); ok {
		if query, ok := v["query"].(map[string]interface{}); ok {
			if pages, ok := query["pages"].(map[string]interface{}); ok {
				for _, page := range pages {
					if pageObject, ok := page.(map[string]interface{}); ok {
						if pageTitle, ok := pageObject["title"].(string); ok {
							return pageTitle, nil
						}
					}
				}
			}
		}
	}
	return "", newError(ResponseError, errors.New("invalid json response"))
}

func getFirstPage(f interface{}) (map[string]interface{}, bool) {
	if v, ok := f.(map[string]interface{}); ok {
		if query, ok := v["query"].(map[string]interface{}); ok {
			if pages, ok := query["pages"].(map[string]interface{}); ok {
				for _, page := range pages {
					if val, ok := page.(map[string]interface{}); ok {
						return val, true
					}
				}
			}
		}
	}
	return nil, false
}

func (page *PageClient) Content() (string, error) {
	k, v := page.queryParam()
	var f interface{}
	err := query(page.wikipedia, map[string][]string{
		"prop":        {"extracts|revisions"},
		"explaintext": {""},
		"rvprop":      {"ids"},
		"redirects":   {""},
		"format":      {"json"},
		"action":      {"query"},
		k:             {v},
	}, &f)
	if err != nil {
		return "", err
	}
	if title, redirect := page.redirect(f); redirect {
		return NewPage(page.wikipedia, title).Content()
	}
	if v, ok := getFirstPage(f); ok {
		if extract, ok := v["extract"].(string); ok {
			return extract, nil
		}
	}
	return "", newError(ResponseError, errors.New("invalid json response"))
}

func (page *PageClient) HtmlContent() (string, error) {
	k, v := page.queryParam()
	var f interface{}
	err := query(page.wikipedia, map[string][]string{
		"prop":        {"revisions"},
		"explaintext": {""},
		"rvprop":      {"content"},
		"rvlimit":     {"1"},
		"rvparse":     {""},
		"redirects":   {""},
		"format":      {"json"},
		"action":      {"query"},
		k:             {v},
	}, &f)
	if err != nil {
		return "", nil
	}
	if title, redirect := page.redirect(f); redirect {
		return NewPage(page.wikipedia, title).HtmlContent()
	}
	if v, ok := getFirstPage(f); ok {
		if revisions, ok := v["revisions"].([]interface{}); ok {
			for _, revisionInterface := range revisions {
				if revision, ok := revisionInterface.(map[string]interface{}); ok {
					if html, ok := revision["*"].(string); ok {
						return html, nil
					}
				}
			}
		}
	}
	return "", newError(ResponseError, errors.New("invalid json response"))
}

func (page *PageClient) Summary() (string, error) {
	k, v := page.queryParam()
	var f interface{}
	err := query(page.wikipedia, map[string][]string{
		"prop":        {"extracts"},
		"explaintext": {""},
		"exintro":     {""},
		"redirects":   {""},
		"format":      {"json"},
		"action":      {"query"},
		k:             {v},
	}, &f)
	if err != nil {
		return "", err
	}
	if title, redirect := page.redirect(f); redirect {
		return NewPage(page.wikipedia, title).Summary()
	}
	if v, ok := getFirstPage(f); ok {
		if extract, ok := v["extract"].(string); ok {
			return extract, nil
		}
	}
	return "", newError(ResponseError, errors.New("invalid json response"))
}

func parseCont(q interface{}) (map[string][]string, error) {
	params := make(map[string][]string)
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
					return nil, errors.New("invalid continue parameter")
				}
			}
		}
	}
	return params, nil
}

func (page *PageClient) requestImages(params map[string][]string) (*ImagesRequest, error) {
	k, v := page.queryParam()
	var f interface{}
	if len(params) == 0 {
		params["continue"] = []string{""}
	}
	for k, v := range map[string][]string{
		"generator": {"images"},
		"gimlimit":  {page.wikipedia.ImagesResults()},
		"prop":      {"imageinfo"},
		"iiprop":    {"url"},
		"format":    {"json"},
		"action":    {"query"},
		k:           {v},
	} {
		params[k] = v
	}
	err := query(page.wikipedia, params, &f)
	if err != nil {
		return nil, err
	}
	imagesRequest := new(ImagesRequest)
	imagesRequest.cont, err = parseCont(f)
	if err != nil {
		return nil, err
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
		return nil, newError(ResponseError, errors.New("invalid json response"))
	}
	return imagesRequest, nil

}

func (page *PageClient) Images() <-chan ImageRequest {
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

func (page *PageClient) requestExtlinks(params map[string][]string) (*ReferencesRequest, error) {
	k, v := page.queryParam()
	var f interface{}
	if len(params) == 0 {
		params["continue"] = []string{""}
	}
	for k, v := range map[string][]string{
		"prop":    {"extlinks"},
		"ellimit": {page.wikipedia.LinksResults()},
		"format":  {"json"},
		"action":  {"query"},
		k:         {v},
	} {
		params[k] = v
	}
	err := query(page.wikipedia, params, &f)
	if err != nil {
		return nil, err
	}
	referencesRequest := new(ReferencesRequest)
	referencesRequest.cont, err = parseCont(f)
	if err != nil {
		return nil, err
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
		return nil, newError(ResponseError, errors.New("invalid json response"))
	}
	return referencesRequest, nil

}

func (page *PageClient) Extlinks() <-chan ReferenceRequest {
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

func (page *PageClient) requestLinks(params map[string][]string) (*LinksRequest, error) {
	k, v := page.queryParam()
	var f interface{}
	if len(params) == 0 {
		params["continue"] = []string{""}
	}
	for k, v := range map[string][]string{
		"prop":        {"links"},
		"pllimit":     {page.wikipedia.LinksResults()},
		"plnamespace": {"0"},
		"format":      {"json"},
		"action":      {"query"},
		k:             {v},
	} {
		params[k] = v
	}
	err := query(page.wikipedia, params, &f)
	if err != nil {
		return nil, err
	}
	linksRequest := new(LinksRequest)
	linksRequest.cont, err = parseCont(f)
	if err != nil {
		return nil, err
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
		return nil, newError(ResponseError, errors.New("invalid json response"))
	}
	return linksRequest, nil

}

func (page *PageClient) Links() <-chan LinkRequest {
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

func (page *PageClient) requestCategories(params map[string][]string) (*CategoriesRequest, error) {
	k, v := page.queryParam()
	var f interface{}
	if len(params) == 0 {
		params["continue"] = []string{""}
	}
	for k, v := range map[string][]string{
		"prop":    {"categories"},
		"cllimit": {page.wikipedia.CategoriesResults()},
		"format":  {"json"},
		"action":  {"query"},
		k:         {v},
	} {
		params[k] = v
	}
	err := query(page.wikipedia, params, &f)
	if err != nil {
		return nil, err
	}
	categoriesRequest := new(CategoriesRequest)
	categoriesRequest.cont, err = parseCont(f)
	if err != nil {
		return nil, err
	}

	if v, ok := f.(map[string]interface{}); ok {
		if query, ok := v["query"].(map[string]interface{}); ok {
			if pages, ok := query["pages"].(map[string]interface{}); ok {
				for _, page := range pages {
					if v, ok := page.(map[string]interface{}); ok {
						if categories, ok := v["categories"].([]interface{}); ok {
							for _, elI := range categories {
								if el, ok := elI.(map[string]interface{}); ok {
									if name, ok := el["title"].(string); ok {
										categoriesRequest.categories = append(categoriesRequest.categories, Category{Name: name})
									}
								}
							}
						}
					}
				}
			}
		}
	}
	if len(categoriesRequest.categories) == 0 {
		return nil, newError(ResponseError, errors.New("invalid json response"))
	}
	return categoriesRequest, nil

}

func (page *PageClient) Categories() <-chan CategoryRequest {
	ch := make(chan CategoryRequest)
	go func() {
		defer close(ch)
		cont := make(map[string][]string)
		for {
			categoriesRequest, err := page.requestCategories(cont)
			if err != nil {
				ch <- CategoryRequest{Err: err}
				return
			}
			for _, category := range categoriesRequest.categories {
				ch <- CategoryRequest{Category: category}
			}
			cont = categoriesRequest.cont
			if len(cont) == 0 {
				break
			}
		}
	}()
	return ch
}

func (page *PageClient) Sections() ([]string, error) {
	id, err := page.Id()
	if err != nil {
		return nil, err
	}
	var f interface{}
	err = query(page.wikipedia, map[string][]string{
		"prop":   {"sections"},
		"format": {"json"},
		"action": {"parse"},
		"pageid": {id},
	}, &f)
	if err != nil {
		return nil, err
	}

	titles := make([]string, 0)
	if v, ok := f.(map[string]interface{}); ok {
		if parse, ok := v["parse"].(map[string]interface{}); ok {
			if sections, ok := parse["sections"].([]interface{}); ok {
				for _, section := range sections {
					if v, ok := section.(map[string]interface{}); ok {
						if line, ok := v["line"].(string); ok {
							titles = append(titles, line)
						}
					}
				}
			}
		}
	}
	if len(titles) == 0 {
		return nil, newError(ResponseError, errors.New("invalid json response"))
	}
	return titles, nil
}

func (page *PageClient) SectionContent(title string) (string, error) {
	content, err := page.Content()
	if err != nil {
		return "", err
	}
	headr := fmt.Sprintf("== %s ==", title)
	index := strings.Index(content, headr)
	if index == -1 {
		return "", nil
	}
	index += len(headr)
	end := strings.Index(content[index:], "==")
	if end == -1 {
		end = len(content)
	} else {
		end += index
	}
	return content[index:end], nil
}
