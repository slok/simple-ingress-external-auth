package auth

import (
	"context"
	"time"

	"github.com/slok/simple-ingress-external-auth/internal/model"
)

const (
	ReasonInvalidToken  = "invalidToken"
	ReasonExpiredToken  = "expiredToken"
	ReasonInvalidURL    = "invalidURL"
	ReasonInvalidMethod = "invalidMethod"
)

type reviewResult struct {
	Valid  bool
	Reason string
}

// Authenticater knows how to authenticate.
type authenticater interface {
	Authenticate(ctx context.Context, r model.TokenReview, t model.StaticTokenValidation) (*reviewResult, error)
}

type authenticaterFunc func(ctx context.Context, r model.TokenReview, t model.StaticTokenValidation) (*reviewResult, error)

func (a authenticaterFunc) Authenticate(ctx context.Context, r model.TokenReview, t model.StaticTokenValidation) (*reviewResult, error) {
	return a(ctx, r, t)
}

func newAuthenticaterChain(auths ...authenticater) authenticater {
	return authenticaterFunc(func(ctx context.Context, r model.TokenReview, t model.StaticTokenValidation) (*reviewResult, error) {
		var res *reviewResult
		var err error
		for _, a := range auths {
			res, err = a.Authenticate(ctx, r, t)
			if err != nil {
				return nil, err
			}

			// If not valid, end chain.
			if !res.Valid {
				return res, nil
			}
		}

		return res, nil
	})
}

func newTokenExistAuthenticator() authenticater {
	return authenticaterFunc(func(ctx context.Context, r model.TokenReview, t model.StaticTokenValidation) (*reviewResult, error) {
		if r.Token == t.Value {
			return &reviewResult{Valid: true}, nil
		}

		return &reviewResult{Valid: false, Reason: ReasonInvalidToken}, nil
	})
}

func newNotExpiredAuthenticator() authenticater {
	return authenticaterFunc(func(ctx context.Context, r model.TokenReview, t model.StaticTokenValidation) (*reviewResult, error) {
		if t.ExpiresAt.IsZero() {
			return &reviewResult{Valid: true}, nil
		}

		if time.Now().Before(t.ExpiresAt) {
			return &reviewResult{Valid: true}, nil
		}

		return &reviewResult{Valid: false, Reason: ReasonExpiredToken}, nil
	})
}

func newValidMethodAuthenticator() authenticater {
	return authenticaterFunc(func(ctx context.Context, r model.TokenReview, t model.StaticTokenValidation) (*reviewResult, error) {
		if t.Common.AllowedMethod == nil {
			return &reviewResult{Valid: true}, nil
		}

		if t.Common.AllowedMethod.MatchString(r.HTTPMethod) {
			return &reviewResult{Valid: true}, nil
		}

		return &reviewResult{Valid: false, Reason: ReasonInvalidMethod}, nil
	})
}

func newValidURLAuthenticator() authenticater {
	return authenticaterFunc(func(ctx context.Context, r model.TokenReview, t model.StaticTokenValidation) (*reviewResult, error) {
		if t.Common.AllowedURL == nil {
			return &reviewResult{Valid: true}, nil
		}

		if t.Common.AllowedURL.MatchString(r.HTTPURL) {
			return &reviewResult{Valid: true}, nil
		}

		return &reviewResult{Valid: false, Reason: ReasonInvalidURL}, nil
	})
}
