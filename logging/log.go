package logging

import (
	"github.com/go-logr/logr"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func CreateLogger() (logr.Logger, error) {
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

	return zap.New(opts...), nil
}
