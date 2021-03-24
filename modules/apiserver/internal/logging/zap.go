package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var atom zap.AtomicLevel

func newZapLogger(opts ...zap.Option) *zap.Logger {
	cfg := zap.NewProductionConfig()

	if out := os.Getenv("KUBERLOGIC_APISERVER_LOG"); out != "" {
		cfg.OutputPaths = []string{
			out,
		}
	}

	atom = zap.NewAtomicLevel()
	cfg.Level = atom

	logger, _ := cfg.Build(opts...)
	return logger
}

func zapLoggerDebug() {
	atom.SetLevel(zapcore.DebugLevel)
}
