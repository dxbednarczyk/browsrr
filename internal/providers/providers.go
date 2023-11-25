package providers

import (
	"net/http"
	"net/http/cookiejar"

	"github.com/PuerkitoBio/goquery"
)

func scrapeSite(url string) (*goquery.Document, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Jar: jar,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
