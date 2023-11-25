package providers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/dxbednarczyk/browsrr/templates"
)

func One337X(w http.ResponseWriter, r *http.Request) {
	query := r.PostForm.Get("query")
	query = strings.Trim(query, " ")

	if len(query) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("query must be longer than 3 characters"))

		return
	}

	formatted := fmt.Sprintf("https://1337x.to/search/%s/1/", query)

	doc, err := scrapeSite(formatted)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("failed to scrape site: %v", err)))

		fmt.Fprintln(os.Stderr, err)
	}

	parsed := parseOne337XDocument(doc)

	w.Header().Add("Content-Type", "text/html")
	templates.One337XResultTemplate(parsed).Render(context.Background(), w)
}

func parseOne337XDocument(doc *goquery.Document) *templates.One337XResult {
	r := templates.One337XResult{}

	// weird that this space is here. catches all error messages.
	if doc.FindMatcher(goquery.Single("h1")).Text() == " Message:" {
		r.Errors = append(r.Errors, errors.New("no results found"))

		return &r
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
			go details(&r, mu, &wg, url)
		}
	})

	wg.Wait()

	return &r
}

func details(ts *templates.One337XResult, mu *sync.Mutex, wg *sync.WaitGroup, url string) {
	defer wg.Done()

	combined := "https://1337x.to" + url

	doc, err := scrapeSite(combined)
	if err != nil {
		ts.Errors = append(ts.Errors, err)
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
	ts.Items = append(ts.Items, t)
	mu.Unlock()
}
