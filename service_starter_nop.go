package rscsrv

type NopStarterReporter struct{}

func (*NopStarterReporter) BeforeBegin(service Service) {}

func (*NopStarterReporter) BeforeLoadConfiguration(service Configurable) {}

func (*NopStarterReporter) AfterLoadConfiguration(service Configurable, conf interface{}, err error) {}

func (*NopStarterReporter) BeforeApplyConfiguration(service Configurable) {}

func (*NopStarterReporter) AfterApplyConfiguration(service Configurable, conf interface{}, err error) {
}

func (*NopStarterReporter) BeforeStart(service Service) {}

func (*NopStarterReporter) AfterStart(service Service, err error) {}

func (*NopStarterReporter) BeforeStop(service Service) {}

func (*NopStarterReporter) AfterStop(service Service, err error) {}
