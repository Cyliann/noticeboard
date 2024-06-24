package app

import (
	"github.com/charmbracelet/log"
	"html/template"
	"joynext/downdetector/internal/db"
	"net/http"
	"path/filepath"
)

func RenderOpenReports(w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join("templates", "index.html")

	tmpl, err := template.ParseFiles(lp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err)
		return
	}

	data, err := db.GetOpenReports()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error(err)
		return
	}

}

func RenderDashboard(w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join("templates", "dashboard.html")

	tmpl, err := template.ParseFiles(lp)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err)
		return
	}

	data, err := db.GetAllReports()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error(err)
		return
	}
}

func ServeLogin(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/login.html")
}

func ServeNewReport(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/newReport.html")
}

func ServeChangePassword(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "./static/changePassword.html")
}
