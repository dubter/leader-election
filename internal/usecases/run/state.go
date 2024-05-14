package run

import (
	"context"
)

type AutomataState interface {
	Run(ctx context.Context) (AutomataState, error)
	String() string
}
