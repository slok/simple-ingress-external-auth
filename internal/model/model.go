package model

import (
	"regexp"
	"time"
)

// Token represents an auth token that can be used to validate an authentication.
// Token value generation example: `openssl rand -base64 32`.
type Token struct {
	Value         string
	Disable       bool
	ExpiresAt     time.Time
	AllowedURL    *regexp.Regexp
	AllowedMethod *regexp.Regexp
}

// TokenReview represents an auth requests sent by the client to be reviewed.
type TokenReview struct {
	Token      string
	HTTPURL    string
	HTTPMethod string
}
