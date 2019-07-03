package main

import (
	"time"

	"github.com/lab259/go-rscsrv"
)

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
	return nil
}

func main() {
	serviceStarter := rscsrv.NewServiceStarter(
		&rscsrv.ColorStarterReporter{},
		&Service1{},
	)
	serviceStarter.Start()
}
