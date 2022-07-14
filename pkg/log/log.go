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

func Infof(format string, args ...any) {
	Logger.Info().Msg(fmt.Sprintf(format, args...))
}

func Info(msg string) {
	Logger.Info().Msg(msg)
}

func Warnf(format string, args ...any) {
	Logger.Warn().Msg(fmt.Sprintf(format, args...))
}

func Warn(msg string) {
	Logger.Warn().Msg(msg)
}

func Errorf(format string, args ...any) {
	Logger.Error().Msg(fmt.Sprintf(format, args...))
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
