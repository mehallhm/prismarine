package environment

import "context"

type Instance interface {
	Type() string

	Config() *Configuration

	//Events() *events.Bus

	// Exists determines whether the Instance exists in the environment
	Exists() (bool, error)

	IsRunning(ctx context.Context) (bool, error)

	Prelude(ctx context.Context) error

	Start(ctx context.Context) error

	Stop(ctx context.Context) error

	Create() error
}
