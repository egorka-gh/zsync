package endpoint

import (
	"context"

	service "github.com/egorka-gh/zbazar/zsync/pkg/service"
	endpoint "github.com/go-kit/kit/endpoint"
)

// ListVersionRequest collects the request parameters for the ListVersion method.
type ListVersionRequest struct {
	Source string `json:"source"`
}

// ListVersionResponse collects the response parameters for the ListVersion method.
type ListVersionResponse struct {
	V0 []service.Version `json:"v0"`
	E1 error             `json:"e1"`
}

// MakeListVersionEndpoint returns an endpoint that invokes ListVersion on the service.
func MakeListVersionEndpoint(s service.ZsyncService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(ListVersionRequest)
		v0, e1 := s.ListVersion(ctx, req.Source)
		return ListVersionResponse{
			E1: e1,
			V0: v0,
		}, nil
	}
}

// Failed implements Failer.
func (r ListVersionResponse) Failed() error {
	return r.E1
}

// PullPackRequest collects the request parameters for the PullPack method.
type PullPackRequest struct {
	Source string `json:"source"`
	Table  string `json:"table"`
	Start  int    `json:"start"`
}

// PullPackResponse collects the response parameters for the PullPack method.
type PullPackResponse struct {
	V0 service.VersionPack `json:"v0"`
	E1 error               `json:"e1"`
}

// MakePullPackEndpoint returns an endpoint that invokes PullPack on the service.
func MakePullPackEndpoint(s service.ZsyncService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(PullPackRequest)
		v0, e1 := s.PullPack(ctx, req.Source, req.Table, req.Start)
		return PullPackResponse{
			E1: e1,
			V0: v0,
		}, nil
	}
}

// Failed implements Failer.
func (r PullPackResponse) Failed() error {
	return r.E1
}

// PushPackRequest collects the request parameters for the PushPack method.
type PushPackRequest struct {
	Pack service.VersionPack `json:"pack"`
}

// PushPackResponse collects the response parameters for the PushPack method.
type PushPackResponse struct {
	E0 error `json:"e0"`
}

// MakePushPackEndpoint returns an endpoint that invokes PushPack on the service.
func MakePushPackEndpoint(s service.ZsyncService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(PushPackRequest)
		e0 := s.PushPack(ctx, req.Pack)
		return PushPackResponse{E0: e0}, nil
	}
}

// Failed implements Failer.
func (r PushPackResponse) Failed() error {
	return r.E0
}

// PackDoneRequest collects the request parameters for the PackDone method.
type PackDoneRequest struct {
	Pack service.VersionPack `json:"pack"`
}

// PackDoneResponse collects the response parameters for the PackDone method.
type PackDoneResponse struct {
	E0 error `json:"e0"`
}

// MakePackDoneEndpoint returns an endpoint that invokes PackDone on the service.
func MakePackDoneEndpoint(s service.ZsyncService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(PackDoneRequest)
		e0 := s.PackDone(ctx, req.Pack)
		return PackDoneResponse{E0: e0}, nil
	}
}

// Failed implements Failer.
func (r PackDoneResponse) Failed() error {
	return r.E0
}

// AddActivityRequest collects the request parameters for the AddActivity method.
type AddActivityRequest struct {
	Activity service.Activity `json:"activity"`
}

// AddActivityResponse collects the response parameters for the AddActivity method.
type AddActivityResponse struct {
	E0 error `json:"e0"`
}

// MakeAddActivityEndpoint returns an endpoint that invokes AddActivity on the service.
func MakeAddActivityEndpoint(s service.ZsyncService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(AddActivityRequest)
		e0 := s.AddActivity(ctx, req.Activity)
		return AddActivityResponse{E0: e0}, nil
	}
}

// Failed implements Failer.
func (r AddActivityResponse) Failed() error {
	return r.E0
}

// GetLevelRequest collects the request parameters for the GetLevel method.
type GetLevelRequest struct {
	Card string `json:"card"`
}

// GetLevelResponse collects the response parameters for the GetLevel method.
type GetLevelResponse struct {
	I0 int   `json:"i0"`
	E1 error `json:"e1"`
}

// MakeGetLevelEndpoint returns an endpoint that invokes GetLevel on the service.
func MakeGetLevelEndpoint(s service.ZsyncService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(GetLevelRequest)
		i0, e1 := s.GetLevel(ctx, req.Card)
		return GetLevelResponse{
			E1: e1,
			I0: i0,
		}, nil
	}
}

// Failed implements Failer.
func (r GetLevelResponse) Failed() error {
	return r.E1
}

// Failer is an interface that should be implemented by response types.
// Response encoders can check if responses are Failer, and if so they've
// failed, and if so encode them using a separate write path based on the error.
type Failure interface {
	Failed() error
}

// ListVersion implements Service. Primarily useful in a client.
func (e Endpoints) ListVersion(ctx context.Context, source string) (v0 []service.Version, e1 error) {
	request := ListVersionRequest{Source: source}
	response, err := e.ListVersionEndpoint(ctx, request)
	if err != nil {
		return
	}
	return response.(ListVersionResponse).V0, response.(ListVersionResponse).E1
}

// PullPack implements Service. Primarily useful in a client.
func (e Endpoints) PullPack(ctx context.Context, source string, table string, start int) (v0 service.VersionPack, e1 error) {
	request := PullPackRequest{
		Source: source,
		Start:  start,
		Table:  table,
	}
	response, err := e.PullPackEndpoint(ctx, request)
	if err != nil {
		return
	}
	return response.(PullPackResponse).V0, response.(PullPackResponse).E1
}

// PushPack implements Service. Primarily useful in a client.
func (e Endpoints) PushPack(ctx context.Context, pack service.VersionPack) (e0 error) {
	request := PushPackRequest{Pack: pack}
	response, err := e.PushPackEndpoint(ctx, request)
	if err != nil {
		return
	}
	return response.(PushPackResponse).E0
}

// PackDone implements Service. Primarily useful in a client.
func (e Endpoints) PackDone(ctx context.Context, pack service.VersionPack) (e0 error) {
	request := PackDoneRequest{Pack: pack}
	response, err := e.PackDoneEndpoint(ctx, request)
	if err != nil {
		return
	}
	return response.(PackDoneResponse).E0
}

// AddActivity implements Service. Primarily useful in a client.
func (e Endpoints) AddActivity(ctx context.Context, activity service.Activity) (e0 error) {
	request := AddActivityRequest{Activity: activity}
	response, err := e.AddActivityEndpoint(ctx, request)
	if err != nil {
		return
	}
	return response.(AddActivityResponse).E0
}

// GetLevel implements Service. Primarily useful in a client.
func (e Endpoints) GetLevel(ctx context.Context, card string) (i0 int, e1 error) {
	request := GetLevelRequest{Card: card}
	response, err := e.GetLevelEndpoint(ctx, request)
	if err != nil {
		return
	}
	return response.(GetLevelResponse).I0, response.(GetLevelResponse).E1
}
