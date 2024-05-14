package depgraph

import (
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/states"
)

type dgEntity[T any] struct {
	sync.Once
	value   T
	initErr error
}

func (e *dgEntity[T]) get(init func() (T, error)) (T, error) {
	e.Do(func() {
		e.value, e.initErr = init()
	})
	if e.initErr != nil {
		return *new(T), e.initErr
	}
	return e.value, nil
}

type DepGraph struct {
	logger         *dgEntity[*slog.Logger]
	stateRunner    *dgEntity[*run.LoopRunner]
	initState      *dgEntity[*states.InitState]
	attempterState *dgEntity[*states.AttempterState]
	leaderState    *dgEntity[*states.LeaderState]
	failoverState  *dgEntity[*states.FailoverState]
	stoppingState  *dgEntity[*states.StoppingState]
}

func New() *DepGraph {
	return &DepGraph{
		logger:         &dgEntity[*slog.Logger]{},
		stateRunner:    &dgEntity[*run.LoopRunner]{},
		initState:      &dgEntity[*states.InitState]{},
		attempterState: &dgEntity[*states.AttempterState]{},
		leaderState:    &dgEntity[*states.LeaderState]{},
		failoverState:  &dgEntity[*states.FailoverState]{},
		stoppingState:  &dgEntity[*states.StoppingState]{},
	}
}

func (dg *DepGraph) GetLogger() (*slog.Logger, error) {
	return dg.logger.get(func() (*slog.Logger, error) {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{})), nil
	})
}

func (dg *DepGraph) GetInitState(args cmdargs.RunArgs) (*states.InitState, error) {
	return dg.initState.get(func() (*states.InitState, error) {
		return states.NewInitState(args, dg)
	})
}

func (dg *DepGraph) GetAttempterState(args cmdargs.RunArgs) (*states.AttempterState, error) {
	return dg.attempterState.get(func() (*states.AttempterState, error) {
		return states.NewAttempterState(args, dg)
	})
}

func (dg *DepGraph) GetLeaderState(args cmdargs.RunArgs) (*states.LeaderState, error) {
	return dg.leaderState.get(func() (*states.LeaderState, error) {
		return states.NewLeaderState(args, dg)
	})
}

func (dg *DepGraph) GetFailoverState(args cmdargs.RunArgs) (*states.FailoverState, error) {
	return dg.failoverState.get(func() (*states.FailoverState, error) {
		return states.NewFailoverState(args, dg)
	})
}

func (dg *DepGraph) GetStoppingState(args cmdargs.RunArgs) (*states.StoppingState, error) {
	return dg.stoppingState.get(func() (*states.StoppingState, error) {
		return states.NewStoppingState(args, dg)
	})
}

func (dg *DepGraph) GetRunner() (run.Runner, error) {
	return dg.stateRunner.get(func() (*run.LoopRunner, error) {
		logger, err := dg.GetLogger()
		if err != nil {
			return nil, fmt.Errorf("get logger: %w", err)
		}
		return run.NewLoopRunner(logger), nil
	})
}
