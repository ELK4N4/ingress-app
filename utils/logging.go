package utils

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

var Logger *zerolog.Logger

func NewLogger() *zerolog.Logger {
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

	logger := zerolog.New(output).With().Caller().Timestamp().Logger()
	return &logger
}

func init() {
	Logger = NewLogger()
}
