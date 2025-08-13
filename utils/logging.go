package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type AppLogger struct {
	zerolog.Logger // Is it better to do emmbedding or private field?
}

var Logger *AppLogger

func NewLogger() *AppLogger {
	output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	output.FormatLevel = func(i any) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}
	output.FormatFieldName = func(i any) string {
		return fmt.Sprintf("%s:", i)
	}
	output.FormatFieldValue = func(i any) string {
		return fmt.Sprintf("%s", i)
	}
	output.FormatErrFieldName = func(i any) string {
		return fmt.Sprintf("%s: ", i)
	}

	zerolog := zerolog.New(output).With().Caller().Timestamp().Logger()
	return &AppLogger{zerolog}
}

func init() {
	Logger = NewLogger()
}

func (l *AppLogger) LogInfo() *zerolog.Event {
	return l.Logger.Info()
}

func (l *AppLogger) LogError() *zerolog.Event {
	return l.Logger.Error()
}

func (l *AppLogger) LogDebug() *zerolog.Event {
	return l.Logger.Debug()
}

func (l *AppLogger) LogWarn() *zerolog.Event {
	return l.Logger.Warn()
}

func (l *AppLogger) LogFatal() *zerolog.Event {
	return l.Logger.Fatal()
}
