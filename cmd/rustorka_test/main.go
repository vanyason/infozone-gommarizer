package main

import (
	"github.com/vanyason/infozone-gommaraizer/pkg/logger"
	"github.com/vanyason/infozone-gommaraizer/pkg/scrapper"
)

func main() {
	logger.Info("Testing Rustorka")

	client := scrapper.NewHTTPClient()
	rustorka, err := scrapper.NewRustorkaScrapper(client)
	if err != nil {
		logger.Error("❌ Error while creating RustorkaScrapper", "error", err.Error())
		return
	}

	logger.Info("Getting records")
	records, err := rustorka.Scrap()
	if err != nil {
		logger.Error("❌ Error while getting records", "error", err.Error())
		return
	}
	for _, r := range records {
		logger.Debug("Record", "title", r.Title)
	}

	logger.Info("Getting records for summary")
	recordsForSummary, err := rustorka.ScrapForSummary()
	if err != nil {
		logger.Error("❌ Error while getting records for summary", "error", err.Error())
		return
	}
	for _, r := range recordsForSummary {
		logger.Debug("Record for summary", "title", r.Title)
	}

	logger.Info("✅ Done")
}
