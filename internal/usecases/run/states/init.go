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

type DepGraph interface {
	GetLogger() (*slog.Logger, error)
	GetAttempterState(args cmdargs.RunArgs) (*AttempterState, error)
	GetLeaderState(args cmdargs.RunArgs) (*LeaderState, error)
	GetFailoverState(args cmdargs.RunArgs) (*FailoverState, error)
	GetStoppingState(args cmdargs.RunArgs) (*StoppingState, error)
}

func NewInitState(args cmdargs.RunArgs, dg DepGraph) (*InitState, error) {
	logger, err := dg.GetLogger()
	if err != nil {
		return nil, fmt.Errorf("get logger: %w", err)
	}

	return &InitState{
		logger:         logger.With("subsystem", "InitState"),
		zkServers:      args.ZkServers,
		sessionTimeout: args.SessionTimeout,
		dg:             dg,
		args:           args,
	}, nil
}

type InitState struct {
	logger         *slog.Logger
	zkServers      []string
	sessionTimeout time.Duration
	args           cmdargs.RunArgs
	dg             DepGraph
}

func (s *InitState) String() string {
	return "InitState"
}

type result struct {
	conn *zk.Conn
	err  error
}

func (s *InitState) Run(ctx context.Context) (run.AutomataState, error) {
	resChan := make(chan result)
	go func() {
		conn, _, err := zk.Connect(s.zkServers, s.sessionTimeout)
		resChan <- result{
			conn: conn,
			err:  err,
		}
	}()

	select {
	case <-ctx.Done():
		return s.dg.GetStoppingState(s.args)

	case res := <-resChan:
		if res.err != nil {
			s.logger.LogAttrs(ctx, slog.LevelError, "can not connect to zookeeper", slog.String("msg", res.err.Error()))

			return s.dg.GetFailoverState(s.args)
		}

		attempterState, err := s.dg.GetAttempterState(s.args)
		if err != nil {
			return nil, err
		}

		return attempterState.WithConnection(res.conn), nil
	}
}
