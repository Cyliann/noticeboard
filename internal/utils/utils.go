package utils

import (
	"io"
	"math/rand"
	"os"
	"time"

	"github.com/charmbracelet/log"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

var NoReportLog = log.New(os.Stderr)

// GenerateRandomString generates a random string of the specified length.
func GenerateRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func SetupLogging() *os.File {
	// log.SetLevel(log.DebugLevel) // for developement purpose
	logFile, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	mw := io.MultiWriter(os.Stdout, logFile)

	log.SetOutput(mw)
	NoReportLog.SetOutput(mw)

	log.SetReportCaller(true)
	NoReportLog.SetReportTimestamp(true)

	return logFile
}
