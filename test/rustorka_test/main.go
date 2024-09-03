package main

import (
	"fmt"
	"net/url"
	"os"
	"sync"
	"sync/atomic"

	"github.com/vanyason/infozone-gommaraizer/pkg/logger"
	"github.com/vanyason/infozone-gommaraizer/pkg/scrapper/rustorka"
	"github.com/vanyason/infozone-gommaraizer/pkg/utils"
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
			logger.Debug("Topic", "URL", topic)
		}
	}

	// Test fetching topics
	// Worker pool pattern
	logger.Info("Fetching topics concurrently")
	logger.Info("----------------------------")

	var wg sync.WaitGroup
	var errorWhileFetching atomic.Bool
	var counter atomic.Int32

	topicDescriptionsChan := make(chan rustorka.TopicDescription, len(topicDescriptions))
	numWorkers := utils.OptimalGoroutines(true)

	for i := 0; i < numWorkers; i++ { //< spawn workers
		wg.Add(1)

		go func() {
			defer wg.Done()

			for topic := range topicDescriptionsChan {
				newJar, err := utils.CopyJar(jar, u)
				if err != nil {
					logger.Error("Error while copying cookie jar", "error", err.Error())
					errorWhileFetching.Store(true)
					continue
				}

				html, err := rustorka.FetchTopicHTML(topic, newJar)

				if err != nil {
					logger.Error("Error while fetching topic HTML", "topic Title", topic.Title, "topic URL", topic.URL, "error", err.Error())
					errorWhileFetching.Store(true)
					continue
				}

				filename := fmt.Sprintf("rustorka-%d.html", counter.Load())
				counter.Add(1)
				logger.Info("Topic HTML fetched. Saving to file", "topic Title", topic.Title, "topic URL", topic.URL, "file", filename)
				if err := os.WriteFile(filename, []byte(html), 0644); err != nil {
					logger.Error("Error while saving topic HTML", "topic Title", topic.Title, "error", err.Error())
					errorWhileFetching.Store(true)
				}
			}
		}()
	}

	for _, topic := range topicDescriptions { //< fill channel
		topicDescriptionsChan <- topic
	}
	close(topicDescriptionsChan)

	wg.Wait() //< wait for workers

	if errorWhileFetching.Load() {
		logger.Error("Error while fetching topics HTML. Return")
		return
	}

	logger.Info("Topics HTML fetched and saved")
	logger.Info("----------------------------")

	// Test parsing topics
}
