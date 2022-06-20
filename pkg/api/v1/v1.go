package v1

import "time"

type Config struct {
	Version string  `json:"version"`
	Tokens  []Token `json:"tokens"`
}

type Token struct {
	Value              string     `json:"value"`
	Disable            bool       `json:"disable,omitempty"`
	ExpiresAt          *time.Time `json:"expires_at,omitempty"`
	AllowedURLRegex    string     `json:"allowed_url,omitempty"`
	AllowedMethodRegex string     `json:"allowed_method,omitempty"`
}
