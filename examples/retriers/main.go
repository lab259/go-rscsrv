package main

import (
	"errors"
	"fmt"
	"math/rand"
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
	if rand.Intn(100) < 80 {
		return errors.New("80% of this error")
	}
	return nil
}

func (*Service1) Stop() error {
	time.Sleep(time.Second)
	return nil
}

func main() {
	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	rand.Seed(time.Now().Unix())

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	retriers := rscsrv.Retriers(rscsrv.StartRetrierOptions{
		MaxTries:          5,
		DelayBetweenTries: time.Second,
		Timeout:           time.Second * 60,
	})

	serviceStarter := rscsrv.DefaultServiceStarter(
		retriers(
			&Service1{},
			&Service2{},
		)...,
	)
	if err := serviceStarter.Start(); err != nil {
		fmt.Printf("error starting services: %s\n", err)
		os.Exit(1)
	}

	go func() {
		sig := <-sigs
		fmt.Printf("Gracefully stopping because of: %s\n", sig)
		serviceStarter.Stop(true)
		done <- true
	}()

	fmt.Println("Hit <Ctrl+C> to stop the service.")
	<-done
	fmt.Println("Done!")
}
