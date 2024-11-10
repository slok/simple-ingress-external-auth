package model

import (
	"regexp"
	"time"
)

// StaticTokenValidation represents an auth static token information that can be used to validate an authentication.
// Token value generation example: `openssl rand -base64 32`.
type StaticTokenValidation struct {
	Value     string
	ClientID  string
	ExpiresAt time.Time
	Common    TokenCommon
}

type TokenCommon struct {
	AllowedURL    *regexp.Regexp
	AllowedMethod *regexp.Regexp
}

// TokenReview represents an auth requests sent by the client to be reviewed.
type TokenReview struct {
	Token      string
	HTTPURL    string
	HTTPMethod string
}
