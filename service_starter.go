package rscsrv

type ServiceStarterReporter interface {
	BeforeBegin(service Service)

	BeforeLoadConfiguration(service Configurable)
	AfterLoadConfiguration(service Configurable, conf interface{}, err error)

	BeforeApplyConfiguration(service Configurable)
	AfterApplyConfiguration(service Configurable, conf interface{}, err error)

	BeforeStart(service Startable)
	AfterStart(service Startable, err error)

	BeforeStop(service Startable)
	AfterStop(service Startable, err error)
}

type NopStarterReporter struct{}

func (*NopStarterReporter) BeforeBegin(service Service) {}

func (*NopStarterReporter) BeforeLoadConfiguration(service Configurable) {}

func (*NopStarterReporter) AfterLoadConfiguration(service Configurable, conf interface{}, err error) {}

func (*NopStarterReporter) BeforeApplyConfiguration(service Configurable) {}

func (*NopStarterReporter) AfterApplyConfiguration(service Configurable, conf interface{}, err error) {
}

func (*NopStarterReporter) BeforeStart(service Startable) {}

func (*NopStarterReporter) AfterStart(service Startable, err error) {}

func (*NopStarterReporter) BeforeStop(service Startable) {}

func (*NopStarterReporter) AfterStop(service Startable, err error) {}

type serviceStarter struct {
	services []Service
	started  []Startable
	reporter ServiceStarterReporter
}

// NewServiceStarter returns a new instace of a `starter`.
func NewServiceStarter(reporter ServiceStarterReporter, services ...Service) *serviceStarter {
	return &serviceStarter{
		services: services,
		started:  make([]Startable, 0, len(services)),
		reporter: reporter,
	}
}

// Start will go through all provided services trying to load and/or start them.
func (engineStarter *serviceStarter) Start() error {
	// Iterate through all services
	for _, srv := range engineStarter.services {
		engineStarter.reporter.BeforeBegin(srv)

		// If the service is Loadible, starts loading the configuration.
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

		// If the service is startable, tries to start the service.
		if startable, ok := srv.(Startable); ok {
			engineStarter.reporter.BeforeStart(startable)
			err := startable.Start()
			engineStarter.reporter.AfterStart(startable, err)
			if err != nil {
				return err
			}

			// Prepend the service to the list of started services.
			// The order is reverse to get the resources unallocated in the reverse order as they started.
			engineStarter.started = append([]Startable{startable}, engineStarter.started...)
		}
	}

	return nil
}

// Stop will stop all started "startable" services.
func (engineStarter *serviceStarter) Stop(keepGoing bool) error {
	for len(engineStarter.started) > 0 {
		srv := engineStarter.started[0]
		engineStarter.reporter.BeforeBegin(srv)

		engineStarter.reporter.BeforeStop(srv)
		err := srv.Stop()
		engineStarter.reporter.AfterStop(srv, err)
		if err != nil && !keepGoing {
			return err
		}

		engineStarter.started = engineStarter.started[1:] // Removes the service from the list of started services.
	}
	return nil
}
