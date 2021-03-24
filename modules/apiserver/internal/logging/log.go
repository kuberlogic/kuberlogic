package logging

import (
	"github.com/kuberlogic/zapsentry"
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

func init() {
	l = newZapLogger().Sugar()
}

func modifyToSentryLogger(log *zap.Logger, dsn string) *zap.Logger {
	cfg := zapsentry.Configuration{
		Level: zapcore.WarnLevel, //when to send message to sentry
		Tags: map[string]string{
			"component": "apiserver",
		},
	}
	core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromDSN(dsn))
	//in case of err it will return noop core. so we can safely attach it
	if err != nil {
		log.Warn("failed to init zap", zap.Error(err))
	}
	return zapsentry.AttachCoreToLogger(core, log)
}

func UseSentry(dsn string) {
	l = modifyToSentryLogger(newZapLogger(), dsn).Sugar()
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
