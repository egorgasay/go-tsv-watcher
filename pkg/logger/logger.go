package logger

import "github.com/rs/zerolog"

// ILogger on case of changing Logger in the future
type ILogger interface {
	Info(msg string)
	Fatal(msg string)
	Debug(msg string)
	Warn(msg string)
}

// Logger is a default logger
type Logger struct {
	l zerolog.Logger
}

// New creates a new logger
func New(logger zerolog.Logger) ILogger {
	return &Logger{l: logger}
}

// Info logs info
func (l Logger) Info(msg string) {
	l.l.Info().Msg(msg)
}

// Fatal logs fatal
func (l Logger) Fatal(msg string) {
	l.l.Fatal().Msg(msg)
}

// Debug logs debug
func (l Logger) Debug(msg string) {
	l.l.Debug().Msg(msg)
}

// Warn logs warn
func (l Logger) Warn(msg string) {
	l.l.Warn().Msg(msg)
}
