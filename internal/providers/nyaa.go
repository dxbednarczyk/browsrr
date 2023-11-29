package providers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
	"github.com/a-h/templ"
	"github.com/dxbednarczyk/browsrr/templates"
)

const (
	category int = iota
	title
	magnet
	size
	date
	seeders
	leechers
)

func Nyaa(w http.ResponseWriter, r *http.Request, sukebei bool) {
	query, err := trimQuery(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))

		return
	}

	formatted := fmt.Sprintf("https://nyaa.si/?q=%s", query)
	if sukebei {
		formatted = fmt.Sprintf("https://sukebei.nyaa.si/?q=%s", query)
	}

	doc, err := scrapeSite(formatted)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to scrape site: %v", err)))

		return
	}

	res := templates.NyaaResult{}

	parseNyaaDocument(doc, &res)

	h := templ.Handler(templates.NyaaResultTemplate(&res))

	if res.Errors != nil {
		h.Status = http.StatusConflict
	}

	h.ServeHTTP(w, r)
}

func parseNyaaDocument(doc *goquery.Document, r *templates.NyaaResult) {
	doc.Find("tr .success, .default, .danger").Each(func(i int, s *goquery.Selection) {
		t := new(templates.NyaaTorrent)

		s.Find("td").Each(func(i int, s *goquery.Selection) {
			switch i {
			case category:
				t[category] = s.Children().First().AttrOr("title", "Unknown")
			case title:
				t[title] = s.Children().Last().AttrOr("title", "Unknown")
			case magnet:
				t[magnet] = s.Children().Last().AttrOr("magnet", "javascript:alert('Magnet link unavailable')")
			case date:
				t[date] = s.AttrOr("data-timestamp", "0")
			case size, seeders, leechers:
				t[i] = s.Text()
			}
		})

		r.Items = append(r.Items, *t)
	})

	if len(r.Items) == 0 {
		r.Errors = append(r.Errors, errors.New("no results found"))
	}
}
