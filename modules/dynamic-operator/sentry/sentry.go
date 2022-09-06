/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 */

package sentry

import (
	"github.com/getsentry/sentry-go"
	"github.com/kuberlogic/zapsentry"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type SentryTags struct {
	Component    string
	Version      string
	DeploymentId string
}

func UseSentryWithLogger(dsn string, log *zap.Logger, tags *SentryTags) *zap.Logger {
	cfg := zapsentry.Configuration{
		Level: zapcore.WarnLevel, //when to send message to sentry
		Tags: map[string]string{
			"component":     tags.Component,
			"version":       tags.Version,
			"deployment_id": tags.DeploymentId,
		},
	}
	core, err := zapsentry.NewCore(cfg, zapsentry.NewSentryClientFromDSN(dsn))
	//in case of err it will return noop core. so we can safely attach it
	if err != nil {
		log.Warn("failed to init zap", zap.Error(err))
	}
	return zapsentry.AttachCoreToLogger(core, log)
}

func InitSentry(dsn string, tags *SentryTags) error {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              dsn,
		AttachStacktrace: true,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1.0,
	}); err != nil {
		return err
	}
	sentry.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetTag("component", tags.Component)
		scope.SetTag("version", tags.Version)
		scope.SetTag("deployment_id", tags.DeploymentId)
	})
	return nil
}
