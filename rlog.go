package rlog

// We only provide a standard interface for logging here, then the extensions in
// one app could have a chance to use the same logging implementation.
//
// Golang gives a proposal to provide a standard interface for structured
// logging with levels, and the logging package in the next generation calls
// 'slog'. Refer to https://go.googlesource.com/proposal/+/master/design/56345-structured-logging.md.
//
// And once the 'slog' is released, we could use it directly and remove this.

const (
	ATTR_KEY_MODULE     = "module"
	DEFAULT_MODULE_NAME = "default"
)

var _default_handler LogHandler
var _default_r_log = &_r_logger{
	module: DEFAULT_MODULE_NAME,
}

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

type _r_logger struct {
	module  string
	handler LogHandler
}

type ILogger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
}

func (l *_r_logger) doLog(msg string, level LogLevel, args ...any) {
	size := len(args)/2 + 1 // 1 for the module name.
	attrs := make([]LogAttr, size)

	for i := 0; i < len(args); i += 2 {
		attrs[i/2] = LogAttr{
			Key:   args[i].(string),
			Value: args[i+1],
		}
	}

	attrs[size-1] = LogAttr{
		Key:   ATTR_KEY_MODULE,
		Value: l.module,
	}

	l.handler.Handle(LogRecord{
		Message: msg,
		Attrs:   attrs,
		Level:   level,
	})
}

func (l *_r_logger) Debug(msg string, args ...any) {
	if l.handler.Enabled(LogLevelDebug) {
		l.doLog(msg, LogLevelDebug, args...)
	}
}

func (l *_r_logger) Info(msg string, args ...any) {
	if l.handler.Enabled(LogLevelInfo) {
		l.doLog(msg, LogLevelInfo, args...)
	}
}

func (l *_r_logger) Warn(msg string, args ...any) {
	if l.handler.Enabled(LogLevelWarn) {
		l.doLog(msg, LogLevelWarn, args...)
	}
}

func (l *_r_logger) Error(msg string, args ...any) {
	if l.handler.Enabled(LogLevelError) {
		l.doLog(msg, LogLevelError, args...)
	}
}

// Set the logging implementation with this function, this function must be
// called before 'GetDefaultLogger()' or 'GetLogger()'. In general, this
// function shall be called in the 'main()' function, before starting the rte
// app.
func SetRLogDefaultHandler(h LogHandler) {
	_default_handler = h
	_default_r_log.handler = h
}

func GetDefaultLogger() ILogger {
	if _default_handler == nil {
		panic("Please set the default handler first.")
	}

	return _default_r_log
}

// Extensions shall use this function to retrieve a logger.
//
// The source code of the extensions are compiled into one executable, so we can
// not distinguish which extension is the source of the logs based on the call
// stacks. So the extension could retrieve its own logger by passing its name
// here, and we will add a 'module' attribute to the log record.
//
// The '_default_handler' is preferred to be set before calling this function,
// so we do not use a mutex lock here.
func GetLogger(module string) ILogger {
	if _default_handler == nil {
		panic("Please set the default handler first.")
	}

	return &_r_logger{
		module:  module,
		handler: _default_handler,
	}
}
