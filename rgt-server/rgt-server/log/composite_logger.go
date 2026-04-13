package log

import "io"

type compositeLogger struct {
	loggers []Logger
}

type multiWriter struct {
	writers []io.Writer
}

func (w *multiWriter) Write(p []byte) (n int, err error) {
	for _, writer := range w.writers {
		n, err = writer.Write(p)
		if err != nil {
			return
		}
	}
	return
}

func NewCompositeLogger(loggers ...Logger) Logger {
	return &compositeLogger{
		loggers: loggers,
	}
}

func (l *compositeLogger) SetLevel(level LogLevel) {
	for _, logger := range l.loggers {
		logger.SetLevel(level)
	}
}

func (l *compositeLogger) GetOutput() io.Writer {
	multiWriter := &multiWriter{
		writers: make([]io.Writer, 0, len(l.loggers)),
	}
	for _, logger := range l.loggers {
		multiWriter.writers = append(multiWriter.writers, logger.GetOutput())
	}
	return multiWriter
}

func (l *compositeLogger) SetOutput(out io.Writer) {
	for _, logger := range l.loggers {
		logger.SetOutput(out)
	}
}

func (l *compositeLogger) SetFormatterf(formatterf func(level LogLevel, format string, a ...any) string) {
	for _, logger := range l.loggers {
		logger.SetFormatterf(formatterf)
	}
}

func (l *compositeLogger) SetFormatter(formatter func(level LogLevel, a ...any) string) {
	for _, logger := range l.loggers {
		logger.SetFormatter(formatter)
	}
}

func (l *compositeLogger) GetLevel() LogLevel {
	return l.loggers[0].GetLevel()
}

func (l *compositeLogger) Error(args ...any) {
	for _, logger := range l.loggers {
		logger.Error(args...)
	}
}

func (l *compositeLogger) Errorf(format string, args ...any) {
	for _, logger := range l.loggers {
		logger.Errorf(format, args...)
	}
}

func (l *compositeLogger) Warn(args ...any) {
	for _, logger := range l.loggers {
		logger.Warn(args...)
	}
}

func (l *compositeLogger) Warnf(format string, args ...any) {
	for _, logger := range l.loggers {
		logger.Warnf(format, args...)
	}
}

func (l *compositeLogger) Info(args ...any) {
	for _, logger := range l.loggers {
		logger.Info(args...)
	}
}

func (l *compositeLogger) Infof(format string, args ...any) {
	for _, logger := range l.loggers {
		logger.Infof(format, args...)
	}
}

func (l *compositeLogger) Debug(args ...any) {
	for _, logger := range l.loggers {
		logger.Debug(args...)
	}
}

func (l *compositeLogger) Debugf(format string, args ...any) {
	for _, logger := range l.loggers {
		logger.Debugf(format, args...)
	}
}

func (l *compositeLogger) Trace(args ...any) {
	for _, logger := range l.loggers {
		logger.Trace(args...)
	}
}

func (l *compositeLogger) Tracef(format string, args ...any) {
	for _, logger := range l.loggers {
		logger.Tracef(format, args...)
	}
}

func (l *compositeLogger) ErrorLevel() bool {
	for _, logger := range l.loggers {
		if logger.GetLevel() >= ERROR {
			return true
		}
	}
	return false
}

func (l *compositeLogger) WarnLevel() bool {
	for _, logger := range l.loggers {
		if logger.GetLevel() >= WARNING {
			return true
		}
	}
	return false
}

func (l *compositeLogger) InfoLevel() bool {
	for _, logger := range l.loggers {
		if logger.GetLevel() >= INFO {
			return true
		}
	}
	return false
}

func (l *compositeLogger) DebugLevel() bool {
	for _, logger := range l.loggers {
		if logger.GetLevel() >= DEBUG {
			return true
		}
	}
	return false
}

func (l *compositeLogger) TraceLevel() bool {
	for _, logger := range l.loggers {
		if logger.GetLevel() >= TRACE {
			return true
		}
	}
	return false
}
