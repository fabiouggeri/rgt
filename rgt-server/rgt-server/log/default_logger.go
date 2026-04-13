package log

import (
	"fmt"
	"io"
	"runtime"
)

type defaultLogger struct {
	output     io.Writer // writer must be thread safe
	formatterf func(level LogLevel, format string, a ...any) string
	formatter  func(level LogLevel, a ...any) string
	level      LogLevel
}

var newLine = windows_newLine

func init() {
	switch runtime.GOOS {
	case "linux":
		newLine = linux_newLine
	case "darwin":
		newLine = macos_newLine
	}
}

func NewLogger(logLevel LogLevel, out io.Writer) Logger {
	return &defaultLogger{level: logLevel, output: out, formatterf: defaultFormatterf, formatter: defaultFormatter}
}

func (l *defaultLogger) GetLevel() LogLevel {
	return l.level
}

func (l *defaultLogger) SetLevel(level LogLevel) {
	l.level = level
}

func (l *defaultLogger) SetFormatterf(p func(level LogLevel, format string, a ...any) string) {
	l.formatterf = p
}

func (l *defaultLogger) SetFormatter(p func(level LogLevel, a ...any) string) {
	l.formatter = p
}

func (l *defaultLogger) GetOutput() io.Writer {
	return l.output
}

func (l *defaultLogger) SetOutput(out io.Writer) {
	l.output = out
}

func (l *defaultLogger) Error(args ...any) {
	if l.level >= ERROR {
		fmt.Fprint(l.output, l.formatter(ERROR, args...), newLine)
	}
}

func (l *defaultLogger) Errorf(format string, args ...any) {
	if l.level >= ERROR {
		fmt.Fprint(l.output, l.formatterf(ERROR, format, args...), newLine)
	}
}

func (l *defaultLogger) Warn(args ...any) {
	if l.level >= WARNING {
		fmt.Fprint(l.output, l.formatter(WARNING, args...), newLine)
	}
}

func (l *defaultLogger) Warnf(format string, args ...any) {
	if l.level >= WARNING {
		fmt.Fprint(l.output, l.formatterf(WARNING, format, args...), newLine)
	}
}

func (l *defaultLogger) Info(args ...any) {
	if l.level >= INFO {
		fmt.Fprint(l.output, l.formatter(INFO, args...), newLine)
	}
}

func (l *defaultLogger) Infof(format string, args ...any) {
	if l.level >= INFO {
		fmt.Fprint(l.output, l.formatterf(INFO, format, args...), newLine)
	}
}

func (l *defaultLogger) Debug(args ...any) {
	if l.level >= DEBUG {
		fmt.Fprint(l.output, l.formatter(DEBUG, args...), newLine)
	}
}

func (l *defaultLogger) Debugf(format string, args ...any) {
	if l.level >= DEBUG {
		fmt.Fprint(l.output, l.formatterf(DEBUG, format, args...), newLine)
	}
}

func (l *defaultLogger) Trace(args ...any) {
	if l.level >= TRACE {
		fmt.Fprint(l.output, l.formatter(TRACE, args...), newLine)
	}
}

func (l *defaultLogger) Tracef(format string, args ...any) {
	if l.level >= TRACE {
		fmt.Fprint(l.output, l.formatterf(TRACE, format, args...), newLine)
	}
}
