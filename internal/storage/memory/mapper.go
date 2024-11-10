package memory

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/drone/envsubst"
	"github.com/ghodss/yaml"

	"github.com/slok/simple-ingress-external-auth/internal/model"
	apiv1 "github.com/slok/simple-ingress-external-auth/pkg/api/v1"
)

func mapJSONV1ToModel(data string) (map[string]model.StaticTokenValidation, error) {
	// Substitute env vars in the required strings.
	envedData, err := envsubst.EvalEnv(data)
	if err != nil {
		return nil, fmt.Errorf("could not substitute env vars into the configuration: %w", err)
	}

	// Try loading first in JSON and then YAML.
	c1 := apiv1.Config{}
	err = json.Unmarshal([]byte(envedData), &c1)
	if err != nil {
		err2 := yaml.Unmarshal([]byte(envedData), &c1)
		if err2 != nil {
			return nil, fmt.Errorf("json and yaml unrmashal failed, json: %q, yaml: %q", err, err2)
		}
	}

	if c1.Version != "v1" {
		return nil, fmt.Errorf("invalid version, expected v1, got %s", c1.Version)
	}

	// Map.
	tokens := map[string]model.StaticTokenValidation{}
	for _, t := range c1.Tokens {
		if t.Value == "" {
			return nil, fmt.Errorf("token value can't be empty")
		}

		if t.Disable {
			continue
		}

		var expiresAt time.Time
		if t.ExpiresAt != nil {
			expiresAt = *t.ExpiresAt
		}

		token := model.StaticTokenValidation{
			Value:     t.Value,
			ClientID:  t.ClientID,
			ExpiresAt: expiresAt,
		}

		if t.AllowedMethodRegex != "" {
			r, err := regexp.Compile(t.AllowedMethodRegex)
			if err != nil {
				return nil, fmt.Errorf("could not compile %s regex: %w", t.AllowedMethodRegex, err)
			}
			token.Common.AllowedMethod = r
		}

		if t.AllowedURLRegex != "" {
			r, err := regexp.Compile(t.AllowedURLRegex)
			if err != nil {
				return nil, fmt.Errorf("could not compile %s regex: %w", t.AllowedURLRegex, err)
			}
			token.Common.AllowedURL = r
		}

		// Check same token is not twice.
		_, ok := tokens[token.Value]
		if ok {
			return nil, fmt.Errorf("a token has been declared multiple times")
		}

		tokens[token.Value] = token
	}

	return tokens, nil
}
