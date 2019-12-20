package rscsrv_test

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/lab259/go-rscsrv"
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

func (reporter *countEngineReporter) BeforeLoadConfiguration(service rscsrv.Configurable) {
	reporter.countBeforeLoadConfiguration++
}

func (reporter *countEngineReporter) AfterLoadConfiguration(service rscsrv.Configurable, conf interface{}, err error) {
	reporter.countAfterLoadConfiguration++
}

func (reporter *countEngineReporter) BeforeApplyConfiguration(service rscsrv.Configurable) {
	reporter.countBeforeApplyConfiguration++
}

func (reporter *countEngineReporter) AfterApplyConfiguration(service rscsrv.Configurable, conf interface{}, err error) {
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
	started                 bool
	errStart                error
	startDuration           time.Duration
	stopped                 bool
	errStop                 error
}

type MockServiceWithCancellation struct {
	MockService
	ctx    context.Context
	cancel context.CancelFunc
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
	time.Sleep(service.startDuration)
	service.started = true
	service.stopped = false
	return service.errStart
}

func (service *MockService) Stop() error {
	service.started = false
	service.stopped = true
	return service.errStop
}

func (service *MockServiceWithCancellation) StartWithContext(ctx context.Context) error {
	service.ctx, service.cancel = context.WithCancel(ctx)

	select {
	case <-time.After(service.startDuration):
		return nil
	case <-service.ctx.Done():
		return service.ctx.Err()
	}
}

func (service *MockServiceWithCancellation) Stop() error {
	service.cancel()
	return nil
}

