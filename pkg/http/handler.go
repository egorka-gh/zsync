package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	endpoint "github.com/egorka-gh/zbazar/zsync/pkg/endpoint"
	http1 "github.com/go-kit/kit/transport/http"
)

//PackPattern url pattern for serving static content (sync pack files)
const PackPattern string = "/pack/"

// makeListVersionHandler creates the handler logic
func makeListVersionHandler(m *http.ServeMux, endpoints endpoint.Endpoints, options []http1.ServerOption) {
	m.Handle("/list-version", http1.NewServer(endpoints.ListVersionEndpoint, decodeListVersionRequest, encodeListVersionResponse, options...))
}

// decodeListVersionResponse  is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeListVersionRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := endpoint.ListVersionRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// encodeListVersionResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer
func encodeListVersionResponse(ctx context.Context, w http.ResponseWriter, response interface{}) (err error) {
	if f, ok := response.(endpoint.Failure); ok && f.Failed() != nil {
		ErrorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(response)
	return
}

// makePullPackHandler creates the handler logic
func makePullPackHandler(m *http.ServeMux, endpoints endpoint.Endpoints, options []http1.ServerOption) {
	m.Handle("/pull-pack", http1.NewServer(endpoints.PullPackEndpoint, decodePullPackRequest, encodePullPackResponse, options...))
}

// decodePullPackResponse  is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodePullPackRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := endpoint.PullPackRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// encodePullPackResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer
func encodePullPackResponse(ctx context.Context, w http.ResponseWriter, response interface{}) (err error) {
	if f, ok := response.(endpoint.Failure); ok && f.Failed() != nil {
		ErrorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(response)
	return
}

// makePushPackHandler creates the handler logic
func makePushPackHandler(m *http.ServeMux, endpoints endpoint.Endpoints, options []http1.ServerOption) {
	m.Handle("/push-pack", http1.NewServer(endpoints.PushPackEndpoint, decodePushPackRequest, encodePushPackResponse, options...))
}

// decodePushPackResponse  is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodePushPackRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := endpoint.PushPackRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// encodePushPackResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer
func encodePushPackResponse(ctx context.Context, w http.ResponseWriter, response interface{}) (err error) {
	if f, ok := response.(endpoint.Failure); ok && f.Failed() != nil {
		ErrorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(response)
	return
}

// makePackDoneHandler creates the handler logic
func makePackDoneHandler(m *http.ServeMux, endpoints endpoint.Endpoints, options []http1.ServerOption) {
	m.Handle("/pack-done", http1.NewServer(endpoints.PackDoneEndpoint, decodePackDoneRequest, encodePackDoneResponse, options...))
}

// decodePackDoneResponse  is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodePackDoneRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := endpoint.PackDoneRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// encodePackDoneResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer
func encodePackDoneResponse(ctx context.Context, w http.ResponseWriter, response interface{}) (err error) {
	if f, ok := response.(endpoint.Failure); ok && f.Failed() != nil {
		ErrorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(response)
	return
}

// makeAddActivityHandler creates the handler logic
func makeAddActivityHandler(m *http.ServeMux, endpoints endpoint.Endpoints, options []http1.ServerOption) {
	m.Handle("/add-activity", http1.NewServer(endpoints.AddActivityEndpoint, decodeAddActivityRequest, encodeAddActivityResponse, options...))
}

// decodeAddActivityResponse  is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeAddActivityRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := endpoint.AddActivityRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// encodeAddActivityResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer
func encodeAddActivityResponse(ctx context.Context, w http.ResponseWriter, response interface{}) (err error) {
	if f, ok := response.(endpoint.Failure); ok && f.Failed() != nil {
		ErrorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(response)
	return
}

// makeGetLevelHandler creates the handler logic
func makeGetLevelHandler(m *http.ServeMux, endpoints endpoint.Endpoints, options []http1.ServerOption) {
	m.Handle("/get-level", http1.NewServer(endpoints.GetLevelEndpoint, decodeGetLevelRequest, encodeGetLevelResponse, options...))
}

// decodeGetLevelResponse  is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded request from the HTTP request body.
func decodeGetLevelRequest(_ context.Context, r *http.Request) (interface{}, error) {
	req := endpoint.GetLevelRequest{}
	err := json.NewDecoder(r.Body).Decode(&req)
	return req, err
}

// encodeGetLevelResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer
func encodeGetLevelResponse(ctx context.Context, w http.ResponseWriter, response interface{}) (err error) {
	if f, ok := response.(endpoint.Failure); ok && f.Failed() != nil {
		ErrorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	err = json.NewEncoder(w).Encode(response)
	return
}
func ErrorEncoder(_ context.Context, err error, w http.ResponseWriter) {
	w.WriteHeader(err2code(err))
	json.NewEncoder(w).Encode(errorWrapper{Error: err.Error()})
}
func ErrorDecoder(r *http.Response) error {
	var w errorWrapper
	if err := json.NewDecoder(r.Body).Decode(&w); err != nil {
		return err
	}
	return errors.New(w.Error)
}

// This is used to set the http status, see an example here :
// https://github.com/go-kit/kit/blob/master/examples/addsvc/pkg/addtransport/http.go#L133
func err2code(err error) int {
	return http.StatusInternalServerError
}

type errorWrapper struct {
	Error string `json:"error"`
}
