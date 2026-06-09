module github.com/tx7do/go-wind-plugins/metrics/datadog

go 1.26.3

require (
	github.com/DataDog/datadog-go/v5 v5.8.3
	github.com/tx7do/go-wind-plugins/metrics v0.0.1
)

require (
	github.com/Microsoft/go-winio v0.6.2 // indirect
	golang.org/x/sys v0.43.0 // indirect
)

replace github.com/tx7do/go-wind-plugins/metrics => ../
