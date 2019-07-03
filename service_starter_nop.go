package rscsrv

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
