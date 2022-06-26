package authenticate

import (
	"fmt"
	"net/http"
	"strings"

	httpmetrics "github.com/slok/go-http-metrics/middleware"
	httpmetricsstd "github.com/slok/go-http-metrics/middleware/std"

	"github.com/slok/simple-ingress-external-auth/internal/app/auth"
	"github.com/slok/simple-ingress-external-auth/internal/log"
	"github.com/slok/simple-ingress-external-auth/internal/metrics"
	"github.com/slok/simple-ingress-external-auth/internal/model"
)

// New returns an HTTP handler that knows how to authenticate external requests.
func New(logger log.Logger, metricRec metrics.Recorder, authAppSvc auth.Service) http.Handler {
	authHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Map request to model.
		review, err := mapRequestToModel(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, err := w.Write([]byte("error mapping request: " + err.Error()))
			if err != nil {
				logger.Warningf("Error writing response body: %s", err)
			}
			return
		}

		// Review authentication.
		resp, err := authAppSvc.Authenticate(r.Context(), *review)
		if err != nil {
			logger.Errorf("auth app error: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			_, err := w.Write([]byte("error authenticating"))
			if err != nil {
				logger.Warningf("Error writing response body: %s", err)
			}
			return
		}

		if !resp.Authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			_, err := w.Write([]byte("invalid token"))
			if err != nil {
				logger.Warningf("Error writing response body: %s", err)
			}
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	// Measure handler.
	metricsMiddleware := httpmetrics.New(httpmetrics.Config{Recorder: metricRec})
	h := httpmetricsstd.Handler("", metricsMiddleware, authHandler)

	return h
}

func mapRequestToModel(r *http.Request) (*auth.AuthenticateRequest, error) {
	// Headers.
	const (
		originalURL         = "X-Original-URL"
		originalMethod      = "X-Original-Method"
		authorization       = "Authorization"
		authorizationBearer = "Bearer"
	)

	// Get token.
	token := r.Header.Get(authorization)
	token = strings.Replace(token, authorizationBearer, "", 1)
	token = strings.TrimSpace(token)

	if token == "" {
		return nil, fmt.Errorf("missing token")
	}

	// Get other properties.
	method := r.Header.Get(originalMethod)
	url := r.Header.Get(originalURL)

	return &auth.AuthenticateRequest{Review: model.TokenReview{
		Token:      token,
		HTTPURL:    url,
		HTTPMethod: method,
	}}, nil
}
