package rscsrv

type ServiceStarterReporter interface {
	BeforeBegin(service Service)

	BeforeLoadConfiguration(service Configurable)
	AfterLoadConfiguration(service Configurable, conf interface{}, err error)

	BeforeApplyConfiguration(service Configurable)
	AfterApplyConfiguration(service Configurable, conf interface{}, err error)

	BeforeStart(service Service)
	AfterStart(service Service, err error)

	BeforeStop(service Service)
	AfterStop(service Service, err error)
}
