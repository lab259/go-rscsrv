# go-rscsrv
Resource/Service for Golang

```go
serviceStarter := rscsrv.NewServiceStarter([]rscsrv.Service{
	&Service1{},
	&Service2{},
	&Service3{},
}, &rscsrv.ColorServiceReporter{})
err := serviceStarter.Start()
if err != nil {
	serviceStarter.Stop(true)
}
// ... Service1, Service2 and Service3 were started
```
