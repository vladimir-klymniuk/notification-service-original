package notify

import (
	"context"
	"encoding/json"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"net/http"
)

func NewHTTPHandler(ep Endpoints) http.Handler {
	m := mux.NewRouter()

	options := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}

	m.Handle("/notify", kithttp.NewServer(
		ep.NotifyEndpoint,
		decodeRequest,
		encodeJSONResponse,
		options...,
	)).Methods("POST")

	return m
}

// decodeRequest decodes request
func decodeRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var rr Request

	if err := validateRequestHeaders(r.Header); err != nil {
		return nil, err
	}

	_ = json.NewDecoder(r.Body).Decode(&rr)

	err := validateRequestBody(rr)
	if err != nil {
		return nil, err
	}

	return rr, err
}

// EncodeJSONResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func encodeJSONResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(http.StatusAccepted)
	return nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	switch errors.Cause(err) {
	case ErrInvalidParameter, ErrRequestBodyMissingParams:
		w.WriteHeader(http.StatusBadRequest)
	case ErrRequestHeaderMissingParams:
		w.WriteHeader(http.StatusForbidden)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func validateRequestHeaders(h http.Header) error {

	e := ErrRequestHeaderMissingParams
	ok := true

	if x := h.Get("X-NS-TENANTID"); x == "" {
		e = errors.Wrap(e, "X-NS-TENANTID")
		ok = false
	}

	if x := h.Get("X-NS-SERVICE"); x == "" {
		e = errors.Wrap(e, "X-NS-SERVICE")
		ok = false
	}

	if ok {
		return nil
	}

	return e
}

func validateRequestBody(rr Request) error {

	e := ErrRequestBodyMissingParams
	ok := true

	if rr.HTTPRequest == "" {
		e = errors.Wrap(e, "http_request")
		ok = false
	}

	if rr.Type == "" {
		e = errors.Wrap(e, "type")
		ok = false
	}

	if ok {
		return nil
	}

	return e
}