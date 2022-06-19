package memory

import (
	"context"
	"fmt"

	"github.com/slok/simple-ingress-external-auth/internal/internalerrors"
	"github.com/slok/simple-ingress-external-auth/internal/log"
	"github.com/slok/simple-ingress-external-auth/internal/model"
)

type TokenRepository struct {
	tokens map[string]model.Token
}

func NewTokenRepository(logger log.Logger, config string) (*TokenRepository, error) {
	tokens, err := mapJSONV1ToModel(config)
	if err != nil {
		return nil, err
	}

	logger.WithValues(log.Kv{"svc": "memory.TokenRepository", "tokens": len(tokens)}).Infof("Tokens loaded")

	return &TokenRepository{tokens: tokens}, nil
}

func (t TokenRepository) GetToken(ctx context.Context, tokenValue string) (*model.Token, error) {
	token, ok := t.tokens[tokenValue]
	if !ok {
		return nil, fmt.Errorf("token not found: %w", internalerrors.ErrNotFound)
	}

	return &token, nil
}
