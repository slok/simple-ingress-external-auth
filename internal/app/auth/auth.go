package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/slok/simple-ingress-external-auth/internal/internalerrors"
	"github.com/slok/simple-ingress-external-auth/internal/log"
	"github.com/slok/simple-ingress-external-auth/internal/model"
)

type TokenGetter interface {
	GetToken(ctx context.Context, tokenValue string) (*model.Token, error)
}

//go:generate mockery --case underscore --output authmock --outpkg authmock --name TokenGetter

type Service struct {
	tokenGetter TokenGetter
	logger      log.Logger

	authenticater authenticater
}

func NewService(logger log.Logger, tokenGetter TokenGetter) Service {
	return Service{
		tokenGetter: tokenGetter,
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
}

func (s Service) Authenticate(ctx context.Context, req AuthenticateRequest) (*AuthenticateResponse, error) {
	if req.Review.Token == "" {
		return nil, fmt.Errorf("token is missing")
	}

	token, err := s.tokenGetter.GetToken(ctx, req.Review.Token)
	if err != nil {

		if errors.Is(err, internalerrors.ErrNotFound) {
			return &AuthenticateResponse{Authenticated: false}, nil
		}

		return nil, fmt.Errorf("could not get token: %w", err)
	}

	valid, err := s.authenticater.Authenticate(ctx, req.Review, *token)
	if err != nil {
		return nil, fmt.Errorf("could not authenticate token: %w", err)
	}

	return &AuthenticateResponse{Authenticated: valid}, nil
}
