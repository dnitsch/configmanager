package log

import (
	"log"
	"os"
)

var (
	info     = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	warn     = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	errorout = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)
)

func Infof(format string, args ...interface{}) {
	info.Printf(format, args...)
}

func Info(msg string) {
	info.Print(msg)
}

func Warnf(format string, args ...interface{}) {
	warn.Printf(format, args...)
}

func Warn(msg string) {
	warn.Print(msg)
}

func Errorf(format string, args ...interface{}) {
	errorout.Printf(format, args...)
}

func Error(msg string) {
	errorout.Print(msg)
}
