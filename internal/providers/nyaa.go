package providers

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/dxbednarczyk/browsrr/templates"
	"github.com/labstack/echo/v4"
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

func Nyaa(ctx echo.Context) error {
	query := ctx.FormValue("query")
	query = strings.Trim(query, " ")

	formatted := fmt.Sprintf("https://nyaa.si/?q=%s", query)

	r := templates.NyaaResult{}

	doc, err := scrapeSite(formatted)
	if err != nil {
		r.Errors = append(r.Errors, fmt.Errorf("failed to scrape site: %v", err))

		return templates.Render(ctx, http.StatusInternalServerError, templates.NyaaResultTemplate(&r))
	}

	parseNyaaDocument(doc, &r)

	statusCode := http.StatusOK
	if r.Errors != nil {
		statusCode = http.StatusConflict
	}

	return templates.Render(ctx, statusCode, templates.NyaaResultTemplate(&r))
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
