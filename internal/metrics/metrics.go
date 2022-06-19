package metrics

import "context"

type Recorder interface {
	TokenReview(ctx context.Context, success, valid bool, invalidReason string)
}

type noop bool

const Noop = noop(false)

func (noop) TokenReview(ctx context.Context, success, valid bool, invalidReason string) {}
