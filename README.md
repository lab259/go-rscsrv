# go-rscsrv

[![CircleCI](https://circleci.com/gh/lab259/go-rscsrv.svg?style=shield)](https://circleci.com/gh/lab259/go-rscsrv) [![codecov](https://codecov.io/gh/lab259/go-rscsrv/branch/master/graph/badge.svg)](https://codecov.io/gh/lab259/go-rscsrv) [![GoDoc](https://godoc.org/github.com/lab259/go-rscsrv?status.svg)](http://godoc.org/github.com/lab259/go-rscsrv) [![Go Report Card](https://goreportcard.com/badge/github.com/lab259/go-rscsrv)](https://goreportcard.com/report/github.com/lab259/go-rscsrv) [![Release](https://img.shields.io/github/release/lab259/go-rscsrv.svg?style=shield)](https://github.com/lab259/go-rscsrv/releases/latest)

> Resource/Service pattern for Go applications.

```go
serviceStarter := rscsrv.NewServiceStarter([]rscsrv.Service{
	&Service1,
	&Service2,
	&Service3,
}, &rscsrv.ColorServiceReporter{})

err := serviceStarter.Start()
if err != nil {
	serviceStarter.Stop(true)
}

// ... Service1, Service2 and Service3 were started
```

See [`/examples`](/examples) for more usage examples.