var _ = Describe("ServiceStarter", func() {
	It("should start all service", func() {
		reporter := &countEngineReporter{}
		engineStarter := rscsrv.NewServiceStarter(reporter, &MockService{})
		err := engineStarter.Start()
		Expect(err).ToNot(HaveOccurred())
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
		engineStarter := rscsrv.NewServiceStarter(
			reporter,
			&MockService{},
		)
		err := engineStarter.Start()
		Expect(err).ToNot(HaveOccurred())
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
		engineStarter := rscsrv.NewServiceStarter(
			reporter,
			&MockService{
				errStop: errors.New("stopping error"),
			},
		)
		err := engineStarter.Start()
		Expect(err).ToNot(HaveOccurred())
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
		engineStarter := rscsrv.NewServiceStarter(
			reporter,
			&MockService{
				errStop: errors.New("stopping error"),
			},
		)
		err := engineStarter.Start()
		Expect(err).ToNot(HaveOccurred())
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
		engineStarter := rscsrv.NewServiceStarter(
			reporter,
			&MockService{
				errLoadingConfiguration: errors.New("loading configuration error"),
			},
		)
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
		engineStarter := rscsrv.NewServiceStarter(
			reporter,
			&MockService{
				errApplyConfiguration: errors.New("applying configuration error"),
			},
		)
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

	It("should cancel the starting a service", func() {
		// Start a service that takes 1 seconds to boot;
		// Stops the service starter after 50 milliseconds;
		// The service start should be cancelled;
		reporter := &countEngineReporter{}
		engineStarter := rscsrv.NewServiceStarter(
			reporter,
			&MockServiceWithCancellation{
				MockService: MockService{
					startDuration: time.Second,
				},
			},
		)
		go func() {
			time.Sleep(time.Millisecond * 50)
			Expect(engineStarter.Stop(true)).To(Succeed())
		}()
		err := engineStarter.Start()
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(context.Canceled))
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

	It("should stop services started before a cancellation", func() {
		// Start services 1 and 2 (they starts immediately);
		// While service 3 is starting (takes a second), Stop the service starter;
		// Both service 1 and 2 should be stopped after cancelling the service 3 starting process;
		reporter := &countEngineReporter{}
		engineStarter := rscsrv.NewServiceStarter(
			reporter,
			&MockService{},
			&MockService{},
			&MockServiceWithCancellation{
				MockService: MockService{
					startDuration: time.Second,
				},
			},
		)
		go func() {
			defer GinkgoRecover()

			time.Sleep(time.Millisecond * 50)
			Expect(engineStarter.Stop(true)).To(Succeed())

			Expect(reporter.countBeforeBegin).To(Equal(3))
			Expect(reporter.countBeforeLoadConfiguration).To(Equal(3))
			Expect(reporter.countAfterLoadConfiguration).To(Equal(3))
			Expect(reporter.countBeforeApplyConfiguration).To(Equal(3))
			Expect(reporter.countAfterApplyConfiguration).To(Equal(3))
			Expect(reporter.countBeforeStart).To(Equal(3))
			Expect(reporter.countAfterStart).To(Equal(3))
			Expect(reporter.countBeforeStop).To(Equal(2))
			Expect(reporter.countAfterStop).To(Equal(2))
		}()
		err := engineStarter.Start()
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(context.Canceled))
	})

	It("should cancel the starting process after starting a service that has a start not cancellable", func() {
		reporter := &countEngineReporter{}

		service1 := &MockService{}
		service2 := &MockServiceWithCancellation{
			MockService: MockService{
				startDuration: time.Millisecond * 100,
			},
		}

		engineStarter := rscsrv.NewServiceStarter(reporter, service1, service2)

		go func() {
			defer GinkgoRecover()

			time.Sleep(time.Millisecond * 50)
			Expect(engineStarter.Stop(true)).To(Succeed())

			Expect(reporter.countBeforeBegin).To(Equal(2))
			Expect(reporter.countBeforeLoadConfiguration).To(Equal(2))
			Expect(reporter.countAfterLoadConfiguration).To(Equal(2))
			Expect(reporter.countBeforeApplyConfiguration).To(Equal(2))
			Expect(reporter.countAfterApplyConfiguration).To(Equal(2))
			Expect(reporter.countBeforeStart).To(Equal(2))
			Expect(reporter.countAfterStart).To(Equal(2))
			Expect(reporter.countBeforeStop).To(Equal(1))
			Expect(reporter.countAfterStop).To(Equal(1))
			Expect(service1.started).To(BeFalse())
			Expect(service2.started).To(BeFalse())
			Expect(service1.stopped).To(BeTrue())
			Expect(service2.stopped).To(BeFalse()) // Yes, its stop was never called...
		}()
		err := engineStarter.Start()
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(context.Canceled))
	})

	It("should cancel the starting process after a non cancellable service", func() {
		reporter := &countEngineReporter{}

		service1 := &MockService{
			startDuration: time.Millisecond * 50,
		}
		service2 := &MockServiceWithCancellation{
			MockService: MockService{
				startDuration: time.Millisecond * 100,
			},
		}

		engineStarter := rscsrv.NewServiceStarter(reporter, service1, service2)

		go func() {
			defer GinkgoRecover()

			time.Sleep(time.Millisecond * 25)
			Expect(engineStarter.Stop(true)).To(Succeed())

			Expect(reporter.countBeforeBegin).To(Equal(1))
			Expect(reporter.countBeforeLoadConfiguration).To(Equal(1))
			Expect(reporter.countAfterLoadConfiguration).To(Equal(1))
			Expect(reporter.countBeforeApplyConfiguration).To(Equal(1))
			Expect(reporter.countAfterApplyConfiguration).To(Equal(1))
			Expect(reporter.countBeforeStart).To(Equal(1))
			Expect(reporter.countAfterStart).To(Equal(1))
			Expect(reporter.countBeforeStop).To(Equal(1))
			Expect(reporter.countAfterStop).To(Equal(1))
			Expect(service1.started).To(BeFalse())
			Expect(service2.started).To(BeFalse())
			Expect(service1.stopped).To(BeTrue())
			Expect(service2.stopped).To(BeFalse()) // Yes, its stop was never called...
		}()
		err := engineStarter.Start()
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(context.Canceled))
	})

	It("should cancel the starting process after cancelling the start at the last non cancellable service", func() {
		// Start service1 immediately;
		// Start service2 taking 100 milliseconds (non cancellable);
		// Cancel after 50 milliseconds (while service2 still starting);
		// Since service2 start not cancellable, waits it to finish before get cancelled.
		reporter := &countEngineReporter{}

		service1 := &MockService{}
		service2 := &MockService{
			startDuration: time.Millisecond * 100,
		}

		engineStarter := rscsrv.NewServiceStarter(reporter, service1, service2)

		go func() {
			defer GinkgoRecover()

			time.Sleep(time.Millisecond * 50)
			Expect(engineStarter.Stop(true)).To(Succeed())

			Expect(reporter.countBeforeBegin).To(Equal(2))
			Expect(reporter.countBeforeLoadConfiguration).To(Equal(2))
			Expect(reporter.countAfterLoadConfiguration).To(Equal(2))
			Expect(reporter.countBeforeApplyConfiguration).To(Equal(2))
			Expect(reporter.countAfterApplyConfiguration).To(Equal(2))
			Expect(reporter.countBeforeStart).To(Equal(2))
			Expect(reporter.countAfterStart).To(Equal(2))
			Expect(reporter.countBeforeStop).To(Equal(2))
			Expect(reporter.countAfterStop).To(Equal(2))
			Expect(service1.started).To(BeFalse())
			Expect(service2.started).To(BeFalse())
			Expect(service1.stopped).To(BeTrue())
			Expect(service2.stopped).To(BeTrue())
		}()
		err := engineStarter.Start()
		Expect(err).To(HaveOccurred())
		Expect(err).To(Equal(context.Canceled))
	})
})
