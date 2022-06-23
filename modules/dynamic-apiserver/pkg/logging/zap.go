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
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
)

var atom zap.AtomicLevel

func newZapLogger(opts ...zap.Option) *zap.Logger {
	// NewDevelopmentConfig -- for production logger
	cfg := zap.NewDevelopmentConfig()

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
