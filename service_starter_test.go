package rscsrv_test

import (
	"errors"

	rscsrv "github.com/lab259/go-rscsrv"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type countEngineReporter struct {
	countBeforeBegin              int
	countBeforeLoadConfiguration  int
	countAfterLoadConfiguration   int
	countBeforeApplyConfiguration int
	countAfterApplyConfiguration  int
	countBeforeStart              int
	countAfterStart               int
	countBeforeStop               int
	countAfterStop                int
}

func (reporter *countEngineReporter) BeforeBegin(service rscsrv.Service) {
	reporter.countBeforeBegin++
}

func (reporter *countEngineReporter) BeforeLoadConfiguration(service rscsrv.Service) {
	reporter.countBeforeLoadConfiguration++
}

func (reporter *countEngineReporter) AfterLoadConfiguration(service rscsrv.Service, conf interface{}, err error) {
	reporter.countAfterLoadConfiguration++
}

func (reporter *countEngineReporter) BeforeApplyConfiguration(service rscsrv.Service) {
	reporter.countBeforeApplyConfiguration++
}

func (reporter *countEngineReporter) AfterApplyConfiguration(service rscsrv.Service, conf interface{}, err error) {
	reporter.countAfterApplyConfiguration++
}

func (reporter *countEngineReporter) BeforeStart(service rscsrv.Service) {
	reporter.countBeforeStart++
}

func (reporter *countEngineReporter) AfterStart(service rscsrv.Service, err error) {
	reporter.countAfterStart++
}

func (reporter *countEngineReporter) BeforeStop(service rscsrv.Service) {
	reporter.countBeforeStop++
}

func (reporter *countEngineReporter) AfterStop(service rscsrv.Service, err error) {
	reporter.countAfterStop++
}

type MockService struct {
	errLoadingConfiguration error
	errApplyConfiguration   error
	errRestart              error
	errStart                error
	errStop                 error
}

func (service *MockService) Name() string {
	return "mock-service"
}

func (service *MockService) LoadConfiguration() (interface{}, error) {
	return nil, service.errLoadingConfiguration
}

func (service *MockService) ApplyConfiguration(interface{}) error {
	return service.errApplyConfiguration
}

func (service *MockService) Restart() error {
	return service.errRestart
}

func (service *MockService) Start() error {
	return service.errStart
}

func (service *MockService) Stop() error {
	return service.errStop
}

var _ = Describe("ServiceStarter", func() {
	It("should start all service", func() {
		reporter := &countEngineReporter{}
		engineStarter := rscsrv.NewServiceStarter([]rscsrv.Service{
			&MockService{},
		}, reporter)
		err := engineStarter.Start()
		Expect(err).To(BeNil())
		Expect(reporter.countBeforeBegin).To(Equal(1))
		Expect(reporter.countBeforeLoadConfiguration).To(Equal(1))
		Expect(reporter.countAfterLoadConfiguration).To(Equal(1))
		Expect(reporter.countBeforeApplyConfiguration).To(Equal(1))
		Expect(reporter.countAfterApplyConfiguration).To(Equal(1))
		Expect(reporter.countBeforeStart).To(Equal(1))
		Expect(reporter.countAfterStart).To(Equal(1))
		Expect(reporter.countBeforeStop).To(Equal(0))
		Expect(reporter.countAfterStop).To(Equal(0))
	})

	It("should stop all service", func() {
		reporter := &countEngineReporter{}
		engineStarter := rscsrv.NewServiceStarter([]rscsrv.Service{
			&MockService{},
		}, reporter)
		err := engineStarter.Start()
		Expect(err).To(BeNil())
		Expect(reporter.countBeforeBegin).To(Equal(1))
		Expect(reporter.countBeforeLoadConfiguration).To(Equal(1))
		Expect(reporter.countAfterLoadConfiguration).To(Equal(1))
		Expect(reporter.countBeforeApplyConfiguration).To(Equal(1))
		Expect(reporter.countAfterApplyConfiguration).To(Equal(1))
		Expect(reporter.countBeforeStart).To(Equal(1))
		Expect(reporter.countAfterStart).To(Equal(1))
		Expect(reporter.countBeforeStop).To(Equal(0))
		Expect(reporter.countAfterStop).To(Equal(0))

		Expect(engineStarter.Stop(false)).To(Succeed())
		Expect(reporter.countBeforeStop).To(Equal(1))
		Expect(reporter.countAfterStop).To(Equal(1))
	})

	It("should fail stopping service and return error", func() {
		reporter := &countEngineReporter{}
		engineStarter := rscsrv.NewServiceStarter([]rscsrv.Service{
			&MockService{
				errStop: errors.New("stopping error"),
			},
		}, reporter)
		err := engineStarter.Start()
		Expect(err).To(BeNil())
		Expect(reporter.countBeforeBegin).To(Equal(1))
		Expect(reporter.countBeforeLoadConfiguration).To(Equal(1))
		Expect(reporter.countAfterLoadConfiguration).To(Equal(1))
		Expect(reporter.countBeforeApplyConfiguration).To(Equal(1))
		Expect(reporter.countAfterApplyConfiguration).To(Equal(1))
		Expect(reporter.countBeforeStart).To(Equal(1))
		Expect(reporter.countAfterStart).To(Equal(1))
		Expect(reporter.countBeforeStop).To(Equal(0))
		Expect(reporter.countAfterStop).To(Equal(0))

		err = engineStarter.Stop(false)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("stopping error"))
		Expect(reporter.countBeforeStop).To(Equal(1))
		Expect(reporter.countAfterStop).To(Equal(1))
	})

	It("should fail stopping service and keep going", func() {
		reporter := &countEngineReporter{}
		engineStarter := rscsrv.NewServiceStarter([]rscsrv.Service{
			&MockService{
				errStop: errors.New("stopping error"),
			},
		}, reporter)
		err := engineStarter.Start()
		Expect(err).To(BeNil())
		Expect(reporter.countBeforeBegin).To(Equal(1))
		Expect(reporter.countBeforeLoadConfiguration).To(Equal(1))
		Expect(reporter.countAfterLoadConfiguration).To(Equal(1))
		Expect(reporter.countBeforeApplyConfiguration).To(Equal(1))
		Expect(reporter.countAfterApplyConfiguration).To(Equal(1))
		Expect(reporter.countBeforeStart).To(Equal(1))
		Expect(reporter.countAfterStart).To(Equal(1))
		Expect(reporter.countBeforeStop).To(Equal(0))
		Expect(reporter.countAfterStop).To(Equal(0))

		err = engineStarter.Stop(true)
		Expect(err).ToNot(HaveOccurred())
		Expect(reporter.countBeforeStop).To(Equal(1))
		Expect(reporter.countAfterStop).To(Equal(1))
	})

	It("should fail loading configuration", func() {
		reporter := &countEngineReporter{}
		engineStarter := rscsrv.NewServiceStarter([]rscsrv.Service{
			&MockService{
				errLoadingConfiguration: errors.New("loading configuration error"),
			},
		}, reporter)
		err := engineStarter.Start()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("loading configuration error"))
		Expect(reporter.countBeforeBegin).To(Equal(1))
		Expect(reporter.countBeforeLoadConfiguration).To(Equal(1))
		Expect(reporter.countAfterLoadConfiguration).To(Equal(1))
		Expect(reporter.countBeforeApplyConfiguration).To(Equal(0))
		Expect(reporter.countAfterApplyConfiguration).To(Equal(0))
		Expect(reporter.countBeforeStart).To(Equal(0))
		Expect(reporter.countAfterStart).To(Equal(0))
		Expect(reporter.countBeforeStop).To(Equal(0))
		Expect(reporter.countAfterStop).To(Equal(0))
	})

	It("should fail applying configuration", func() {
		reporter := &countEngineReporter{}
		engineStarter := rscsrv.NewServiceStarter([]rscsrv.Service{
			&MockService{
				errApplyConfiguration: errors.New("applying configuration error"),
			},
		}, reporter)
		err := engineStarter.Start()
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("applying configuration error"))
		Expect(reporter.countBeforeBegin).To(Equal(1))
		Expect(reporter.countBeforeLoadConfiguration).To(Equal(1))
		Expect(reporter.countAfterLoadConfiguration).To(Equal(1))
		Expect(reporter.countBeforeApplyConfiguration).To(Equal(1))
		Expect(reporter.countAfterApplyConfiguration).To(Equal(0))
		Expect(reporter.countBeforeStart).To(Equal(0))
		Expect(reporter.countAfterStart).To(Equal(0))
		Expect(reporter.countBeforeStop).To(Equal(0))
		Expect(reporter.countAfterStop).To(Equal(0))
	})
})
