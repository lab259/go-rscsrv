package rscsrv_test

import (
	"errors"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/lab259/go-rscsrv"
)

type retrierMockReporter struct {
	count int
}

func (reporter *retrierMockReporter) ReportRetrier(retrier *rscsrv.StartRetrier, err error) error {
	if err != nil {
		reporter.count++
	}
	return err
}

type retrierMockService struct {
	startDelay time.Duration
	startCount int
	successAt  int
}

func (service *retrierMockService) Name() string {
	return "retrierMockService"
}

func (service *retrierMockService) Restart() error {
	return nil
}

func (service *retrierMockService) Start() error {
	service.startCount++
	time.Sleep(service.startDelay)
	if service.successAt == service.startCount {
		return nil
	}
	return errors.New("failed to start")
}

func (service *retrierMockService) Stop() error {
	return nil
}

type retrierMockPanicService struct {
	startDelay  time.Duration
	startCount  int
	successAt   int
	dataToPanic interface{}
}

func (service *retrierMockPanicService) Name() string {
	return "retrierMockPanicService"
}

func (service *retrierMockPanicService) Restart() error {
	return nil
}

func (service *retrierMockPanicService) Start() error {
	service.startCount++
	time.Sleep(service.startDelay)
	if service.successAt == service.startCount {
		return nil
	}
	p := service.dataToPanic
	if p == nil {
		p = errors.New("panicked error")
	}
	panic(p)
	return nil
}

func (service *retrierMockPanicService) Stop() error {
	return nil
}

type retrierMockServiceNonStartable struct {
}

func (service *retrierMockServiceNonStartable) Name() string {
	return "retrierMockServiceNonStartable"
}

