package transport

import (
	"context"
	"encoding/json"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	goZipkin "github.com/openzipkin/zipkin-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"os"
	endpoints "secKill/sk-admin/endpoint"
	"secKill/sk-admin/model"
)

var (
//ErrorBadRequest = errors.New("invalid request parameter")
)

// MakeHttpHandler make http handler use mux
func MakeHttpHandler(_ context.Context, endpoints endpoints.SkAdminEndpoints, zipkinTracer *goZipkin.Tracer, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	zipkinServer := zipkin.HTTPServerTrace(zipkinTracer, zipkin.Name("http-transport"))

	options := []kitHttp.ServerOption{
		//kitHttp.ServerErrorLogger(logger),
		kitHttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		//kitHttp.ServerErrorEncoder(kitHttp.DefaultErrorEncoder),
		kitHttp.ServerErrorEncoder(encodeError),
		zipkinServer,
	}

	r.Methods("GET").Path("/product/list").Handler(kitHttp.NewServer(
		endpoints.GetProductEndpoint,
		decodeGetListRequest,
		encodeResponse,
		options...,
	))

	r.Methods("POST").Path("/product/create").Handler(kitHttp.NewServer(
		endpoints.GetProductEndpoint,
		decodeCreateProductCheckRequest,
		encodeResponse,
		options...,
	))

	r.Methods("POST").Path("/activity/create").Handler(kitHttp.NewServer(
		endpoints.CreateActivityEndpoint,
		decodeCreateActivityCheckRequest,
		encodeResponse,
		options...,
	))

	r.Methods("GET").Path("/activity/list").Handler(kitHttp.NewServer(
		endpoints.GetActivityEndpoint,
		decodeGetListRequest,
		encodeResponse,
		options...,
	))

	r.Path("/metrics").Handler(promhttp.Handler())

	// create health check handler
	r.Methods("GET").Path("/health").Handler(kitHttp.NewServer(
		endpoints.HealthCheckEndpoint,
		decodeHealthCheckRequest,
		encodeResponse,
		options...,
	))

	loggedRouter := handlers.LoggingHandler(os.Stdout, r)

	return loggedRouter
}

// decodeUserRequest decode request params to struct
func decodeGetListRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return endpoints.GetListRequest{}, nil
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

// decodeHealthCheckRequest decode request
func decodeHealthCheckRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return endpoints.HealthRequest{}, nil
}

func decodeCreateProductCheckRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var product model.Product
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		return nil, err
	}
	return product, nil
}

func decodeCreateActivityCheckRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var activity model.Activity
	if err := json.NewDecoder(r.Body).Decode(&activity); err != nil {
		return nil, err
	}
	return activity, nil
}
