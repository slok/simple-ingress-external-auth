package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/slok/simple-ingress-external-auth/internal/internalerrors"
	"github.com/slok/simple-ingress-external-auth/internal/log"
	"github.com/slok/simple-ingress-external-auth/internal/metrics"
	"github.com/slok/simple-ingress-external-auth/internal/model"
)

type TokenGetter interface {
	GetToken(ctx context.Context, tokenValue string) (*model.Token, error)
}

//go:generate mockery --case underscore --output authmock --outpkg authmock --name TokenGetter

type Service struct {
	tokenGetter TokenGetter
	metricsRec  metrics.Recorder
	logger      log.Logger

	authenticater authenticater
}

func NewService(logger log.Logger, metricsRec metrics.Recorder, tokenGetter TokenGetter) Service {
	return Service{
		tokenGetter: tokenGetter,
		metricsRec:  metricsRec,
		logger:      logger,

		authenticater: newAuthenticaterChain(
			newTokenExistAuthenticator(),
			newDisabledAuthenticator(),
			newNotExpiredAuthenticator(),
			newValidMethodAuthenticator(),
			newValidURLAuthenticator(),
		),
	}
}

type AuthenticateRequest struct {
	Review model.TokenReview
}
type AuthenticateResponse struct {
	Authenticated bool
	Reason        string
}

func (s Service) Authenticate(ctx context.Context, req AuthenticateRequest) (resp *AuthenticateResponse, err error) {
	defer func() {
		var auth, reason = false, ""
		if resp != nil {
			auth = resp.Authenticated
			reason = resp.Reason
		}
		s.metricsRec.TokenReview(ctx, err == nil, auth, reason)
	}()

	if req.Review.Token == "" {
		return nil, fmt.Errorf("token is missing")
	}

	// Get token and its properties.
	token, err := s.tokenGetter.GetToken(ctx, req.Review.Token)
	if err != nil {
		if errors.Is(err, internalerrors.ErrNotFound) {
			return &AuthenticateResponse{Authenticated: false, Reason: ReasonInvalidToken}, nil
		}

		return nil, fmt.Errorf("could not get token: %w", err)
	}

	// Token review.
	res, err := s.authenticater.Authenticate(ctx, req.Review, *token)
	if err != nil {
		return nil, fmt.Errorf("could not authenticate token: %w", err)
	}

	return &AuthenticateResponse{Authenticated: res.Valid, Reason: res.Reason}, nil
}
