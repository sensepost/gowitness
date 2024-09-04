package log

import (
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

// LLogger is a charmbracelet logger type redefinition
type LLogger = log.Logger

// Logger is this package level logger
var Logger *LLogger

func init() {
	styles := log.DefaultStyles()
	styles.Keys["err"] = lipgloss.NewStyle().Foreground(lipgloss.Color("204"))
	styles.Values["err"] = lipgloss.NewStyle().Bold(true)

	Logger = log.NewWithOptions(os.Stderr, log.Options{
		ReportTimestamp: true,
	})
	Logger.SetStyles(styles)
	Logger.SetLevel(log.InfoLevel)
}

// EnableDebug enabled debug logging and caller reporting
func EnableDebug() {
	Logger.SetLevel(log.DebugLevel)
	Logger.SetReportCaller(true)
}

// EnableSilence will silence most logs, except this written with Print
func EnableSilence() {
	Logger.SetLevel(log.FatalLevel + 100)
}

// Debug logs debug messages
func Debug(msg string, keyvals ...interface{}) {
	Logger.Helper()
	Logger.Debug(msg, keyvals...)
}

// Info logs info messages
func Info(msg string, keyvals ...interface{}) {
	Logger.Helper()
	Logger.Info(msg, keyvals...)
}

// Warn logs warning messages
func Warn(msg string, keyvals ...interface{}) {
	Logger.Helper()
	Logger.Warn(msg, keyvals...)
}

// Error logs error messages
func Error(msg string, keyvals ...interface{}) {
	Logger.Helper()
	Logger.Error(msg, keyvals...)
}

// Fatal logs fatal messages and panics
func Fatal(msg string, keyvals ...interface{}) {
	Logger.Helper()
	Logger.Fatal(msg, keyvals...)
}

// Print logs messages regardless of level
func Print(msg string, keyvals ...interface{}) {
	Logger.Helper()
	Logger.Print(msg, keyvals...)
}

// With returns a sublogger with a prefix
func With(keyvals ...interface{}) *LLogger {
	return Logger.With(keyvals...)
}
