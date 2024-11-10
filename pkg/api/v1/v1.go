package v1

import "time"

type Config struct {
	Version string  `json:"version"`
	Tokens  []Token `json:"tokens"`
}

type Common struct {
	Disable            bool   `json:"disable,omitempty"`
	AllowedURLRegex    string `json:"allowed_url,omitempty"`
	AllowedMethodRegex string `json:"allowed_method,omitempty"`
}

type Token struct {
	Common

	Value     string     `json:"value"`
	ClientID  string     `json:"client_id"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}
