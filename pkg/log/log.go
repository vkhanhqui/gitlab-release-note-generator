package log

import (
	"gitLab-rls-note/pkg/errors"
	"log"
	"os"
)

type Logger interface {
	// Error will also send to Sentry if available
	Error(msg string, err error, args ...interface{})
	// Info logs at info level
	Info(msg string, args ...interface{})
	// Debug logs at debug level
	Debug(msg string, args ...interface{})

	// With adds structured context to the logger
	With(args ...interface{}) Logger
}

// PanicRecover captures the panic value, log it as error and then exit with code 1
func PanicRecover(logger Logger) {
	r := recover()
	if r == nil {
		return
	}

	err, ok := r.(error)
	if err != nil && ok {
		logger.Error("Panic with error", err)
		os.Exit(1)
		return
	}

	logger.Error("Panic occurred!", errors.Errorf("%s", r))
	os.Exit(1)
}

type LogLevel int

const (
	ErrorLevel LogLevel = 1
	InfoLevel  LogLevel = 2
	DebugLevel LogLevel = 3
)

// Print outputs a simple log line
func Print(args ...interface{}) {
	log.Print(args...)
}

func Printf(template string, args ...interface{}) {
	log.Printf(template, args...)
}
