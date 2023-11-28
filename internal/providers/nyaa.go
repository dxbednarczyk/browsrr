package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

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

func Nyaa(ctx echo.Context, sukebei bool) error {
	query, err := trimQuery(ctx)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	formatted := fmt.Sprintf("https://nyaa.si/?q=%s", query)
	if sukebei {
		formatted = fmt.Sprintf("https://sukebei.nyaa.si/?q=%s", query)
	}

	doc, err := scrapeSite(formatted)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to scrape site: %v", err))
	}


	r := templates.NyaaResult{}

	parseNyaaDocument(doc, &r)

	statusCode := http.StatusOK
	if r.Errors != nil {
		statusCode = http.StatusConflict
	}

	ctx.Response().Status = statusCode
	return templates.NyaaResultTemplate(&r).Render(context.Background(), ctx.Response().Writer)
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
