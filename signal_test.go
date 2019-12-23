package rscsrv

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"go.uber.org/atomic"
)

type MockService struct {
	name                    string
	errLoadingConfiguration error
	errApplyConfiguration   error
	errRestart              error
	started                 atomic.Bool
	errStart                error
	startDuration           time.Duration
	stopped                 atomic.Bool
	stopDuration            time.Duration
	errStop                 error
}

func (service *MockService) Name() string {
	if service.name != "" {
		return service.name
	}
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
	service.started.Store(true)
	service.stopped.Store(false)
	time.Sleep(service.startDuration)
	return service.errStart
}

func (service *MockService) Stop() error {
	service.started.Store(false)
	service.stopped.Store(true)
	time.Sleep(service.stopDuration)
	return service.errStop
}

var _ = Describe("Signal", func() {
	It("should stop services when receiving a signal", func() {
		// Start service1 and service2;
		// Start process done;
		// The signal arrives and the services get stopped.
		service1 := &MockService{}
		service2 := &MockService{}
		starter := SignalStarter(NewServiceStarter(&NopStarterReporter{}, service1, service2))
		Expect(starter.Start()).To(Succeed())
		signalStarter := starter.(*signalServiceStarter)
		signalStarter.signals <- os.Interrupt
		time.Sleep(time.Millisecond * 25)
		Expect(service1.stopped.Load()).To(BeTrue())
		Expect(service2.stopped.Load()).To(BeTrue())
	})

	It("should cancel starting process when receiving a signal", func(done Done) {
		// Start service1 taking 50 milliseconds;
		// Concurrently, after 25 milliseconds, cancels the starting process;
		service1 := &MockService{
			name:          "service1",
			startDuration: time.Millisecond * 50,
		}
		service2 := &MockService{
			name: "service2",
		}
		starter := SignalStarter(NewServiceStarter(&NopStarterReporter{}, service1, service2))
		go func() {
			defer GinkgoRecover()

			signalStarter := starter.(*signalServiceStarter)
			time.Sleep(time.Millisecond * 25)
			signalStarter.signals <- os.Interrupt
			time.Sleep(time.Millisecond * 50)
			Expect(service1.stopped.Load()).To(BeTrue())
			Expect(service2.stopped.Load()).To(BeFalse())

			close(done)
		}()
		Expect(starter.Start()).To(Equal(context.Canceled))
	}, 1)
})
