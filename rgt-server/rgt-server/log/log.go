package log

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type LogLevel uint8

type Logger interface {
	GetLevel() LogLevel
	SetLevel(level LogLevel)
	SetFormatterf(p func(level LogLevel, format string, a ...any) string)
	SetFormatter(p func(level LogLevel, a ...any) string)
	GetOutput() io.Writer
	SetOutput(out io.Writer)
	Error(args ...any)
	Errorf(format string, args ...any)
	Warn(args ...any)
	Warnf(format string, args ...any)
	Info(args ...any)
	Infof(format string, args ...any)
	Debug(args ...any)
	Debugf(format string, args ...any)
	Trace(args ...any)
	Tracef(format string, args ...any)
}

const (
	OFF     LogLevel = 0
	ERROR   LogLevel = 1
	WARNING LogLevel = 2
	INFO    LogLevel = 3
	DEBUG   LogLevel = 4
	TRACE   LogLevel = 5
)

const (
	windows_newLine = "\r\n"
	linux_newLine   = "\n"
	macos_newLine   = "\r"
)

var (
	loggerDefault Logger = &defaultLogger{
		level:      INFO,
		output:     os.Stdout,
		formatterf: defaultFormatterf,
		formatter:  defaultFormatter,
	}
)

func (l LogLevel) Name() string {
	return logLevelName(l)
}

func (l LogLevel) String() string {
	return logLevelName(l)
}

func (l LogLevel) GoString() string {
	return logLevelName(l)
}

func SetDefaultLogger(logger Logger) {
	loggerDefault = logger
}

func GetDefaultLogger() Logger {
	return loggerDefault
}

func defaultFormatterf(level LogLevel, format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

func defaultFormatter(level LogLevel, a ...any) string {
	return fmt.Sprint(a...)
}

func SetLevel(level LogLevel) {
	if level >= OFF && level <= TRACE {
		loggerDefault.SetLevel(level)
	}
}

func SetOutput(out io.Writer) {
	if out != nil {
		loggerDefault.SetOutput(out)
	}
}

func SetFormatterf(p func(level LogLevel, format string, a ...any) string) {
	if p != nil {
		loggerDefault.SetFormatterf(p)
	}
}

func SetFormatter(p func(level LogLevel, a ...any) string) {
	if p != nil {
		loggerDefault.SetFormatter(p)
	}
}

func GetLevel() LogLevel {
	return loggerDefault.GetLevel()
}

func Error(args ...any) {
	loggerDefault.Error(args...)
}

func Errorf(format string, args ...any) {
	loggerDefault.Errorf(format, args...)
}

func Warn(args ...any) {
	loggerDefault.Warn(args...)
}

func Warnf(format string, args ...any) {
	loggerDefault.Warnf(format, args...)
}

func Info(args ...any) {
	loggerDefault.Info(args...)
}

func Infof(format string, args ...any) {
	loggerDefault.Infof(format, args...)
}

func Debug(args ...any) {
	loggerDefault.Debug(args...)
}

func Debugf(format string, args ...any) {
	loggerDefault.Debugf(format, args...)
}

func Trace(args ...any) {
	loggerDefault.Trace(args...)
}

func Tracef(format string, args ...any) {
	loggerDefault.Tracef(format, args...)
}

func ErrorLevel() bool {
	return IsLevel(ERROR)
}

func WarningLevel() bool {
	return IsLevel(WARNING)
}

func InfoLevel() bool {
	return IsLevel(INFO)
}

func DebugLevel() bool {
	return IsLevel(DEBUG)
}

func TraceLevel() bool {
	return IsLevel(TRACE)
}

func IsLevel(level LogLevel) bool {
	return loggerDefault.GetLevel() >= level
}

func LogLevelFromName(name string) LogLevel {
	upperValue := strings.ToUpper(strings.TrimSpace(name))
	switch upperValue {
	case "ERROR":
		return ERROR
	case "WARN", "WARNING":
		return WARNING
	case "INFO":
		return INFO
	case "DEBUG":
		return DEBUG
	case "TRACE":
		return TRACE
	default:
		return OFF
	}
}

func logLevelName(level LogLevel) string {
	switch level {
	case ERROR:
		return "ERROR"
	case WARNING:
		return "WARNING"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	case TRACE:
		return "TRACE"
	default:
		return "OFF"
	}
}

func TimestampLogPrintf(level LogLevel, format string, a ...any) string {
	return fmt.Sprintf("%s - %-5s: %s", time.Now().Format(time.RFC3339), logLevelName(level), fmt.Sprintf(format, a...))
}

func TimestampLogPrintln(level LogLevel, a ...any) string {
	return fmt.Sprintf("%s - %-5s: %s", time.Now().Format(time.RFC3339), logLevelName(level), fmt.Sprint(a...))
}
