package memory

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/drone/envsubst"
	"github.com/ghodss/yaml"

	"github.com/slok/simple-ingress-external-auth/internal/model"
)

type configV1 struct {
	Version string          `json:"version"`
	Tokens  []configV1Token `json:"tokens"`
}

type configV1Token struct {
	Value              string    `json:"value"`
	Disable            bool      `json:"disable,omitempty"`
	ExpiresAt          time.Time `json:"expires_at,omitempty"`
	AllowedURLRegex    string    `json:"allowed_url,omitempty"`
	AllowedMethodRegex string    `json:"allowed_method,omitempty"`
}

func mapJSONV1ToModel(data string) (map[string]model.Token, error) {
	// Substitute env vars in the required strings.
	envedData, err := envsubst.EvalEnv(data)
	if err != nil {
		return nil, fmt.Errorf("could not substitute env vars into the configuration: %w", err)
	}

	// Try loading first in JSON and then YAML.
	c1 := configV1{}
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
	tokens := map[string]model.Token{}
	for _, t := range c1.Tokens {
		if t.Value == "" {
			return nil, fmt.Errorf("token value can't be empty")
		}

		token := model.Token{
			Value:     t.Value,
			Disable:   t.Disable,
			ExpiresAt: t.ExpiresAt,
		}

		if t.AllowedMethodRegex != "" {
			r, err := regexp.Compile(t.AllowedMethodRegex)
			if err != nil {
				return nil, fmt.Errorf("could not compile %s regex: %w", t.AllowedMethodRegex, err)
			}
			token.AllowedMethod = r
		}

		if t.AllowedURLRegex != "" {
			r, err := regexp.Compile(t.AllowedURLRegex)
			if err != nil {
				return nil, fmt.Errorf("could not compile %s regex: %w", t.AllowedURLRegex, err)
			}
			token.AllowedURL = r
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
