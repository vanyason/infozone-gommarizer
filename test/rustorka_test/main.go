package main

import (
	"fmt"
	"net/url"
	"os"

	"github.com/vanyason/infozone-gommaraizer/pkg/logger"
	"github.com/vanyason/infozone-gommaraizer/pkg/scrapper/rustorka"
)

func main() {
	logger.Info("Rustorka test started")

	// Test Authentication
	jar, err := rustorka.GetRustorkaAuthToken()
	u, _ := url.Parse(rustorka.ForumURL)

	if err != nil {
		logger.Error("Error while getting auth cookie", "error", err.Error())
		return
	} else {
		logger.Info("Cookie retrieved")

		if len(jar.Cookies(u)) != 2 {
			logger.Warn("Cookies unexpected length", "expected", 2, "current", len(jar.Cookies(u)))
		}

		logger.Debug("CookieJar Dump", "cookie", jar)
		for _, cookie := range jar.Cookies(u) {
			logger.Debug("Cookie", "name", cookie.Name, "value", cookie.Value)
		}
	}

	// Test getting 30 days top html page
	html, err := rustorka.FetchLast30DaysHTML(jar)
	if err != nil {
		logger.Error("Error while fetching HTML", "error", err.Error())
		return
	} else {
		filename := "rustorka-top30.html"
		logger.Info("HTML fetched. Saving to file", "file", filename)
		err := os.WriteFile(filename, []byte(html), 0644)
		if err != nil {
			logger.Error("Error while saving HTML", "error", err.Error())
			return
		}
	}

	// Test parsing 30 days top html page - retrieve topic descriptions
	topicDescriptions, err := rustorka.ParseLast30DaysHTML(html)
	if err != nil {
		logger.Error("Error while parsing HTML", "error", err.Error())
		return
	} else {
		logger.Info("HTML parsed")
		logger.Info("Topic descriptions", "count", len(topicDescriptions))

		logger.Debug("As a manual tester, compare topic descriptions against original urls MANUALLY", "url", rustorka.Top30URL)
		for _, topic := range topicDescriptions {
			logger.Debug("Topic", "url", topic)
		}
	}

	// Test fetching topics
	errorWhileFetching := false
	for i, td := range topicDescriptions {
		html, err := rustorka.FetchTopicHTML(td, jar)
		if err != nil {
			logger.Error("Error while fetching topic HTML", "topicHeader", td.Header, "error", err.Error())
			errorWhileFetching = true
		} else {
			filename := fmt.Sprintf("rustorka-%d.html", i)
			logger.Info("Topic HTML fetched. Saving to file", "file", filename)
			err := os.WriteFile(filename, []byte(html), 0644)
			if err != nil {
				logger.Error("Error while saving topic HTML", "topicHeader", td.Header, "error", err.Error())
				errorWhileFetching = true
			}
		}
	}

	if errorWhileFetching {
		logger.Error("Error while fetching topics HTML. Return")
		return
	}

	// Test parsing topics
}
