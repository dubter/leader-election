package cmdargs

import (
	"time"
)

type RunArgs struct {
	ZkServers        []string
	LeaderTimeout    time.Duration
	SessionTimeout   time.Duration
	AttempterTimeout time.Duration
	FileDir          string
	StorageCapacity  int
	ZKEphemeralPath  string
}
