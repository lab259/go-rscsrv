package rscsrv

import "errors"

var (
	// ErrWrongConfigurationInformed is the error returned when a configuration
	// object is loaded but it is not valid.
	ErrWrongConfigurationInformed = errors.New("wrong configuration informed")

	// ErrServiceNotRunning is the error returned when a non started server is
	// stopped or restarted.
	ErrServiceNotRunning = errors.New("service not running")
)

// Nameable abstracts the precense of a the name in a possible `Startable` or
// `Configurable`.
type Service interface {
	// Name identifies the service.
	Name() string
}

// Startable is an abtraction for implementing parts that can be started,
// restarted and stopped.
type Startable interface {
	Service
	startable
}

type startable interface {
	// Restarts the service. If successful nil will be returned, otherwise the
	// error.
	Restart() error

	// Start starts the service. If successful nil will be returned, otherwise
	// the error.
	Start() error

	// Stop stops the service. If successful nil will be returned, otherwise the
	// error.
	Stop() error
}

type Configurable interface {
	Service
	configurable
}

type configurable interface {
	// Loads the configuration. If successful nil will be returned, otherwise
	// the error.
	LoadConfiguration() (interface{}, error)

	// Applies a given configuration object to the service. If successful nil
	// will be returned, otherwise the error.
	ApplyConfiguration(interface{}) error
}
