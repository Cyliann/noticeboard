package db

import (
	"database/sql"
	"example/downdetector/internal/utils"

	"github.com/charmbracelet/log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

// creates a default user with password 'changeme' if users table is empty
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

INSERT INTO users (username, password, salt)
SELECT 'admin', '9ca53ef06fbb9b87ddb126147bf346adbf6e79691073b19c5c07bfec1f384b2d', 'salt' 
WHERE NOT EXISTS (SELECT 1 FROM users);
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
