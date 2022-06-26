package metrics

import (
	"context"
	"time"

	httpmetrics "github.com/slok/go-http-metrics/metrics"
)

type Recorder interface {
	TokenReview(ctx context.Context, success, valid bool, invalidReason string)

	// Metrics.
	httpmetrics.Recorder
}

type noop bool

const Noop = noop(false)

func (noop) TokenReview(ctx context.Context, success, valid bool, invalidReason string) {}
func (noop) ObserveHTTPRequestDuration(ctx context.Context, h httpmetrics.HTTPReqProperties, t time.Duration) {
}
func (noop) ObserveHTTPResponseSize(ctx context.Context, h httpmetrics.HTTPReqProperties, t int64) {}
func (noop) AddInflightRequests(ctx context.Context, h httpmetrics.HTTPProperties, t int)          {}
