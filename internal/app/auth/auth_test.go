package auth_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/slok/simple-ingress-external-auth/internal/app/auth"
	"github.com/slok/simple-ingress-external-auth/internal/app/auth/authmock"
	"github.com/slok/simple-ingress-external-auth/internal/internalerrors"
	"github.com/slok/simple-ingress-external-auth/internal/log"
	"github.com/slok/simple-ingress-external-auth/internal/model"
)

func TestServiceAuth(t *testing.T) {
	tests := map[string]struct {
		mock    func(mtg *authmock.TokenGetter)
		req     auth.AuthenticateRequest
		expResp *auth.AuthenticateResponse
		expErr  bool
	}{
		"A token review without token should fail.": {
			mock: func(mtg *authmock.TokenGetter) {},
			req: auth.AuthenticateRequest{Review: model.TokenReview{
				Token: "",
			}},
			expErr: true,
		},

		"A token review with fails while getting the token it should fail.": {
			mock: func(mtg *authmock.TokenGetter) {
				mtg.On("GetToken", mock.Anything, "sometoken").Once().Return(nil, fmt.Errorf("something"))
			},
			req: auth.AuthenticateRequest{Review: model.TokenReview{
				Token: "sometoken",
			}},
			expErr: true,
		},

		"A token review with a valid token should return that is valid.": {
			mock: func(mtg *authmock.TokenGetter) {
				mtg.On("GetToken", mock.Anything, "missing").Once().Return(nil, internalerrors.ErrNotFound)
			},
			req: auth.AuthenticateRequest{Review: model.TokenReview{
				Token: "missing",
			}},
			expResp: &auth.AuthenticateResponse{Authenticated: false},
		},

		"A token review with that is disabled should be invalid.": {
			mock: func(mtg *authmock.TokenGetter) {
				mtg.On("GetToken", mock.Anything, "token0").Once().Return(&model.Token{
					Value:   "token0",
					Disable: true,
				}, nil)
			},
			req: auth.AuthenticateRequest{Review: model.TokenReview{
				Token: "token0",
			}},
			expResp: &auth.AuthenticateResponse{Authenticated: false},
		},

		"A token review that has expired shoud be invalid.": {
			mock: func(mtg *authmock.TokenGetter) {
				mtg.On("GetToken", mock.Anything, "token0").Once().Return(&model.Token{
					Value:     "token0",
					ExpiresAt: time.Now().Add(-24 * time.Hour),
				}, nil)
			},
			req: auth.AuthenticateRequest{Review: model.TokenReview{
				Token: "token0",
			}},
			expResp: &auth.AuthenticateResponse{Authenticated: false},
		},

		"A token review with an invalid URL should be invalid.": {
			mock: func(mtg *authmock.TokenGetter) {
				mtg.On("GetToken", mock.Anything, "token0").Once().Return(&model.Token{
					Value:      "token0",
					AllowedURL: regexp.MustCompile("https://something.com/.*"),
				}, nil)
			},
			req: auth.AuthenticateRequest{Review: model.TokenReview{
				Token:   "token0",
				HTTPURL: "https://otherthing.com/api/v1",
			}},
			expResp: &auth.AuthenticateResponse{Authenticated: false},
		},

		"A token review with an invalid method should be invalid.": {
			mock: func(mtg *authmock.TokenGetter) {
				mtg.On("GetToken", mock.Anything, "token0").Once().Return(&model.Token{
					Value:         "token0",
					AllowedMethod: regexp.MustCompile("POST"),
				}, nil)
			},
			req: auth.AuthenticateRequest{Review: model.TokenReview{
				Token:      "token0",
				HTTPMethod: "GET",
			}},
			expResp: &auth.AuthenticateResponse{Authenticated: false},
		},

		"A token review  that is valid, should be authenticated.": {
			mock: func(mtg *authmock.TokenGetter) {
				mtg.On("GetToken", mock.Anything, "token0").Once().Return(&model.Token{
					Value: "token0",
				}, nil)
			},
			req: auth.AuthenticateRequest{Review: model.TokenReview{
				Token: "token0",
			}},
			expResp: &auth.AuthenticateResponse{Authenticated: true},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			assert := assert.New(t)

			mtg := &authmock.TokenGetter{}
			test.mock(mtg)

			svc := auth.NewService(log.Noop, mtg)

			gotResp, err := svc.Authenticate(context.TODO(), test.req)

			if test.expErr {
				assert.Error(err)
			} else if assert.NoError(err) {
				assert.Equal(test.expResp, gotResp)
			}
		})
	}
}
