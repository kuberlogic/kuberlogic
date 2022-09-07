/*
 * CloudLinux Software Inc 2019-2021 All Rights Reserved
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *    http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package logging

import (
	"github.com/kuberlogic/kuberlogic/modules/dynamic-operator/sentry"
	"go.uber.org/zap"
)

type Logger interface {
	// using only key/value methods for the correct scribing records for the sentry

	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
	Infof(template string, args ...interface{})
	Sync() error
}

var l *zap.SugaredLogger

func init() {
	l = newZapLogger().Sugar()
}

func UseSentry(dsn, component, version, deploymentId string) {
	sentryTags := &sentry.SentryTags{
		Component:    component,
		Version:      version,
		DeploymentId: deploymentId,
	}
	l = sentry.UseSentryWithLogger(dsn, newZapLogger(), sentryTags).Sugar()
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
