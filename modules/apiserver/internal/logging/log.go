package logging

import (
	"github.com/getsentry/sentry-go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Sync() error
}

var l *zap.SugaredLogger

func entryToEvent(entry zapcore.Entry) *sentry.Event {
	var level sentry.Level
	switch entry.Level {
	case zap.DebugLevel:
		level = sentry.LevelDebug
	case zap.InfoLevel:
		level = sentry.LevelInfo
	case zap.WarnLevel:
		level = sentry.LevelWarning
	case zap.ErrorLevel, zap.DPanicLevel, zap.PanicLevel:
		level = sentry.LevelError
	case zap.FatalLevel:
		level = sentry.LevelFatal
	}

	event := sentry.NewEvent()
	event.Level = level
	event.Message = entry.Message
	event.Logger = entry.LoggerName
	event.Timestamp = entry.Time
	return event
}

func init() {
	sentryOptions := zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.RegisterHooks(core, func(entry zapcore.Entry) error {
			// sending all events to sentry above warn level
			if entry.Level >= zap.WarnLevel {
				sentry.CaptureEvent(entryToEvent(entry))
			}
			return nil
		})
	})

	l = newZapLogger(sentryOptions)
}

func WithComponentLogger(component string) Logger {
	return l.With(
		"component",
		component,
	)
}

func With(args ...interface{}) Logger {
	return l.With(args...)
}

func DebugLevel(enable bool) {
	l.Infof("debug logs are enabled %v", enable)
	if enable {
		zapLoggerDebug()
	}
}
