package states

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/extra"
	"github.com/go-zookeeper/zk"
)

func NewAttempterState(args cmdargs.RunArgs, dg DepGraph) (*AttempterState, error) {
	logger, err := dg.GetLogger()
	if err != nil {
		return nil, fmt.Errorf("get logger: %w", err)
	}

	return &AttempterState{
		logger:          logger.With("subsystem", "AttempterState"),
		zkEphemeralPath: args.ZKEphemeralPath,
		ticker:          extra.NewTicker(args.AttempterTimeout),
		args:            args,
		dg:              dg,
	}, nil
}

type AttempterState struct {
	logger          *slog.Logger
	zkEphemeralPath string
	conn            *zk.Conn
	ticker          extra.Ticker
	args            cmdargs.RunArgs
	dg              DepGraph
}

func (s *AttempterState) WithConnection(conn *zk.Conn) *AttempterState {
	s.conn = conn
	return s
}

func (s *AttempterState) Stop() {
	s.ticker.Stop()
}

func (s *AttempterState) String() string {
	return "AttempterState"
}

func (s *AttempterState) Run(ctx context.Context) (run.AutomataState, error) {
	if s.conn == nil {
		return s.dg.GetFailoverState(s.args)
	}

	resChan := make(chan error)
	go func() {
		for range s.ticker.Chan() {
			_, err := s.conn.Create(s.zkEphemeralPath, []byte(fmt.Sprintf("hostname: %s, time: %s", hostname(), time.Now())), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
			if !errors.Is(err, zk.ErrNodeExists) {
				resChan <- err
				return
			}

			s.logger.LogAttrs(ctx, slog.LevelInfo, "failed attempt: node is already exist",
				slog.String("time", time.Now().String()),
				slog.String("hostname", hostname()),
			)
		}
	}()

	select {
	case <-ctx.Done():
		return s.dg.GetStoppingState(s.args)

	case err := <-resChan:
		if err != nil {
			s.logger.LogAttrs(ctx, slog.LevelError, "error occurred", slog.String("msg", err.Error()))
			return s.dg.GetFailoverState(s.args)
		}

		leaderState, err := s.dg.GetLeaderState(s.args)
		if err != nil {
			return nil, err
		}

		return leaderState.WithConnection(s.conn), nil
	}
}
