package rscsrv

type ServiceStarterReporter interface {
	BeforeBegin(service Service)

	BeforeLoadConfiguration(service Service)
	AfterLoadConfiguration(service Service, conf interface{}, err error)

	BeforeApplyConfiguration(service Service)
	AfterApplyConfiguration(service Service, conf interface{}, err error)

	BeforeStart(service Service)
	AfterStart(service Service, err error)

	BeforeStop(service Service)
	AfterStop(service Service, err error)
}

type NopServiceReporter struct{}

func (*NopServiceReporter) BeforeBegin(service Service) {}

func (*NopServiceReporter) BeforeLoadConfiguration(service Service) {}

func (*NopServiceReporter) AfterLoadConfiguration(service Service, conf interface{}, err error) {}

func (*NopServiceReporter) BeforeApplyConfiguration(service Service) {}

func (*NopServiceReporter) AfterApplyConfiguration(service Service, conf interface{}, err error) {}

func (*NopServiceReporter) BeforeStart(service Service) {}

func (*NopServiceReporter) AfterStart(service Service, err error) {}

func (*NopServiceReporter) BeforeStop(service Service) {}

func (*NopServiceReporter) AfterStop(service Service, err error) {}

type ServiceStarter struct {
	services []Service
	started  []Service
	reporter ServiceStarterReporter
}

func NewServiceStarter(services []Service, reporter ServiceStarterReporter) *ServiceStarter {
	return &ServiceStarter{
		services: services,
		started:  make([]Service, 0, len(services)),
		reporter: reporter,
	}
}

func (engineStarter *ServiceStarter) Start() error {
	for _, srv := range engineStarter.services {
		engineStarter.reporter.BeforeBegin(srv)
		engineStarter.reporter.BeforeLoadConfiguration(srv)
		conf, err := srv.LoadConfiguration()
		engineStarter.reporter.AfterLoadConfiguration(srv, conf, err)
		if err != nil {
			return err
		}

		engineStarter.reporter.BeforeApplyConfiguration(srv)
		err = srv.ApplyConfiguration(conf)
		if err != nil {
			return err
		}
		engineStarter.reporter.AfterApplyConfiguration(srv, conf, err)

		engineStarter.reporter.BeforeStart(srv)
		err = srv.Start()
		engineStarter.reporter.AfterStart(srv, err)
		if err != nil {
			return err
		}

		engineStarter.started = append([]Service{srv}, engineStarter.started...)
	}

	return nil
}

func (engineStarter *ServiceStarter) Stop(keepGoing bool) error {
	for len(engineStarter.started) > 0 {
		srv := engineStarter.started[0]
		engineStarter.reporter.BeforeBegin(srv)

		engineStarter.reporter.BeforeStop(srv)
		err := srv.Stop()
		engineStarter.reporter.AfterStop(srv, err)
		if err != nil && !keepGoing {
			return err
		}

		engineStarter.started = engineStarter.started[1:]
	}
	return nil
}
