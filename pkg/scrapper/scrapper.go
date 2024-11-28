package scrapper

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/vanyason/infozone-gommaraizer/pkg/logger"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type ScrapRecord struct {
	Title            string
	Text             string
	MainPictureURL   string
	OtherPicturesURL []string
	Extra            any
}

type Scrapper interface {
	Scrap() ([]ScrapRecord, error)
	ScrapForSummary() ([]ScrapRecord, error)
}

func NewHTTPClient() *http.Client {
	jar, _ := cookiejar.New(nil)

	timeout := 30 * time.Second

	transport := &http.Transport{MaxIdleConns: 50, IdleConnTimeout: timeout}

	checkRedirect := func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 { // Limit the number of redirects to 10
			return http.ErrUseLastResponse
		}

		if len(via) > 0 { //< Merge cookies from previous responses into the new request
			previousResponse := via[len(via)-1]
			mergedCookies := jar.Cookies(previousResponse.URL)
			for _, cookie := range mergedCookies {
				req.AddCookie(cookie)
			}
		}

		return nil //< Follow the redirect
	}

	return &http.Client{
		Jar:           jar,
		Transport:     transport,
		CheckRedirect: checkRedirect,
		Timeout:       timeout,
	}
}

func fetchHTML(client *http.Client, req *http.Request) (string, error) {
	if client == nil {
		panic("client is nil")
	}

	if req == nil {
		panic("req is nil")
	}

	logger.Info("Fetching HTML from", "url", req.URL.String())

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	// Convert the response body from windows-1251 to UTF-8 if needed
	var reader io.Reader = resp.Body
	if resp.Header.Get("Content-Type") == "text/html; charset=windows-1251" {
		// logger.Debug("Converting response body from windows-1251 to UTF-8")
		reader = transform.NewReader(resp.Body, charmap.Windows1251.NewDecoder())
	}

	// Read the response body
	body, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("error reading the response body: %w", err)
	}

	// Return the HTML as a string
	return string(body), nil
}

type linkToPage struct {
	Title string
	URL   string
}
