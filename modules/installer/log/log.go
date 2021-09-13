package log

import "github.com/sirupsen/logrus"

type Logger interface {
	Infof(string, ...interface{})
	Debugf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})
}

func NewLogger() Logger {
	l := logrus.New()
	l.SetLevel(logrus.DebugLevel)

	return l
}
