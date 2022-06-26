package prometheus

import (
	"context"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	httpmetrics "github.com/slok/go-http-metrics/metrics"
	httpmetricsprometheus "github.com/slok/go-http-metrics/metrics/prometheus"
)

const prefix = "simple_ingress_external_auth"

type Recorder struct {
	httpmetrics.Recorder

	tokenReview *prometheus.CounterVec
}

func NewRecorder(reg prometheus.Registerer) Recorder {
	// Create HTTP metrics recorder.
	rec := httpmetricsprometheus.NewRecorder(httpmetricsprometheus.Config{
		Prefix:   prefix,
		Registry: reg,
	})

	r := Recorder{
		Recorder: rec,

		tokenReview: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: prefix,
			Subsystem: "token",
			Name:      "reviews_total",
			Help:      "The number of token reviews.",
		}, []string{"success", "valid", "invalid_reason"}),
	}

	reg.MustRegister(
		r.tokenReview,
	)

	return r
}

func (r Recorder) TokenReview(ctx context.Context, success, valid bool, invalidReason string) {
	r.tokenReview.WithLabelValues(
		strconv.FormatBool(success),
		strconv.FormatBool(valid),
		invalidReason).Inc()
}
