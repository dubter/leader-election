package states

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run"
	"github.com/go-zookeeper/zk"
)

func NewFailoverState(args cmdargs.RunArgs, dg DepGraph) (*FailoverState, error) {
	logger, err := dg.GetLogger()
	if err != nil {
		return nil, fmt.Errorf("get logger: %w", err)
	}

	return &FailoverState{
		logger:         logger.With("subsystem", "FailoverState"),
		zkServers:      args.ZkServers,
		sessionTimeout: args.SessionTimeout,
		dg:             dg,
		args:           args,
	}, nil
}

type FailoverState struct {
	logger         *slog.Logger
	zkServers      []string
	sessionTimeout time.Duration
	args           cmdargs.RunArgs
	dg             DepGraph
}

func (s *FailoverState) String() string {
	return "FailoverState"
}

func (s *FailoverState) connectWithExponentialBackoff(ctx context.Context, resChan chan result) {
	const maxRetries = 5
	const initialDelay = time.Second
	delay := initialDelay

	for attempt := 0; attempt < maxRetries; attempt++ {
		conn, _, err := zk.Connect(s.zkServers, s.sessionTimeout)
		if err == nil {
			resChan <- result{
				conn: conn,
			}
			return
		}

		s.logger.LogAttrs(ctx, slog.LevelError, fmt.Sprintf("Error connecting to Zookeeper on attempt %d", attempt+1), slog.String("msg", err.Error()))

		// Increase delay exponentially
		delay *= 2
		time.Sleep(delay)
	}

	resChan <- result{
		err: fmt.Errorf("unable to connect to Zookeeper after %d attempts", maxRetries),
	}
}

func (s *FailoverState) Run(ctx context.Context) (run.AutomataState, error) {
	resChan := make(chan result)
	go s.connectWithExponentialBackoff(ctx, resChan)

	select {
	case <-ctx.Done():
		return s.dg.GetStoppingState(s.args)

	case res := <-resChan:
		if res.err != nil {
			s.logger.LogAttrs(ctx, slog.LevelError, "can not connect to zookeeper", slog.String("msg", res.err.Error()))
			return s.dg.GetStoppingState(s.args)
		}

		attempterState, err := s.dg.GetAttempterState(s.args)
		if err != nil {
			return nil, err
		}

		return attempterState.WithConnection(res.conn), nil
	}
}
