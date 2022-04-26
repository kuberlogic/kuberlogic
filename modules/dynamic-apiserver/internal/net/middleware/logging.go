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

package middleware

import (
	"github.com/go-chi/chi/middleware"
	"github.com/kuberlogic/kuberlogic/modules/dynamic-apiserver/internal/logging"
	"net/http"
	"time"
)

func NewLoggingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
		t1 := time.Now()
		defer func() {
			l := logging.With(
				"path", r.URL.Path,
				"proto", r.Proto,
				"took", time.Since(t1),
				"status", ww.Status())
			l.Infow("request processed")
		}()

		next.ServeHTTP(ww, r)
	}
	return http.HandlerFunc(fn)
}
