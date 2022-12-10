package authenticate_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appauth "github.com/slok/simple-ingress-external-auth/internal/app/auth"
	httpauthenticate "github.com/slok/simple-ingress-external-auth/internal/http/authenticate"
	"github.com/slok/simple-ingress-external-auth/internal/log"
	"github.com/slok/simple-ingress-external-auth/internal/metrics"
	"github.com/slok/simple-ingress-external-auth/internal/storage/memory"
)

var tokens = `
{
	"version": "v1",
	"tokens": [
		{"value": "token0", "client_id": "foo"},
		{"value": "token1", "disable": true}
	]
}
`

func TestIntegrationAuthenticate(t *testing.T) {
	tests := map[string]struct {
		tokens         string
		httpHeaders    map[string]string
		expCode        int
		clientIdHeader string
	}{
		"A request without token, should return 404": {
			tokens:         tokens,
			expCode:        http.StatusBadRequest,
			clientIdHeader: "",
		},

		"A request with an invalid token, should return 401": {
			tokens: tokens,
			httpHeaders: map[string]string{
				"Authorization": "Bearer token1",
			},
			expCode:        http.StatusUnauthorized,
			clientIdHeader: "",
		},

		"A request with a valid token, should return 200": {
			tokens: tokens,
			httpHeaders: map[string]string{
				"Authorization": "Bearer token0",
			},
			expCode:        http.StatusOK,
			clientIdHeader: "",
		},
		"A request with a valid token, should return 200 with a ClientID response header": {
			tokens: tokens,
			httpHeaders: map[string]string{
				"Authorization": "Bearer token0",
			},
			expCode:        http.StatusOK,
			clientIdHeader: "X-Client-Id",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			// Create dependencies.
			repo, err := memory.NewTokenRepository(log.Noop, test.tokens)
			require.NoError(err)
			svc := appauth.NewService(log.Noop, metrics.Noop, repo)

			// Run server.
			handler := httpauthenticate.New(log.Noop, metrics.Noop, svc, test.clientIdHeader)
			server := httptest.NewServer(handler)
			defer server.Close()

			// Make request.
			req, _ := http.NewRequest(http.MethodGet, server.URL, nil)
			for k, v := range test.httpHeaders {
				req.Header.Add(k, v)
			}
			resp, err := http.DefaultClient.Do(req)
			require.NoError(err)

			// Check Status Code
			assert.Equal(test.expCode, resp.StatusCode)
			// Check ClientID Header value
			if test.clientIdHeader != "" {
				assert.Equal("foo", resp.Header[http.CanonicalHeaderKey(test.clientIdHeader)][0])
			}
		})
	}
}
