package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/go-kit/kit/log"
	"goji.io"
	"goji.io/pat"

	"github.com/clocklear/plexus/pkg/plex"
	"github.com/clocklear/plexus/pkg/plex/schema"
)

// DefaultRequestHandler creates an instance of the default HTTP request handler
func DefaultRequestHandler(logger log.Logger, store *plex.ActivityStore) (*goji.Mux, error) {

	v, err := schema.NewValidator()
	if err != nil {
		return nil, err
	}

	mux := goji.NewMux()

	mux.HandleFunc(pat.Get("/health"), handleHealthCheck())
	mux.HandleFunc(pat.Post("/hook"), handlePlexWebhook(v, store))
	mux.HandleFunc(pat.Get("/hook"), handleGetAllHooks(store))

	mux.Use(loggerMiddleware(logger))
	return mux, nil
}

func handleHealthCheck() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(keyLogger).(log.Logger)
		Ok(w, messageResponse{Message: "Ok"}, logger)
	}
}

func handlePlexWebhook(v *schema.Validator, store *plex.ActivityStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(keyLogger).(log.Logger)

		// Extract request body
		payload, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()

		// Validate the request
		err = v.Validate(payload)
		if err != nil {
			Failure(w, err, http.StatusBadRequest, logger)
			return
		}

		// Store result
		pl := plex.WebhookPayload{}
		if err = json.Unmarshal(payload, &pl); err != nil {
			Failure(w, err, http.StatusInternalServerError, logger)
			return
		}
		err = store.Add(pl)
		if err != nil {
			Failure(w, err, http.StatusInternalServerError, logger)
			return
		}

		Ok(w, messageResponse{Message: "Ok"}, logger)
	}
}

func handleGetAllHooks(store *plex.ActivityStore) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(keyLogger).(log.Logger)
		Ok(w, store.GetAll(), logger)
	}
}
