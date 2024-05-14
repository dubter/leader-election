package states

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run"
)

func NewStoppingState(args cmdargs.RunArgs, dg DepGraph) (*StoppingState, error) {
	logger, err := dg.GetLogger()
	if err != nil {
		return nil, fmt.Errorf("get logger: %w", err)
	}

	return &StoppingState{
		logger: logger.With("subsystem", "StoppingState"),
		dg:     dg,
		args:   args,
	}, nil
}

type StoppingState struct {
	logger *slog.Logger
	dg     DepGraph
	args   cmdargs.RunArgs
}

func (s *StoppingState) String() string {
	return "StoppingState"
}

func (s *StoppingState) Run(ctx context.Context) (run.AutomataState, error) {
	attempterState, err := s.dg.GetAttempterState(s.args)
	if err != nil {
		return nil, err
	}

	leaderState, err := s.dg.GetLeaderState(s.args)
	if err != nil {
		return nil, err
	}

	attempterState.Stop()
	leaderState.Stop()

	if ctx.Err() != nil {
		s.logger.LogAttrs(ctx, slog.LevelWarn, "the server is stopped", slog.String("error", ctx.Err().Error()))

		return nil, ctx.Err()
	}

	s.logger.LogAttrs(ctx, slog.LevelWarn, "the server is stopped")

	return nil, nil
}
