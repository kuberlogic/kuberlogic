package logging

import (
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	zapr "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

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

func CreateLogger() (logr.Logger, error) {
	sentryOptions := zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return zapcore.RegisterHooks(core, func(entry zapcore.Entry) error {
			// sending all events to sentry above warn level
			if entry.Level >= zap.WarnLevel {
				fmt.Println("---->", entry.Stack)
				fmt.Println("======>", entry.Level, entry.Message)
				sentry.CaptureEvent(entryToEvent(entry))
			}
			return nil
		})
	})

	opts := []zapr.Opts{
		zapr.UseDevMode(true),
		zapr.RawZapOpts(sentryOptions),
	}

	if logName := os.Getenv("KUBERLOGIC_OPERATOR_LOG"); logName != "" {
		file, err := os.Create(logName)
		if err != nil {
			return nil, err
		}
		opts = append(opts, zapr.WriteTo(file))
	}

	return zapr.New(opts...), nil
}
