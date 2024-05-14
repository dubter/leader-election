package run

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	InitState = iota
	AttempterState
	LeaderState
	FailoverState
	StoppingState
)

var mappedStates = map[string]int{
	"InitState":      InitState,
	"AttempterState": AttempterState,
	"LeaderState":    LeaderState,
	"FailoverState":  FailoverState,
	"StoppingState":  StoppingState,
}

var (
	stateChangesTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "state_changes_total",
		Help: "Total number of state changes",
	})
	stateDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "state_duration_seconds",
		Help:    "Duration of states",
		Buckets: prometheus.DefBuckets,
	})
	currentState = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "current_state",
		Help: "Current state",
	})
)

func metrics(ctx context.Context, logger *slog.Logger) {
	prometheus.MustRegister(stateChangesTotal)
	prometheus.MustRegister(stateDuration)
	prometheus.MustRegister(currentState)

	http.Handle("/metrics", promhttp.Handler())
	logger.Info("Starting HTTP metrics server on :8080")
	defer logger.Info("HTTP metrics server is closed")

	httpCh := make(chan error)
	go func() {
		var err error
		if err = http.ListenAndServe(":8080", nil); err != nil {
			logger.Error("Failed to start HTTP metrics server: %v", err)
		}

		httpCh <- err
	}()

	select {
	case <-ctx.Done():
	case <-httpCh:
	}
}
