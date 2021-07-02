package zerolog

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

// list of log level
const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// DefaultTimeFormat of logger
const DefaultTimeFormat = time.RFC3339

var _ = (*Logger)(nil)

type (
	Level int

	// KV is a type for logging with more information
	// this used by with function
	KV map[string]interface{}

	Logger struct {
		logger zerolog.Logger
		config Config
		valid  bool
	}

	Config struct {
		Level      Level
		AppName    string
		LogFile    string
		TimeFormat string
		CallerSkip int
		Caller     bool
		UseColor   bool
		UseJSON    bool
		StdLog     bool
	}
)

// OpenLogFile tries to open the log file (creates it if not exists) in write-only/append mode and return it
// Note: the func return nil for both *os.File and error if the file name is empty string
func (c *Config) OpenLogFile() (*os.File, error) {
	if c.LogFile == "" {
		return nil, nil
	}

	err := os.MkdirAll(filepath.Dir(c.LogFile), 0755)
	if err != nil && err != os.ErrExist {
		return nil, err
	}

	return os.OpenFile(c.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
}

// DefaultLogger return default value of logger
func DefaultLogger() *Logger {
	l := Logger{
		config: Config{
			Level:      InfoLevel,
			TimeFormat: DefaultTimeFormat,
		},
		valid: true,
	}

	lgr := zerolog.New(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		NoColor:    !l.config.UseColor,
		TimeFormat: l.config.TimeFormat,
	})
	lgr = setLevel(lgr, l.config.Level)

	l.logger = lgr
	return &l
}

// New logger
func New(config *Config, opts ...func(*Config)) (*Logger, error) {
	if config == nil {
		config = &Config{
			Level:      InfoLevel,
			TimeFormat: DefaultTimeFormat,
		}
	}

	if config.TimeFormat == "" {
		config.TimeFormat = DefaultTimeFormat
	}
	for _, opt := range opts {
		opt(config)
	}

	lgr, err := newLogger(config)
	if err != nil {
		return nil, err
	}
	l := Logger{
		logger: lgr,
		config: *config,
		valid:  true,
	}
	return &l, nil
}

func newLogger(config *Config) (zerolog.Logger, error) {
	var (
		lgr zerolog.Logger
	)

	zerolog.TimeFieldFormat = config.TimeFormat
	zerolog.CallerSkipFrameCount = 4 + config.CallerSkip

	var writer io.Writer = os.Stderr
	file, err := config.OpenLogFile()
	if err != nil {
		return lgr, err
	} else if file != nil {
		writer = file
		config.UseColor = false
	}

	if !config.UseJSON {
		writer = zerolog.ConsoleWriter{
			Out:        writer,
			NoColor:    !config.UseColor,
			TimeFormat: config.TimeFormat,
		}
	}

	if config.StdLog {
		// Avoiding breaking changes for std log
		zerolog.TimestampFieldName = "time"
		zerolog.LevelFieldName = "lvl"
		zerolog.MessageFieldName = "msg"
		zerolog.CallerFieldName = "line"
		zerolog.ErrorFieldName = "err"
	}

	lgr = zerolog.New(writer).With().Str("app", config.AppName).Logger()
	lgr = setLevel(lgr, config.Level)
	if config.Caller {
		lgr = lgr.With().Caller().Logger()
	}
	return lgr, nil
}

func setLevel(lgr zerolog.Logger, level Level) zerolog.Logger {
	switch level {
	case TraceLevel:
		lgr = lgr.Level(zerolog.TraceLevel)
	case DebugLevel:
		lgr = lgr.Level(zerolog.DebugLevel)
	case InfoLevel:
		lgr = lgr.Level(zerolog.InfoLevel)
	case WarnLevel:
		lgr = lgr.Level(zerolog.WarnLevel)
	case ErrorLevel:
		lgr = lgr.Level(zerolog.ErrorLevel)
	case FatalLevel:
		lgr = lgr.Level(zerolog.FatalLevel)
	default:
		lgr = lgr.Level(zerolog.InfoLevel)
	}
	return lgr
}

