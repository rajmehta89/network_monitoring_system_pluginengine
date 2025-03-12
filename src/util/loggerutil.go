package util

import (
	"log"
	"os"
	"sync"
)

type Logger struct {
	instance *log.Logger
}

var (
	once        sync.Once
	logInstance *Logger
)

// InitializeLogger initializes logging to a file
func InitializeLogger() *Logger {

	once.Do(func() {

		if _, err := os.Stat("logs"); os.IsNotExist(err) {

			os.Mkdir("logs", 0755)

		}

		// Open or create the log file
		logFile, err := os.OpenFile("logs/app.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

		if err != nil {

			log.Fatalf("Failed to open log file: %v", err)

		}

		// Initialize logger without filename and line number
		logger := log.New(logFile, "", log.Ldate|log.Ltime|log.Lmicroseconds)

		logInstance = &Logger{instance: logger}

	})

	return logInstance

}

func (l *Logger) LogInfo(message string) {

	l.instance.Println("INFO:", message)

}

func (l *Logger) LogError(err error) {

	if err != nil {

		l.instance.Println("ERROR:", err)

	}

}

func (l *Logger) LogWarning(message string) {

	l.instance.Println("WARNING:", message)

}
