package prometheus

import (
	"context"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const prefix = "simple_ingress_external_auth"

type Recorder struct {
	tokenReview *prometheus.CounterVec
}

func NewRecorder(reg prometheus.Registerer) Recorder {
	r := Recorder{
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