// SetLevel for setting log level
func (l *Logger) SetLevel(level Level) {
	if level < DebugLevel || level > FatalLevel {
		level = InfoLevel
	}
	if level != l.config.Level {
		l.logger = setLevel(l.logger, level)
		l.config.Level = level
	}
}

// IsValid check if Logger is created using constructor
func (l *Logger) IsValid() bool {
	return l.valid
}

// Debug function
func (l *Logger) Debug(args ...interface{}) {
	l.logger.Debug().Timestamp().Msg(fmt.Sprint(args...))
}

// Debugln function
func (l *Logger) Debugln(args ...interface{}) {
	l.logger.Debug().Timestamp().Msg(fmt.Sprintln(args...))
}

// Debugf function
func (l *Logger) Debugf(format string, v ...interface{}) {
	l.logger.Debug().Timestamp().Msgf(format, v...)
}

// DebugWithFields function
func (l *Logger) DebugWithFields(msg string, KV KV) {
	l.logger.Debug().Timestamp().Fields(KV).Msg(msg)
}

// Info function
func (l *Logger) Info(args ...interface{}) {
	l.logger.Info().Timestamp().Msg(fmt.Sprint(args...))
}

// Infoln function
func (l *Logger) Infoln(args ...interface{}) {
	l.logger.Info().Timestamp().Msg(fmt.Sprintln(args...))
}

// Infof function
func (l *Logger) Infof(format string, v ...interface{}) {
	l.logger.Info().Timestamp().Msgf(format, v...)
}

// InfoWithFields function
func (l *Logger) InfoWithFields(msg string, KV KV) {
	l.logger.Info().Timestamp().Fields(KV).Msg(msg)
}

// Warn function
func (l *Logger) Warn(args ...interface{}) {
	l.logger.Warn().Timestamp().Msg(fmt.Sprint(args...))
}

// Warnln function
func (l *Logger) Warnln(args ...interface{}) {
	l.logger.Warn().Timestamp().Msg(fmt.Sprintln(args...))
}

// Warnf function
func (l *Logger) Warnf(format string, v ...interface{}) {
	l.logger.Warn().Timestamp().Msgf(format, v...)
}

// WarnWithFields function
func (l *Logger) WarnWithFields(msg string, KV KV) {
	l.logger.Warn().Timestamp().Fields(KV).Msg(msg)
}

// Error function
func (l *Logger) Error(args ...interface{}) {
	l.logger.Error().Timestamp().Msg(fmt.Sprint(args...))
}

// Errorln function
func (l *Logger) Errorln(args ...interface{}) {
	l.logger.Error().Timestamp().Msg(fmt.Sprintln(args...))
}

// Errorf function
func (l *Logger) Errorf(format string, v ...interface{}) {
	l.logger.Error().Timestamp().Msgf(format, v...)
}

// ErrorWithFields function
func (l *Logger) ErrorWithFields(msg string, KV KV) {
	l.logger.Error().Timestamp().Fields(KV).Msg(msg)
}

// Errors function to log errors package
func (l *Logger) Errors(err error) {
	l.logger.Error().Timestamp().Msg(err.Error())
}

// Fatal function
func (l *Logger) Fatal(args ...interface{}) {
	l.logger.Fatal().Timestamp().Msg(fmt.Sprint(args...))
}

// Fatalln function
func (l *Logger) Fatalln(args ...interface{}) {
	l.logger.Fatal().Timestamp().Msg(fmt.Sprintln(args...))
}

// Fatalf function
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.logger.Fatal().Timestamp().Msgf(format, v...)
}

// FatalWithFields function
func (l *Logger) FatalWithFields(msg string, KV KV) {
	l.logger.Fatal().Timestamp().Fields(KV).Msg(msg)
}

const (
	contextFieldName  = "ctx_id"
	metadataFieldName = "metadata"
	requestFieldName  = "req_id"
)

