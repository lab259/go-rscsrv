package rscsrv

import (
	"errors"
	"sync"
	"time"
)

var (
	// ErrMaxTriesExceeded is the error returned when the maximum number of
	// failed tries is reached.
	ErrMaxTriesExceeded = errors.New("too many tries")

	// ErrStartTimeout is the error returned when the `StartRetrier` reaches its
	// time limit (defined by `options.FailAfter`).
	ErrStartTimeout = errors.New("start timeout")

	// ErrUnknownPanic is the error returned the original service panics and
	// the error recovered is not an actual `error`.
	ErrUnknownPanic = errors.New("unknown panic")

	// ErrStartCancelled is the error returned the start process is cancelled
	// by a `Stop` call before it gets finished.
	ErrStartCancelled = errors.New("start cancelled by a stop")
)

type startRetrierReporter interface {
	// Error receives an error handles it. Then, the same error should be
	// returned. This is useful for logging and if some special logic error
	// handling.
	Error(err error) error
}

// NopStarRetrierReporter is a reporter with an empty implementation.
type NopStarRetrierReporter struct{}

// Error will just return the input param without doing anything.
func (*NopStarRetrierReporter) Error(err error) error {
	return err
}

// StartRetrierOptions defines the options for the `StartRetrier`.
type StartRetrierOptions struct {
	// MaxTries is the number of failures before giving up. 0 means it will be
	// trying to start eternally.
	MaxTries int

	// DelayBetweenTries is the time the `StartRetrier` will wait between tries.
	// If the duration is 0, the `StartRetrier` will use 5 second as a default
	// value.
	DelayBetweenTries time.Duration

	// Timeout configures for how long the system should be trying to start
	// a service before gives it up.
	Timeout time.Duration

	// Reporter configures a receiver for all start errors that might happen.
	Reporter startRetrierReporter
}

// StartRetrier is a helper that implements retrying start the service.
type StartRetrier struct {
	Service
	starting     bool
	startingM    sync.Mutex
	startingDone chan bool
	options      StartRetrierOptions
}

// NewStartRetrier configures
func NewStartRetrier(service Service, options StartRetrierOptions) *StartRetrier {
	if options.DelayBetweenTries == 0 {
		options.DelayBetweenTries = 5 * time.Second
	}
	return &StartRetrier{
		Service: service,
		options: options,
	}
}

// Retrier creates a 2nd order function to conveniently wrap `Service`s.
func Retrier(options StartRetrierOptions) func(Service) *StartRetrier {
	return func(s Service) *StartRetrier {
		return NewStartRetrier(s, options)
	}
}

func (retrier *StartRetrier) getStarting() (starting bool) {
	retrier.startingM.Lock()
	starting = retrier.starting
	retrier.startingM.Unlock()
	return
}

func (retrier *StartRetrier) setStarting(starting bool) {
	retrier.startingM.Lock()
	retrier.starting = starting
	retrier.startingM.Unlock()
}

// Start starts the provided service. If it fails, the retrier will try it again
// according to the options provided.
func (retrier *StartRetrier) Start() error {
	startable, ok := retrier.Service.(Startable)
	if !ok { // No need to do anything...
		return nil
	}

	retrier.startingDone = make(chan bool)
	defer func() {
		retrier.setStarting(false)
		close(retrier.startingDone)
	}()

	startedAt := time.Now()

	try := 0
	retrier.setStarting(true)
	for retrier.getStarting() {
		err := func() (err error) {
			defer func() {
				r := recover()
				if r == nil {
					return
				}
				if eee, ok := r.(error); ok {
					err = eee
				} else {
					err = ErrUnknownPanic
				}
				if retrier.options.Reporter != nil { // If we have a reporter, report the error
					err = retrier.options.Reporter.Error(err)
				}
			}()

			err = startable.Start()
			if retrier.options.Reporter != nil { // If we have a reporter, report the error
				err = retrier.options.Reporter.Error(err)
			}
			return
		}()
		if err == nil { // If there is no error, no need to retry anything. Done.
			return nil
		}

		try++

		// If there is a maximum number of tries defined and it was reached ...
		if retrier.options.MaxTries > 0 && retrier.options.MaxTries <= try {
			return ErrMaxTriesExceeded
		}

		// If there is a maximum number of time defined and it was reached ...
		if retrier.options.Timeout > 0 && retrier.options.Timeout <= time.Since(startedAt) {
			return ErrStartTimeout
		}

		// Waits a little bit
		time.Sleep(retrier.options.DelayBetweenTries)
	}
	return ErrStartCancelled
}

// Stop stops the provided service. If it still starting, the starting process
// is cancelled.
func (retrier *StartRetrier) Stop() error {
	startable, ok := retrier.Service.(Startable)
	if !ok { // No need to do anything...
		return nil
	}

	if retrier.getStarting() {
		retrier.setStarting(false)
		<-retrier.startingDone
		return nil
	}
	return startable.Stop()
}

// Restart restarts the provided service.
func (retrier *StartRetrier) Restart() error {
	err := retrier.Stop()
	if err != nil {
		return err
	}
	return retrier.Start()
}
