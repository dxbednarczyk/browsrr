package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/PuerkitoBio/goquery"
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

func Nyaa(w http.ResponseWriter, r *http.Request) {
	query := r.PostForm.Get("query")
	query = strings.Trim(query, " ")

	formatted := fmt.Sprintf("https://nyaa.si/?q=%s", query)

	doc, err := scrapeSite(formatted)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to scrape site: %v", err)))

		fmt.Fprintln(os.Stderr, err)
	}

	parsed := parseNyaaDocument(doc)

	w.Header().Add("Content-Type", "text/html")
	templates.NyaaResultTemplate(parsed).Render(context.Background(), w)
}

func parseNyaaDocument(doc *goquery.Document) *templates.NyaaResult {
	r := templates.NyaaResult{}

	// if no results found error text
	if doc.FindMatcher(goquery.Single("h3")) != nil {
		r.Errors = append(r.Errors, errors.New("no results found"))

		return &r
	}

	doc.Find("tr .success, .default, .danger").Each(func(i int, s *goquery.Selection) {
		t := new(templates.NyaaTorrent)

		s.Find("td").Each(func(i int, s *goquery.Selection) {
			switch i {
			case category:
				t[i] = s.Children().First().AttrOr("title", "Unknown")
			case title:
				t[i] = s.Children().Last().AttrOr("title", "Unknown")
			case magnet:
				t[i] = s.Children().Last().AttrOr("magnet", "javascript:alert('Magnet link unavailable')")
			case date:
				t[i] = s.AttrOr("data-timestamp", "0")
			case size, seeders, leechers:
				t[i] = s.Text()
			}
		})

		r.Items = append(r.Items, *t)
	})

	return &r
}
