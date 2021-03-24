package middleware

import (
	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
	"net/http"
)

func SentryLogPanic(next http.Handler) http.Handler {
	return sentryhttp.New(sentryhttp.Options{
		Repanic:         true,
		WaitForDelivery: true,
	}).Handle(next)
}

func SetSentryRequestScope(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sentry.ConfigureScope(func(scope *sentry.Scope) {
			scope.SetRequest(r)
		})

		next.ServeHTTP(w, r)
	})
}
