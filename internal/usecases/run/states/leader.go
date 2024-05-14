package states

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/usecases/run/extra"
	"github.com/go-zookeeper/zk"
)

const (
	layout = "2006-01-02_15-04-05"
)

func NewLeaderState(args cmdargs.RunArgs, dg DepGraph) (*LeaderState, error) {
	logger, err := dg.GetLogger()
	if err != nil {
		return nil, fmt.Errorf("get logger: %w", err)
	}

	return &LeaderState{
		logger:          logger.With("subsystem", "LeaderState"),
		fileDir:         args.FileDir,
		storageCapacity: args.StorageCapacity,
		zkEphemeralPath: args.ZKEphemeralPath,
		dg:              dg,
		ticker:          extra.NewTicker(args.LeaderTimeout),
	}, nil
}

type LeaderState struct {
	logger          *slog.Logger
	ticker          extra.Ticker
	fileDir         string
	storageCapacity int
	zkEphemeralPath string
	conn            *zk.Conn
	dg              DepGraph
	args            cmdargs.RunArgs
}

func (s *LeaderState) WithConnection(conn *zk.Conn) *LeaderState {
	s.conn = conn
	return s
}

func (s *LeaderState) Stop() {
	s.ticker.Stop()
}

func (s *LeaderState) String() string {
	return "LeaderState"
}

func (s *LeaderState) Run(ctx context.Context) (run.AutomataState, error) {
	if s.conn == nil {
		return s.dg.GetFailoverState(s.args)
	}

	failChan := make(chan error)
	go func() {
		for range s.ticker.Chan() {
			if s.conn.State() != zk.StateHasSession {
				failChan <- nil
				return
			}

			fileCount, err := countFiles(s.fileDir)
			if err != nil {
				failChan <- err
				return
			}

			if fileCount >= s.storageCapacity {
				err := cleanDirectory(s.fileDir)
				if err != nil {
					failChan <- err
					return
				}
			}

			fileName := fmt.Sprintf("%s_%s.txt", hostname(), time.Now().Format(layout))
			filePath := filepath.Join(s.fileDir, fileName)
			_, err = os.Create(filePath)
			if err != nil {
				failChan <- err
				break
			}

			s.logger.LogAttrs(ctx, slog.LevelInfo, "created new file",
				slog.String("filePath", filePath))
		}
	}()

	select {
	case <-ctx.Done():
		return s.dg.GetStoppingState(s.args)

	case err := <-failChan:
		if err != nil {
			s.logger.LogAttrs(ctx, slog.LevelError, fmt.Sprintf("Error from leader file system in directory %s: %v", s.fileDir, err))
			return s.dg.GetStoppingState(s.args)
		}

		return s.dg.GetFailoverState(s.args)
	}
}

func countFiles(dirPath string) (int, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return 0, err
	}
	return len(files), nil
}

func cleanDirectory(dirPath string) error {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	for _, file := range files {
		err := os.Remove(filepath.Join(dirPath, file.Name()))
		if err != nil {
			return err
		}
	}
	return nil
}

func hostname() string {
	host, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return host
}
