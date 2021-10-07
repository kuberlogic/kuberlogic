package middleware

import (
	"github.com/go-chi/cors"
	"github.com/kuberlogic/kuberlogic/modules/apiserver/internal/config"
	"net/http"
)

func NewCorsMiddleware(cfg *config.Config) func(handler http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   cfg.Cors.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	})
}
