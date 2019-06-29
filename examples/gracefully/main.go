package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lab259/go-rscsrv"
)

type Service2 struct {
	Service1
}

func (*Service2) Name() string {
	return "service2"
}

type Service1 struct{}

func (*Service1) Name() string {
	return "service1"
}

func (*Service1) LoadConfiguration() (interface{}, error) {
	time.Sleep(time.Millisecond * 300)
	return map[string]interface{}{}, nil
}

func (*Service1) ApplyConfiguration(interface{}) error {
	time.Sleep(time.Millisecond * 300)
	return nil
}

func (*Service1) Restart() error {
	return nil
}

func (*Service1) Start() error {
	time.Sleep(time.Second)
	return nil
}

func (*Service1) Stop() error {
	time.Sleep(time.Second)
	return nil
}

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	serviceStarter := rscsrv.NewServiceStarter(
		&rscsrv.ColorServiceReporter{},
		&Service1{},
		&Service2{},
	)
	serviceStarter.Start()

	go func() {
		sig := <-sigs
		fmt.Printf("Gracefully stopping because of: %s\n", sig)
		serviceStarter.Stop(true)
		done <- true
	}()

	<-done
	fmt.Println("Done!")
}
