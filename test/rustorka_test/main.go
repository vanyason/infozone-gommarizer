package main

import (
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

	// Test getting html page
	html, err := rustorka.FetchHTML("https://rustorka.com/forum/top.php?mode=release&stat=30days", jar)
	if err != nil {
		logger.Error("Error while fetching HTML", "error", err.Error())
		return
	} else {
		filename := "rustorka.html"
		logger.Info("HTML fetched. Saving to file", "file", filename)
		err := os.WriteFile(filename, []byte(html), 0644)
		if err != nil {
			logger.Error("Error while saving HTML", "error", err.Error())
			return
		}
	}

}
