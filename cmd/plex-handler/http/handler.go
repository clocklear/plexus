package http

import (
	"net/http"

	"github.com/go-kit/kit/log"
	"goji.io"
	"goji.io/pat"
)

// DefaultRequestHandler creates an instance of the default HTTP request handler
func DefaultRequestHandler(logger log.Logger) *goji.Mux {
	mux := goji.NewMux()

	mux.HandleFunc(pat.Get("/health"), handleHealthCheck())
	mux.HandleFunc(pat.Post("/hook"), handlePlexWebhook())

	mux.Use(loggerMiddleware(logger))
	return mux
}

func handleHealthCheck() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(keyLogger).(log.Logger)
		Ok(w, messageResponse{Message: "Ok"}, logger)
	}
}

func handlePlexWebhook() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// logger := r.Context().Value(keyLogger).(log.Logger)

	}
}
