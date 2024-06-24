package app

import (
	"context"
	"errors"
	_ "joynext/downdetector/docs"
	"joynext/downdetector/internal/db"
	"joynext/downdetector/internal/utils"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MadAppGang/httplog"
	"github.com/charmbracelet/log"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// SetupServer sets up the HTTP server and routes.
func SetupServer() *http.Server {
	// Serve static files from the ./static directory.
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("GET /static/", http.StripPrefix("/static/", fs))

	// Serve swagger documentation under /docs/
	http.Handle("GET /docs/*", httpSwagger.WrapHandler)

	// Set up static endpoint
	http.Handle("GET /", httplog.Logger(http.HandlerFunc(RenderOpenReports)))
	http.Handle("GET /dashboard", httplog.Logger(db.CheckIfUserLoggedIn(RenderDashboard)))
	http.Handle("GET /login", httplog.Logger(http.HandlerFunc(ServeLogin)))
	http.Handle("GET /zglos", httplog.Logger(db.CheckIfUserLoggedIn(ServeNewReport)))
	http.Handle("GET /changepassword", httplog.Logger(db.CheckIfUserLoggedIn(ServeChangePassword)))

	//Set up API endpoints
	// GET
	http.Handle("GET /api/logout", httplog.Logger(db.CheckIfUserLoggedIn(db.LogoutHandler)))
	http.Handle("GET /api/salt", httplog.Logger(http.HandlerFunc(db.GetSaltHandler)))
	http.Handle("GET /api/pepper", httplog.Logger(http.HandlerFunc(db.GetPepperHandler)))

	// POST, PUT and DELETE
	http.Handle("POST /api/reports", httplog.Logger(db.CheckIfUserLoggedIn(db.AddReportHandler)))
	http.Handle("POST /api/login", httplog.Logger(db.LoginMiddleware(db.SessionHandler)))
	http.Handle("PUT /api/reports/{id}", httplog.Logger(db.CheckIfUserLoggedIn(db.EditReportHandler)))
	http.Handle("PUT /api/changepassword", httplog.Logger(db.CheckIfUserLoggedIn(db.ChangePasswordHandler)))
	http.Handle("DELETE /api/reports/{id}", httplog.Logger(db.CheckIfUserLoggedIn(db.DeleteReportHandler)))

	// Initialize the HTTP server.
	srv := &http.Server{
		Handler: nil,
		Addr:    ":8080",
	}

	return srv
}

// ConnectDB initializes the database connection.
func ConnectDB() error {
	return db.Connect()
}

// GracefulShutdown handles server and database shutdown gracefully.
func GracefulShutdown(srv *http.Server, logFile *os.File) {
	idleConnsClosed := make(chan struct{})
	go func() {
		// Create channel for shutdown signals.
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt)
		signal.Notify(stop, syscall.SIGTERM)

		// Receive shutdown signals.
		<-stop
		utils.NoReportLog.Warn("Received interrupt")

		// Attempt to gracefully shut down the server.
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Error("Error shutting down server: %v", err)
		} else {
			utils.NoReportLog.Info("Server gracefully stopped")
		}

		// Close the database connection.
		if err := db.DB.Close(); err != nil {
			log.Error("Error closing database: %v", err)
		} else {
			utils.NoReportLog.Info("Database closed")
		}

		close(idleConnsClosed)
	}()

	utils.NoReportLog.Info("Serving http...")
	if err := srv.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed: %v", err)
		}
	}

	<-idleConnsClosed
	utils.NoReportLog.Info("Shutdown complete")
	logFile.WriteString("--------\n")
	logFile.Close()
}
