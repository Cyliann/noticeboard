package main

import (
	"example/downdetector/internal/app"
	"example/downdetector/internal/utils"

	"github.com/charmbracelet/log"
)

// @title Downtetector
// @version 1.0
// @description API for Downdetector website

// @contact.name Maksymilian Cych
// @contact.email maksymilian@cych.eu

// @license.name GPLv3
// @license.url https://www.gnu.org/licenses/gpl-3.0.en.html
// @BasePath /api
// @Router /api

func main() {
	logFile := utils.SetupLogging()
	// Initialize the HTTP server.
	srv := app.SetupServer()

	// Connect to the database.
	err := app.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}

	// Handle graceful shutdown.
	app.GracefulShutdown(srv, logFile)
}
