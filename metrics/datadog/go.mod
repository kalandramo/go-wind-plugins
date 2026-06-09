module github.com/tx7do/go-wind-plugins/metrics/datadog

go 1.26.3

require (
	github.com/DataDog/datadog-go/v5 v5.6.0
	github.com/tx7do/go-wind-plugins/metrics v0.0.1
)

require (
	github.com/Microsoft/go-winio v0.5.0 // indirect
	golang.org/x/sys v0.0.0-20210510120138-977fb7262007 // indirect
)

replace github.com/tx7do/go-wind-plugins/metrics => ../
