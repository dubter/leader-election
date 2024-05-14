package run

import (
	"context"
	"fmt"
	"log/slog"
	"time"
)

var _ Runner = &LoopRunner{}

type Runner interface {
	Run(ctx context.Context, state AutomataState) error
}

func NewLoopRunner(logger *slog.Logger) *LoopRunner {
	logger = logger.With("subsystem", "StateRunner")
	return &LoopRunner{
		logger: logger,
	}
}

type LoopRunner struct {
	logger *slog.Logger
}

func (r *LoopRunner) Run(ctx context.Context, state AutomataState) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go metrics(ctx, r.logger)

	for state != nil {
		r.logger.LogAttrs(ctx, slog.LevelInfo, "start running state", slog.String("state", state.String()))

		start := time.Now()
		currentState.Set(float64(mappedStates[state.String()]))

		var err error
		state, err = state.Run(ctx)
		stateChangesTotal.Inc()
		stateDuration.Observe(time.Since(start).Seconds())

		if err != nil {
			return fmt.Errorf("state %s run: %w", state.String(), err)
		}
	}
	r.logger.LogAttrs(ctx, slog.LevelInfo, "no new state, finish")
	return nil
}
