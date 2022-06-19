package auth

import (
	"context"
	"time"

	"github.com/slok/simple-ingress-external-auth/internal/model"
)

// Authenticater knows how to authenticate.
type authenticater interface {
	Authenticate(ctx context.Context, r model.TokenReview, t model.Token) (valid bool, err error)
}

type authenticaterFunc func(ctx context.Context, r model.TokenReview, t model.Token) (bool, error)

func (a authenticaterFunc) Authenticate(ctx context.Context, r model.TokenReview, t model.Token) (bool, error) {
	return a(ctx, r, t)
}

func newAuthenticaterChain(auths ...authenticater) authenticater {
	return authenticaterFunc(func(ctx context.Context, r model.TokenReview, t model.Token) (bool, error) {
		for _, a := range auths {
			valid, err := a.Authenticate(ctx, r, t)
			if err != nil {
				return false, err
			}

			// If not valid, end chain.
			if !valid {
				return false, nil
			}
		}

		// Valid.
		return true, nil
	})
}

func newTokenExistAuthenticator() authenticater {
	return authenticaterFunc(func(ctx context.Context, r model.TokenReview, t model.Token) (bool, error) {
		valid := r.Token == t.Value

		return valid, nil
	})
}

func newNotExpiredAuthenticator() authenticater {
	return authenticaterFunc(func(ctx context.Context, r model.TokenReview, t model.Token) (bool, error) {
		if t.ExpiresAt.IsZero() {
			return true, nil
		}

		valid := time.Now().Before(t.ExpiresAt)

		return valid, nil
	})
}

func newValidMethodAuthenticator() authenticater {
	return authenticaterFunc(func(ctx context.Context, r model.TokenReview, t model.Token) (bool, error) {
		if t.AllowedMethod == nil {
			return true, nil
		}

		valid := t.AllowedMethod.MatchString(r.HTTPMethod)

		return valid, nil
	})
}

func newValidURLAuthenticator() authenticater {
	return authenticaterFunc(func(ctx context.Context, r model.TokenReview, t model.Token) (bool, error) {
		if t.AllowedURL == nil {
			return true, nil
		}

		valid := t.AllowedURL.MatchString(r.HTTPURL)

		return valid, nil
	})
}

func newDisabledAuthenticator() authenticater {
	return authenticaterFunc(func(ctx context.Context, r model.TokenReview, t model.Token) (bool, error) {
		valid := !t.Disable

		return valid, nil
	})
}
