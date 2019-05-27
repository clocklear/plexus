package http

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/pborman/uuid"
)

// Error represents a WTF error.
type Error string

type contextKey int

// Error returns the error message.
func (e Error) Error() string {
	return string(e)
}

const (
	// ErrInternal represents some sort of internal error
	errInternal = Error("Internal")
	// ErrBadData represents invalid input
	errBadData = Error("invalid input data")

	keyLogger contextKey = iota
)

type responseSpy struct {
	http.ResponseWriter
	code int
}

func (s *responseSpy) WriteHeader(code int) {
	s.code = code
	s.ResponseWriter.WriteHeader(code)
}

// Middleware aliases the functions that take in handlers
// and return handlers to form a middleware stack
type Middleware func(http.Handler) http.Handler

func loggerMiddleware(baseLogger log.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		h := func(w http.ResponseWriter, rq *http.Request) {

			logger := log.With(baseLogger, "request_id", uuid.NewRandom().String())

			rq = rq.WithContext(
				context.WithValue(rq.Context(), keyLogger, logger),
			)

			begin := time.Now()
			spy := responseSpy{
				ResponseWriter: w,
			}

			w = &spy

			next.ServeHTTP(w, rq)

			// Add headers to the log output
			// var hs []string
			// for k, v := range rq.Header {

			// 	params := []interface{}{k}

			// 	// if there's just one value for this header,
			// 	// just add it to the params rather than the array itself...
			// 	if len(v) == 1 {
			// 		params = append(params, v[0])
			// 	}

			// 	hs = append(hs, fmt.Sprintf("%s:%s", params...))
			// }

			logger.Log(
				"uri", rq.RequestURI,
				"ip", rq.RemoteAddr,
				// "headers", strings.Join(hs, ","),
				"response_code", spy.code,
				"latency", time.Now().Sub(begin).String(),
			)
		}

		return http.HandlerFunc(h)
	}
}

// writeJSON encodes v to w in JSON format. Error() is called if encoding fails.
func writeJSON(code int, w http.ResponseWriter, v interface{}, logger log.Logger) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Log("err", err.Error())
	}
}

// Ok writes an success API message to the response
func Ok(w http.ResponseWriter, v interface{}, logger log.Logger) {
	writeJSON(http.StatusOK, w, v, logger)
}

// Failure writes an API error message to the response and logger.
func Failure(w http.ResponseWriter, err error, code int, logger log.Logger) {
	if logger != nil {
		logger.Log("err", fmt.Sprintf("http error: %s (code=%d)", err, code))
	}

	if code == http.StatusInternalServerError {
		err = errInternal
	}

	writeJSON(code, w, &failureResponse{Error: err.Error()}, logger)
}

type failureResponse struct {
	Error string `json:"error,omitempty"`
}

type messageResponse struct {
	Message string `json:"msg"`
}
