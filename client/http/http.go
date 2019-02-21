package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	http1 "net/http"
	"net/url"
	"strings"

	endpoint1 "github.com/egorka-gh/zbazar/zsync/pkg/endpoint"
	http2 "github.com/egorka-gh/zbazar/zsync/pkg/http"
	service "github.com/egorka-gh/zbazar/zsync/pkg/service"
	endpoint "github.com/go-kit/kit/endpoint"
	http "github.com/go-kit/kit/transport/http"
)

// New returns an AddService backed by an HTTP server living at the remote
// instance. We expect instance to come from a service discovery system, so
// likely of the form "host:port".
func New(instance string, options map[string][]http.ClientOption) (service.ZsyncService, error) {
	if !strings.HasPrefix(instance, "http") {
		instance = "http://" + instance
	}
	u, err := url.Parse(instance)
	if err != nil {
		return nil, err
	}
	var listVersionEndpoint endpoint.Endpoint
	{
		listVersionEndpoint = http.NewClient("POST", copyURL(u, "/list-version"), encodeHTTPGenericRequest, decodeListVersionResponse, options["ListVersion"]...).Endpoint()
	}

	var pullPackEndpoint endpoint.Endpoint
	{
		pullPackEndpoint = http.NewClient("POST", copyURL(u, "/pull-pack"), encodeHTTPGenericRequest, decodePullPackResponse, options["PullPack"]...).Endpoint()
	}

	var pushPackEndpoint endpoint.Endpoint
	{
		pushPackEndpoint = http.NewClient("POST", copyURL(u, "/push-pack"), encodeHTTPGenericRequest, decodePushPackResponse, options["PushPack"]...).Endpoint()
	}

	var packDoneEndpoint endpoint.Endpoint
	{
		packDoneEndpoint = http.NewClient("POST", copyURL(u, "/pack-done"), encodeHTTPGenericRequest, decodePackDoneResponse, options["PackDone"]...).Endpoint()
	}

	var addActivityEndpoint endpoint.Endpoint
	{
		addActivityEndpoint = http.NewClient("POST", copyURL(u, "/add-activity"), encodeHTTPGenericRequest, decodeAddActivityResponse, options["AddActivity"]...).Endpoint()
	}

	var getLevelEndpoint endpoint.Endpoint
	{
		getLevelEndpoint = http.NewClient("POST", copyURL(u, "/get-level"), encodeHTTPGenericRequest, decodeGetLevelResponse, options["GetLevel"]...).Endpoint()
	}

	return endpoint1.Endpoints{
		AddActivityEndpoint: addActivityEndpoint,
		GetLevelEndpoint:    getLevelEndpoint,
		ListVersionEndpoint: listVersionEndpoint,
		PackDoneEndpoint:    packDoneEndpoint,
		PullPackEndpoint:    pullPackEndpoint,
		PushPackEndpoint:    pushPackEndpoint,
	}, nil
}

// EncodeHTTPGenericRequest is a transport/http.EncodeRequestFunc that
// SON-encodes any request to the request body. Primarily useful in a client.
func encodeHTTPGenericRequest(_ context.Context, r *http1.Request, request interface{}) error {
	var buf bytes.Buffer

	if err := json.NewEncoder(&buf).Encode(request); err != nil {
		return err
	}
	r.Body = ioutil.NopCloser(&buf)
	return nil
}

// decodeListVersionResponse is a transport/http.DecodeResponseFunc that decodes
// a JSON-encoded concat response from the HTTP response body. If the response
// as a non-200 status code, we will interpret that as an error and attempt to
//  decode the specific error message from the response body.
func decodeListVersionResponse(_ context.Context, r *http1.Response) (interface{}, error) {
	if r.StatusCode != http1.StatusOK {
		return nil, http2.ErrorDecoder(r)
	}
	var resp endpoint1.ListVersionResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

// decodePullPackResponse is a transport/http.DecodeResponseFunc that decodes
// a JSON-encoded concat response from the HTTP response body. If the response
// as a non-200 status code, we will interpret that as an error and attempt to
//  decode the specific error message from the response body.
func decodePullPackResponse(_ context.Context, r *http1.Response) (interface{}, error) {
	if r.StatusCode != http1.StatusOK {
		return nil, http2.ErrorDecoder(r)
	}
	var resp endpoint1.PullPackResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

// decodePushPackResponse is a transport/http.DecodeResponseFunc that decodes
// a JSON-encoded concat response from the HTTP response body. If the response
// as a non-200 status code, we will interpret that as an error and attempt to
//  decode the specific error message from the response body.
func decodePushPackResponse(_ context.Context, r *http1.Response) (interface{}, error) {
	if r.StatusCode != http1.StatusOK {
		return nil, http2.ErrorDecoder(r)
	}
	var resp endpoint1.PushPackResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

// decodePackDoneResponse is a transport/http.DecodeResponseFunc that decodes
// a JSON-encoded concat response from the HTTP response body. If the response
// as a non-200 status code, we will interpret that as an error and attempt to
//  decode the specific error message from the response body.
func decodePackDoneResponse(_ context.Context, r *http1.Response) (interface{}, error) {
	if r.StatusCode != http1.StatusOK {
		return nil, http2.ErrorDecoder(r)
	}
	var resp endpoint1.PackDoneResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

// decodeAddActivityResponse is a transport/http.DecodeResponseFunc that decodes
// a JSON-encoded concat response from the HTTP response body. If the response
// as a non-200 status code, we will interpret that as an error and attempt to
//  decode the specific error message from the response body.
func decodeAddActivityResponse(_ context.Context, r *http1.Response) (interface{}, error) {
	if r.StatusCode != http1.StatusOK {
		return nil, http2.ErrorDecoder(r)
	}
	var resp endpoint1.AddActivityResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}

// decodeGetLevelResponse is a transport/http.DecodeResponseFunc that decodes
// a JSON-encoded concat response from the HTTP response body. If the response
// as a non-200 status code, we will interpret that as an error and attempt to
//  decode the specific error message from the response body.
func decodeGetLevelResponse(_ context.Context, r *http1.Response) (interface{}, error) {
	if r.StatusCode != http1.StatusOK {
		return nil, http2.ErrorDecoder(r)
	}
	var resp endpoint1.GetLevelResponse
	err := json.NewDecoder(r.Body).Decode(&resp)
	return resp, err
}
func copyURL(base *url.URL, path string) (next *url.URL) {
	n := *base
	n.Path = path
	next = &n
	return
}
