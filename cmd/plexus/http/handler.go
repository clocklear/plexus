package http

import (
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"strings"
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

func hasContentType(r *http.Request, mimetype string) bool {
	contentType := r.Header.Get("Content-type")
	if contentType == "" {
		return mimetype == "application/octet-stream"
	}

	for _, v := range strings.Split(contentType, ",") {
		t, _, err := mime.ParseMediaType(v)
		if err != nil {
			break
		}
		if t == mimetype {
			return true
		}
	}
	return false
}

func handlePlexWebhook(v *schema.Validator, store *plex.Store, cfg plex.Config) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := r.Context().Value(keyLogger).(log.Logger)
		reqID := r.Context().Value(keyRequestID).(string)

		// https://support.plex.tv/articles/115002267687-webhooks/
		// Per their documentation, Plex will send a multipart form request, and 'payload' is the JSON of the hook
		// We want to be flexible (makes testing easier), so lets see if we can handle both scenarios (multipart vs raw JSON post)
		payload := []byte{}
		if hasContentType(r, "multipart/form-data") {
			err := r.ParseMultipartForm(10 * 1024 * 1024) // 10mb
			if err != nil {
				Failure(w, err, http.StatusBadRequest, logger)
				return
			}
			payload = []byte(r.FormValue("payload"))
		} else {
			// Assume raw JSON post
			pl, err := ioutil.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				Failure(w, err, http.StatusInternalServerError, logger)
				return
			}
			payload = pl
		}

		// Should have JSON bytes by this point
		// Validate the request
		err := v.Validate(payload)
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
