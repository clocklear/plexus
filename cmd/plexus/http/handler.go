package http

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"goji.io"
	"goji.io/pat"

	"github.com/clocklear/plexus/pkg/plex"
	"github.com/clocklear/plexus/pkg/plex/schema"
)

// DefaultRequestHandler creates an instance of the default HTTP request handler
func DefaultRequestHandler(logger log.Logger, store *plex.Store, cfg plex.Config) (*goji.Mux, error) {

	v, err := schema.NewValidator()
	if err != nil {
		return nil, err
	}

	mux := goji.NewMux()

	mux.HandleFunc(pat.Get("/health"), handleHealthCheck())
	mux.HandleFunc(pat.Post("/hook"), handlePlexWebhook(v, store, cfg))
	mux.HandleFunc(pat.Get("/activity"), handleGetAllHooks(store))

	mux.Use(loggerMiddleware(logger))
	return mux, nil
}

func handleHealthCheck() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(keyLogger).(log.Logger)
		Ok(w, messageResponse{Message: "Ok"}, logger)
	}
}

func handlePlexWebhook(v *schema.Validator, store *plex.Store, cfg plex.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(keyLogger).(log.Logger)
		reqID := r.Context().Value(keyRequestID).(string)

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
		err = store.AddActivity(plex.Activity{
			RequestID:  reqID,
			ReceivedAt: time.Now(),
			Payload:    pl,
		})
		if err != nil {
			Failure(w, err, http.StatusInternalServerError, logger)
			return
		}

		// Pass payload to configuration handler
		err = cfg.Handle(logger, pl, payload)
		if err != nil {
			Failure(w, err, http.StatusInternalServerError, logger)
		}

		Ok(w, messageResponse{Message: "Ok"}, logger)
	}
}

func handleGetAllHooks(store *plex.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(keyLogger).(log.Logger)
		act, err := store.GetAllActivity()
		if err != nil {
			Failure(w, err, http.StatusInternalServerError, logger)
			return
		}
		Ok(w, act, logger)
	}
}
