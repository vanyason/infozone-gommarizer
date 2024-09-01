package main

import "github.com/vanyason/infozone-gommaraizer/pkg/logger"

func main() {
	logger.Debug("Testing DEBUG logger")
	logger.Info("Info level")
	logger.Warn("Warn level")
	logger.Error("Error", "error", "unknown")
}
