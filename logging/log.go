package logging

import (
	"github.com/go-logr/logr"
	"io/ioutil"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type fileWriter struct {
	out string
}

func (fw fileWriter) Write(p []byte) (n int, err error) {
	err = ioutil.WriteFile(fw.out, p, os.ModePerm)
	return 0, err
}

func CreateLogger() logr.Logger {
	opts := []zap.Opts{
		zap.UseDevMode(true),
	}

	if out := os.Getenv("KUBERLOGIC_OPERATOR_LOG"); out != "" {
		writer := fileWriter{out}
		opts = append(opts, zap.WriteTo(writer))
	}
	return zap.New(opts...)
}
