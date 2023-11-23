package providers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/dxbednarczyk/browsrr/templates"
)

func One337X(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	query := r.PostForm.Get("query")

	formatted := fmt.Sprintf("https://1337x.to/search/%s/1/", query)

	fmt.Println(formatted)

	resp, err := http.Get(formatted)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error fetching result"))

		return
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error creating document"))

		return
	}

	parsed := parseDocument(doc)

	w.Header().Add("Content-Type", "text/html")
	templates.ResultsTemplate(parsed).Render(context.Background(), w)
}

func parseDocument(doc *goquery.Document) *templates.Results {
	r := templates.Results{}

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

func details(ts *templates.Results, mu *sync.Mutex, wg *sync.WaitGroup, url string) {
	defer wg.Done()

	combined := "https://1337x.to" + url

	resp, err := http.Get(combined)
	if err != nil {
		ts.Errors = append(ts.Errors, err)
		return
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		ts.Errors = append(ts.Errors, err)
		return
	}

	t := templates.Torrent{
		Info: make(map[string]string),
	}

	untrimmedName := doc.Find("h1").First().Text()
	t.Name = strings.Trim(untrimmedName, " ")

	t.Magnet, _ = doc.Find("a[href^=magnet]").First().Attr("href")

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
