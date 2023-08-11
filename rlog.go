package rlog

import "sync"

// We only provide a standard interface for logging here, then the extensions in
// one app could have a chance to use the same logging implementation.
//
// Golang gives a proposal to provide a standard interface for structured
// logging with levels, and the logging package in the next generation calls
// 'slog'. Refer to https://go.googlesource.com/proposal/+/master/design/56345-structured-logging.md.
//
// And once the 'slog' is released, we could use it directly and remove this.

const (
	KEY_DEFAULT_LOGGER = "default"
)

var loggers = sync.Map{} // map[string]ILogger

type LogLevel int8

const (
	LogLevelDebug LogLevel = iota - 1
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

type LogAttr struct {
	Key   string
	Value any
}

type LogRecord struct {
	Message string
	Attrs   []LogAttr
	Level   LogLevel
}

type LogHandler interface {
	Enabled(l LogLevel) bool
	Handle(r LogRecord)
}

type r_logger struct {
	handler LogHandler
}

type ILogger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func (l *r_logger) doLog(msg string, level LogLevel, args ...any) {
	size := len(args)/2 + 1 // 1 for the module name.
	attrs := make([]LogAttr, size)

	for i := 0; i < len(args); i += 2 {
		attrs[i/2] = LogAttr{
			Key:   args[i].(string),
			Value: args[i+1],
		}
	}

	l.handler.Handle(LogRecord{
		Message: msg,
		Attrs:   attrs,
		Level:   level,
	})
}

func (l *r_logger) Debug(msg string, args ...any) {
	if l.handler.Enabled(LogLevelDebug) {
		l.doLog(msg, LogLevelDebug, args...)
	}
}

func (l *r_logger) Info(msg string, args ...any) {
	if l.handler.Enabled(LogLevelInfo) {
		l.doLog(msg, LogLevelInfo, args...)
	}
}

func (l *r_logger) Warn(msg string, args ...any) {
	if l.handler.Enabled(LogLevelWarn) {
		l.doLog(msg, LogLevelWarn, args...)
	}
}

func (l *r_logger) Error(msg string, args ...any) {
	if l.handler.Enabled(LogLevelError) {
		l.doLog(msg, LogLevelError, args...)
	}
}

// Register the handler if not absent.
//
// Set the logging implementation with this function, this function must be
// called before 'GetDefaultLogger()' or 'GetLogger()'. In general, this
// function shall be called in the 'main()' function, before starting the rte
// app.
func RegisterLogHandler(name string, h LogHandler) (ok bool) {
	_, loaded := loggers.LoadOrStore(name, h)

	if loaded {
		return false
	}

	return true
}

func GetDefaultLogger() ILogger {
	return GetLogger(KEY_DEFAULT_LOGGER)
}

func GetLogger(handler string) ILogger {
	logger, ok := loggers.Load(handler)

	if ok {
		h := logger.(LogHandler)
		return &r_logger{handler: h}
	} else {
		return nil
	}
}
