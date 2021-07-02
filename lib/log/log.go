package log

import (
	"errors"
	"os"

	"github.com/kecci/go-toolkit/lib/log/internal/zerolog"
)

const (
	EnvName        = "APPENV"
	DevelopmentEnv = "development"
	StagingEnv     = "staging"
	ProductionEnv  = "production"
)

type (
	// Logger interface
	Log = zerolog.Logger
	// Level is leveling log error
	Level int
	// Logger is interface all function need to have  for lib logger
	Logger interface {
		SetLevel(level Level)
		Debug(args ...interface{})
		Debugln(args ...interface{})
		Debugf(format string, args ...interface{})
		DebugWithFields(msg string, KV zerolog.KV)
		Info(args ...interface{})
		Infoln(args ...interface{})
		Infof(format string, args ...interface{})
		InfoWithFields(msg string, KV zerolog.KV)
		Warn(args ...interface{})
		Warnln(args ...interface{})
		Warnf(format string, args ...interface{})
		WarnWithFields(msg string, KV zerolog.KV)
		Error(args ...interface{})
		Errorln(args ...interface{})
		Errorf(format string, args ...interface{})
		ErrorWithFields(msg string, KV zerolog.KV)
		Errors(err error)
		Fatal(args ...interface{})
		Fatalln(args ...interface{})
		Fatalf(format string, args ...interface{})
		FatalWithFields(msg string, KV zerolog.KV)
		IsValid() bool // IsValid check if Logger is created using constructor

		StdTrace(requestID string, contextID string, err error, metadata interface{}, message string)
		StdTracef(requestID string, contextID string, err error, metadata interface{}, format string, args ...interface{})
		StdDebug(requestID string, contextID string, err error, metadata interface{}, message string)
		StdDebugf(requestID string, contextID string, err error, metadata interface{}, format string, args ...interface{})
		StdInfo(requestID string, contextID string, err error, metadata interface{}, message string)
		StdInfof(requestID string, contextID string, err error, metadata interface{}, format string, args ...interface{})
		StdWarn(requestID string, contextID string, err error, metadata interface{}, message string)
		StdWarnf(requestID string, contextID string, err error, metadata interface{}, format string, args ...interface{})
		StdError(requestID string, contextID string, err error, metadata interface{}, message string)
		StdErrorf(requestID string, contextID string, err error, metadata interface{}, format string, args ...interface{})
		StdFatal(requestID string, contextID string, err error, metadata interface{}, message string)
		StdFatalf(requestID string, contextID string, err error, metadata interface{}, format string, args ...interface{})
	}
)

var (
	isDev         = isDevelopment()
	infoLogger, _ = zerolog.New(&zerolog.Config{Level: zerolog.InfoLevel, UseColor: isDev, UseJSON: true, Caller: true})
	traceLogger   = infoLogger
	debugLogger   = infoLogger
	warnLogger    = infoLogger
	errorLogger   = infoLogger
	fatalLogger   = infoLogger
	loggers       = [6]*Log{
		traceLogger,
		debugLogger,
		infoLogger,
		warnLogger,
		errorLogger,
		fatalLogger,
	}

	errInvalidLogger = errors.New("invalid logger")
	errInvalidLevel  = errors.New("invalid log level")
)

func isDevelopment() bool {
	e := os.Getenv(EnvName)
	if e == "" {
		e = DevelopmentEnv
	}

	if e == DevelopmentEnv {
		return true
	}
	return false
}

// Debug prints debug level log like log.Print
func Debug(args ...interface{}) {
	debugLogger.Debug(args...)
}

// Debugln prints debug level log like log.Println
func Debugln(args ...interface{}) {
	debugLogger.Debugln(args...)
}

// Debugf prints debug level log like log.Printf
func Debugf(format string, v ...interface{}) {
	debugLogger.Debugf(format, v...)
}

// DebugWithFields prints debug level log with additional fields.
// useful when output is in json format
func DebugWithFields(msg string, fields zerolog.KV) {
	debugLogger.DebugWithFields(msg, fields)
}

// Print info level log like log.Print
func Print(v ...interface{}) {
	infoLogger.Info(v...)
}

// Println info level log like log.Println
func Println(v ...interface{}) {
	infoLogger.Infoln(v...)
}

// Printf info level log like log.Printf
func Printf(format string, v ...interface{}) {
	infoLogger.Infof(format, v...)
}

// Info prints info level log like log.Print
func Info(args ...interface{}) {
	infoLogger.Info(args...)
}

// Infoln prints info level log like log.Println
func Infoln(args ...interface{}) {
	infoLogger.Infoln(args...)
}

// Infof prints info level log like log.Printf
func Infof(format string, v ...interface{}) {
	infoLogger.Infof(format, v...)
}

// InfoWithFields prints info level log with additional fields.
// useful when output is in json format
func InfoWithFields(msg string, fields zerolog.KV) {
	infoLogger.InfoWithFields(msg, fields)
}

// Warn prints warn level log like log.Print
func Warn(args ...interface{}) {
	warnLogger.Warn(args...)
}

// Warnln prints warn level log like log.Println
func Warnln(args ...interface{}) {
	warnLogger.Warnln(args...)
}

// Warnf prints warn level log like log.Printf
func Warnf(format string, v ...interface{}) {
	warnLogger.Warnf(format, v...)
}

// WarnWithFields prints warn level log with additional fields.
// useful when output is in json format
func WarnWithFields(msg string, fields zerolog.KV) {
	warnLogger.WarnWithFields(msg, fields)
}

// Error prints error level log like log.Print
func Error(args ...interface{}) {
	errorLogger.Error(args...)
}

// Errorln prints error level log like log.Println
func Errorln(args ...interface{}) {
	errorLogger.Errorln(args...)
}

// Errorf prints error level log like log.Printf
func Errorf(format string, v ...interface{}) {
	errorLogger.Errorf(format, v...)
}

// ErrorWithFields prints error level log with additional fields.
// useful when output is in json format
func ErrorWithFields(msg string, fields zerolog.KV) {
	errorLogger.ErrorWithFields(msg, fields)
}

// Errors can handle error from tdk/x/go/errors package
func Errors(err error) {
	errorLogger.Errors(err)
}

// Fatal prints fatal level log like log.Print
func Fatal(args ...interface{}) {
	fatalLogger.Fatal(args...)
}

// Fatalln prints fatal level log like log.Println
func Fatalln(args ...interface{}) {
	fatalLogger.Fatalln(args...)
}

// Fatalf prints fatal level log like log.Printf
func Fatalf(format string, v ...interface{}) {
	fatalLogger.Fatalf(format, v...)
}

// FatalWithFields prints fatal level log with additional fields.
// useful when output is in json format
func FatalWithFields(msg string, fields zerolog.KV) {
	fatalLogger.FatalWithFields(msg, fields)
}
