package middleware

import (
	"github.com/go-chi/chi/middleware"
	"github.com/kuberlogic/operator/modules/apiserver/internal/logging"
	"net/http"
	"time"
)

func NewLoggingMiddleware(log logging.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
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
}