var _ = Describe("StartRetrier", func() {
	It("should start a service with no failures", func() {
		// This tests just starts a service that starts at first, no problems.

		service := &retrierMockService{
			successAt: 1,
		}

		reporter := &retrierMockReporter{}

		engineStarter := rscsrv.NewServiceStarter(
			&rscsrv.NopStarterReporter{},
			rscsrv.NewStartRetrier(service, rscsrv.StartRetrierOptions{
				MaxTries:          5,
				DelayBetweenTries: time.Millisecond, // If not defined, it will wait 5 seconds default...
				Reporter:          reporter,
			}),
		)
		Expect(engineStarter.Start()).To(Succeed())
		Expect(service.startCount).To(Equal(1))
		Expect(reporter.count).To(Equal(0))
		Expect(engineStarter.Stop(true)).To(Succeed())
	})

	It("should cancel starting when asked to stop", func() {
		// The retrier have a 5 maximum tries.
		// The service fails always, taking 0.1 second each try.
		// After 2nd try, the service gets stopped.
		// The starting process should be cancelled.

		service := &retrierMockService{
			startDelay: time.Millisecond * 100,
		}

		reporter := &retrierMockReporter{}

		retrier := rscsrv.NewStartRetrier(service, rscsrv.StartRetrierOptions{
			DelayBetweenTries: time.Millisecond, // If not defined, it will wait 5 seconds default...
			Reporter:          reporter,
		})

		engineStarter := rscsrv.NewServiceStarter(
			&rscsrv.NopStarterReporter{},
			retrier,
		)
		go func() {
			defer GinkgoRecover()

			time.Sleep(time.Millisecond * 150)
			Expect(retrier.Stop()).To(Succeed())
		}()
		Expect(engineStarter.Start()).To(Equal(rscsrv.ErrStartCancelled))
		Expect(service.startCount).To(Equal(2))
		Expect(reporter.count).To(Equal(2))
	})

	It("should cancel starting when asked stop with no reporter", func() {
		// The retrier have a 5 maximum tries.
		// The service fails always, taking 0.1 second each try.
		// After 2nd try, the service gets stopped.
		// The starting process should be cancelled.

		service := &retrierMockService{
			startDelay: time.Millisecond * 100,
		}

		retrier := rscsrv.NewStartRetrier(service, rscsrv.StartRetrierOptions{
			DelayBetweenTries: time.Millisecond, // If not defined, it will wait 5 seconds default...
			Reporter:          &retrierMockReporter{},
		})

		engineStarter := rscsrv.NewServiceStarter(
			&rscsrv.NopStarterReporter{},
			retrier,
		)
		go func() {
			defer GinkgoRecover()

			time.Sleep(time.Millisecond * 150)
			Expect(retrier.Stop()).To(Succeed())
		}()
		Expect(engineStarter.Start()).To(Equal(rscsrv.ErrStartCancelled))
		Expect(service.startCount).To(Equal(2))
	})

	It("should not error when handling a service not Startable", func() {
		// This tests just starts a service that starts at first, no problems.

		service := &retrierMockServiceNonStartable{}

		engineStarter := rscsrv.NewServiceStarter(
			&rscsrv.NopStarterReporter{},
			rscsrv.NewStartRetrier(service, rscsrv.StartRetrierOptions{
				MaxTries:          5,
				DelayBetweenTries: time.Millisecond, // If not defined, it will wait 5 seconds default...
				Reporter:          &retrierMockReporter{},
			}),
		)
		Expect(engineStarter.Start()).To(Succeed())
		Expect(engineStarter.Stop(true)).To(Succeed())
	})

	Context("MaxTries", func() {
		It("should start a service at the last chance", func() {
			// The retrier have a 5 maximum tries.
			// The service fails 4 times before succeed.

			service := &retrierMockService{
				successAt: 5,
			}

			engineStarter := rscsrv.NewServiceStarter(
				&rscsrv.NopStarterReporter{},
				rscsrv.NewStartRetrier(service, rscsrv.StartRetrierOptions{
					MaxTries:          5,
					DelayBetweenTries: time.Millisecond, // If not defined, it will wait 5 seconds default...
					Reporter:          &retrierMockReporter{},
				}),
			)
			Expect(engineStarter.Start()).To(Succeed())
			Expect(service.startCount).To(Equal(5))
		})

		It("should start a service at the last chance with error panic", func() {
			// The retrier have a 5 maximum tries.
			// The service fails by panicking an `error` interface.
			// The service fails 4 times before succeed.

			service := &retrierMockPanicService{
				successAt: 5,
			}

			reporter := &retrierMockReporter{}

			engineStarter := rscsrv.NewServiceStarter(
				&rscsrv.NopStarterReporter{},
				rscsrv.NewStartRetrier(service, rscsrv.StartRetrierOptions{
					MaxTries:          5,
					DelayBetweenTries: time.Millisecond, // If not defined, it will wait 5 seconds default...
					Reporter:          reporter,
				}),
			)
			Expect(engineStarter.Start()).To(Succeed())
			Expect(service.startCount).To(Equal(5))
			Expect(reporter.count).To(Equal(4))
		})

		It("should start a service at the last chance with non error panic", func() {
			// The retrier have a 5 maximum tries.
			// The service fails by panicking an `error` interface.
			// The service fails 4 times before succeed.

			service := &retrierMockPanicService{
				successAt:   5,
				dataToPanic: "non error interface panic info",
			}

			reporter := &retrierMockReporter{}

			engineStarter := rscsrv.NewServiceStarter(
				&rscsrv.NopStarterReporter{},
				rscsrv.NewStartRetrier(service, rscsrv.StartRetrierOptions{
					MaxTries:          5,
					DelayBetweenTries: time.Millisecond, // If not defined, it will wait 5 seconds default...
					Reporter:          reporter,
				}),
			)
			Expect(engineStarter.Start()).To(Succeed())
			Expect(service.startCount).To(Equal(5))
			Expect(reporter.count).To(Equal(4))
		})

		It("should fail starting by maximum failures count", func() {
			// The retrier have a 5 maximum tries.
			// The service always fails.

			service := &retrierMockService{
				successAt: 6,
			}

			engineStarter := rscsrv.NewServiceStarter(
				&rscsrv.NopStarterReporter{},
				rscsrv.NewStartRetrier(service, rscsrv.StartRetrierOptions{
					MaxTries:          5,
					DelayBetweenTries: time.Millisecond, // If not defined, it will wait 5 seconds default...
					Reporter:          &retrierMockReporter{},
				}),
			)
			err := engineStarter.Start()
			Expect(err).To(Equal(rscsrv.ErrMaxTriesExceeded))
			Expect(service.startCount).To(Equal(5))
		})
	})

	Context("Timeout", func() {
		It("should start a service right on time", func() {
			// The retrier have timeout of 0.5 seconds.
			// The service fails once (takes 0.4 seconds).
			// The retrier still have 0.1 seconds, so it tries again.
			// The service now succeeds starting.

			service := &retrierMockService{
				startDelay: time.Millisecond * 400,
				successAt:  2,
			}

			engineStarter := rscsrv.NewServiceStarter(
				&rscsrv.NopStarterReporter{},
				rscsrv.NewStartRetrier(service, rscsrv.StartRetrierOptions{
					Timeout:           time.Millisecond * 500,
					DelayBetweenTries: time.Millisecond, // If not defined, it will wait 5 seconds default...
					Reporter:          &retrierMockReporter{},
				}),
			)
			Expect(engineStarter.Start()).To(Succeed())
			Expect(service.startCount).To(Equal(2))
		})

		It("should start the service with overtime", func() {
			// The retrier have timeout of 0.5 seconds.
			// The service fails once (takes 0.3 seconds).
			// The retrier still have 0.2 seconds, so it tries again.
			// The service starts with a 0.1 seconds overtime, no problem.

			service := &retrierMockService{
				startDelay: time.Millisecond * 300,
				successAt:  2,
			}

			engineStarter := rscsrv.NewServiceStarter(
				&rscsrv.NopStarterReporter{},
				rscsrv.NewStartRetrier(service, rscsrv.StartRetrierOptions{
					Timeout:           time.Millisecond * 500,
					DelayBetweenTries: time.Millisecond, // If not defined, it will wait 5 seconds default...
					Reporter:          &retrierMockReporter{},
				}),
			)
			Expect(engineStarter.Start()).To(Succeed())
			Expect(service.startCount).To(Equal(2))
		})

		It("should fail starting by timeout", func() {
			// The retrier have timeout of 0.2 seconds.
			// The service fails once (takes 0.15 seconds).
			// The retrier still have 0.05 seconds, so it tries again.
			// The service fails again.
			// The retrier checks 0.3 seconds with a 0.1 seconds overtime, so if fails.

			service := &retrierMockService{
				startDelay: time.Millisecond * 150,
			}

			engineStarter := rscsrv.NewServiceStarter(
				&rscsrv.NopStarterReporter{},
				rscsrv.NewStartRetrier(service, rscsrv.StartRetrierOptions{
					Timeout:           time.Millisecond * 200,
					DelayBetweenTries: time.Millisecond, // If not defined, it will wait 5 seconds default...
					Reporter:          &retrierMockReporter{},
				}),
			)
			Expect(engineStarter.Start()).To(Equal(rscsrv.ErrStartTimeout))
			Expect(service.startCount).To(Equal(2))
		})
	})

	Context("Retrier", func() {
		It("should start a service at the last chance", func() {
			// The retrier have a 5 maximum tries.
			// The service fails 4 times before succeed.

			service := &retrierMockService{
				successAt: 5,
			}

			retrier := rscsrv.Retrier(rscsrv.StartRetrierOptions{
				MaxTries:          5,
				DelayBetweenTries: time.Millisecond, // If not defined, it will wait 5 seconds default...
				Reporter:          &retrierMockReporter{},
			})

			engineStarter := rscsrv.NewServiceStarter(
				&rscsrv.NopStarterReporter{},
				retrier(service),
			)
			Expect(engineStarter.Start()).To(Succeed())
			Expect(service.startCount).To(Equal(5))
		})
	})

	Context("Stop", func() {
		It("should cancel stop the service while waiting the delay between tries", func(done Done) {
			// The retrier have timeout of 1 second.
			// The service fails after 150 milliseconds.
			// The retrier waits 250 milliseconds to try again.
			// The retrier gets cancelled before the above timeout finishes.
			// An ErrStartCancelled is returned.

			service := &retrierMockService{
				startDelay: time.Millisecond * 150,
				successAt:  2,
			}

			retrier := rscsrv.NewStartRetrier(service, rscsrv.StartRetrierOptions{
				Timeout:           time.Second,
				DelayBetweenTries: time.Millisecond * 250, // If not defined, it will wait 5 seconds default...
				Reporter:          &retrierMockReporter{},
			})

			go func() {
				time.Sleep(time.Millisecond * 200)
				Expect(retrier.Stop()).To(Succeed())
				close(done)
			}()
			Expect(retrier.Start()).To(Equal(rscsrv.ErrStartCancelled))
			Expect(service.startCount).To(Equal(1))
		}, 0.5)
	})
})
