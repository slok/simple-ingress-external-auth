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
	GetStaticTokenValidation(ctx context.Context, tokenValue string) (*model.StaticTokenValidation, error)
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
	ClientID      string
	Authenticated bool
	Reason        string
}

func (s Service) Authenticate(ctx context.Context, req AuthenticateRequest) (resp *AuthenticateResponse, err error) {
	defer func() {
		var auth, reason, clientID = false, "", ""
		if resp != nil {
			auth = resp.Authenticated
			reason = resp.Reason
			clientID = resp.ClientID
		}
		s.metricsRec.TokenReview(ctx, err == nil, auth, clientID, reason)
	}()

	if req.Review.Token == "" {
		return nil, fmt.Errorf("token is missing")
	}

	logger := s.logger.WithValues(log.Kv{"url": req.Review.HTTPURL, "method": req.Review.HTTPMethod})

	// Get token and its properties.
	token, err := s.tokenGetter.GetStaticTokenValidation(ctx, req.Review.Token)
	if err != nil {
		if errors.Is(err, internalerrors.ErrNotFound) {
			logger.Infof("Unknown token")
			return &AuthenticateResponse{Authenticated: false, Reason: ReasonInvalidToken}, nil
		}

		return nil, fmt.Errorf("could not get token: %w", err)
	}

	// Token review.
	res, err := s.authenticater.Authenticate(ctx, req.Review, *token)
	if err != nil {
		return nil, fmt.Errorf("could not authenticate token: %w", err)
	}

	if !res.Valid {
		logger.WithValues(log.Kv{"client": token.ClientID, "reason": res.Reason}).Infof("Token unauthorized")
	}

	return &AuthenticateResponse{
		ClientID:      token.ClientID,
		Authenticated: res.Valid,
		Reason:        res.Reason,
	}, nil
}
