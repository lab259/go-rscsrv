# go-rscsrv

[![CircleCI](https://circleci.com/gh/lab259/go-rscsrv.svg?style=shield)](https://circleci.com/gh/lab259/go-rscsrv) [![codecov](https://codecov.io/gh/lab259/go-rscsrv/branch/master/graph/badge.svg)](https://codecov.io/gh/lab259/go-rscsrv) [![GoDoc](https://godoc.org/github.com/lab259/go-rscsrv?status.svg)](http://godoc.org/github.com/lab259/go-rscsrv) [![Go Report Card](https://goreportcard.com/badge/github.com/lab259/go-rscsrv)](https://goreportcard.com/report/github.com/lab259/go-rscsrv) [![Release](https://img.shields.io/github/release/lab259/go-rscsrv.svg?style=shield)](https://github.com/lab259/go-rscsrv/releases/latest)

> Resource/Service pattern for Go applications.

```go
serviceStarter := rscsrv.NewServiceStarter(
	&rscsrv.ColorServiceReporter{},  // First, the reporter
	&Service1, &Service2, &Service3, // Here all services that should be started.
)

err := serviceStarter.Start()
if err != nil {
	serviceStarter.Stop(true)
}

// ... Service1, Service2 and Service3 were started
```

See [`/examples`](/examples) for more usage examples.

## Retrier

The `StartRetrier` is a mechanism that retries starting a `Service` when it
fails. No special `Service` implementation is required. The `StartRetrier` is a
proxy that wraps the real `Service`.

The `NewStartRetrier` wraps the `Service` providing the desired repeatability.

Example:

```go
serviceStarter := rscsrv.NewServiceStarter(
	&rscsrv.ColorServiceReporter{},  // First, the reporter
	rscsrv.NewStartRetrier(&Service1, rscsrv.StartRetrierOptions{
		MaxTries:          5,
		DelayBetweenTries: time.Second * 5,
		Timeout:           time.Second * 60,
	}), &Service2, &Service3, // Here all services that should be started.
)

err := serviceStarter.Start()
if err != nil {
	serviceStarter.Stop(true)
}
```

In this example, the `Service1` is wrapped by a `StartRetrier`. The retrier will
keep trying to start `Service1` until it reaches 5 failures. Between each try,
the retrier will wait 5 seconds before try again.

### Retrier helpers

The way `StartRetrier` was designed is for `opt-in`, so when the library gets
updated, the behaviour do not change. So a helper was designed to 

Retriers will apply the `StartRetrier` to many services at once:

```go
retriers := rscsrv.Retriers(rscsrv.StartRetrierOptions{
	MaxTries:          5,
	DelayBetweenTries: time.Second,
	Timeout:           time.Second * 60,
})

serviceStarter := rscsrv.NewServiceStarter(
	&rscsrv.ColorServiceReporter{},  // First, the reporter
	retriers(&Service1, &Service2, &Service3)...,
)

err := serviceStarter.Start()
if err != nil {
	serviceStarter.Stop(true)
}
```