// StdTrace zerolog implementation for trace level log
func (l *Logger) StdTrace(requestID string, contextID string, err error, metadata interface{}, message string) {
	l.logger.Trace().Timestamp().Str(requestFieldName, requestID).Str(contextFieldName, contextID).Err(err).Interface(metadataFieldName, metadata).Msg(message)
}

// StdTracef zerolog implementation for trace level log
func (l *Logger) StdTracef(requestID string, contextID string, err error, metadata interface{}, format string, args ...interface{}) {
	l.logger.Trace().Timestamp().Str(requestFieldName, requestID).Str(contextFieldName, contextID).Err(err).Interface(metadataFieldName, metadata).Msgf(format, args...)
}

// StdDebug zerolog implementation for trace level log
func (l *Logger) StdDebug(requestID string, contextID string, err error, metadata interface{}, message string) {
	l.logger.Debug().Timestamp().Str(requestFieldName, requestID).Str(contextFieldName, contextID).Err(err).Interface(metadataFieldName, metadata).Msg(message)
}

// StdDebugf zerolog implementation for trace level log
func (l *Logger) StdDebugf(requestID string, contextID string, err error, metadata interface{}, format string, args ...interface{}) {
	l.logger.Debug().Timestamp().Str(requestFieldName, requestID).Str(contextFieldName, contextID).Err(err).Interface(metadataFieldName, metadata).Msgf(format, args...)
}

// StdInfo zerolog implementation for trace level log
func (l *Logger) StdInfo(requestID string, contextID string, err error, metadata interface{}, message string) {
	l.logger.Info().Timestamp().Str(requestFieldName, requestID).Str(contextFieldName, contextID).Err(err).Interface(metadataFieldName, metadata).Msg(message)
}

// StdInfof zerolog implementation for trace level log
func (l *Logger) StdInfof(requestID string, contextID string, err error, metadata interface{}, format string, args ...interface{}) {
	l.logger.Info().Timestamp().Str(requestFieldName, requestID).Str(contextFieldName, contextID).Err(err).Interface(metadataFieldName, metadata).Msgf(format, args...)
}

// StdWarn zerolog implementation for trace level log
func (l *Logger) StdWarn(requestID string, contextID string, err error, metadata interface{}, message string) {
	l.logger.Warn().Timestamp().Str(requestFieldName, requestID).Str(contextFieldName, contextID).Err(err).Interface(metadataFieldName, metadata).Msg(message)
}

// StdWarnf zerolog implementation for trace level log
func (l *Logger) StdWarnf(requestID string, contextID string, err error, metadata interface{}, format string, args ...interface{}) {
	l.logger.Warn().Timestamp().Str(requestFieldName, requestID).Str(contextFieldName, contextID).Err(err).Interface(metadataFieldName, metadata).Msgf(format, args...)
}

// StdError zerolog implementation for trace level log
func (l *Logger) StdError(requestID string, contextID string, err error, metadata interface{}, message string) {
	l.logger.Error().Timestamp().Str(requestFieldName, requestID).Str(contextFieldName, contextID).Err(err).Interface(metadataFieldName, metadata).Msg(message)
}

// StdErrorf zerolog implementation for trace level log
func (l *Logger) StdErrorf(requestID string, contextID string, err error, metadata interface{}, format string, args ...interface{}) {
	l.logger.Error().Timestamp().Str(requestFieldName, requestID).Str(contextFieldName, contextID).Err(err).Interface(metadataFieldName, metadata).Msgf(format, args...)
}

// StdFatal zerolog implementation for trace level log
func (l *Logger) StdFatal(requestID string, contextID string, err error, metadata interface{}, message string) {
	l.logger.Fatal().Timestamp().Str(requestFieldName, requestID).Str(contextFieldName, contextID).Err(err).Interface(metadataFieldName, metadata).Msg(message)
}

// StdFatalf zerolog implementation for trace level log
func (l *Logger) StdFatalf(requestID string, contextID string, err error, metadata interface{}, format string, args ...interface{}) {
	l.logger.Fatal().Timestamp().Str(requestFieldName, requestID).Str(contextFieldName, contextID).Err(err).Interface(metadataFieldName, metadata).Msgf(format, args...)
}
