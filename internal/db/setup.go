package db

import (
	"database/sql"
	"joynext/downdetector/internal/utils"

	"github.com/charmbracelet/log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

var schema = `
CREATE TABLE IF NOT EXISTS reports (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  title TEXT,
  content TEXT,
  isSolved BOOLEAN
);

CREATE TABLE IF NOT EXISTS users (
  username TEXT NOT NULL PRIMARY KEY,
  password TEXT NOT NULL,
  salt TEXT NOT NULL
);
`

// Sets up a connection to the database
func Connect() error {
	utils.NoReportLog.Info("Connecting to db...")
	var err error
	DB, err = sql.Open("sqlite3", "reports.db")
	if err != nil {
		return err
	}

	_, err = DB.Exec(schema)
	if err != nil {
		log.Error("Error executing schema", "err", err)
		return err
	}

	utils.NoReportLog.Info("Connected")
	return nil
}
