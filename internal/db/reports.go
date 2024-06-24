package db

import (
	"encoding/json"
	"io"
	"joynext/downdetector/internal/utils"
	"net/http"

	"github.com/charmbracelet/log"
)

type Report struct {
	ID       uint   `db:"id"`
	Title    string `db:"title"`
	Content  string `db:"content"`
	IsSolved bool   `db:"isSolved"`
}

type ReportList struct {
	Reports []Report
	IsEmpty bool
}

type NewReport struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// GetOpenReports retrieves a list of all open reports.
func GetOpenReports() (ReportList, error) {
	reports := ReportList{}
	rows, err := DB.Query("SELECT * FROM reports WHERE isSolved=false")
	if err != nil {
		return ReportList{}, err
	}

	for rows.Next() {
		report := Report{}
		err := rows.Scan(&report.ID, &report.Title, &report.Content, &report.IsSolved)
		if err != nil {
			return ReportList{}, err
		}
		reports.Reports = append(reports.Reports, report)
	}
	reports.IsEmpty = len(reports.Reports) == 0

	return reports, nil
}

// GetAllReports retrieves a list of all reports.
func GetAllReports() ([]Report, error) {
	var reports []Report
	rows, err := DB.Query("SELECT * FROM reports")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		report := Report{}
		err := rows.Scan(&report.ID, &report.Title, &report.Content, &report.IsSolved)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}

	return reports, nil
}

// AddReportHandler adds a new report.
//
// @Summary Add a new report
// @Description Adds a new report to the system
// @Tags reports
// @Accept json
// @Produce plain
// @Param report body NewReport true "New Report"
// @Success 201
// @Failure 400
// @Failure 500
// @Router /reports [post]
func AddReportHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Failed to read request body", "err", err)
		return
	}

	newReport := NewReport{}
	err = json.Unmarshal(body, &newReport)
	if err != nil {
		if _, ok := err.(*json.SyntaxError); ok {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Failed to unmarshall report", "err", err)
		return
	}

	if newReport.Title == "" || newReport.Content == "" {
		http.Error(w, "No title nor content can be empty", http.StatusBadRequest)
		return
	}

	_, err = DB.Exec("INSERT INTO reports (title, content, isSolved) VALUES (?, ?, ?)", newReport.Title, newReport.Content, false)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Failed to insert report", "err", err)
		return
	}

	ip := r.RemoteAddr
	utils.NoReportLog.Infof("%s created a report", ip)
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// EditReportHandler edits an existing report.
//
// @Summary Edit an existing report
// @Description Edits the details of an existing report
// @Tags reports
// @Accept mpfd
// @Produce plain
// @Param id path int true "Report ID"
// @Param title formData string true "Title"
// @Param content formData string true "Content"
// @Param isSolved formData bool true "Is Solved"
// @Success 200
// @Failure 500
// @Router /reports/{id} [put]
func EditReportHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	r.ParseMultipartForm(4 << 20) // max memory 4MB
	title := r.Form.Get("title")
	content := r.Form.Get("content")
	isSolved := r.Form.Get("isSolved") != "" // when submitting a form if a checkbox is unchecked it's not included in the payload instead of being false

	_, err := DB.Exec("UPDATE reports SET title=?, content=?, isSolved=? WHERE id=?", title, content, isSolved, id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Failed to update report", "err", err)
		return
	}

	ip := r.RemoteAddr
	utils.NoReportLog.Infof("%s edited report %s", ip, id)
	w.WriteHeader(http.StatusOK)
}

// DeleteReportHandler deletes an existing report.
//
// @Summary Delete a report
// @Description Deletes a report from the system
// @Tags reports
// @Param id path int true "Report ID"
// @Produce plain
// @Success 200
// @Failure 500
// @Router /reports/{id} [delete]
func DeleteReportHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, err := DB.Exec("DELETE FROM reports WHERE id=?", id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("Failed to delete user", "err", err)
		return
	}

	ip := r.RemoteAddr
	utils.NoReportLog.Infof("%s deleted report %s", ip, id)
	w.WriteHeader(http.StatusOK)
}
