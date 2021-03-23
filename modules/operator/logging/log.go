package logging

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/kuberlogic/zapsentry"
	zap2 "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	zap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func modifyToSentryLogger(log *zap2.Logger, dsn string) *zap2.Logger {
	cfg := zapsentry.Configuration{
		Level: zapcore.WarnLevel, //when to send message to sentry
		Tags: map[string]string{
			"component": "operator",
		},
	}
	core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromDSN(dsn))
	//in case of err it will return noop core. so we can safely attach it
	if err != nil {
		log.Warn("failed to init zap", zap2.Error(err))
	}
	return zapsentry.AttachCoreToLogger(core, log)
}

func CreateZapLogger() (*zap2.Logger, error) {
	opts := []zap.Opts{
		zap.UseDevMode(true),
	}

	if logName := os.Getenv("KUBERLOGIC_OPERATOR_LOG"); logName != "" {
		file, err := os.Create(logName)
		if err != nil {
			return nil, err
		}
		opts = append(opts, zap.WriteTo(file))
	}

	return zap.NewRaw(opts...), nil
}

func GetLogger(logger *zap2.Logger) logr.Logger {
	return zapr.NewLogger(logger)
}

func UseSentry(dsn string, logger *zap2.Logger) logr.Logger {
	return zapr.NewLogger(modifyToSentryLogger(logger, dsn))
}
