package rscsrv

import (
	"os"
	"os/signal"
)

type signalServiceStarter struct {
	ServiceStarter
	signalList []os.Signal

	signals chan os.Signal
}

func SignalStarter(serviceStarter ServiceStarter, signals ...os.Signal) ServiceStarter {
	if len(signals) == 0 {
		signals = []os.Signal{os.Interrupt}
	}
	return &signalServiceStarter{
		ServiceStarter: serviceStarter,
		signalList:     signals,
		signals:        make(chan os.Signal, 1),
	}
}

func (starter *signalServiceStarter) Start() error {
	signal.Notify(starter.signals, starter.signalList...)
	go func() {
		<-starter.signals
		starter.Stop(true)
	}()
	return starter.ServiceStarter.Start()
}

func (starter *signalServiceStarter) Stop(keepGoing bool) error {
	err := starter.ServiceStarter.Stop(keepGoing)
	if err != nil {
		return err
	}
	close(starter.signals)
	return nil
}

func (starter *signalServiceStarter) Wait() {
	starter.ServiceStarter.Wait()
}
