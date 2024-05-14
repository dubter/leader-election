package commands

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/commands/cmdargs"
	"github.com/central-university-dev/2024-spring-go-course-lesson8-leader-election/internal/depgraph"
	"github.com/spf13/cobra"
)

var (
	defaultZKServers        = []string{"zoo1:2181", "zoo2:2181", "zoo3:2181"} // Default Zookeeper servers
	defaultLeaderTimeout    = time.Second * 10                                // Default Leader Timeout
	defaultAttempterTimeout = time.Second * 10                                // Default Attempter Timeout
	defaultSessionTimeout   = time.Second * 2                                 // Default Session Timeout
	defaultFileDir          = "/tmp/election"                                 // Default File Directory
	defaultStorageCapacity  = 40                                              // Default Storage Capacity
	defaultZKEphemeralPath  = "/app_ephemeral"
)

func InitRunCommand() (cobra.Command, error) {
	cmdArgs := cmdargs.RunArgs{}
	cmd := cobra.Command{
		Use:   "run",
		Short: "Starts a leader election node",
		Long: `This command starts the leader election node that connects to zookeeper
		and starts to try to acquire leadership by creation of ephemeral node`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			dg := depgraph.New()
			logger, err := dg.GetLogger()
			if err != nil {
				return fmt.Errorf("get logger: %w", err)
			}
			logger.Info("args received",
				slog.String("zk-servers", strings.Join(cmdArgs.ZkServers, ", ")),
				slog.Duration("leader-timeout", cmdArgs.LeaderTimeout),
				slog.Duration("attempter-timeout", cmdArgs.AttempterTimeout),
				slog.Duration("session-timeout", cmdArgs.SessionTimeout),
				slog.String("file-dir", cmdArgs.FileDir),
				slog.Int("storage-capacity", cmdArgs.StorageCapacity),
			)

			_, err = os.ReadDir(cmdArgs.FileDir)
			if err != nil {
				logger.Error(fmt.Sprintf("'file-dir' %s can not be read: %v", cmdArgs.FileDir, err))
				os.Exit(1)
			}

			runner, err := dg.GetRunner()
			if err != nil {
				return fmt.Errorf("get runner: %w", err)
			}

			firstState, err := dg.GetInitState(cmdArgs)
			if err != nil {
				return fmt.Errorf("get first state: %w", err)
			}
			err = runner.Run(cmd.Context(), firstState)
			if err != nil {
				return fmt.Errorf("run states: %w", err)
			}
			return nil
		},
	}

	// priority: flag -> env -> default
	cmd.Flags().StringSliceVarP(&(cmdArgs.ZkServers), "zk-servers", "s", []string{}, "Set the zookeeper servers.")
	cmd.Flags().DurationVarP(&(cmdArgs.LeaderTimeout), "leader-timeout", "l", 0, "Set the frequency at which the leader writes the file to disk.")
	cmd.Flags().DurationVarP(&(cmdArgs.AttempterTimeout), "attempter-timeout", "a", 0, "Set the frequency with which an attempter tries to become a leader.")
	cmd.Flags().DurationVarP(&(cmdArgs.SessionTimeout), "session-timeout", "t", 0, "Set the session timeout with zookeeper.")
	cmd.Flags().StringVarP(&(cmdArgs.FileDir), "file-dir", "f", "", "Set the directory to leader writing files.")
	cmd.Flags().IntVarP(&(cmdArgs.StorageCapacity), "storage-capacity", "c", 0, "Maximum count of files in 'file-dir'.")
	cmd.Flags().StringVarP(&(cmdArgs.FileDir), "zk-path", "p", "", "Set the ephemeral directory in zookeeper for leader election.")

	if len(cmdArgs.ZkServers) == 0 {
		cmdArgs.ZkServers = getEnvStrings("ZK_SERVERS", defaultZKServers)
	}

	if cmdArgs.LeaderTimeout == 0 {
		cmdArgs.LeaderTimeout = getEnvDuration("LEADER_TIMEOUT", defaultLeaderTimeout)
	}

	if cmdArgs.AttempterTimeout == 0 {
		cmdArgs.AttempterTimeout = getEnvDuration("ATTEMPTER_TIMEOUT", defaultAttempterTimeout)
	}

	if cmdArgs.SessionTimeout == 0 {
		cmdArgs.SessionTimeout = getEnvDuration("SESSION_TIMEOUT", defaultSessionTimeout)
	}

	if cmdArgs.FileDir == "" {
		cmdArgs.FileDir = getEnvString("FILE_DIR", defaultFileDir)
	}

	if cmdArgs.StorageCapacity == 0 {
		cmdArgs.StorageCapacity = getEnvInt("STORAGE_CAPACITY", defaultStorageCapacity)
	}

	if cmdArgs.ZKEphemeralPath == "" {
		cmdArgs.ZKEphemeralPath = getEnvString("ZK_PATH", defaultZKEphemeralPath)
	}

	return cmd, nil
}

func getEnvString(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvStrings(key string, defaultValue []string) []string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return strings.Split(value, ",")
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := time.ParseDuration(valueStr)
	if err != nil {
		fmt.Printf("error parsing duration for %s: %v\n", key, err)
		return defaultValue
	}
	return value
}

func getEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		fmt.Printf("error parsing int for %s: %v\n", key, err)
		return defaultValue
	}
	return value
}
