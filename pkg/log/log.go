package log

import (
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

var (
	Logger zerolog.Logger
)

func Debugf(format string, args ...any) {
	Debug(fmt.Sprintf(format, args...))
}

func Debug(msg string) {
	Logger.Debug().Msg(msg)
}

func Infof(format string, args ...any) {
	Info(fmt.Sprintf(format, args...))
}

func Info(msg string) {
	Logger.Info().Msg(msg)
}

func Warnf(format string, args ...any) {
	Warn(fmt.Sprintf(format, args...))
}

func Warn(msg string) {
	Logger.Warn().Msg(msg)
}

func Errorf(format string, args ...any) {
	Error(fmt.Errorf(format, args...))
}

func Error(err error) {
	Logger.Error().Err(err).Msg("")
}

func init() {
	logLevel := "error"
	// Set global log level
	if level, found := os.LookupEnv("CONFIGMANAGER_LOG_LEVEL"); found {
		logLevel = strings.ToLower(level)
	}
	lvl, err := zerolog.ParseLevel(logLevel)
	if err != nil {
		panic(fmt.Errorf("StartUpLoggerFailed: %v", err))
	}
	Logger = zerolog.New(os.Stderr).With().Timestamp().Logger().Level(lvl)
}
