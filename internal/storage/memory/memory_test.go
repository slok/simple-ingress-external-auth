package memory_test

import (
	"context"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/slok/simple-ingress-external-auth/internal/log"
	"github.com/slok/simple-ingress-external-auth/internal/model"
	"github.com/slok/simple-ingress-external-auth/internal/storage/memory"
)

var (
	goodJSONConfig = `
{
	"version": "v1",
	"tokens": [
		{
			"value": "t0",
			"client_id": "c0"
		},
		{
			"value": "t1",
			"client_id": "c1",
			"disable": true,
			"expires_at": "2022-07-04T14:21:22.52Z",
			"allowed_url": "https://custom.host.slok.dev/.*",
			"allowed_method": "(GET|POST)"
		},
		{
			"value": "t2",
			"client_id": "c2",
			"allowed_method": "PUT"
		}
	]
}
`
	goodYAMLConfig = `
version: v1
tokens:
- value: t0
  client_id: c0

- value: t1
  client_id: c1
  disable: true
  expires_at: 2022-07-04T14:21:22.52Z
  allowed_url: https://custom.host.slok.dev/.*
  allowed_method: (GET|POST)

- value: t2
  client_id: c2
  allowed_method: PUT
`
)

func TestTokenRepositoryGetToken(t *testing.T) {
	tests := map[string]struct {
		config   string
		env      map[string]string
		token    string
		expToken *model.Token
		expErr   bool
	}{
		"If the token is missing, it should fail": {
			config: goodJSONConfig,
			token:  "t3",
			expErr: true,
		},

		"An existing token, should be returned (basic)": {
			config: goodJSONConfig,
			token:  "t0",
			expToken: &model.Token{
				Value:    "t0",
				ClientID: "c0",
			},
		},

		"An existing token, should be returned (full)": {
			config: goodJSONConfig,
			token:  "t1",
			expToken: &model.Token{
				Value:         "t1",
				ClientID:      "c1",
				Disable:       true,
				ExpiresAt:     time.Date(2022, time.Month(7), 4, 14, 21, 22, 520000000, time.UTC),
				AllowedURL:    regexp.MustCompile(`https://custom.host.slok.dev/.*`),
				AllowedMethod: regexp.MustCompile(`(GET|POST)`),
			},
		},

		"An existing token, should be returned (full YAML)": {
			config: goodYAMLConfig,
			token:  "t1",
			expToken: &model.Token{
				Value:         "t1",
				ClientID:      "c1",
				Disable:       true,
				ExpiresAt:     time.Date(2022, time.Month(7), 4, 14, 21, 22, 520000000, time.UTC),
				AllowedURL:    regexp.MustCompile(`https://custom.host.slok.dev/.*`),
				AllowedMethod: regexp.MustCompile(`(GET|POST)`),
			},
		},

		"A token form the env vars should be set correctly.": {
			env: map[string]string{
				"TEST_TOKEN": "1234567890",
			},
			config: `
			{
				"version": "v1",
				"tokens": [
					{
						"value": "${TEST_TOKEN}",
						"client_id": "test-env-client"
					}
				]
			}
			`,
			token: "1234567890",
			expToken: &model.Token{
				Value:    "1234567890",
				ClientID: "test-env-client",
			},
		},

		"A token form the env vars should be set correctly (YAML).": {
			env: map[string]string{
				"TEST_TOKEN": "1234567890",
			},
			config: `
version: v1
tokens: 
- value: ${TEST_TOKEN}
  client_id: test-env-client`,
			token: "1234567890",
			expToken: &model.Token{
				Value:    "1234567890",
				ClientID: "test-env-client",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			require := require.New(t)
			assert := assert.New(t)

			// Prepare env vars.
			for k, v := range test.env {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range test.env {
					os.Unsetenv(k)
				}
			}()

			repo, err := memory.NewTokenRepository(log.Noop, test.config)
			require.NoError(err)

			token, err := repo.GetToken(context.TODO(), test.token)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expToken, token)
			}
		})
	}
}
