package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/dxbednarczyk/browsrr/templates"
	"github.com/labstack/echo/v4"
)

func One337X(ctx echo.Context) error {
	query, err := trimQuery(ctx)
	if err != nil {
		return ctx.String(http.StatusBadRequest, err.Error())
	}

	formatted := fmt.Sprintf("https://1337x.to/search/%s/1/", query)

	doc, err := scrapeSite(formatted)
	if err != nil {
		return ctx.String(http.StatusInternalServerError, fmt.Sprintf("failed to scrape site: %v", err))
	}

	r := templates.One337XResult{}

	parseOne337XDocument(doc, &r)

	statusCode := http.StatusOK
	if r.Errors != nil {
		statusCode = http.StatusConflict
	}

	ctx.Response().Status = statusCode
	return templates.One337XResultTemplate(&r).Render(context.Background(), ctx.Response().Writer)
}

func parseOne337XDocument(doc *goquery.Document, r *templates.One337XResult) {
	// _____________________________________ weird that ↓ this space is here.
	if doc.FindMatcher(goquery.Single("h1")).Text() == " Message:" {
		r.Errors = append(r.Errors, errors.New("no results found"))

		return
	}

	mu := &sync.Mutex{}

	var wg sync.WaitGroup

	doc.Find(".name a").Each(func(i int, s *goquery.Selection) {
		url, exists := s.Attr("href")
		if !exists {
			return
		}

		if strings.HasPrefix(url, "/torrent") {
			wg.Add(1)
			go details(r, mu, &wg, url)
		}
	})

	wg.Wait()
}

func details(r *templates.One337XResult, mu *sync.Mutex, wg *sync.WaitGroup, url string) {
	defer wg.Done()

	combined := "https://1337x.to" + url

	doc, err := scrapeSite(combined)
	if err != nil {
		r.Errors = append(r.Errors, err)

		return
	}

	t := templates.One337XTorrent{
		Info: make(map[string]string),
	}

	untrimmedName := doc.FindMatcher(goquery.Single("h1")).Text()
	t.Name = strings.Trim(untrimmedName, " ")

	t.Magnet, _ = doc.FindMatcher(goquery.Single("a[href^=magnet]")).Attr("href")

	doc.Find(".list").Each(func(i int, s *goquery.Selection) {
		if i == 0 {
			return
		}

		s.Find("li").Each(func(i int, s *goquery.Selection) {
			key := s.Find("strong").First().Text()
			value := s.Find("span").First().Text()

			t.Info[key] = strings.Trim(value, " ")
		})
	})

	mu.Lock()
	r.Items = append(r.Items, t)
	mu.Unlock()
}
