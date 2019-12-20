package rscsrv

import (
	"context"
	"sync"
)

// ServiceStarter is an abtraction for service starter which is responsible
// for starting and stopping services.
//
// See Also
//
// `NewServiceStarter`, `DefaultServiceStarter`
type ServiceStarter interface {
	Start() error
	Stop(keepGoing bool) error
	Wait()
}

type serviceStarter struct {
	ctx        context.Context
	cancelFunc context.CancelFunc

	chMutex     sync.RWMutex
	startDoneCh chan bool
	stopDoneCh  chan bool
	services    []Service
	started     []Service
	reporter    ServiceStarterReporter
}

// DefaultServiceStarter returns a default ServiceStarter integrated
// with the ColorStarterReporter.
func DefaultServiceStarter(services ...Service) ServiceStarter {
	return NewServiceStarter(DefaultColorStarterReporter, services...)
}

// QuietServiceStarter returns a default ServiceStarter integrated
// with the NopStarterReporter.
func QuietServiceStarter(services ...Service) ServiceStarter {
	return NewServiceStarter(&NopStarterReporter{}, services...)
}

// NewServiceStarter returns a new instace of a `ServiceStarter`.
func NewServiceStarter(reporter ServiceStarterReporter, services ...Service) ServiceStarter {
	return &serviceStarter{
		services: services,
		started:  make([]Service, 0, len(services)),
		reporter: reporter,
	}
}

// Start will go through all provided services trying to load and/or start them.
func (engineStarter *serviceStarter) Start() error {
	engineStarter.chMutex.Lock()
	engineStarter.ctx, engineStarter.cancelFunc = context.WithCancel(context.Background())
	engineStarter.startDoneCh = make(chan bool)
	engineStarter.stopDoneCh = make(chan bool)
	engineStarter.chMutex.Unlock()
	defer func() {
		close(engineStarter.startDoneCh)
		engineStarter.cancelFunc()
	}()

	// Iterate through all services
	for _, srv := range engineStarter.services {
		engineStarter.chMutex.RLock()
		// Ensure the context is not cancelled:
		select {
		case <-engineStarter.ctx.Done():
			engineStarter.chMutex.RUnlock()
			// Broadcast the start is done...
			return engineStarter.ctx.Err()
		default:
			// Not cancelled ... everything must go on.
		}
		engineStarter.chMutex.RUnlock()

		engineStarter.reporter.BeforeBegin(srv)

		// If the service is Configurable, starts loading the configuration.
		if configurable, ok := srv.(Configurable); ok {
			engineStarter.reporter.BeforeLoadConfiguration(configurable)

			// Loads configuration
			conf, err := configurable.LoadConfiguration()
			engineStarter.reporter.AfterLoadConfiguration(configurable, conf, err)
			if err != nil {
				return err
			}
			engineStarter.reporter.BeforeApplyConfiguration(configurable)

			// Applies the configuration to the service.
			err = configurable.ApplyConfiguration(conf)
			if err != nil {
				return err
			}
			engineStarter.reporter.AfterApplyConfiguration(configurable, conf, err)
		}

		var err error
		switch startable := srv.(type) {
		case StartableWithContext:
			// If the service is Startable, tries to start the service.
			engineStarter.reporter.BeforeStart(srv)
			err = startable.StartWithContext(engineStarter.ctx)
		case Startable:
			// If the service is Startable, tries to start the service.
			engineStarter.reporter.BeforeStart(srv)
			err = startable.Start()
		default:
			continue
		}

		engineStarter.reporter.AfterStart(srv, err)
		if err != nil {
			return err
		}
		// Prepend the service to the list of started services.
		// The order is reverse to get the resources unallocated in the reverse order as they started.
		engineStarter.started = append([]Service{srv}, engineStarter.started...)
	}
	select {
	case <-engineStarter.ctx.Done():
		// Broadcast the start is done...
		return engineStarter.ctx.Err()
	default:
		// Not cancelled ... everything must go on.
	}
	return nil
}

// Stop will stop all started "startable" services.
func (engineStarter *serviceStarter) Stop(keepGoing bool) error {
	defer func() {
		engineStarter.chMutex.Lock()
		close(engineStarter.stopDoneCh)
		engineStarter.chMutex.Unlock()
	}()

	engineStarter.chMutex.RLock()
	if engineStarter.ctx != nil {
		engineStarter.cancelFunc()
		engineStarter.chMutex.RUnlock()
		<-engineStarter.startDoneCh
	} else {
		engineStarter.chMutex.RUnlock()
	}

	for len(engineStarter.started) > 0 {
		srv := engineStarter.started[0]
		// If the service is Stoppable, tries to stop the service.
		if stoppable, ok := srv.(Stoppable); ok {
			engineStarter.reporter.BeforeStop(srv)
			err := stoppable.Stop()
			engineStarter.reporter.AfterStop(srv, err)
			if err != nil && !keepGoing {
				return err
			}
		}

		// Removes the service from the list of started services.
		engineStarter.started = engineStarter.started[1:]
	}
	return nil
}

// Wait will keep waiting until the Stop be finished.
func (engineStarter *serviceStarter) Wait() {
	<-engineStarter.stopDoneCh
}
