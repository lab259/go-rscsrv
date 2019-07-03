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
