package providers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-chi/chi/v5"
)

type One337XResult struct {
	mu     *sync.Mutex
	wg     sync.WaitGroup
	Items  []torrent `json:"items"`
	Errors []error   `json:"errors"`
}

type torrent struct {
	Name   string
	Magnet string
}

func One337X(w http.ResponseWriter, r *http.Request) {
	encoded := chi.URLParam(r, "encoded")

	decodedBytes, err := base64.URLEncoding.DecodeString(encoded)
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error decoding url"))
		return
	}

	decoded := string(decodedBytes)

	resp, err := http.Get(decoded)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("error fetching result"))
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error creating document"))
		return
	}

	parsed := parseDocument(doc)

	j, err := json.Marshal(&parsed)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("error marshaling response"))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(j)
}

func parseDocument(doc *goquery.Document) *One337XResult {
	r := One337XResult{
		mu: &sync.Mutex{},
	}

	doc.Find(".name a").Each(func(i int, s *goquery.Selection) {
		url, exists := s.Attr("href")
		if !exists {
			return
		}

		if strings.HasPrefix(url, "/torrent/") {
			r.wg.Add(1)
			go details(&r, url)
		}
	})

	r.wg.Wait()

	return &r
}

func details(ts *One337XResult, url string) {
	defer ts.wg.Done()

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

	untrimmedName := doc.Find("h1").First().Text()
	name := strings.Trim(untrimmedName, " ")

	magnet, _ := doc.Find("a[href^=magnet]").First().Attr("href")

	t := torrent{
		Name:   name,
		Magnet: magnet,
	}

	ts.mu.Lock()
	ts.Items = append(ts.Items, t)
	ts.mu.Unlock()
}
