package rscsrv

// ServiceStarter is an abtraction for service starter which is responsible
// for starting and stopping services.
//
// See Also
//
// `NewServiceStarter`, `DefaultServiceStarter`
type ServiceStarter interface {
	Start() error
	Stop(keepGoing bool) error
}

type serviceStarter struct {
	services []Service
	started  []Service
	reporter ServiceStarterReporter
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
	// Iterate through all services
	for _, srv := range engineStarter.services {
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

		// If the service is Startable, tries to start the service.
		if startable, ok := srv.(Startable); ok {
			engineStarter.reporter.BeforeStart(startable)
			err := startable.Start()
			engineStarter.reporter.AfterStart(startable, err)
			if err != nil {
				return err
			}

			// Prepend the service to the list of started services.
			// The order is reverse to get the resources unallocated in the reverse order as they started.
			engineStarter.started = append([]Service{srv}, engineStarter.started...)
		}
	}

	return nil
}

// Stop will stop all started "startable" services.
func (engineStarter *serviceStarter) Stop(keepGoing bool) error {
	for len(engineStarter.started) > 0 {
		srv := engineStarter.started[0]
		engineStarter.reporter.BeforeBegin(srv)

		// If the service is Startable, tries to stop the service.
		if startable, ok := srv.(Startable); ok {
			engineStarter.reporter.BeforeStop(startable)
			err := startable.Stop()
			engineStarter.reporter.AfterStop(startable, err)
			if err != nil && !keepGoing {
				return err
			}
		}

		// Removes the service from the list of started services.
		engineStarter.started = engineStarter.started[1:]
	}
	return nil
}
