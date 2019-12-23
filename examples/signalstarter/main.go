package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/lab259/go-rscsrv"
)

type FakeService struct {
	name          string
	startDuration time.Duration
	stopDuration  time.Duration
}

func (service *FakeService) Name() string {
	return service.name
}

func (service *FakeService) LoadConfiguration() (interface{}, error) {
	time.Sleep(time.Millisecond * 300)
	return map[string]interface{}{}, nil
}

func (service *FakeService) ApplyConfiguration(interface{}) error {
	time.Sleep(time.Millisecond * 300)
	return nil
}

func (service *FakeService) Start() error {
	time.Sleep(service.startDuration)
	return nil
}

func (service *FakeService) Stop() error {
	time.Sleep(service.stopDuration)
	return nil
}

type CancellableFakeService struct {
	FakeService
}

func (service *CancellableFakeService) StartWithContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(service.startDuration):
		return nil
	}
}

func main() {
	serviceStarter := rscsrv.SignalStarter(rscsrv.DefaultServiceStarter(
		&CancellableFakeService{
			FakeService{
				name:          "Service 1 (start cancellable)",
				startDuration: time.Second,
				stopDuration:  time.Second,
			},
		},
		&FakeService{
			name:          "Service 2",
			startDuration: time.Second * 2,
			stopDuration:  time.Second * 2,
		},
		&CancellableFakeService{
			FakeService{
				name:          "Service 3 (start cancellable)",
				startDuration: time.Second * 3,
				stopDuration:  time.Second * 3,
			},
		},
		&FakeService{
			name:          "Service 4",
			startDuration: time.Second * 4,
			stopDuration:  time.Second * 4,
		},
	))
	if err := serviceStarter.Start(); err != nil {
		serviceStarter.Wait()
		if err == context.Canceled {
			fmt.Println("starting process aborted by signal")
			os.Exit(1)
		} else {
			fmt.Printf("error starting services: %s\n", err)
			os.Exit(2)
		}
	}
	fmt.Println("Hit <Ctrl+C> to stop the service.")
	serviceStarter.Wait()
}
