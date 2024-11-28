package scrapper

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/vanyason/infozone-gommaraizer/pkg/logger"
	"github.com/vanyason/infozone-gommaraizer/pkg/utils"
	"golang.org/x/net/html"
)

const (
	ForumURL    = "https://rustorka.com/forum/"
	LoginURL    = "https://rustorka.com/forum/login.php"
	Top30URL    = "https://rustorka.com/forum/top.php?mode=release&stat=30days"
	UserAgent   = "Mozilla/5.0 (X11; Linux x86_64; rv:129.0) Gecko/20100101 Firefox/129.0"
	AuthPayload = "login_username=asdasd&login_password=asdasd&autologin=1&login=%C2%F5%EE%E4"
)

type RustorkaScrapper struct {
	Scrapper
	client *http.Client
}

func NewRustorkaScrapper(client *http.Client) (*RustorkaScrapper, error) {
	if client == nil {
		panic("client is nil")
	}

	r := &RustorkaScrapper{
		client: client,
	}

	if err := r.getRustorkaAuthToken(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *RustorkaScrapper) Scrap() ([]ScrapRecord, error) {
	// Get top 30 downloads html
	top30HTML, err := r.getTop30()
	if err != nil {
		return nil, err
	}

	// Save the html
	if err := utils.Save("rustorka_top30.html", top30HTML); err != nil {
		logger.Warn("Failed to save Rustorka top 30 downloads HTML", "error", err.Error())
	}

	// Extract links from the html
	links, err := r.parseTop30(top30HTML)
	if err != nil {
		return nil, err
	}

	// Scrape each link via separate goroutine
	htmlChan := make(chan string, len(links))
	errorsChan := make(chan error, len(links))
	wg := sync.WaitGroup{}
	wg.Add(len(links))

	logger.Info("Getting Rustorka topics pages", "count", len(links))

	for idx, link := range links {
		go func(idx int, link linkToPage) {
			defer wg.Done()

			logger.Info("Getting", "title", link.Title, "url", link.URL)

			req, _ := http.NewRequest("GET", link.URL, nil)
			req.Header.Add("User-Agent", UserAgent)

			html, err := fetchHTML(r.client, req)
			if err != nil {
				errorsChan <- err
			} else {
				htmlChan <- html
				if err := utils.Save(fmt.Sprintf("rustorka_topic_%d.html", idx), html); err != nil {
					logger.Warn("Failed to save Rustorka page HTML", "error", err.Error())
				}
			}
		}(idx, link)
	}

	wg.Wait()
	close(htmlChan)
	close(errorsChan)

	if len(errorsChan) > 0 {
		for err := range errorsChan {
			logger.Error(err.Error())
		}
		return nil, fmt.Errorf("russtorka: %d errors while scraping", len(errorsChan))
	}

	// Parse each topic page
	records := make([]ScrapRecord, 0, len(htmlChan))

	logger.Info("Parsing Rustorka topics pages", "count", len(links))
	for html := range htmlChan {
		rec, err := r.parseTopic(html)
		if err != nil {
			return nil, err
		}

		records = append(records, *rec)
	}

	return records, nil
}

func (r *RustorkaScrapper) ScrapForSummary() ([]ScrapRecord, error) {
	return nil, nil
}

// To parse Rustorka - auth token is needed.
// This method visits login page and gets the auth token.
// Should be called before any further actions.
func (r *RustorkaScrapper) getRustorkaAuthToken() error {
	logger.Info("Getting Rustorka auth token")

	req, _ := http.NewRequest("POST", LoginURL, strings.NewReader(AuthPayload))
	req.Header.Add("User-Agent", UserAgent)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Send the request
	resp, err := r.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %s", resp.Status)
	}

	return nil
}

func (r *RustorkaScrapper) getTop30() (string, error) {
	logger.Info("Getting Rustorka last 30 days top downloads")

	req, _ := http.NewRequest("GET", Top30URL, nil)
	req.Header.Add("User-Agent", UserAgent)

	return fetchHTML(r.client, req)
}

func (r *RustorkaScrapper) parseTop30(htmlPage string) ([]linkToPage, error) {
	logger.Info("Parsing Rustorka's Top 30 downloads HTML")

	doc, err := html.Parse(strings.NewReader(htmlPage))
	if err != nil {
		return nil, err
	}

	links := make([]linkToPage, 0, 30)

	var parse func(n *html.Node)
	parse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && strings.Contains(attr.Val, "viewtopic.php?t=") {
					links = append(links, linkToPage{n.FirstChild.Data, ForumURL + attr.Val})
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parse(c)
		}
	}

	parse(doc)

	return links, nil
}

func (r *RustorkaScrapper) parseTopic(htmlPage string) (*ScrapRecord, error) {
	logger.Info("Parsing Rustorka's topic")

	doc, err := html.Parse(strings.NewReader(htmlPage))
	if err != nil {
		return nil, err
	}

	record := &ScrapRecord{}

	var parse func(n *html.Node)
	parse = func(n *html.Node) {
		// Parse Title
		// Parse Image
		// Parse Description
		// Parse Other images
		// Parse Extra (downloads)
		if n.Type == html.ElementNode && n.Data == "span" {
			for _, attr := range n.Attr {
				if attr.Key == "title" && strings.Contains(attr.Val, "Скачано") {
					num, err := stringToIntFirstSlpit(n.FirstChild.Data)
					if err != nil {
						logger.Warn("Failed to convert string to int", "error", err.Error(), "string", n.FirstChild.Data)
					} else {
						record.Extra = num
					}
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			parse(c)
		}
	}

	parse(doc)
	return record, nil
}

func stringToIntFirstSlpit(input string) (int, error) {
	// Split the string by spaces to separate the number from the unit
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return 0, fmt.Errorf("input string is empty or invalid")
	}

	// Parse the first part as an integer
	number, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, fmt.Errorf("failed to convert string to int: %w", err)
	}

	return number, nil
}
