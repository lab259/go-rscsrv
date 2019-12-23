package rscsrv

import "errors"

import "context"

var (
	// ErrWrongConfigurationInformed is the error returned when a configuration
	// object is loaded but it is not valid.
	ErrWrongConfigurationInformed = errors.New("wrong configuration informed")

	// ErrServiceNotRunning is the error returned when a non started server is
	// stopped or restarted.
	ErrServiceNotRunning = errors.New("service not running")
)

// Service abstracts the precense of a the name in a possible `Startable` or
// `Configurable`.
type Service interface {
	// Name identifies the service.
	Name() string
}

// Startable is an abstraction for implementing parts that can be started,
// restarted and stopped.
type Startable interface {
	// Start starts the service. If successful nil will be returned, otherwise
	// the error.
	Start() error
}

// Stoppable is an abstraction with the Stop behavior.
type Stoppable interface {
	// Stop stops the service. If successful nil will be returned, otherwise the
	// error.
	Stop() error
}

// StartableWithContext is an abstraction for implementing services that
// can have its Start process cancelled.
type StartableWithContext interface {
	// StartWithContext starts the service with a context that can be cancelled.
	StartWithContext(ctx context.Context) error
}

// Configurable is an abstraction for implement loading and applying
// configuration.
type Configurable interface {
	// Loads the configuration. If successful nil will be returned, otherwise
	// the error.
	LoadConfiguration() (interface{}, error)

	// Applies a given configuration object to the service. If successful nil
	// will be returned, otherwise the error.
	ApplyConfiguration(interface{}) error
}
