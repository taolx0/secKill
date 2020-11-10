package transport

import (
	"context"
	"encoding/json"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	goZipkin "github.com/openzipkin/zipkin-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	endpoints "secKill/sk-app/endpoint"
	"secKill/sk-app/model"
)

//var (
//	ErrorBadRequest = errors.New("invalid request parameter")
//)

// MakeHttpHandler make http handler use mux
func MakeHttpHandler(_ context.Context, endpoints endpoints.SkAppEndpoints, zipkinTracer *goZipkin.Tracer, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	zipkinServer := zipkin.HTTPServerTrace(zipkinTracer, zipkin.Name("http-transport"))

	options := []kitHttp.ServerOption{
		//kitHttp.ServerErrorLogger(logger),
		kitHttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		//kitHttp.ServerErrorEncoder(kitHttp.DefaultErrorEncoder),
		kitHttp.ServerErrorEncoder(encodeError),
		zipkinServer,
	}

	r.Methods("GET").Path("/sec/info").Handler(kitHttp.NewServer(
		endpoints.GetSecInfoEndpoint,
		decodeSecInfoRequest,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Path("/sec/list").Handler(kitHttp.NewServer(
		endpoints.GetSecInfoListEndpoint,
		decodeSecInfoListRequest,
		encodeResponse,
		options...,
	))

	r.Methods("POST").Path("/sec/kill").Handler(kitHttp.NewServer(
		endpoints.SecKillEndpoint,
		decodeSecKillRequest,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Path("/sec/test").Handler(kitHttp.NewServer(
		endpoints.TestEndpoint,
		decodeSecInfoListRequest,
		encodeResponse,
		options...,
	))

	r.Path("/metrics").Handler(promhttp.Handler())

	// create health check handler
	r.Methods("GET").Path("/health").Handler(kitHttp.NewServer(
		endpoints.HeathCheckEndpoint,
		decodeTestRequest,
		encodeResponse,
		options...,
	))

	return r
}

// decodeUserRequest decode request params to struct
func decodeSecInfoRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var secInfoRequest endpoints.SecInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&secInfoRequest); err != nil {
		return nil, err
	}
	return secInfoRequest, nil
}

// encode errors from business-logic
func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

// encodeArithmeticResponse encode response to return
func encodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

//NOT USE
// decodeHealthCheckRequest decode request
/*func decodeHealthCheckRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return endpoints.HealthRequest{}, nil
}*/

func decodeTestRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return endpoints.HealthRequest{}, nil
}

func decodeSecInfoListRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return nil, nil
}

func decodeSecKillRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var secRequest model.SecRequest
	if err := json.NewDecoder(r.Body).Decode(&secRequest); err != nil {
		return nil, err
	}
	return secRequest, nil
}
