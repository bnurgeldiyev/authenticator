package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
)

// Logger -.
type Logger struct {
	*zerolog.Logger
}

// New function returns new Logger instance.
func New(level string, dir string) (*zerolog.Logger, error) {
	logger := zerolog.New(os.Stdout)

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return &logger, err
	}

	if lvl == zerolog.NoLevel {
		logger = zerolog.Nop()
		return &logger, nil
	}

	logPath := getPath(dir)

	var logFile *os.File
	logFile, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return &logger, errors.Wrap(err, fmt.Sprintf("logPath: %s", logPath))
	}

	mw := io.MultiWriter(os.Stdout, logFile)

	customFormatter := zerolog.ConsoleWriter{
		Out:        mw,
		TimeFormat: time.RFC3339,
		FormatLevel: func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("[%s]", i))
		},
		FormatCaller: func(i interface{}) string {
			if i == nil {
				return ""
			}
			return filepath.Base(fmt.Sprintf("%s", i))
		},
		FormatMessage: func(i interface{}) string {
			return fmt.Sprintf("| %s", i)
		},
		FormatFieldName: func(i interface{}) string {
			return fmt.Sprintf("| %s: ", i)
		},
		FormatFieldValue: func(i interface{}) string {
			return fmt.Sprintf("%s", i)
		},
	}

	logger = zerolog.New(customFormatter).
		Level(lvl).
		With().
		CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount).
		Timestamp().
		Logger()

	return &logger, nil
}

func getPath(dir string) string {
	if len(dir) == 0 {
		dir = "."
	}
	_ = os.MkdirAll(dir, 0700)
	return filepath.Join(dir, fmt.Sprintf("daemon.%s.log", time.Now().Local().Format("20060102")))
}
