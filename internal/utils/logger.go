package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

var Logger *logrus.Entry

func InitLogs() *os.File {
	// open a file
	f, err := os.OpenFile("errors.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		logrus.Error(err)
	}

	logrus.SetReportCaller(true)
	// Log as JSON instead of the default ASCII formatter.
	logrus.SetFormatter(&logrus.TextFormatter{})

	// Output to stderr instead of stdout, could also be a file.
	logrus.SetOutput(f)

	// Only log the warning severity or above.
	logrus.SetLevel(logrus.DebugLevel)

	Logger = logrus.NewEntry(logrus.StandardLogger())
	return f
}
func LoggerDesc(desc string) *logrus.Entry {
	Logger = Logger.WithField("desc", desc)
	return Logger
}

func GetLoggerDesc() string {
	desc := ""
	if v, ok := Logger.Data["desc"].(string); ok {
		desc = v
	}
	return desc
}
