package transport

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/zipkin"
	"github.com/go-kit/kit/transport"
	kitHttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	goZipkin "github.com/openzipkin/zipkin-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"secKill/oauth-service/endpoint"
	"secKill/oauth-service/service"
)

var (
	//ErrorBadRequest         = errors.New("invalid request parameter")
	ErrorGrantTypeRequest   = errors.New("invalid request grant type")
	ErrorTokenRequest       = errors.New("invalid request token")
	ErrInvalidClientRequest = errors.New("invalid client message")
)

// MakeHttpHandler make http handler use mux
func MakeHttpHandler(_ context.Context, endpoints endpoint.OAuth2Endpoints, _ service.TokenService, clientService service.ClientDetailsService, zipkinTracer *goZipkin.Tracer, logger log.Logger) http.Handler {
	r := mux.NewRouter()
	zipkinServer := zipkin.HTTPServerTrace(zipkinTracer, zipkin.Name("http-transport"))

	options := []kitHttp.ServerOption{
		kitHttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kitHttp.ServerErrorEncoder(encodeError),
		zipkinServer,
	}
	r.Path("/metrics").Handler(promhttp.Handler())

	clientAuthorizationOptions := []kitHttp.ServerOption{
		kitHttp.ServerBefore(makeClientAuthorizationContext(clientService, logger)),
		kitHttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kitHttp.ServerErrorEncoder(encodeError),
		zipkinServer,
	}

	r.Methods("POST").Path("/oauth/token").Handler(kitHttp.NewServer(
		endpoints.TokenEndpoint,
		decodeTokenRequest,
		encodeJsonResponse,
		clientAuthorizationOptions...,
	))

	r.Methods("POST").Path("/oauth/check_token").Handler(kitHttp.NewServer(
		endpoints.CheckTokenEndpoint,
		decodeCheckTokenRequest,
		encodeJsonResponse,
		clientAuthorizationOptions...,
	))

	// create health check handler
	r.Methods("GET").Path("/health").Handler(kitHttp.NewServer(
		endpoints.HealthCheckEndpoint,
		decodeHealthCheckRequest,
		encodeJsonResponse,
		options...,
	))

	return r
}

func makeClientAuthorizationContext(clientDetailsService service.ClientDetailsService, _ log.Logger) kitHttp.RequestFunc {

	return func(ctx context.Context, r *http.Request) context.Context {

		if clientId, clientSecret, ok := r.BasicAuth(); ok {
			clientDetails, err := clientDetailsService.GetClientDetailByClientId(ctx, clientId, clientSecret)
			if err == nil {
				return context.WithValue(ctx, endpoint.OAuth2ClientDetailsKey, clientDetails)
			}
		}
		return context.WithValue(ctx, endpoint.OAuth2ErrorKey, ErrInvalidClientRequest)
	}
}

func decodeTokenRequest(_ context.Context, r *http.Request) (interface{}, error) {
	grantType := r.URL.Query().Get("grant_type")
	if grantType == "" {
		return nil, ErrorGrantTypeRequest
	}
	return &endpoint.TokenRequest{
		GrantType: grantType,
		Reader:    r,
	}, nil

}

func decodeCheckTokenRequest(_ context.Context, r *http.Request) (interface{}, error) {
	tokenValue := r.URL.Query().Get("token")
	if tokenValue == "" {
		return nil, ErrorTokenRequest
	}

	return &endpoint.CheckTokenRequest{
		Token: tokenValue,
	}, nil

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

func encodeJsonResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// decodeHealthCheckRequest decode request
func decodeHealthCheckRequest(_ context.Context, _ *http.Request) (interface{}, error) {
	return endpoint.HealthRequest{}, nil
}
