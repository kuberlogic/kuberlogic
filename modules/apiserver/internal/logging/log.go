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
	hub := sentry.CurrentHub()

	event := sentry.NewEvent()
	event.Level = sentry.Level(entry.Level.String())
	event.Message = entry.Message
	event.Logger = entry.LoggerName
	event.Timestamp = entry.Time

	if hub.Client().Options().AttachStacktrace {
		event.Threads = []sentry.Thread{{
			Stacktrace: sentry.NewStacktrace(),
			Crashed:    false,
			Current:    true,
		}}
	}
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
