package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"
)

type LogLevel uint8

type Logger struct {
	output     io.Writer // writer must be thread safe
	formatterf func(level LogLevel, format string, a ...any) string
	formatter  func(level LogLevel, a ...any) string
	level      LogLevel
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
	DEFAULT_LOG_ID  = "default"
)

var (
	loggers map[string]*Logger = make(map[string]*Logger)
	newLine                    = windows_newLine
)

func init() {
	switch runtime.GOOS {
	case "linux":
		newLine = linux_newLine
	case "darwin":
		newLine = macos_newLine
	}
	loggers[DEFAULT_LOG_ID] = &Logger{level: INFO, output: os.Stdout, formatterf: defaultFormatterf, formatter: defaultFormatter}
}

func (l LogLevel) Name() string {
	return logLevelName(l)
}

func (l LogLevel) String() string {
	return logLevelName(l)
}

func (l LogLevel) GoString() string {
	return logLevelName(l)
}

func defaultFormatterf(level LogLevel, format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

func defaultFormatter(level LogLevel, a ...any) string {
	return fmt.Sprint(a...)
}

func RemoveAllLoggers() {
	loggers = make(map[string]*Logger)
}

func SetLevel(id string, level LogLevel) {
	if level >= OFF && level <= TRACE {
		l, found := loggers[id]
		if found {
			l.SetLevel(level)
		}
	}
}

func SetOutput(id string, out io.Writer) {
	if out != nil {
		l, found := loggers[id]
		if found {
			l.SetOutput(out)
		}
	}
}

func SetFormatterf(id string, p func(level LogLevel, format string, a ...any) string) {
	if p != nil {
		l, found := loggers[id]
		if found {
			l.SetFormatterf(p)
		}
	}
}

func SetFormatter(id string, p func(level LogLevel, a ...any) string) {
	if p != nil {
		l, found := loggers[id]
		if found {
			l.SetFormatter(p)
		}
	}
}

func AddLogger(id string, logLevel LogLevel, out io.Writer) {
	loggers[id] = &Logger{level: logLevel, output: out, formatterf: defaultFormatterf, formatter: defaultFormatter}
}

func RemoveLogger(id string) {
	delete(loggers, id)
}

func GetLevel(id string) LogLevel {
	l, found := loggers[id]
	if found {
		return l.level
	}
	return OFF
}

func Error(args ...any) {
	for _, v := range loggers {
		v.Error(args...)
	}
}

func Errorf(format string, args ...any) {
	for _, v := range loggers {
		v.Errorf(format, args...)
	}
}

func Warn(args ...any) {
	for _, v := range loggers {
		v.Warn(args...)
	}
}

func Warnf(format string, args ...any) {
	for _, v := range loggers {
		v.Warnf(format, args...)
	}
}

func Info(args ...any) {
	for _, v := range loggers {
		v.Info(args...)
	}
}

func Infof(format string, args ...any) {
	for _, v := range loggers {
		v.Infof(format, args...)
	}
}

func Debug(args ...any) {
	for _, v := range loggers {
		v.Debug(args...)
	}
}

func Debugf(format string, args ...any) {
	for _, v := range loggers {
		v.Debugf(format, args...)
	}
}

func Trace(args ...any) {
	for _, v := range loggers {
		v.Trace(args...)
	}
}

func Tracef(format string, args ...any) {
	for _, v := range loggers {
		v.Tracef(format, args...)
	}
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
	for _, v := range loggers {
		if v.level >= level {
			return true
		}
	}
	return false
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

func (l *Logger) GetLevel() LogLevel {
	return l.level
}

func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *Logger) SetFormatterf(p func(level LogLevel, format string, a ...any) string) {
	l.formatterf = p
}

func (l *Logger) SetFormatter(p func(level LogLevel, a ...any) string) {
	l.formatter = p
}

func (l *Logger) GetOutput() io.Writer {
	return l.output
}

func (l *Logger) SetOutput(out io.Writer) {
	l.output = out
}

func (l *Logger) Error(args ...any) {
	if l.level >= ERROR {
		fmt.Fprint(l.output, l.formatter(ERROR, args...), newLine)
	}
}

func (l *Logger) Errorf(format string, args ...any) {
	if l.level >= ERROR {
		fmt.Fprint(l.output, l.formatterf(ERROR, format, args...), newLine)
	}
}

func (l *Logger) Warn(args ...any) {
	if l.level >= WARNING {
		fmt.Fprint(l.output, l.formatter(WARNING, args...), newLine)
	}
}

func (l *Logger) Warnf(format string, args ...any) {
	if l.level >= WARNING {
		fmt.Fprint(l.output, l.formatterf(WARNING, format, args...), newLine)
	}
}

func (l *Logger) Info(args ...any) {
	if l.level >= INFO {
		fmt.Fprint(l.output, l.formatter(INFO, args...), newLine)
	}
}

func (l *Logger) Infof(format string, args ...any) {
	if l.level >= INFO {
		fmt.Fprint(l.output, l.formatterf(INFO, format, args...), newLine)
	}
}

func (l *Logger) Debug(args ...any) {
	if l.level >= DEBUG {
		fmt.Fprint(l.output, l.formatter(DEBUG, args...), newLine)
	}
}

func (l *Logger) Debugf(format string, args ...any) {
	if l.level >= DEBUG {
		fmt.Fprint(l.output, l.formatterf(DEBUG, format, args...), newLine)
	}
}

func (l *Logger) Trace(args ...any) {
	if l.level >= TRACE {
		fmt.Fprint(l.output, l.formatter(TRACE, args...), newLine)
	}
}

func (l *Logger) Tracef(format string, args ...any) {
	if l.level >= TRACE {
		fmt.Fprint(l.output, l.formatterf(TRACE, format, args...), newLine)
	}
}
